// Copyright 2019 The Prometheus Authors
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
// http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package main

import (
	"bufio"
	"io"

	"github.com/prometheus/client_golang/prometheus"
)

// Exporter TODO
type Exporter struct {
	registry        prometheus.Registerer
	registered      bool
	up              prometheus.Gauge
	updated         prometheus.Gauge
	batterySOC      prometheus.Gauge
	batteryCharging prometheus.Gauge
	batteryCurrent  prometheus.Gauge
	cellVoltage     *prometheus.GaugeVec
	pvCurrent       *prometheus.GaugeVec
	loadCurrent     prometheus.Gauge
	tempInternal    prometheus.Gauge
	tempExternal    prometheus.Gauge
	adc             *prometheus.GaugeVec
	heat            *prometheus.GaugeVec
	status          prometheus.Gauge
}

// NewExporter TODO
func NewExporter(registry prometheus.Registerer) *Exporter {
	m := &Exporter{
		registry:   registry,
		registered: false,
		up: prometheus.NewGauge(prometheus.GaugeOpts{
			Namespace: "sbms",
			Name:      "up",
			Help:      "Was the last scrape of sbms successful.",
		}),
		updated: prometheus.NewGauge(prometheus.GaugeOpts{
			Namespace: "sbms",
			Name:      "updated",
			Help:      "The date the data was last updated (elapsed time since the Unix epoch).",
		}),
		batterySOC: prometheus.NewGauge(prometheus.GaugeOpts{
			Namespace: "sbms",
			Subsystem: "battery",
			Name:      "soc",
			Help:      "Battery state of charge (%).",
		}),
		batteryCharging: prometheus.NewGauge(prometheus.GaugeOpts{
			Namespace: "sbms",
			Subsystem: "battery",
			Name:      "charging",
			Help:      "Is the battery currently charging or discharging?",
		}),
		batteryCurrent: prometheus.NewGauge(prometheus.GaugeOpts{
			Namespace: "sbms",
			Subsystem: "battery",
			Name:      "current",
			Help:      "Battery current (positive means charging, negative means discharging).",
		}),
		cellVoltage: prometheus.NewGaugeVec(prometheus.GaugeOpts{
			Namespace: "sbms",
			Subsystem: "battery",
			Name:      "voltage",
			Help:      "battery cell voltage.",
		}, []string{"cell"}),
		pvCurrent: prometheus.NewGaugeVec(prometheus.GaugeOpts{
			Namespace: "sbms",
			Subsystem: "photovoltaic_array",
			Name:      "current",
			Help:      "Array current.",
		}, []string{"pv"}),
		loadCurrent: prometheus.NewGauge(prometheus.GaugeOpts{
			Namespace: "sbms",
			Subsystem: "external_load",
			Name:      "current",
			Help:      "Load current.",
		}),
		tempInternal: prometheus.NewGauge(prometheus.GaugeOpts{
			Namespace: "sbms",
			Subsystem: "device",
			Name:      "internal_temperature",
			Help:      "Device internal temperature.",
		}),
		tempExternal: prometheus.NewGauge(prometheus.GaugeOpts{
			Namespace: "sbms",
			Subsystem: "device",
			Name:      "external_temperature",
			Help:      "Device external temperature.",
		}),
		adc: prometheus.NewGaugeVec(prometheus.GaugeOpts{
			Namespace: "sbms",
			Subsystem: "device",
			Name:      "adc_value",
			Help:      "Device ADC value.",
		}, []string{"adc"}),
		heat: prometheus.NewGaugeVec(prometheus.GaugeOpts{
			Namespace: "sbms",
			Subsystem: "device",
			Name:      "heat_value",
			Help:      "Device heat value.",
		}, []string{"heat"}),
		status: prometheus.NewGauge(prometheus.GaugeOpts{
			Namespace: "sbms",
			Subsystem: "device",
			Name:      "status",
			Help:      "Device status number.",
		}),
	}

	// up is the only metric always registered
	m.registry.MustRegister(m.up)
	m.up.Set(0)

	return m
}

