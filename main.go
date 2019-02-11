package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"net/http"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

var addr = flag.String("listen-address", ":9190", "The address to listen on for HTTP requests.")

func metricsHandler(w http.ResponseWriter, r *http.Request) {
	// Parse requests
	params := r.URL.Query()
	target := params.Get("target")
	if target == "" {
		http.Error(w, "Target parameter is missing", http.StatusBadRequest)
		return
	}

	// Prepare metrics
	gaugeLabels := []string{"flight", "hex", "squawk"}
	latitudeGauge := prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Name: "dump1090_latitude",
		Help: "Latitude of the aircraft",
	}, gaugeLabels)
	longitudeGauge := prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Name: "dump1090_longitude",
		Help: "Longitude of the aircraft",
	}, gaugeLabels)
	altitudeGauge := prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Name: "dump1090_altitude",
		Help: "Altitude of the aircraft",
	}, gaugeLabels)
	verticalRateGauge := prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Name: "dump1090_vertical_rate",
		Help: "Vertical rate of the aircraft",
	}, gaugeLabels)
	trackGauge := prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Name: "dump1090_track",
		Help: "Track of the aircraft",
	}, gaugeLabels)
	speedGauge := prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Name: "dump1090_speed",
		Help: "Speed of the aircraft",
	}, gaugeLabels)
	messagesGauge := prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Name: "dump1090_messages",
		Help: "Messages of the aircraft",
	}, gaugeLabels)
	seenGauge := prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Name: "dump1090_seen",
		Help: "Seen of the aircraft",
	}, gaugeLabels)
	registry := prometheus.NewRegistry()
	registry.MustRegister(latitudeGauge)
	registry.MustRegister(longitudeGauge)
	registry.MustRegister(altitudeGauge)
	registry.MustRegister(verticalRateGauge)
	registry.MustRegister(trackGauge)
	registry.MustRegister(speedGauge)
	registry.MustRegister(messagesGauge)
	registry.MustRegister(seenGauge)

	// Collect metrics
	client, err := NewClient("http://"+target, nil)
	if err != nil {
		http.Error(w, "Target parameter is wrong", http.StatusBadRequest)
		return
	}
	ctx := context.Background()
	records, err := client.GetRecords(ctx)
	if err != nil {
		http.Error(w, "Failed to get data from dump1090", http.StatusInternalServerError)
		log.Printf("%+v\n", err)
		return
	}
	for _, record := range *records {
		if !record.ValidPosition || !record.ValidTrack {
			continue
		}
		fmt.Printf("%+v\n", record)
		flight := record.Flight
		if flight == "" {
			flight = "UNKNOWN"
		}
		labels := prometheus.Labels{
			"flight": flight,
			"hex":    record.Hex,
			"squawk": record.Squawk,
		}
		latitudeGauge.With(labels).Set(record.Latitude)
		longitudeGauge.With(labels).Set(record.Longitude)
		altitudeGauge.With(labels).Set(float64(record.Altitude))
		verticalRateGauge.With(labels).Set(float64(record.VerticalRate))
		trackGauge.With(labels).Set(float64(record.Track))
		speedGauge.With(labels).Set(float64(record.Speed))
		messagesGauge.With(labels).Set(float64(record.Messages))
		seenGauge.With(labels).Set(float64(record.Seen))
	}

	// Respond
	h := promhttp.HandlerFor(registry, promhttp.HandlerOpts{})
	h.ServeHTTP(w, r)
}

func main() {
	flag.Parse()

	http.HandleFunc("/metrics", func(w http.ResponseWriter, r *http.Request) {
		metricsHandler(w, r)
	})

	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/html")
		w.Write([]byte(`<html>
	<head><title>dump1090 Exporter</title></head>
	<body>
	<h1>dump1090 Exporter</h1>
	</body>
	</html>
	`))
	})

	log.Printf("Listening on %s\n", *addr)
	// The Handler function provides a default handler to expose metrics
	// via an HTTP server. "/metrics" is the usual endpoint for that.
	// http.Handle("/metrics", promhttp.Handler())
	log.Fatal(http.ListenAndServe(*addr, nil))
}
