// Copyright 2019 Mike Gleason jr Couturier
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
	"errors"
	"math"
	"time"
)

// Various errors.
var (
	ErrDataLength = errors.New("invalid data length")
)

// Values TODO
type Values struct {
	Date           time.Time
	StateOfCharge  int
	Cell1Voltage   float64
	Cell2Voltage   float64
	Cell3Voltage   float64
	Cell4Voltage   float64
	Cell5Voltage   float64
	Cell6Voltage   float64
	Cell7Voltage   float64
	Cell8Voltage   float64
	InternalTemp   float64
	ExternalTemp   float64
	Charging       bool
	BatteryCurrent float64
	PV1Current     float64
	PV2Current     float64
	ExtLoadCurrent float64
	ADC2           int
	ADC3           int
	ADC4           int
	Heat1          int
	Heat2          int
	Status         int
}

// ReadFrom TODO
func (v *Values) ReadFrom(b []byte) error {
	if len(b) != 59 {
		return ErrDataLength
	}

	v.Date = time.Date(2000+v.unpackBase91(b, 0, 1), time.Month(v.unpackBase91(b, 1, 1)), v.unpackBase91(b, 2, 1), v.unpackBase91(b, 3, 1), v.unpackBase91(b, 4, 1), v.unpackBase91(b, 5, 1), 0, time.UTC)
	v.StateOfCharge = v.unpackBase91(b, 6, 2)
	v.Cell1Voltage = float64(v.unpackBase91(b, 8, 2)) / 1000.0
	v.Cell2Voltage = float64(v.unpackBase91(b, 10, 2)) / 1000.0
	v.Cell3Voltage = float64(v.unpackBase91(b, 12, 2)) / 1000.0
	v.Cell4Voltage = float64(v.unpackBase91(b, 14, 2)) / 1000.0
	v.Cell5Voltage = float64(v.unpackBase91(b, 16, 2)) / 1000.0
	v.Cell6Voltage = float64(v.unpackBase91(b, 18, 2)) / 1000.0
	v.Cell7Voltage = float64(v.unpackBase91(b, 20, 2)) / 1000.0
	v.Cell8Voltage = float64(v.unpackBase91(b, 22, 2)) / 1000.0
	v.InternalTemp = float64(v.unpackBase91(b, 24, 2)-450) / 10.0
	v.ExternalTemp = float64(v.unpackBase91(b, 26, 2)-450) / 10.0
	v.Charging = b[28] == '+'
	v.BatteryCurrent = float64(v.unpackBase91(b, 29, 3)) / 1000.0
	v.PV1Current = float64(v.unpackBase91(b, 32, 3)) / 1000.0
	v.PV2Current = float64(v.unpackBase91(b, 35, 3)) / 1000.0
	v.ExtLoadCurrent = float64(v.unpackBase91(b, 38, 3)) / 1000.0
	v.ADC2 = v.unpackBase91(b, 41, 3)
	v.ADC3 = v.unpackBase91(b, 44, 3)
	v.ADC4 = v.unpackBase91(b, 47, 3)
	v.Heat1 = v.unpackBase91(b, 50, 3)
	v.Heat2 = v.unpackBase91(b, 53, 3)
	v.Status = v.unpackBase91(b, 56, 3)

	if !v.Charging {
		v.BatteryCurrent = -v.BatteryCurrent
	}

	return nil
}

func (v *Values) unpackBase91(b []byte, pos, size int) int {
	n := 0
	for i := 0; i < size; i++ {
		n = n + ((int(b[(pos+size-1)-i]) - 35) * int(math.Pow(91, float64(i))))
	}
	return n
}
