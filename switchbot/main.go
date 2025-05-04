package main

import (
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"math/rand"
	"net/http"
	"os"
	"strconv"
	"time"

	influxdb2 "github.com/influxdata/influxdb-client-go/v2"
	"github.com/influxdata/influxdb-client-go/v2/api/write"
)

type DeviceStatusResponse struct {
	StatusCode int    `json:"statusCode"`
	Message    string `json:"message"`
	Body       struct {
		DeviceId         string  `json:"deviceId"`
		DeviceType       string  `json:"deviceType"`
		HubDeviceId      string  `json:"hubDeviceId"`
		Voltage          float64 `json:"voltage"`
		Weight           float64 `json:"weight"`
		ElectricityOfDay int     `json:"electricityOfDay"`
		ElectricCurrent  float64 `json:"electricCurrent"`
	} `json:"body"`
}

var (
	TOKEN     = os.Getenv("TOKEN")
	SECRET    = os.Getenv("SECRET")
	DEVICE_ID = "3C8427A20152"

	INFLUXDB_TOKEN = os.Getenv("INFLUXDB_TOKEN")
	INFLUXDB_URL   = "http://gmktec01.lan:8086"
	influxclient   = influxdb2.NewClient(INFLUXDB_URL, INFLUXDB_TOKEN)
	writeAPI       = influxclient.WriteAPIBlocking("bamboo", "homelab")
)

func main() {
	// Generate signature, nonce, and timestamp
	sign, nonce, timestamp := generateHMACSignature(TOKEN, SECRET)

	// Set up the API request
	url := fmt.Sprintf("https://api.switch-bot.com/v1.1/devices/%s/status", DEVICE_ID)
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		log.Fatal("Error creating request:", err)
	}

	req.Header.Set("Authorization", TOKEN)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("t", timestamp)
	req.Header.Set("sign", sign)
	req.Header.Set("nonce", nonce)

	// Make the HTTP request
	client := &http.Client{}
	defer client.CloseIdleConnections()

	resp, err := client.Do(req)
	if err != nil {
		log.Fatal("Error making request:", err)
	}
	defer resp.Body.Close()

	// Read and parse the response
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		log.Fatal("Error reading response body:", err)
	}

	if resp.StatusCode != http.StatusOK {
		log.Fatalf("Error: HTTP %d - %s\n", resp.StatusCode, string(body))
	}

	var statusResponse DeviceStatusResponse
	if err := json.Unmarshal(body, &statusResponse); err != nil {
		log.Fatal("Error unmarshalling JSON:", err)
	}

	// Print the status information
	fmt.Println("Device ID:", statusResponse.Body.DeviceId)
	fmt.Println("Device Type:", statusResponse.Body.DeviceType)
	fmt.Println("Hub Device ID:", statusResponse.Body.HubDeviceId)
	fmt.Println("Voltage:", statusResponse.Body.Voltage)
	fmt.Println("Weight:", statusResponse.Body.Weight)
	fmt.Println("Electricity of Day:", statusResponse.Body.ElectricityOfDay)
	fmt.Println("Electric Current:", statusResponse.Body.ElectricCurrent)

	influxDBsample(statusResponse.Body.DeviceId, statusResponse.Body.DeviceType, statusResponse.Body.ElectricCurrent, statusResponse.Body.Voltage)

}

func influxDBsample(deviceID, deviceType string, current, voltage float64) {
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

	if err := writeAPI.WritePoint(context.Background(), point); err != nil {
		log.Fatal(err)
	}
}

func generateHMACSignature(token, secret string) (string, string, string) {
	// Generate nonce and timestamp
	nonce := strconv.Itoa(rand.Int())
	timestamp := strconv.FormatInt(time.Now().UnixMilli(), 10)

	// Create the string to sign
	stringToSign := fmt.Sprintf("%s%s%s", token, timestamp, nonce)

	// Generate HMAC SHA256 signature
	mac := hmac.New(sha256.New, []byte(secret))
	mac.Write([]byte(stringToSign))
	sign := base64.StdEncoding.EncodeToString(mac.Sum(nil))

	return sign, nonce, timestamp
}
