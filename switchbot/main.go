package main

import (
	"fmt"
	"log"
	"os"

	"github.com/ktakenaka/homelab/switchbot/pkg/influxdb"
	"github.com/ktakenaka/homelab/switchbot/pkg/switchbot"
)

const (
	DEVICE_ID     = "3C8427A20152"
	INFLUXDB_URL  = "http://gmktec01.lan:8086"
	INFLUXDB_ORG  = "bamboo"
	INFLUXDB_BUCKET = "homelab"
)

func main() {
	// Get environment variables
	token := os.Getenv("TOKEN")
	secret := os.Getenv("SECRET")
	influxdbToken := os.Getenv("INFLUXDB_TOKEN")

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
	statusResponse, err := switchbotClient.GetDeviceStatus(DEVICE_ID)
	if err != nil {
		log.Fatalf("Failed to get device status: %v", err)
	}

	// Print the status information
	fmt.Println("Device ID:", statusResponse.Body.DeviceId)
	fmt.Println("Device Type:", statusResponse.Body.DeviceType)
	fmt.Println("Hub Device ID:", statusResponse.Body.HubDeviceId)
	fmt.Println("Voltage:", statusResponse.Body.Voltage)
	fmt.Println("Weight:", statusResponse.Body.Weight)
	fmt.Println("Electricity of Day:", statusResponse.Body.ElectricityOfDay)
	fmt.Println("Electric Current:", statusResponse.Body.ElectricCurrent)

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

	fmt.Println("Successfully wrote data to InfluxDB")
}
