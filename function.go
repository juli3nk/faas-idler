package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/openfaas/faas-provider/types"
)

type FunctionConfig struct {
	URL      string
	Username string
	Password string
}

func NewFunction(url, username, password string) (*FunctionConfig, error) {
	cfg := FunctionConfig{
		URL:      url,
		Username: username,
		Password: password,
	}

	return &cfg, nil
}

func (c *FunctionConfig) ListScalableFunctions() ([]string, error) {
	var result []string

	response, err := c.request("GET", "/system/functions", nil)
	if err != nil {
		return nil, err
	}

	var functions []types.FunctionStatus
	if err := json.Unmarshal(response, &functions); err != nil {
		return nil, err
	}

	for _, f := range functions {
		if f.Labels != nil {
			v, ok := (*f.Labels)["com.openfaas.scale.zero"]
			if ok {
				if v == "true" && f.Replicas > 0 {
					result = append(result, f.Name)
				}
			}
		}
	}

	return result, nil
}

func (c *FunctionConfig) ScaleToZero(functionName string) error {
	payload := struct {
		Replicas int
	}{Replicas: 0}

	body, err := json.Marshal(payload)
	if err != nil {
		return err
	}

	path := fmt.Sprintf("/system/scale-function/%s", functionName)

	if _, err = c.request(http.MethodPost, path, body); err != nil {
		return err
	}

	return nil
}

func (c *FunctionConfig) request(method, path string, body []byte) ([]byte, error) {
	url := fmt.Sprintf("%s%s", c.URL, path)

	client := &http.Client{
		Timeout: time.Second * 10,
	}

	req, err := http.NewRequest(method, url, bytes.NewBuffer(body))
	if err != nil {
		return nil, err
	}
	req.SetBasicAuth(c.Username, c.Password)

	response, err := client.Do(req)
	if err != nil {
		return nil, err
	}

	respBody, err := io.ReadAll(response.Body)
	if err != nil {
		return nil, err
	}

	if err := response.Body.Close(); err != nil {
		return nil, err
	}

	return respBody, nil
}
