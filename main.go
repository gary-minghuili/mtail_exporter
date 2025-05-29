package main

import (
	"encoding/json"
	"fmt"
	syslog "log"
	"net/http"
	"os"
	"sort"
	"strings"
	
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	dto "github.com/prometheus/client_model/go"
	"github.com/prometheus/prom2json"
	"golang.org/x/exp/maps"
)

const (
	mtailMetricsUrl = "http://localhost:3903/metrics"
	metricsUrl      = "0.0.0.0:3904"
	logPath         = "./mtail_exporter.log"
)

func main() {
	
	err := InitLogger(logPath)
	if err != nil {
		syslog.Println("Failed to initialize logger")
		os.Exit(-1)
	}
	GetLogger().Info("starting exporter ......")
	http.Handle("/metrics", http.HandlerFunc(MetricsHandler))
	err = http.ListenAndServe(metricsUrl, nil)
	if err != nil {
		GetLogger().Error("http.ListenAndServe error")
		os.Exit(-1)
	}
}

func MetricsHandler(writer http.ResponseWriter, request *http.Request) {
	if mtailMetrics, err := getMetrics(); err == nil {
		updateMetrics(mtailMetrics)
	}
	promhttp.Handler().ServeHTTP(writer, request)
}

func getMetrics() ([]MtailMetric, error) {
	var mtailMetrics []MtailMetric
	mfChan := make(chan *dto.MetricFamily, 1024)
	transport, err := makeTransport("", "", true)
	if err != nil {
		GetLogger().Error("transport error")
		return nil, err
	}
	err = prom2json.FetchMetricFamilies(mtailMetricsUrl, mfChan, transport)
	if err != nil {
		GetLogger().Error("prom2json.FetchMetricFamilies error")
		return nil, err
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
		return nil, err
	}
	if err = json.Unmarshal(jsonText, &mtailMetrics); err != nil {
		GetLogger().Error("json unmarshal error")
	}
	return mtailMetrics, nil
}

func updateMetrics(mtailMetrics []MtailMetric) {
	msg := "duplicate metrics collector registration attempted"
	for _, mtailMetric := range mtailMetrics {
		labels := make([]string, 0)
		for _, metricInfo := range mtailMetric.Metrics {
			labels = maps.Keys(metricInfo.Labels)
			break
		}
		sort.Strings(labels)
		// // TODO update labels
		// indexKey := "xxx"
		// if value, ok := switchMap[indexKey]; ok {
		// 	labels = append(labels, value)
		// }
		switch mtailMetric.Type {
		case "GAUGE":
			metric := prometheus.NewGaugeVec(
				prometheus.GaugeOpts{Name: mtailMetric.Name, Help: mtailMetric.Help},
				labels,
			)
			if err := prometheus.Register(metric); err != nil {
				if !strings.Contains(err.Error(), msg) {
					GetLogger().Error("metric collector registration error: " + err.Error())
				}
			}
			for _, metricInfo := range mtailMetric.Metrics {
				// TODO update labels value
				// values := append(maps.Values(metricInfo.Labels), "xxx")
				values := make([]string, 0)
				for _, label := range labels {
					value := metricInfo.Labels[label]
					values = append(values, value)
				}
				metric.WithLabelValues(values...).Set(toFloat64(metricInfo.Value))
			}
		case "COUNTER":
			metric := prometheus.NewCounterVec(
				prometheus.CounterOpts{Name: mtailMetric.Name, Help: mtailMetric.Help},
				labels,
			)
			if err := prometheus.Register(metric); err != nil {
				if !strings.Contains(err.Error(), msg) {
					GetLogger().Error("metric collector registration error: " + err.Error())
				}
			}
			metric.Reset()
			for _, metricInfo := range mtailMetric.Metrics {
				// TODO update labels
				// values := append(maps.Values(metricInfo.Labels), "xxx")
				values := make([]string, 0)
				for _, label := range labels {
					value := metricInfo.Labels[label]
					values = append(values, value)
				}
				metric.WithLabelValues(values...).Add(toFloat64(metricInfo.Value))
			}
		default:
			GetLogger().Error(fmt.Sprintf("unknown metrics type: %s", mtailMetric.Type))
		}
	}
}
