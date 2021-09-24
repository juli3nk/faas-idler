package main

import (
	"log"
	"os"
	"os/signal"
	"strings"
	"time"

	"github.com/kelseyhightower/envconfig"
)

func main() {
	ecfg := new(Config)
	if err := envconfig.Process("app_config", ecfg); err != nil {
		log.Fatal(err)
	}

	egw := new(Gateway)
	if err := envconfig.Process("app_gateway", egw); err != nil {
		log.Fatal(err)
	}

	emetric := new(Metric)
	if err := envconfig.Process("app_metric", emetric); err != nil {
		log.Fatal(err)
	}

	gwUsername, err := readSecret(egw.UsernameFile)
	if err != nil {
		log.Fatal(err)
	}

	gwPassword, err := readSecret(egw.PasswordFile)
	if err != nil {
		log.Fatal(err)
	}

	fc, err := NewFunction(egw.URL, gwUsername, gwPassword)
	if err != nil {
		log.Fatal(err)
	}

	mc, err := NewMetric(emetric.Host, emetric.Port, emetric.InactivityDuration)
	if err != nil {
		log.Fatal(err)
	}

	ticker := time.NewTicker(time.Second * time.Duration(ecfg.Interval))
	done := make(chan bool)

	go func() {
		for {
			select {
			case <-done:
				return
			case <-ticker.C:
				if err := reconcile(fc, mc); err != nil {
					log.Println(err)
				}
			}
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, os.Interrupt)
	<-quit

	ticker.Stop()
	done <- true
}

func readSecret(filename string) (string, error) {
	b, err := os.ReadFile(filename)
	if err != nil {
		return "", err
	}
	result := strings.TrimSpace(string(b))

	return result, nil
}

func reconcile(fc *FunctionConfig, mc *MetricConfig) error {
	// Get functions
	functions, err := fc.ListScalableFunctions()
	if err != nil {
		return err
	}

	// Get metrics
	functionMetrics := make(map[string]float64)

	for _, f := range functions {
		metric, err := mc.Get(f)
		if err != nil {
			log.Println(err)
			continue
		}

		functionMetrics[f] = metric
	}

	for f, m := range functionMetrics {
		if m == 0 {
			if err := fc.ScaleToZero(f); err != nil {
				log.Println(err)
				continue
			}
		}
	}

	return nil
}
