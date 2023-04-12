package unbroken

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"net/http"
	"strings"

	"github.com/samber/lo"
	"go.uber.org/zap"
)

type PrometheusMetric struct {
	Name   string            `json:"name"`
	Value  string            `json:"value"`
	Labels map[string]string `json:"labels"`
}

func (m *PrometheusMetric) String() string {
	metric := m.Name

	labels := lo.MapToSlice(m.Labels, func(key, value string) string {
		return key + "=" + value
	})

	return strings.Join(append([]string{metric}, labels...), ",") + " metric=" + m.Value
}

func PushToGrafana(metrics []*PrometheusMetric, url, key string) error {
	body := strings.Join(lo.Map(metrics, func(m *PrometheusMetric, _ int) string {
		return m.String()
	}), "\n")

	log.Info("body", zap.Any("body", body))

	req, err := http.NewRequest("POST", url, bytes.NewBuffer([]byte(body)))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "text/plain")
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", key))

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	respBody, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return err
	}

	log.Info("grafana push done", zap.Any("status", resp.Status), zap.Any("body", string(respBody)))

	return nil
}
