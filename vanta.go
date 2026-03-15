package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"time"
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

// DeactivateMetadata holds information about why and when a vulnerability was deactivated.
type DeactivateMetadata struct {
	DeactivatedBy                 string     `json:"deactivatedBy"`
	DeactivatedOnDate             time.Time  `json:"deactivatedOnDate"`
	DeactivationReason            string     `json:"deactivationReason"`
	DeactivatedUntilDate          *time.Time `json:"deactivatedUntilDate"`
	IsVulnDeactivatedIndefinitely bool       `json:"isVulnDeactivatedIndefinitely"`
}

// Vulnerability represents a single vulnerability returned by the Vanta API.
type Vulnerability struct {
	ID                 string              `json:"id"`
	Name               string              `json:"name"`
	Description        string              `json:"description"`
	IntegrationID      string              `json:"integrationId"`
	PackageIdentifier  string              `json:"packageIdentifier"`
	VulnerabilityType  string              `json:"vulnerabilityType"`
	TargetID           string              `json:"targetId"`
	FirstDetectedDate  time.Time           `json:"firstDetectedDate"`
	SourceDetectedDate *time.Time          `json:"sourceDetectedDate"`
	LastDetectedDate   *time.Time          `json:"lastDetectedDate"`
	Severity           string              `json:"severity"`
	CVSSSeverityScore  float64             `json:"cvssSeverityScore"`
	ScannerScore       *int                `json:"scannerScore"`
	IsFixable          bool                `json:"isFixable"`
	RemediateByDate    time.Time           `json:"remediateByDate"`
	RelatedVulns       []string            `json:"relatedVulns"`
	RelatedURLs        []string            `json:"relatedUrls"`
	ExternalURL        string              `json:"externalURL"`
	ScanSource         string              `json:"scanSource"`
	DeactivateMetadata *DeactivateMetadata `json:"deactivateMetadata"`
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

// Vulnerabilities fetches all vulnerabilities from the Vanta API.
func (c *Client) Vulnerabilities() ([]Vulnerability, error) {
	req, err := http.NewRequest("GET", c.BaseURL+"/vulnerabilities", nil)
	if err != nil {
		return nil, err
	}
	q := req.URL.Query()
	q.Set("pageSize", "100")
	req.URL.RawQuery = q.Encode()
	req.Header.Add("accept", "application/json")
	req.Header.Add("authorization", "Bearer "+c.token)

	res, err := c.httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()

	var wrapper struct {
		Results struct {
			PageInfo struct {
				EndCursor       string `json:"endCursor"`
				HasNextPage     bool   `json:"hasNextPage"`
				HasPreviousPage bool   `json:"hasPreviousPage"`
				StartCursor     string `json:"startCursor"`
			} `json:"pageInfo"`
			Data []Vulnerability `json:"data"`
		} `json:"results"`
	}
	if err := json.NewDecoder(res.Body).Decode(&wrapper); err != nil {
		return nil, err
	}
	return wrapper.Results.Data, nil
}

func main() {
	client, err := NewClient(os.Getenv("VANTA_CLIENT_ID"), os.Getenv("VANTA_CLIENT_SECRET"))
	if err != nil {
		log.Fatal(err)
	}

	vulns, err := client.Vulnerabilities()
	if err != nil {
		log.Fatal(err)
	}

	fmt.Printf("%+v\n", vulns)
}
