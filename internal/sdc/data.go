package sdc

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
)

const dataBasePath = "data"

type DataService interface {
	Get(context.Context, *GetDataRequest) (*GetDataResponse, *Response, error)
	Metrics(context.Context) (Metrics, *Response, error)
}

// DataServiceOp handles communication with Data methods of the Sysdig Cloud
// API.
type DataServiceOp struct {
	client *Client
}

var _ DataService = &DataServiceOp{}

type Metric struct {
	ID           string            `json:"metric"`
	Aggregations MetricAggregation `json:"aggregations,omitempty"`
}

type MetricAggregation struct {
	Time  string `json:"time"`
	Group string `json:"group"`
}

type GetDataRequest struct {
	Metrics        []Metric `json:"metrics"`
	DataSourceType string   `json:"dataSourceType,omitempty"`
	Start          int      `json:"start,omitempty"`
	End            int      `json:"end,omitempty"`
	Last           int      `json:"last,omitempty"`
	Filter         string   `json:"filter,omitempty"`
	Paging         string   `json:"paging,omitempty"`
	Sampling       int      `json:"sampling,omitempty"`
}

func (gdr *GetDataRequest) WithMetric(id string, aggregation *MetricAggregation) *GetDataRequest {
	m := &Metric{ID: id}
	if aggregation != nil {
		m.Aggregations = *aggregation
	}
	gdr.Metrics = append(gdr.Metrics, *m)
	return gdr
}

func (gdr *GetDataRequest) WithFilter(filter string) *GetDataRequest {
	gdr.Filter = filter
	return gdr
}

type GetDataResponse struct {
	// A list of time samples.
	Samples []TimeSample `json:"data"`
	Start   Timestamp    `json:"start"`
	End     Timestamp    `json:"end"`
}

type TimeSample struct {
	Time   Timestamp         `json:"t"`
	Values []json.RawMessage `json:"d"`
}

func (gdr *GetDataResponse) FirstValue() (json.RawMessage, error) {
	if len(gdr.Samples) < 1 {
		return nil, errors.New("zero time samples")
	}
	sample := gdr.Samples[0]
	if len(sample.Values) < 1 {
		return nil, errors.New("zero values found in the first sample")
	}
	return sample.Values[0], nil
}

func (s *DataServiceOp) Get(ctx context.Context, gdr *GetDataRequest) (*GetDataResponse, *Response, error) {
	path := fmt.Sprintf("%s/", dataBasePath)
	req, err := s.client.NewRequest(ctx, http.MethodPost, path, gdr)
	if err != nil {
		return nil, nil, err
	}
	data := &GetDataResponse{}
	resp, err := s.client.Do(ctx, req, data)
	if err != nil {
		return nil, resp, err
	}
	return data, resp, nil
}

type Metrics map[string]MetricDefinition

type MetricDefinition struct {
	ID          string   `json:"id"`
	Name        string   `json:"name"`
	Description string   `json:"description"`
	CanMonitor  bool     `json:"canMonitor"`
	Hidden      bool     `json:"hidden"`
	GroupBy     []string `json:"groupBy"`
	Namespaces  []string `json:"namespaces"`

	// Possible values:
	// - "%" (percentage)
	// - "byte"
	// - "date"
	// - "double"
	// - "int"
	// - "number"
	// - "relativeTime"
	// - "string"
	Type string `json:"type"`

	// Possible values:
	// - "counter"
	// - "gauge"
	// - "none", e.g.: id=kubernetes.service.name
	// - "segmentBy", e.g.: id=host, id=port
	MetricType string `json:"metricType"`
}

type MetricsList struct {
	Total             int                 `json:"total"`
	Offset            int                 `json:"offset"`
	MetricDescriptors []MetricDescriptors `json:"metricDescriptors"`
}

type MetricDescriptors struct {
	ID                string        `json:"id"`
	MetricType        string        `json:"metricType"`
	Type              string        `json:"type"`
	Scale             float64       `json:"scale"`
	Category          string        `json:"category"`
	Namespaces        []string      `json:"namespaces"`
	Scopes            []interface{} `json:"scopes"`
	TimeAggregations  []string      `json:"timeAggregations"`
	GroupAggregations []string      `json:"groupAggregations"`
	Identity          bool          `json:"identity"`
	CanMonitor        bool          `json:"canMonitor"`
	CanGroupBy        bool          `json:"canGroupBy"`
	CanFilter         bool          `json:"canFilter"`
	Heuristic         bool          `json:"heuristic"`
}

func (s *DataServiceOp) Metrics(ctx context.Context) (Metrics, *Response, error) {
	initialPath := fmt.Sprintf("v2/metrics/descriptors")
	limit := 5000
	offset := 0
	path := initialPath + "?limit=" + fmt.Sprint(limit) + "&offset=" + fmt.Sprint(offset)
	req, err := s.client.NewRequest(ctx, http.MethodGet, path, nil)
	if err != nil {
		return nil, nil, err
	}
	metrics := MetricsList{}
	resp, err := s.client.Do(ctx, req, &metrics)
	cont := true
	metricsOld := Metrics{}
	for cont == true {
		for _, m := range metrics.MetricDescriptors {
			metricDefinition := MetricDefinition{
				ID:          m.ID,
				Name:        "",
				Description: "",
				CanMonitor:  false,
				Hidden:      false,
				GroupBy:     nil,
				Namespaces:  m.Namespaces,
				Type:        m.Type,
				MetricType:  m.MetricType,
			}
			//We only save the metrics with the label kubernetes.cluster
			if hasNamespace(metricDefinition.Namespaces, "kubernetes.cluster") {
				metricsOld[m.ID] = metricDefinition
			}
		}
		if len(metrics.MetricDescriptors) == limit {
			offset = offset + limit
		} else {
			cont = false
			return metricsOld, resp, nil
		}
		path = initialPath + "?limit=" + fmt.Sprint(limit) + "&offset=" + fmt.Sprint(offset)
		req, err = s.client.NewRequest(ctx, http.MethodGet, path, nil)
		resp, err = s.client.Do(ctx, req, &metrics)
		if err != nil {
			return nil, nil, err
		}
	}

	return metricsOld, resp, nil
}

func hasNamespace(namespaces []string, wanted string) bool {
	for _, item := range namespaces {
		if item == wanted {
			return true
		}
	}
	return false
}
