// Package prometheus is a minimal client for the Prometheus HTTP API.
//
// CubeCOS serves Prometheus behind haproxy at /prometheus: the local server on a
// single node, the deduplicating thanos queriers in HA. Pointing at that route,
// not at :9091, follows whichever one is wired.
package prometheus

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strconv"
	"time"
)

type Client struct {
	url  string
	http *http.Client
}

type Sample struct {
	Time  time.Time
	Value float64
}

type response struct {
	Status string `json:"status"`
	Error  string `json:"error"`
	Data   struct {
		Result []struct {
			Values [][2]any `json:"values"`
		} `json:"result"`
	} `json:"data"`
}

var global *Client

func NewClient(url string) *Client {
	return &Client{url: url, http: &http.Client{Timeout: 30 * time.Second}}
}

func NewGlobalClient(url string) {
	global = NewClient(url)
}

func QueryRange(query string, start, end time.Time, step time.Duration) ([]Sample, error) {
	if global == nil {
		return nil, fmt.Errorf("prometheus: client is not initialized")
	}

	return global.QueryRange(query, start, end, step)
}

// QueryRange runs a range query and returns the samples of the first series.
// Callers aggregate in PromQL (sum(...)), so the result holds one series at most.
func (c *Client) QueryRange(query string, start, end time.Time, step time.Duration) ([]Sample, error) {
	params := url.Values{}
	params.Set("query", query)
	params.Set("start", strconv.FormatInt(start.Unix(), 10))
	params.Set("end", strconv.FormatInt(end.Unix(), 10))
	params.Set("step", strconv.FormatInt(int64(step.Seconds()), 10))

	resp, err := c.http.Get(c.url + "/api/v1/query_range?" + params.Encode())
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body := response{}
	err = json.NewDecoder(resp.Body).Decode(&body)
	if err != nil {
		return nil, fmt.Errorf("prometheus: failed to decode response(%s): %v", resp.Status, err)
	}
	if body.Status != "success" {
		return nil, fmt.Errorf("prometheus: query failed(%s): %s", resp.Status, body.Error)
	}
	if len(body.Data.Result) == 0 {
		return []Sample{}, nil
	}

	return parseValues(body.Data.Result[0].Values)
}

// parseValues reads [<unix seconds>, "<value>"] pairs. Prometheus sends the value
// as a string so it can carry NaN and Inf.
func parseValues(values [][2]any) ([]Sample, error) {
	samples := make([]Sample, 0, len(values))
	for _, v := range values {
		ts, ok := v[0].(float64)
		if !ok {
			return nil, fmt.Errorf("prometheus: unexpected timestamp %v", v[0])
		}
		str, ok := v[1].(string)
		if !ok {
			return nil, fmt.Errorf("prometheus: unexpected value %v", v[1])
		}
		value, err := strconv.ParseFloat(str, 64)
		if err != nil {
			return nil, err
		}

		samples = append(samples, Sample{Time: time.Unix(int64(ts), 0), Value: value})
	}

	return samples, nil
}
