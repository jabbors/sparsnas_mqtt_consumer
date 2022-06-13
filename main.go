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
)

const (
	appName = "sparsnas_mqtt_consumer"
)

var (
	version        string
	influxForward  bool
	influxAddr     string
	influxDatabase string
)

func setupMqttClient(broker string, port int) (mqtt.Client, error) {
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
	// opts.SetUsername("")
	// opts.SetPassword("")
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
	var broker string
	var port int
	var topic string
	var verFlag bool
	flag.BoolVar(&influxForward, "influx-forward", false, "forward messages to influx")
	flag.StringVar(&influxAddr, "influx-addr", "http://localhost:8086", "address to influxdb for storing measurements")
	flag.StringVar(&influxDatabase, "influx-db", "sparsnas", "name of the influx database")
	flag.StringVar(&broker, "broker", "localhost", "IP to the MQTT broker")
	flag.IntVar(&port, "port", 1883, "port to the MQTT broker")
	flag.StringVar(&topic, "topic", "#", "topic to subscribe to")
	flag.BoolVar(&verFlag, "version", false, "print version and exit")
	flag.Parse()

	if verFlag {
		fmt.Printf("%s: version=%s\n", appName, version)
		os.Exit(0)
	}

	if !influxForward {
		log.Printf("Message forwarding to InfluxDB is not enabled. No measurements will be dispatached!")
	} else {
		// TODO: check that we can connect to influx
	}

	// setup go channels for pipeing parsed measurements and errors from mqtt topic
	measurementsCh := make(chan *Measurement, 10)
	errorCh := make(chan error, 5)

	client, err := setupMqttClient(broker, port)
	if err != nil {
		panic(err)
	}

	client.Subscribe(topic, 0, func(client mqtt.Client, msg mqtt.Message) {
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
			if influxForward {
				log.Printf("influxdb: dispatching measurement")
				buf := new(bytes.Buffer)
				_, err := buf.WriteString(m.InfluxLineProtocol())
				if err != nil {
					log.Println("influxdb: failed to write line to buffer", err)
					return
				}
				resp, err := http.Post(fmt.Sprintf("%s/write?db=%s", influxAddr, influxDatabase), "text/plain", buf)
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
