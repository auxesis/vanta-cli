package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
)

// TokenResponse represents an authentication token returned by the /oauth/token endpoint
type TokenResponse struct {
	AccessToken string `json:"access_token"`
	ExpiresIn   int    `json:"expires_in"`
	TokenType   string `json:"token_type"`
}

func getOAuthToken(clientID, clientSecret string) (*TokenResponse, error) {
	payload, err := json.Marshal(map[string]string{
		"client_id":     clientID,
		"client_secret": clientSecret,
		"scope":         "vanta-api.all:read",
		"grant_type":    "client_credentials",
	})
	if err != nil {
		return nil, err
	}

	resp, err := http.Post("https://api.vanta.com/oauth/token", "application/json", bytes.NewReader(payload))
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var token TokenResponse
	if err := json.NewDecoder(resp.Body).Decode(&token); err != nil {
		return nil, err
	}
	return &token, nil
}

func main() {
	clientID := os.Getenv("VANTA_CLIENT_ID")
	clientSecret := os.Getenv("VANTA_CLIENT_SECRET")

	token, err := getOAuthToken(clientID, clientSecret)
	if err != nil {
		log.Fatal(err)
	}

	//url := "https://api.vanta.com/v1/customer-trust/accounts"
	url := "https://api.vanta.com/v1/vulnerabilities"

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		log.Fatal(err)
	}

	req.Header.Add("accept", "application/json")
	req.Header.Add("authorization", "Bearer "+token.AccessToken)

	res, err := http.DefaultClient.Do(req)
	if err != nil {
		log.Fatal(err)
	}
	defer res.Body.Close()

	body, err := io.ReadAll(res.Body)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println(string(body))
}
