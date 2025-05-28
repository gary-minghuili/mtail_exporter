package main

import (
	"encoding/json"
	"fmt"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	dto "github.com/prometheus/client_model/go"
	"github.com/prometheus/prom2json"
	"golang.org/x/exp/maps"
	syslog "log"
	"net/http"
	"os"
	"strings"
)

func main() {
	
	err := InitLogger("./mtail_exporter.log")
	if err != nil {
		syslog.Println("Failed to initialize logger")
		os.Exit(-1)
	}
	GetLogger().Info("starting exporter ......")
	http.Handle("/metrics", http.HandlerFunc(MetricsHandler))
	err = http.ListenAndServe("0.0.0.0:3904", nil)
	if err != nil {
		GetLogger().Error("http.ListenAndServe error")
		panic(err)
	}
}

func MetricsHandler(writer http.ResponseWriter, request *http.Request) {
	mfChan := make(chan *dto.MetricFamily, 1024)
	transport, err := makeTransport("", "", true)
	if err != nil {
		GetLogger().Error("transport error")
		os.Exit(1)
	}
	err = prom2json.FetchMetricFamilies("http://localhost:3903/metrics", mfChan, transport)
	if err != nil {
		GetLogger().Error("prom2json.FetchMetricFamilies error")
		os.Exit(1)
	}
	var result []*prom2json.Family
	
	for mf := range mfChan {
		if *mf.Type == dto.MetricType_COUNTER || *mf.Type == dto.MetricType_GAUGE {
			if !strings.HasPrefix(*mf.Name, "go_") && !strings.HasPrefix(*mf.Name, "process_") {
				result = append(result, prom2json.NewFamily(mf))
			}
		}
	}
	
	jsonText, err := json.Marshal(result)
	if err != nil {
		GetLogger().Error("json marshal error")
		os.Exit(1)
	}
	var mtailMetrics []MtailMetric
	if err = json.Unmarshal(jsonText, &mtailMetrics); err != nil {
		GetLogger().Error("json unmarshal error")
	}
	for _, mtailMetric := range mtailMetrics {
		labels := make([]string, 0)
		for _, metricInfo := range mtailMetric.Metrics {
			labels = maps.Keys(metricInfo.Labels)
			break
		}
		switch mtailMetric.Type {
		case "GAUGE":
			metric := prometheus.NewGaugeVec(
				prometheus.GaugeOpts{Name: mtailMetric.Name, Help: mtailMetric.Help},
				labels,
			)
			prometheus.MustRegister(metric)
			for _, metricInfo := range mtailMetric.Metrics {
				var labelsValue = metricInfo.Labels
				metric.WithLabelValues(maps.Values(labelsValue)...)
			}
		case "COUNTER":
			metric := prometheus.NewCounterVec(
				prometheus.CounterOpts{Name: mtailMetric.Name, Help: mtailMetric.Help},
				labels,
			)
			prometheus.MustRegister(metric)
			for _, metricInfo := range mtailMetric.Metrics {
				metric.WithLabelValues(maps.Values(metricInfo.Labels)...)
			}
		default:
			fmt.Fprintln(os.Stderr, "unknown metrics type:", mtailMetric.Type)
			GetLogger().Error(fmt.Sprintf("unknown metrics type: %s", mtailMetric.Type))
		}
	}
	promhttp.Handler().ServeHTTP(writer, request)
}
