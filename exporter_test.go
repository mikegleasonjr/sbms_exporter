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
	"bytes"
	"flag"
	"io"
	"io/ioutil"
	"net"
	"os"
	"sync"
	"testing"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/common/expfmt"
)

var update = flag.Bool("update", false, "update .metrics files")

func TestMonitor(t *testing.T) {
	reg := prometheus.NewRegistry()
	exp := NewExporter(reg)
	wg := sync.WaitGroup{}
	w, r := net.Pipe()

	wg.Add(1)
	go func() {
		err := exp.Export(r)
		if err != io.EOF {
			t.Errorf("unexpected error: %q", err)
		}
		wg.Done()
	}()

	ensureMetricsEquals(t, reg, `testdata/down.metrics`)
	receiveData(t, w, `testdata/example1.sbms`)
	ensureMetricsEquals(t, reg, `testdata/example1.metrics`)
	receiveData(t, w, `testdata/too-short.sbms`)
	ensureMetricsEquals(t, reg, `testdata/down.metrics`)
	receiveData(t, w, `testdata/example2.sbms`)
	ensureMetricsEquals(t, reg, `testdata/example2.metrics`)

	w.Close()
	wg.Wait()

	ensureMetricsEquals(t, reg, `testdata/down.metrics`)
}

func TestWhitespace(t *testing.T) {
	reg := prometheus.NewRegistry()
	exp := NewExporter(reg)
	wg := sync.WaitGroup{}
	w, r := net.Pipe()

	wg.Add(1)
	go func() {
		err := exp.Export(r)
		if err != io.EOF {
			t.Errorf("unexpected error: %q", err)
		}
		wg.Done()
	}()

	receiveData(t, w, `testdata/whitespace.sbms`)
	ensureMetricsEquals(t, reg, `testdata/whitespace.metrics`)

	w.Close()
	wg.Wait()
}

func receiveData(t *testing.T, w io.Writer, sbms string) {
	t.Helper()

	f, err := os.Open(sbms)
	if err != nil {
		t.Fatal(err)
		return
	}
	defer f.Close()

	if _, err := io.Copy(w, f); err != nil {
		t.Fatalf("unexpected error: %q", err)
	}
	time.Sleep(time.Millisecond)
}

func ensureMetricsEquals(t *testing.T, reg *prometheus.Registry, golden string) {
	t.Helper()

	b := new(bytes.Buffer)
	enc := expfmt.NewEncoder(b, expfmt.FmtText)
	mfs, _ := reg.Gather()

	for _, mf := range mfs {
		if err := enc.Encode(mf); err != nil {
			t.Fatalf("unexpected error: %q", err)
		}
	}

	got := b.Bytes()
	if *update {
		if err := ioutil.WriteFile(golden, got, 0644); err != nil {
			t.Fatalf("unexpected error: %q", err)
		}
	}

	want, _ := ioutil.ReadFile(golden)
	if !bytes.Equal(got, want) {
		t.Errorf("string does not match golden file %s: got '%s', want '%s'", golden, got, want)
	}
}
