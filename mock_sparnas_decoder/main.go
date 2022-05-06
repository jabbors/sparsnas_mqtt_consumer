package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"math/rand"
	"time"

	mqtt "github.com/eclipse/paho.mqtt.golang"
)

var (
	kwhCounter float64
)

func randInt(min, max int) int {
	return rand.Int()%max + min
}

func randFloat(min, max int) float64 {
	return float64(randInt(min*100, max*100) / 100)
}

// Reading represents a sensor reading transmitted by an IKEA Sparsnas
// captured by https://github.com/tubalainen/sparsnas_decoder
//{"Sequence": 26960, "Watt": 4656.00, "kWh": 1849.536, "battery": 100, "FreqErr":-0.13, "Effect":  194, "Data4": 2, "Sensor": 671150}
type Reading struct {
	Sequence int     `json:"Sequence"`
	Watt     float64 `json:"Watt"`
	Kwh      float64 `json:"kWh"`
	Battery  float64 `json:"battery"`
	FreqErr  float64 `json:"FreqErr"`
	Effect   int     `json:"Effect"`
	Data4    int     `json:"Data4"`
	Sensor   int     `json:"Sensor"`
}

func NewReading() *Reading {
	r := Reading{
		Sequence: int(time.Now().Unix()),
		Watt:     randFloat(3000, 4000),
		Kwh:      kwhCounter + rand.Float64(),
		Battery:  100,
		FreqErr:  rand.Float64(),
		Effect:   randInt(100, 200),
		Data4:    2,
		Sensor:   123456,
	}

	return &r
}

func (r *Reading) Serialize() string {
	s, _ := json.Marshal(r)
	return string(s)
}

func main() {
	var broker string
	var port int
	flag.StringVar(&broker, "broker", "localhost", "IP to the MQTT broker")
	flag.IntVar(&port, "port", 1883, "port to the MQTT broker")
	flag.Parse()

	rand.Seed(time.Now().Unix())
	kwhCounter = randFloat(1000, 2000)

	var connectHandler mqtt.OnConnectHandler = func(client mqtt.Client) {
		fmt.Println("Connected")
	}

	var connectLostHandler mqtt.ConnectionLostHandler = func(client mqtt.Client, err error) {
		fmt.Printf("Connection lost: %v", err)
	}

	opts := mqtt.NewClientOptions()
	opts.AddBroker(fmt.Sprintf("tcp://%s:%d", broker, port))
	opts.SetClientID("go_mqtt_client_fake_publisher")
	// opts.SetUsername("")
	// opts.SetPassword("")
	opts.OnConnect = connectHandler
	opts.OnConnectionLost = connectLostHandler
	client := mqtt.NewClient(opts)
	if token := client.Connect(); token.Wait() && token.Error() != nil {
		panic(token.Error())
	}

	defer client.Disconnect(0)

	for {
		r := NewReading()
		topic := fmt.Sprintf("sparsnas/%d", r.Sensor)
		message := r.Serialize()
		token := client.Publish(topic, 0, false, message)
		fmt.Printf("Publising message '%s', to topic '%s'\n", message, topic)
		token.Wait()

		time.Sleep(5 * time.Second)
	}
}
