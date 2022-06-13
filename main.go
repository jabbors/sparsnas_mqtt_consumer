package main

import (
	"bytes"
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"time"

	mqtt "github.com/eclipse/paho.mqtt.golang"
	"github.com/kelseyhightower/envconfig"
)

const (
	appName = "sparsnas_mqtt_consumer"
)

var (
	version string
)

// appConfig represents the configuration.
type appConfig struct {
	MQTTBroker     string `default:"localhost" split_words:"true" desc:"IP or hostaname to the MQTT broker"`
	MQTTPort       int    `default:"1883" split_words:"true" desc:"port to the MQTT broker"`
	MQTTUsername   string `default:"" split_words:"true" desc:"username for MQTT broker"`
	MQTTPassword   string `default:"" split_words:"true" desc:"password for MQTT broker"`
	MQTTTopic      string `default:"#" split_words:"true", decs:"topic to subscribe to"`
	InfluxAddr     string `default:"http://localhost:8086" split_words:"true" desc:"address to influxdb for storing measurements"`
	InfluxDatabase string `default:"sparsnas" split_words:"true" desc:"name of the influx database"`
	InfluxForward  bool   `default:"false" split_words:"true" desc:"forward messages to influx""`
}

// parse options from the environment. Return an error if parsing fails.
func (a *appConfig) parse() {
	defaultUsage := flag.Usage
	flag.Usage = func() {
		// Show default usage for the app (lists flags, etc).
		defaultUsage()
		fmt.Fprint(os.Stderr, "\n")

		err := envconfig.Usage("", a)
		if err != nil {
			fmt.Fprintf(os.Stderr, "error: %s\n\n", err.Error())
			os.Exit(1)
		}
	}

	var verFlag bool
	flag.BoolVar(&verFlag, "version", false, "print version and exit")
	flag.Parse()

	// Print version and exit if -version flag is passed.
	if verFlag {
		fmt.Printf("%s: version=%s\n", appName, version)
		os.Exit(0)
	}

	err := envconfig.Process("", a)
	if err != nil {
		fmt.Fprintf(os.Stderr, "error: %s\n\n", err.Error())
		flag.Usage()
		os.Exit(1)
	}
}

func setupMqttClient(broker string, port int, username, password string) (mqtt.Client, error) {
	var messagePubHandler mqtt.MessageHandler = func(client mqtt.Client, msg mqtt.Message) {
		log.Printf("mqtt: received message from topic: %s", msg.Topic())
	}

	var connectHandler mqtt.OnConnectHandler = func(client mqtt.Client) {
		log.Println("mqtt: connected")
	}

	var connectLostHandler mqtt.ConnectionLostHandler = func(client mqtt.Client, err error) {
		log.Printf("mqtt: connection lost: %v", err)
	}

	opts := mqtt.NewClientOptions()
	opts.AddBroker(fmt.Sprintf("tcp://%s:%d", broker, port))
	opts.SetClientID(fmt.Sprintf("go_mqtt_client_%d", time.Now().Unix()))
	opts.SetUsername(username)
	opts.SetPassword(password)
	opts.SetDefaultPublishHandler(messagePubHandler)
	opts.OnConnect = connectHandler
	opts.OnConnectionLost = connectLostHandler
	client := mqtt.NewClient(opts)
	if token := client.Connect(); token.Wait() && token.Error() != nil {
		return nil, token.Error()
	}

	return client, nil
}

func main() {
	config := &appConfig{}
	config.parse()

	if !config.InfluxForward {
		log.Printf("Message forwarding to InfluxDB is not enabled. No measurements will be dispatached!")
	} else {
		resp, err := http.Get(fmt.Sprintf("%s/ping", config.InfluxAddr))
		if err != nil {
			panic(err)
		}

		if resp.StatusCode != http.StatusNoContent {
			log.Printf("influxdb: unexpected return code, expected %d but got %d", http.StatusNoContent, resp.StatusCode)
			reasons, err := ioutil.ReadAll(resp.Body)
			if err != nil {
				log.Println("influxdb: could not read response body")
				return
			}
			log.Println("influxdb:", string(reasons))
		}
	}

	// setup go channels for pipeing parsed measurements and errors from mqtt topic
	measurementsCh := make(chan *Measurement, 10)
	errorCh := make(chan error, 5)

	client, err := setupMqttClient(config.MQTTBroker, config.MQTTPort, config.MQTTUsername, config.MQTTPassword)
	if err != nil {
		panic(err)
	}

	client.Subscribe(config.MQTTTopic, 0, func(client mqtt.Client, msg mqtt.Message) {
		log.Printf("mqtt: received message: %s from topic: %s\n", msg.Payload(), msg.Topic())

		m, err := NewMeasurement(msg.Payload())
		if err != nil {
			errorCh <- fmt.Errorf("Unmarshal JSON payload %s failed with error: %s", msg.Payload(), err)
		}

		measurementsCh <- m
	})

	defer client.Disconnect(0)

	for {
		select {
		// TODO: handle shutdown

		case err := <-errorCh:
			log.Println(err)
		case m := <-measurementsCh:
			if config.InfluxForward {
				log.Printf("influxdb: dispatching measurement")
				buf := new(bytes.Buffer)
				_, err := buf.WriteString(m.InfluxLineProtocol())
				if err != nil {
					log.Println("influxdb: failed to write line to buffer", err)
					return
				}
				resp, err := http.Post(fmt.Sprintf("%s/write?db=%s", config.InfluxAddr, config.InfluxDatabase), "text/plain", buf)
				if err != nil {
					log.Println("influxdb: failed to POST data:", err)
					return
				}
				defer resp.Body.Close()

				if resp.StatusCode != http.StatusNoContent {
					log.Printf("influxdb: unexpected return code, expected %d but got %d", http.StatusNoContent, resp.StatusCode)
					reasons, err := ioutil.ReadAll(resp.Body)
					if err != nil {
						log.Println("influxdb: could not read response body")
						return
					}
					log.Println("influxdb:", string(reasons))
				}
			}
		}
	}
}
