# sparsnas_mqtt_consumer

sparsnas_mqtt_consumer is an application subscribing for decoded messages (by <https://github.com/tubalainen/sparsnas_decoder>) published to a MQTT topic and dispatching them forward to other data sources. Currently InfluxDB is the only supported output data source, but the code is design such that new data sources can be added easily.

Additionally the repo contains example Grafana dashboards for plotting the data collected data.

## How to use?

1. Build and run the consumer

```
go build && ./sparsnas_mqtt_consumer -influx-forward
```

Runtime configuration flags

```
Usage of ./sparsnas_mqtt_consumer:
  -broker string
     IP to the MQTT broker (default "localhost")
  -influx-addr string
     address to influxdb for storing measurements (default "http://localhost:8086")
  -influx-db string
     name of the influx database (default "sparsnas")
  -influx-forward
     forward messages to influx
  -port int
     port to the MQTT broker (default 1883)
  -topic string
     topic to subscribe to (default "#")
  -version
     print version and exit
```

## Development

1. Fire up the backend services and setup the influx database

```
docker-compose up -d
./create-database.sh
```

2. Build and run the mock sparsnas_decoder publishing generated readings to MQTT

```
cd mock_sparsnas_decoder
go build
./mock_sparnas_decoder
```

3. Build and run the consumer

```
go build
./sparsnas_mqtt_consumer
```
