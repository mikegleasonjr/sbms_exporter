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
	"context"
	"net/http"
	"sync"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/prometheus/common/log"
	"github.com/prometheus/common/version"
	"gopkg.in/alecthomas/kingpin.v2"
)

func main() {
	metricsPath := kingpin.Flag("telemetry-path", "Path under which to expose metrics.").Default("/metrics").String()
	listenAddress := kingpin.Flag("listen-address", "Address to listen on for web interface and telemetry.").Default(":9101").String()
	serialPort := kingpin.Flag("serial-port", "The serial port to read metrics from.").Required().File()

	log.AddFlags(kingpin.CommandLine)
	kingpin.Version(version.Print("sbms_exporter"))
	kingpin.HelpFlag.Short('h')
	kingpin.Parse()

	http.Handle("/metrics", promhttp.Handler())
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte(`<html>
             <head><title>SBMS Exporter</title></head>
             <body>
             <h1>SBMS Exporter</h1>
             <p><a href='` + *metricsPath + `'>Metrics</a></p>
             </body>
             </html>`))
	})

	srv := &http.Server{Addr: *listenAddress}
	var wg sync.WaitGroup
	log.Infoln("Starting sbms_exporter", version.Info())
	log.Infoln("Listening on", *listenAddress)
	wg.Add(1)
	go func() {
		log.Errorln(srv.ListenAndServe())
		(*serialPort).Close()
		wg.Done()
	}()

	log.Errorln(NewExporter(prometheus.DefaultRegisterer).Export(*serialPort))
	srv.Shutdown(context.Background())
	wg.Wait()
}
