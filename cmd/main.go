package main

import (
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/gorilla/mux"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

func recordMetrics() {
	go func() {
		for {
			opsProcessed.Inc()
			time.Sleep(2 * time.Second)
		}
	}()
}

var (
	opsProcessed = promauto.NewCounter(prometheus.CounterOpts{
		Name: "myapp_processed_ops_total",
		Help: "The total number of processed events",
	})
)

func main() {
	recordMetrics()

	// Prometheus: Histogram to collect required metrics
	histogram := prometheus.NewHistogramVec(prometheus.HistogramOpts{
		Name:    "say_hi_seconds",
		Help:    "Time take to say hi",
		Buckets: []float64{1, 2, 5, 6, 10},
	}, []string{"code"}) // this will be partitioned by the HTTP code.

	router := mux.NewRouter()
	router.Handle("/hi/{name}", SayHello(histogram))
	router.Handle("/metrics", promhttp.Handler())

	//Registering the metric with Prometheus
	prometheus.Register(histogram)

	log.Fatal(http.ListenAndServe(":8009", router))
}

// SayHello :: /hi/{name}
func SayHello(histogram *prometheus.HistogramVec) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {

		//monitoring how long it takes to respond
		start := time.Now()
		defer r.Body.Close()
		code := 500

		defer func() {
			httpDuration := time.Since(start)
			histogram.WithLabelValues(fmt.Sprintf("%d", code)).Observe(httpDuration.Seconds())
		}()

		code = http.StatusBadRequest
		if r.Method == "GET" {
			code = http.StatusOK
			vars := mux.Vars(r)
			name := vars["name"]

			greet := fmt.Sprintf("Hello %s \n", name)
			w.Write([]byte(greet))
		} else {
			w.WriteHeader(code)
		}
	}
}
