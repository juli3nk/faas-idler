package main

import (
	"fmt"
	"log"
	"net/http"
	"net/url"
	"strconv"
	"time"

	"github.com/openfaas/faas/gateway/metrics"
)

type MetricConfig struct {
	Host               string
	Port               int
	InactivityDuration uint32
}

func NewMetric(host string, port int, inactivityDuration uint32) (*MetricConfig, error) {
	return &MetricConfig{
		Host:               host,
		Port:               port,
		InactivityDuration: inactivityDuration,
	}, nil
}

func (c *MetricConfig) Get(functionName string) (float64, error) {
	var metric float64

	duration := fmt.Sprintf("%dm", c.InactivityDuration)

	queryStr := url.QueryEscape(`sum(rate(gateway_function_invocation_total{function_name="` + functionName + `"}[` + duration + `]))`)

	client := &http.Client{
		Timeout: time.Second * 10,
	}

	query := metrics.NewPrometheusQuery(c.Host, c.Port, client)
	response, err := query.Fetch(queryStr)
	if err != nil {
		return -1.0, err
	}

	for _, v := range response.Data.Result {
		if v.Metric.FunctionName == functionName {
			metricValue := v.Value[1]

			switch metricValue.(type) {
			case string:
				f, err := strconv.ParseFloat(metricValue.(string), 64)
				if err != nil {
					log.Printf("unable to convert value for metric: %s\n", err)
					continue
				}

				metric = f
				break
			}
		}
	}

	return metric, nil
}
