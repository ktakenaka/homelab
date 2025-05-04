package influxdb

import (
	"context"
	"fmt"
	"time"

	influxdb2 "github.com/influxdata/influxdb-client-go/v2"
	"github.com/influxdata/influxdb-client-go/v2/api"
	"github.com/influxdata/influxdb-client-go/v2/api/write"
)

// Client represents an InfluxDB client
type Client struct {
	client   influxdb2.Client
	writeAPI api.WriteAPIBlocking
}

// NewClient creates a new InfluxDB client
func NewClient(url, token, org, bucket string) *Client {
	client := influxdb2.NewClient(url, token)
	writeAPI := client.WriteAPIBlocking(org, bucket)

	return &Client{
		client:   client,
		writeAPI: writeAPI,
	}
}

// WriteConsumptionData writes consumption data to InfluxDB
func (c *Client) WriteConsumptionData(deviceID, deviceType string, current, voltage float64) error {
	tags := map[string]string{
		"deviceType": deviceType,
	}
	fields := map[string]interface{}{
		"deviceID": deviceID,
		"current":  current,
		"voltage":  voltage,
		"watt":     current * voltage,
		"kW":       current * voltage / 1000,
	}
	point := write.NewPoint("consumption", tags, fields, time.Now())

	if err := c.writeAPI.WritePoint(context.Background(), point); err != nil {
		return fmt.Errorf("failed to write point to InfluxDB: %w", err)
	}

	return nil
}

// Close closes the InfluxDB client
func (c *Client) Close() {
	c.client.Close()
}