// Export TODO
func (m *Exporter) Export(r io.Reader) error {
	s := bufio.NewScanner(r)
	v := new(Values)
	down := func() {
		m.up.Set(0)
		m.ensureExporterCleared()
	}
	up := func() {
		m.up.Set(1)
		m.ensureExporterRegistered()
	}

	defer down()

	for s.Scan() {
		if err := v.ReadFrom(s.Bytes()); err != nil {
			down()
			continue
		}

		m.updated.Set(float64(v.Date.UnixNano()) / 1e9)
		m.batterySOC.Set(float64(v.StateOfCharge))
		m.batteryCharging.Set(boolAsFloat(v.Charging))
		m.batteryCurrent.Set(v.BatteryCurrent)
		m.cellVoltage.With(prometheus.Labels{"cell": "1"}).Set(v.Cell1Voltage)
		m.cellVoltage.With(prometheus.Labels{"cell": "2"}).Set(v.Cell2Voltage)
		m.cellVoltage.With(prometheus.Labels{"cell": "3"}).Set(v.Cell3Voltage)
		m.cellVoltage.With(prometheus.Labels{"cell": "4"}).Set(v.Cell4Voltage)
		m.cellVoltage.With(prometheus.Labels{"cell": "5"}).Set(v.Cell5Voltage)
		m.cellVoltage.With(prometheus.Labels{"cell": "6"}).Set(v.Cell6Voltage)
		m.cellVoltage.With(prometheus.Labels{"cell": "7"}).Set(v.Cell7Voltage)
		m.cellVoltage.With(prometheus.Labels{"cell": "8"}).Set(v.Cell8Voltage)
		m.pvCurrent.With(prometheus.Labels{"pv": "1"}).Set(v.PV1Current)
		m.pvCurrent.With(prometheus.Labels{"pv": "2"}).Set(v.PV2Current)
		m.loadCurrent.Set(v.ExtLoadCurrent)
		m.tempInternal.Set(v.InternalTemp)
		m.tempExternal.Set(v.ExternalTemp)
		m.adc.With(prometheus.Labels{"adc": "2"}).Set(float64(v.ADC2))
		m.adc.With(prometheus.Labels{"adc": "3"}).Set(float64(v.ADC3))
		m.adc.With(prometheus.Labels{"adc": "4"}).Set(float64(v.ADC4))
		m.heat.With(prometheus.Labels{"heat": "1"}).Set(float64(v.Heat1))
		m.heat.With(prometheus.Labels{"heat": "2"}).Set(float64(v.Heat2))
		m.status.Set(float64(v.Status))
		up()
	}

	if s.Err() != nil {
		return s.Err()
	}

	return io.EOF
}

func (m *Exporter) ensureExporterRegistered() {
	if m.registered {
		return
	}
	m.registry.MustRegister(m.updated)
	m.registry.MustRegister(m.batterySOC)
	m.registry.MustRegister(m.batteryCharging)
	m.registry.MustRegister(m.batteryCurrent)
	m.registry.MustRegister(m.cellVoltage)
	m.registry.MustRegister(m.pvCurrent)
	m.registry.MustRegister(m.loadCurrent)
	m.registry.MustRegister(m.tempInternal)
	m.registry.MustRegister(m.tempExternal)
	m.registry.MustRegister(m.adc)
	m.registry.MustRegister(m.heat)
	m.registry.MustRegister(m.status)
	m.registered = true
}

func (m *Exporter) ensureExporterCleared() {
	if !m.registered {
		return
	}
	m.registry.Unregister(m.updated)
	m.registry.Unregister(m.batterySOC)
	m.registry.Unregister(m.batteryCharging)
	m.registry.Unregister(m.batteryCurrent)
	m.registry.Unregister(m.cellVoltage)
	m.registry.Unregister(m.pvCurrent)
	m.registry.Unregister(m.loadCurrent)
	m.registry.Unregister(m.tempInternal)
	m.registry.Unregister(m.tempExternal)
	m.registry.Unregister(m.adc)
	m.registry.Unregister(m.heat)
	m.registry.Unregister(m.status)
	m.registered = false
}

func boolAsFloat(b bool) float64 {
	if b {
		return 1
	}
	return 0
}
