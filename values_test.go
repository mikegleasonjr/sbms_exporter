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
	"github.com/google/go-cmp/cmp"

	"testing"
	"time"
)

func TestValuesReadFrom(t *testing.T) {
	testCases := []struct {
		data   []byte
		values *Values
		err    error
	}{
		{
			data:   nil,
			values: &Values{},
			err:    ErrDataLength,
		},
		{
			data:   make([]byte, 60),
			values: &Values{},
			err:    ErrDataLength,
		},
		{
			data: []byte("3';2LD$,I)I*I+I+H}I%I+I**h##+#)P####->##################%N("),
			values: &Values{
				Date:           time.Date(2016, 4, 24, 15, 41, 33, 0, time.UTC),
				StateOfCharge:  100,
				Cell1Voltage:   3.464,
				Cell2Voltage:   3.465,
				Cell3Voltage:   3.466,
				Cell4Voltage:   3.466,
				Cell5Voltage:   3.457,
				Cell6Voltage:   3.460,
				Cell7Voltage:   3.466,
				Cell8Voltage:   3.465,
				InternalTemp:   25.6,
				ExternalTemp:   -45.0,
				Charging:       true,
				BatteryCurrent: 0.591,
				PV1Current:     0,
				PV2Current:     0.937,
				ExtLoadCurrent: 0,
				ADC2:           0,
				ADC3:           0,
				ADC4:           0,
				Heat1:          0,
				Heat2:          0,
				Status:         20480,
			},
			err: nil,
		},
	}
	for _, tC := range testCases {
		t.Run(string(tC.data), func(t *testing.T) {
			values := new(Values)
			err := values.ReadFrom(tC.data)

			if got, want := err, tC.err; got != want {
				t.Errorf("unexpected error: got %q, want %q", got, want)
			}

			if diff := cmp.Diff(tC.values, values); diff != "" {
				t.Errorf("values mismatch (-want +got):\n%s", diff)
			}
		})
	}
}
