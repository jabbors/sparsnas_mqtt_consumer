package main

import (
	"encoding/json"
	"fmt"
)

// Measurement represents sensor reading transmitted by a IKEA Sparsnas
//{"Sequence": 26960, "Watt": 4656.00, "kWh": 1849.536, "battery": 100, "FreqErr":-0.13, "Effect":  194, "Data4": 2, "Sensor": 671150}
type Measurement struct {
	Sequence int     `json:"Sequence"`
	Watt     float64 `json:"Watt"`
	Kwh      float64 `json:"kWh"`
	Battery  float64 `json:"battery"`
	FreqErr  float64 `json:"FreqErr"`
	Effect   int     `json:"Effect"`
	Data4    int     `json:"Data4"`
	Sensor   int     `json:"Sensor"`
}

func NewMeasurement(b []byte) (*Measurement, error) {
	m := Measurement{}
	err := json.Unmarshal(b, &m)
	if err != nil {
		return nil, err
	}

	return &m, nil
}

func (m *Measurement) InfluxLineProtocol() string {
	return fmt.Sprintf("reading,sensor=%d sequence=%d,watt=%f,kwh=%f,battery=%f,freqerr=%f,effect=%d",
		m.Sensor,
		m.Sequence,
		m.Watt,
		m.Kwh,
		m.Battery,
		m.FreqErr,
		m.Effect)
}
