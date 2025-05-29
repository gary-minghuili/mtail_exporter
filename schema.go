package main

type MtailMetric struct {
	Name    string `json:"name"`
	Help    string `json:"help"`
	Type    string `json:"type"`
	Metrics []struct {
		Labels map[string]string `json:"labels"`
		Value  string            `json:"value"`
	} `json:"metrics"`
}
