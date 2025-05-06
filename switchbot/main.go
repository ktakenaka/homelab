package main

import (
	"fmt"
	"log"
	"os"

	"github.com/ktakenaka/homelab/switchbot/pkg/influxdb"
	"github.com/ktakenaka/homelab/switchbot/pkg/switchbot"
)

const (
	DEVICE_ID_PLUG  = "3C8427A20152"
	DEVICE_ID_WS    = "6055F933608A"
	DEVICE_ID_THERM = "D53038356238"
	INFLUXDB_URL    = "http://gmktec01.lan:8086"
	INFLUXDB_ORG    = "bamboo"
	INFLUXDB_BUCKET = "homelab"
)

var (
	token         = ""
	secret        = ""
	influxdbToken = ""
)

func main() {
	if token == "" {
		token = os.Getenv("TOKEN")
	}
	if secret == "" {
		secret = os.Getenv("SECRET")
	}
	if influxdbToken == "" {
		influxdbToken = os.Getenv("INFLUXDB_TOKEN")
	}

	if token == "" || secret == "" {
		log.Fatal("TOKEN and SECRET environment variables must be set")
	}

	if influxdbToken == "" {
		log.Fatal("INFLUXDB_TOKEN environment variable must be set")
	}

	// Create SwitchBot client
	switchbotClient := switchbot.NewClient(token, secret)
	defer switchbotClient.Close()

	// Create InfluxDB client
	influxdbClient := influxdb.NewClient(INFLUXDB_URL, influxdbToken, INFLUXDB_ORG, INFLUXDB_BUCKET)
	defer influxdbClient.Close()

	// Get device status
	statusResponse, err := switchbotClient.GetDeviceStatus(DEVICE_ID_PLUG)
	if err != nil {
		log.Fatalf("Failed to get device status: %v", err)
	}

	// Write data to InfluxDB
	err = influxdbClient.WriteConsumptionData(
		statusResponse.Body.DeviceId,
		statusResponse.Body.DeviceType,
		statusResponse.Body.ElectricCurrent,
		statusResponse.Body.Voltage,
	)
	if err != nil {
		log.Fatalf("Failed to write data to InfluxDB: %v", err)
	}

	statusResponse, err = switchbotClient.GetDeviceStatus(DEVICE_ID_THERM)
	if err != nil {
		log.Fatalf("Failed to get device status: %v", err)
	}

	// Write data to InfluxDB
	err = influxdbClient.WriteThermoData(
		statusResponse.Body.DeviceId,
		statusResponse.Body.DeviceType,
		statusResponse.Body.Temperature,
		statusResponse.Body.Humidity,
	)
	if err != nil {
		log.Fatalf("Failed to write data to InfluxDB: %v", err)
	}

	// Get device status for DEVICE_ID_WS
	statusResponse, err = switchbotClient.GetDeviceStatus(DEVICE_ID_WS)
	if err != nil {
		log.Fatalf("Failed to get device status for WS: %v", err)
	}

	// Write consumption data to InfluxDB for DEVICE_ID_WS
	err = influxdbClient.WriteConsumptionData(
		statusResponse.Body.DeviceId,
		statusResponse.Body.DeviceType,
		statusResponse.Body.ElectricCurrent,
		statusResponse.Body.Voltage,
	)
	if err != nil {
		log.Fatalf("Failed to write WS data to InfluxDB: %v", err)
	}

	fmt.Println("Successfully wrote data to InfluxDB")
}
