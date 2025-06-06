package main

import (
	"bytes"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"
)

const (
	apiBaseURL    = "http://localhost:5000/api"
	apiKey        = "secure-api-key-123"
	heartbeatFreq = 15 * time.Second
	clientIP      = "192.168.1.102" // 设置客户端IP，实际使用时应该动态获取
)

type ClientStatus struct {
	Online        bool   `json:"online"`
	LastHeartbeat string `json:"last_heartbeat"`
	IPAddress     string `json:"ip_address"`
}

type Response struct {
	Status       string       `json:"status"`
	Message      string       `json:"message"`
	ClientStatus ClientStatus `json:"client_status,omitempty"`
	Timestamp    string       `json:"timestamp,omitempty"`
}

type RequestData struct {
	IPAddress string `json:"ip_address"`
}

func sendRequest(method, endpoint string, data interface{}) (*Response, error) {
	url := apiBaseURL + endpoint

	var reqBody []byte
	var err error

	if data != nil {
		reqBody, err = json.Marshal(data)
		if err != nil {
			return nil, fmt.Errorf("error marshaling request body: %v", err)
		}
	}

	req, err := http.NewRequest(method, url, bytes.NewBuffer(reqBody))
	if err != nil {
		return nil, fmt.Errorf("error creating request: %v", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+apiKey)
	req.Header.Set("X-Client-IP", clientIP)

	client := &http.Client{
		Timeout: 30 * time.Second,
		Transport: &http.Transport{
			TLSClientConfig:    &tls.Config{InsecureSkipVerify: true},
			MaxIdleConns:       10,
			IdleConnTimeout:    30 * time.Second,
			DisableCompression: true,
		},
	}

	var resp *http.Response
	maxRetries := 3
	for i := 0; i < maxRetries; i++ {
		resp, err = client.Do(req)
		if err == nil {
			break
		}

		if i == maxRetries-1 {
			return nil, fmt.Errorf("after %d attempts, last error: %v", maxRetries, err)
		}

		log.Printf("Attempt %d failed: %v, retrying...", i+1, err)
		time.Sleep(1 * time.Second)
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("error reading response body: %v", err)
	}

	log.Printf("Response from %s: Status=%d, Body=%s", url, resp.StatusCode, string(body))

	if resp.StatusCode >= 400 {
		return nil, fmt.Errorf("server returned error status: %d, body: %s", resp.StatusCode, string(body))
	}

	var response Response
	if err := json.Unmarshal(body, &response); err != nil {
		return nil, fmt.Errorf("error unmarshaling response (body: %s): %v", string(body), err)
	}

	return &response, nil
}

func registerClient() error {
	data := RequestData{IPAddress: clientIP}
	_, err := sendRequest("POST", "/register", data)
	return err
}

func sendHeartbeat() error {
	data := RequestData{IPAddress: clientIP}
	_, err := sendRequest("POST", "/heartbeat", data)
	return err
}

func unregisterClient() error {
	data := RequestData{IPAddress: clientIP}
	_, err := sendRequest("POST", "/unregister", data)
	return err
}

func getStatus() (*ClientStatus, error) {
	data := RequestData{IPAddress: clientIP}
	response, err := sendRequest("GET", "/status", data)
	if err != nil {
		return nil, err
	}

	return &response.ClientStatus, nil
}

func heartbeatLoop() {
	ticker := time.NewTicker(heartbeatFreq)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			if err := sendHeartbeat(); err != nil {
				log.Printf("Error sending heartbeat: %v\n", err)
			} else {
				log.Println("Heartbeat sent successfully")
			}
		}
	}
}

func setupSignalHandling() {
	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		<-sigs
		log.Println("Received termination signal, unregistering client...")
		if err := unregisterClient(); err != nil {
			log.Printf("Error unregistering client: %v\n", err)
		}
		os.Exit(0)
	}()
}

func main() {
	log.SetFlags(log.LstdFlags | log.Lshortfile)

	setupSignalHandling()

	if err := registerClient(); err != nil {
		log.Printf("Failed to register client: %v\n", err)
	} else {
		log.Println("Client registered successfully")
	}

	go heartbeatLoop()

	for {
		status, err := getStatus()
		if err != nil {
			log.Printf("Error getting status: %v\n", err)
		} else {
			log.Printf("Current client status: %+v\n", status)
		}
		time.Sleep(30 * time.Second)
	}
}
