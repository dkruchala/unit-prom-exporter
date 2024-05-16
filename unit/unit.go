package unit

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"net"
	"net/http"
)

var socketPath string = "/var/run/control.unit.sock"

type ControlApi struct {
	Client *http.Client
}

type UnitMetrics struct {
	Connections  ConnectionsData            `json:"connections"`
	Requests     RequestsData               `json:"requests"`
	Applications map[string]ApplicationData `json:"applications"`
}

type ConnectionsData struct {
	Accepted float64 `json:"accepted"`
	Active   float64 `json:"active"`
	Closed   float64 `json:"closed"`
	Idle     float64 `json:"idle"`
}

type RequestsData struct {
	Total float64 `json:"total"`
}

type ApplicationData struct {
	Requests  ApplicationRequestsData   `json:"requests"`
	Processes ApplicationsProcessesData `json:"processes"`
}

type ApplicationRequestsData struct {
	Active float64 `json:"active"`
}

type ApplicationsProcessesData struct {
	Idle     float64 `json:"idle"`
	Running  float64 `json:"running"`
	Starting float64 `json:"starting"`
}

func NewControlApiConnection() *ControlApi {
	return &ControlApi{
		Client: &http.Client{
			Transport: &http.Transport{
				DialContext: func(_ context.Context, _, _ string) (net.Conn, error) {
					return net.Dial("unix", socketPath)
				},
			},
		},
	}
}

func (c *ControlApi) GetStatus() (UnitMetrics, error) {
	response, err := c.Client.Get("http://unix/status")
	if err != nil {
		return UnitMetrics{}, errors.New("cannot retrieve Unit status")
	}

	var metrics UnitMetrics
	body, err := io.ReadAll(response.Body)

	if err != nil {
		return UnitMetrics{}, err
	}

	err = json.Unmarshal(body, &metrics)

	if err != nil {
		return UnitMetrics{}, err
	}

	return metrics, nil
}
