package main

import (
	"crypto/tls"
	"net/http"
	"strconv"
	"time"
)

var (
	switchMap = map[string]string{
		"GAUGE": "gauge",
	}
)

func makeTransport(
	certificate string, key string,
	skipServerCertCheck bool,
) (*http.Transport, error) {
	transport := http.DefaultTransport.(*http.Transport).Clone()
	transport.DisableKeepAlives = true
	transport.ResponseHeaderTimeout = time.Minute
	tlsConfig := &tls.Config{InsecureSkipVerify: skipServerCertCheck}
	if certificate != "" && key != "" {
		cert, err := tls.LoadX509KeyPair(certificate, key)
		if err != nil {
			return nil, err
		}
		tlsConfig.Certificates = []tls.Certificate{cert}
	}
	transport.TLSClientConfig = tlsConfig
	return transport, nil
}

func toFloat64(s string) float64 {
	if f, err := strconv.ParseFloat(s, 64); err == nil {
		return f
	}
	return 0
}
