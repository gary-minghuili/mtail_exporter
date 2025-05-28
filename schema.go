package main

type MtailMetric struct {
	Name    string `json:"name"`
	Help    string `json:"help"`
	Type    string `json:"type"`
	Metrics []struct {
		Labels map[string]string `json:"labels"`
		Value  float64           `json:"value"`
	} `json:"metrics"`
}
