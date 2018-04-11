package sdc

import "context"

const dataBasePath = "/data/"

type DataService interface {
	// https://github.com/draios/python-sdc-client/blob/master/sdcclient/_client.py#L403-L451
	Get(context.Context, *GetDataRequest) (*GetDataResponse, *Response, error)
}

// DataServiceOp handles communication with Data methods of the Sysdig Cloud
// API.
type DataServiceOp struct {
	client *Client
}

var _ DataService = &DataServiceOp{}

type GetDataRequest struct {
	Metrics        string `json:"metrics"`
	DataSourceType string `json:"dataSourceType"`
	Start          string `json:"start"`
	End            string `json:"end"`
	Last           string `json:"last"`
	Filter         string `json:"filter"`
	Paging         string `json:"paging"`
	Sampling       string `json:"sampling"`
}

type GetDataResponse struct{}

func (op *DataServiceOp) Get(ctx context.Context, req *GetDataRequest) (*GetDataResponse, *Response, error) {
	return nil, nil, nil
}
