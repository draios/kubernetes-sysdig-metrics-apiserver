package sdc

import (
	"context"
	"fmt"
	"net/http"
)

const dataBasePath = "data"

type DataService interface {
	Get(context.Context, *GetDataRequest) (*GetDataResponse, *Response, error)
	Metrics(context.Context) (*GetDataMetricsResponse, *Response, error)
}

// DataServiceOp handles communication with Data methods of the Sysdig Cloud
// API.
type DataServiceOp struct {
	client *Client
}

var _ DataService = &DataServiceOp{}

type Metric struct {
	ID           string            `json:"id"`
	Aggregations MetricAggregation `json:"aggregations"`
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

type GetDataResponse struct {
	Data []DataItem `json:"data"`
}

type DataItem struct {
	Points []interface{} `json:"d"`
	Time   Timestamp     `json:"t"`
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

type GetDataMetricsResponse map[string]MetricDefinition

type MetricDefinition struct {
	Name        string   `json:"name"`
	Description string   `json:"description"`
	CanMonitor  bool     `json:"canMonitor"`
	Hidden      bool     `json:"hidden`
	GroupBy     []string `json:"groupBy"`
	Namespaces  []string `json:"namespaces"`
	Type        string   `json:"type"`

	// Possible values:
	// - "%" (percentage)
	// - "byte"
	// - "date"
	// - "double"
	// - "int"
	// - "number"
	// - "relativeTime"
	// - "string"
	MetricType string `json:"metricType"`
}

func (s *DataServiceOp) Metrics(ctx context.Context) (*GetDataMetricsResponse, *Response, error) {
	path := fmt.Sprintf("%s/metrics", dataBasePath)

	req, err := s.client.NewRequest(ctx, http.MethodGet, path, nil)
	if err != nil {
		return nil, nil, err
	}

	metrics := &GetDataMetricsResponse{}
	resp, err := s.client.Do(ctx, req, metrics)
	if err != nil {
		return nil, resp, err
	}

	return metrics, resp, nil
}
