package switchbot

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"math/rand"
	"net/http"
	"strconv"
	"time"
)

// Client represents a SwitchBot API client
type Client struct {
	Token  string
	Secret string
	HTTP   *http.Client
}

// DeviceStatusResponse represents the response from the SwitchBot API for device status
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
		Temperature      float64 `json:"temperature"`
		Humidity         float64 `json:"humidity"`
	} `json:"body"`
}

// NewClient creates a new SwitchBot client
func NewClient(token, secret string) *Client {
	return &Client{
		Token:  token,
		Secret: secret,
		HTTP:   &http.Client{},
	}
}

// GetDeviceStatus retrieves the status of a device
func (c *Client) GetDeviceStatus(deviceID string) (*DeviceStatusResponse, error) {
	// Generate signature, nonce, and timestamp
	sign, nonce, timestamp := c.generateHMACSignature()

	// Set up the API request
	url := fmt.Sprintf("https://api.switch-bot.com/v1.1/devices/%s/status", deviceID)
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("error creating request: %w", err)
	}

	req.Header.Set("Authorization", c.Token)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("t", timestamp)
	req.Header.Set("sign", sign)
	req.Header.Set("nonce", nonce)

	// Make the HTTP request
	resp, err := c.HTTP.Do(req)
	if err != nil {
		return nil, fmt.Errorf("error making request: %w", err)
	}
	defer resp.Body.Close()

	// Read and parse the response
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("error reading response body: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("HTTP error: %d - %s", resp.StatusCode, string(body))
	}

	var statusResponse DeviceStatusResponse
	if err := json.Unmarshal(body, &statusResponse); err != nil {
		return nil, fmt.Errorf("error unmarshalling JSON: %w", err)
	}

	return &statusResponse, nil
}

// Close closes any idle connections
func (c *Client) Close() {
	c.HTTP.CloseIdleConnections()
}

// generateHMACSignature generates the HMAC signature for SwitchBot API authentication
func (c *Client) generateHMACSignature() (string, string, string) {
	// Generate nonce and timestamp
	nonce := strconv.Itoa(rand.Int())
	timestamp := strconv.FormatInt(time.Now().UnixMilli(), 10)

	// Create the string to sign
	stringToSign := fmt.Sprintf("%s%s%s", c.Token, timestamp, nonce)

	// Generate HMAC SHA256 signature
	mac := hmac.New(sha256.New, []byte(c.Secret))
	mac.Write([]byte(stringToSign))
	sign := base64.StdEncoding.EncodeToString(mac.Sum(nil))

	return sign, nonce, timestamp
}
