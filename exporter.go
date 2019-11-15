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
	"bytes"
	"io"

	"github.com/prometheus/client_golang/prometheus"
)

// Exporter TODO
type Exporter struct {
	registry          prometheus.Registerer
	registered        bool
	up                prometheus.Gauge
	updated           prometheus.Gauge
	status            prometheus.Gauge
	batteryCharging   prometheus.Gauge
	batterySOC        prometheus.Gauge
	batteryVolts      prometheus.Gauge
	batteryAmperes    prometheus.Gauge
	batteryWatts      prometheus.Gauge
	cellVolts         *prometheus.GaugeVec
	pvVolts           prometheus.Gauge
	pvAmperes         *prometheus.GaugeVec
	pvWatts           *prometheus.GaugeVec
	pvAmperesCombined prometheus.Gauge
	pvWattsCombined   prometheus.Gauge
	thermistorCelsius *prometheus.GaugeVec
	adcValues         *prometheus.GaugeVec
	heatValues        *prometheus.GaugeVec
	extLoadVolts      prometheus.Gauge
	extLoadAmperes    prometheus.Gauge
	extLoadWatts      prometheus.Gauge
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
			Subsystem: "updated",
			Name:      "unix",
			Help:      "The unix date the data was last updated (number of seconds elapsed since January 1, 1970 UTC).",
		}),
		status: prometheus.NewGauge(prometheus.GaugeOpts{
			Namespace: "sbms",
			Subsystem: "device",
			Name:      "status",
			Help:      "Device status number.",
		}),
		batteryCharging: prometheus.NewGauge(prometheus.GaugeOpts{
			Namespace: "sbms",
			Subsystem: "battery",
			Name:      "charging",
			Help:      "Is the battery currently charging or discharging?",
		}),
		batterySOC: prometheus.NewGauge(prometheus.GaugeOpts{
			Namespace: "sbms",
			Subsystem: "battery",
			Name:      "soc",
			Help:      "Battery state of charge (%).",
		}),
		batteryVolts: prometheus.NewGauge(prometheus.GaugeOpts{
			Namespace: "sbms",
			Subsystem: "battery",
			Name:      "volts",
			Help:      "Battery voltage.",
		}),
		batteryAmperes: prometheus.NewGauge(prometheus.GaugeOpts{
			Namespace: "sbms",
			Subsystem: "battery",
			Name:      "amperes",
			Help:      "Battery current (positive means charging, negative means discharging).",
		}),
		batteryWatts: prometheus.NewGauge(prometheus.GaugeOpts{
			Namespace: "sbms",
			Subsystem: "battery",
			Name:      "watts",
			Help:      "Battery power (positive means charging, negative means discharging).",
		}),
		cellVolts: prometheus.NewGaugeVec(prometheus.GaugeOpts{
			Namespace: "sbms",
			Subsystem: "cell",
			Name:      "volts",
			Help:      "Battery cell voltage.",
		}, []string{"cell"}),
		pvVolts: prometheus.NewGauge(prometheus.GaugeOpts{
			Namespace: "sbms",
			Subsystem: "pv",
			Name:      "volts",
			Help:      "Array voltage.",
		}),
		pvAmperes: prometheus.NewGaugeVec(prometheus.GaugeOpts{
			Namespace: "sbms",
			Subsystem: "pv",
			Name:      "amperes",
			Help:      "Array current.",
		}, []string{"pv"}),
		pvWatts: prometheus.NewGaugeVec(prometheus.GaugeOpts{
			Namespace: "sbms",
			Subsystem: "pv",
			Name:      "watts",
			Help:      "Array power.",
		}, []string{"pv"}),
		pvAmperesCombined: prometheus.NewGauge(prometheus.GaugeOpts{
			Namespace: "sbms",
			Subsystem: "pv",
			Name:      "amperes_combined",
			Help:      "Arrays total current.",
		}),
		pvWattsCombined: prometheus.NewGauge(prometheus.GaugeOpts{
			Namespace: "sbms",
			Subsystem: "pv",
			Name:      "watts_combined",
			Help:      "Arrays total power.",
		}),
		thermistorCelsius: prometheus.NewGaugeVec(prometheus.GaugeOpts{
			Namespace: "sbms",
			Subsystem: "thermistor",
			Name:      "celsius",
			Help:      "Device thermistor temperature.",
		}, []string{"sensor"}),
		adcValues: prometheus.NewGaugeVec(prometheus.GaugeOpts{
			Namespace: "sbms",
			Subsystem: "adc",
			Name:      "values",
			Help:      "Device ADC value.",
		}, []string{"adc"}),
		heatValues: prometheus.NewGaugeVec(prometheus.GaugeOpts{
			Namespace: "sbms",
			Subsystem: "heat",
			Name:      "values",
			Help:      "Device heat value.",
		}, []string{"heat"}),
		extLoadVolts: prometheus.NewGauge(prometheus.GaugeOpts{
			Namespace: "sbms",
			Subsystem: "external_load",
			Name:      "volts",
			Help:      "External load voltage.",
		}),
		extLoadAmperes: prometheus.NewGauge(prometheus.GaugeOpts{
			Namespace: "sbms",
			Subsystem: "external_load",
			Name:      "amperes",
			Help:      "External load current.",
		}),
		extLoadWatts: prometheus.NewGauge(prometheus.GaugeOpts{
			Namespace: "sbms",
			Subsystem: "external_load",
			Name:      "watts",
			Help:      "External load power.",
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
		if err := v.ReadFrom(bytes.TrimSpace(s.Bytes())); err != nil {
			down()
			continue
		}

		battVolts := v.Cell1Voltage + v.Cell2Voltage + v.Cell3Voltage + v.Cell4Voltage + v.Cell5Voltage + v.Cell6Voltage + v.Cell7Voltage + v.Cell8Voltage

		up()

		m.updated.Set(float64(v.Date.Unix()))
		m.status.Set(float64(v.Status))
		m.batteryCharging.Set(boolAsFloat(v.Charging))
		m.batterySOC.Set(float64(v.StateOfCharge))
		m.batteryVolts.Set(battVolts)
		m.batteryAmperes.Set(v.BatteryCurrent)
		m.batteryWatts.Set(v.BatteryCurrent * battVolts)
		m.cellVolts.With(prometheus.Labels{"cell": "1"}).Set(v.Cell1Voltage)
		m.cellVolts.With(prometheus.Labels{"cell": "2"}).Set(v.Cell2Voltage)
		m.cellVolts.With(prometheus.Labels{"cell": "3"}).Set(v.Cell3Voltage)
		m.cellVolts.With(prometheus.Labels{"cell": "4"}).Set(v.Cell4Voltage)
		m.cellVolts.With(prometheus.Labels{"cell": "5"}).Set(v.Cell5Voltage)
		m.cellVolts.With(prometheus.Labels{"cell": "6"}).Set(v.Cell6Voltage)
		m.cellVolts.With(prometheus.Labels{"cell": "7"}).Set(v.Cell7Voltage)
		m.cellVolts.With(prometheus.Labels{"cell": "8"}).Set(v.Cell8Voltage)
		m.pvVolts.Set(battVolts)
		m.pvAmperes.With(prometheus.Labels{"pv": "1"}).Set(v.PV1Current)
		m.pvAmperes.With(prometheus.Labels{"pv": "2"}).Set(v.PV2Current)
		m.pvWatts.With(prometheus.Labels{"pv": "1"}).Set(v.PV1Current * battVolts)
		m.pvWatts.With(prometheus.Labels{"pv": "2"}).Set(v.PV2Current * battVolts)
		m.pvAmperesCombined.Set(v.PV1Current + v.PV2Current)
		m.pvWattsCombined.Set(v.PV1Current*battVolts + v.PV2Current*battVolts)
		m.thermistorCelsius.With(prometheus.Labels{"sensor": "internal"}).Set(v.InternalTemp)
		m.thermistorCelsius.With(prometheus.Labels{"sensor": "external"}).Set(v.ExternalTemp)
		m.adcValues.With(prometheus.Labels{"adc": "2"}).Set(float64(v.ADC2))
		m.adcValues.With(prometheus.Labels{"adc": "3"}).Set(float64(v.ADC3))
		m.adcValues.With(prometheus.Labels{"adc": "4"}).Set(float64(v.ADC4))
		m.heatValues.With(prometheus.Labels{"heat": "1"}).Set(float64(v.Heat1))
		m.heatValues.With(prometheus.Labels{"heat": "2"}).Set(float64(v.Heat2))
		m.extLoadVolts.Set(battVolts)
		m.extLoadAmperes.Set(v.ExtLoadCurrent)
		m.extLoadWatts.Set(v.ExtLoadCurrent * battVolts)
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
	m.registry.MustRegister(m.status)
	m.registry.MustRegister(m.batteryCharging)
	m.registry.MustRegister(m.batterySOC)
	m.registry.MustRegister(m.batteryVolts)
	m.registry.MustRegister(m.batteryAmperes)
	m.registry.MustRegister(m.batteryWatts)
	m.registry.MustRegister(m.cellVolts)
	m.registry.MustRegister(m.pvVolts)
	m.registry.MustRegister(m.pvAmperes)
	m.registry.MustRegister(m.pvWatts)
	m.registry.MustRegister(m.pvAmperesCombined)
	m.registry.MustRegister(m.pvWattsCombined)
	m.registry.MustRegister(m.thermistorCelsius)
	m.registry.MustRegister(m.adcValues)
	m.registry.MustRegister(m.heatValues)
	m.registry.MustRegister(m.extLoadVolts)
	m.registry.MustRegister(m.extLoadAmperes)
	m.registry.MustRegister(m.extLoadWatts)
	m.registered = true
}

func (m *Exporter) ensureExporterCleared() {
	if !m.registered {
		return
	}
	m.registry.Unregister(m.updated)
	m.registry.Unregister(m.status)
	m.registry.Unregister(m.batteryCharging)
	m.registry.Unregister(m.batterySOC)
	m.registry.Unregister(m.batteryVolts)
	m.registry.Unregister(m.batteryAmperes)
	m.registry.Unregister(m.batteryWatts)
	m.registry.Unregister(m.cellVolts)
	m.registry.Unregister(m.pvVolts)
	m.registry.Unregister(m.pvAmperes)
	m.registry.Unregister(m.pvWatts)
	m.registry.Unregister(m.pvAmperesCombined)
	m.registry.Unregister(m.pvWattsCombined)
	m.registry.Unregister(m.thermistorCelsius)
	m.registry.Unregister(m.adcValues)
	m.registry.Unregister(m.heatValues)
	m.registry.Unregister(m.extLoadVolts)
	m.registry.Unregister(m.extLoadAmperes)
	m.registry.Unregister(m.extLoadWatts)
	m.registered = false
}

func boolAsFloat(b bool) float64 {
	if b {
		return 1
	}
	return 0
}
