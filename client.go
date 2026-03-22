package main

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/url"
	"os"

	"log"
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

// Client manages communication with the Vanta API.
type Client struct {
	BaseURL    string
	httpClient *http.Client
	token      string
}

// NewClient authenticates with the Vanta API and returns a ready-to-use Client.
func NewClient(clientID, clientSecret string) (*Client, error) {
	token, err := getOAuthToken(clientID, clientSecret)
	if err != nil {
		return nil, err
	}
	return &Client{
		BaseURL:    "https://api.vanta.com/v1",
		httpClient: http.DefaultClient,
		token:      token.AccessToken,
	}, nil
}

func fetchAll[T any](c *Client, path string, params ...url.Values) ([]T, error) {
	var all []T
	cursor := ""

	for {
		req, err := http.NewRequest("GET", c.BaseURL+path, nil)
		if err != nil {
			return nil, err
		}
		q := req.URL.Query()
		q.Set("pageSize", "100")
		if cursor != "" {
			q.Set("pageCursor", cursor)
		}
		if len(params) > 0 {
			for key, vals := range params[0] {
				for _, v := range vals {
					q.Add(key, v)
				}
			}
		}
		req.URL.RawQuery = q.Encode()
		req.Header.Add("accept", "application/json")
		req.Header.Add("authorization", "Bearer "+c.token)

		res, err := c.httpClient.Do(req)
		if err != nil {
			return nil, err
		}

		var wrapper struct {
			Results struct {
				PageInfo struct {
					EndCursor   string `json:"endCursor"`
					HasNextPage bool   `json:"hasNextPage"`
				} `json:"pageInfo"`
				Data []T `json:"data"`
			} `json:"results"`
		}
		err = json.NewDecoder(res.Body).Decode(&wrapper)
		res.Body.Close()
		if err != nil {
			return nil, err
		}

		all = append(all, wrapper.Results.Data...)

		if !wrapper.Results.PageInfo.HasNextPage {
			break
		}
		cursor = wrapper.Results.PageInfo.EndCursor
	}

	return all, nil
}

func fetchByID[T any](c *Client, path string) (T, error) {
	var zero T
	req, err := http.NewRequest("GET", c.BaseURL+path, nil)
	if err != nil {
		return zero, err
	}
	req.Header.Add("accept", "application/json")
	req.Header.Add("authorization", "Bearer "+c.token)

	res, err := c.httpClient.Do(req)
	if err != nil {
		return zero, err
	}
	defer res.Body.Close()

	var result T
	if err := json.NewDecoder(res.Body).Decode(&result); err != nil {
		return zero, err
	}
	return result, nil
}

func newClient() *Client {
	client, err := NewClient(os.Getenv("VANTA_CLIENT_ID"), os.Getenv("VANTA_CLIENT_SECRET"))
	if err != nil {
		log.Fatal(err)
	}
	return client
}
