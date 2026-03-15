package main

import (
	"bytes"
	"encoding/csv"
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
	"strconv"
	"strings"
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

// PolicyDocument represents a single document attachment on a policy version.
type PolicyDocument struct {
	Language string `json:"language"`
	SlugID   string `json:"slugId"`
	URL      string `json:"url"`
}

// PolicyVersion represents a version of a policy.
type PolicyVersion struct {
	Status    string           `json:"status"`
	Documents []PolicyDocument `json:"documents"`
}

// Policy represents a single policy returned by the Vanta API.
type Policy struct {
	ID                    string         `json:"id"`
	Name                  string         `json:"name"`
	Description           string         `json:"description"`
	Status                string         `json:"status"`
	ApprovedAtDate        *time.Time     `json:"approvedAtDate"`
	LatestVersion         *PolicyVersion `json:"latestVersion"`
	LatestApprovedVersion *PolicyVersion `json:"latestApprovedVersion"`
}

// Document represents a single document returned by the Vanta API.
type Document struct {
	ID               string     `json:"id"`
	OwnerID          string     `json:"ownerId"`
	Category         string     `json:"category"`
	Description      string     `json:"description"`
	IsSensitive      bool       `json:"isSensitive"`
	Title            string     `json:"title"`
	UploadStatus     string     `json:"uploadStatus"`
	UploadStatusDate *time.Time `json:"uploadStatusDate"`
	URL              string     `json:"url"`
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

func fetchAll[T any](c *Client, path string) ([]T, error) {
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

// Vulnerabilities fetches all vulnerabilities from the Vanta API, following pagination.
func (c *Client) Vulnerabilities() ([]Vulnerability, error) {
	return fetchAll[Vulnerability](c, "/vulnerabilities")
}

// Policies fetches all policies from the Vanta API, following pagination.
func (c *Client) Policies() ([]Policy, error) {
	return fetchAll[Policy](c, "/policies")
}

// Documents fetches all documents from the Vanta API, following pagination.
func (c *Client) Documents() ([]Document, error) {
	return fetchAll[Document](c, "/documents")
}

var headers = []string{
	"id", "name", "severity", "cvssSeverityScore", "isFixable",
	"firstDetectedDate", "remediateByDate", "packageIdentifier",
	"vulnerabilityType", "integrationId", "targetId", "scanSource", "externalURL",
}

func vulnRow(v Vulnerability) []string {
	return []string{
		v.ID,
		v.Name,
		v.Severity,
		strconv.FormatFloat(v.CVSSSeverityScore, 'f', -1, 64),
		strconv.FormatBool(v.IsFixable),
		v.FirstDetectedDate.Format(time.RFC3339),
		v.RemediateByDate.Format(time.RFC3339),
		v.PackageIdentifier,
		v.VulnerabilityType,
		v.IntegrationID,
		v.TargetID,
		v.ScanSource,
		v.ExternalURL,
	}
}

func printJSON(v any) error {
	enc := json.NewEncoder(os.Stdout)
	enc.SetIndent("", "  ")
	return enc.Encode(v)
}

func printCSV(vulns []Vulnerability) error {
	w := csv.NewWriter(os.Stdout)
	if err := w.Write(headers); err != nil {
		return err
	}
	for _, v := range vulns {
		if err := w.Write(vulnRow(v)); err != nil {
			return err
		}
	}
	w.Flush()
	return w.Error()
}

func printTSV(vulns []Vulnerability) error {
	w := csv.NewWriter(os.Stdout)
	w.Comma = '\t'
	if err := w.Write(headers); err != nil {
		return err
	}
	for _, v := range vulns {
		if err := w.Write(vulnRow(v)); err != nil {
			return err
		}
	}
	w.Flush()
	return w.Error()
}

func printMarkdown(vulns []Vulnerability) {
	fmt.Println("| " + strings.Join(headers, " | ") + " |")
	fmt.Println("|" + strings.Repeat(" --- |", len(headers)))
	for _, v := range vulns {
		row := vulnRow(v)
		fmt.Println("| " + strings.Join(row, " | ") + " |")
	}
}

var policyHeaders = []string{
	"id", "name", "status", "approvedAtDate", "latestVersionStatus",
}

func policyRow(p Policy) []string {
	approvedAt := ""
	if p.ApprovedAtDate != nil {
		approvedAt = p.ApprovedAtDate.Format(time.RFC3339)
	}
	latestStatus := ""
	if p.LatestVersion != nil {
		latestStatus = p.LatestVersion.Status
	}
	return []string{p.ID, p.Name, p.Status, approvedAt, latestStatus}
}

func printPoliciesCSV(policies []Policy) error {
	w := csv.NewWriter(os.Stdout)
	if err := w.Write(policyHeaders); err != nil {
		return err
	}
	for _, p := range policies {
		if err := w.Write(policyRow(p)); err != nil {
			return err
		}
	}
	w.Flush()
	return w.Error()
}

func printPoliciesTSV(policies []Policy) error {
	w := csv.NewWriter(os.Stdout)
	w.Comma = '\t'
	if err := w.Write(policyHeaders); err != nil {
		return err
	}
	for _, p := range policies {
		if err := w.Write(policyRow(p)); err != nil {
			return err
		}
	}
	w.Flush()
	return w.Error()
}

func printPoliciesMarkdown(policies []Policy) {
	fmt.Println("| " + strings.Join(policyHeaders, " | ") + " |")
	fmt.Println("|" + strings.Repeat(" --- |", len(policyHeaders)))
	for _, p := range policies {
		row := policyRow(p)
		fmt.Println("| " + strings.Join(row, " | ") + " |")
	}
}

var documentHeaders = []string{
	"id", "ownerId", "category", "title", "uploadStatus", "uploadStatusDate", "isSensitive", "url",
}

func documentRow(d Document) []string {
	uploadStatusDate := ""
	if d.UploadStatusDate != nil {
		uploadStatusDate = d.UploadStatusDate.Format(time.RFC3339)
	}
	return []string{d.ID, d.OwnerID, d.Category, d.Title, d.UploadStatus, uploadStatusDate, strconv.FormatBool(d.IsSensitive), d.URL}
}

func printDocumentsCSV(docs []Document) error {
	w := csv.NewWriter(os.Stdout)
	if err := w.Write(documentHeaders); err != nil {
		return err
	}
	for _, d := range docs {
		if err := w.Write(documentRow(d)); err != nil {
			return err
		}
	}
	w.Flush()
	return w.Error()
}

func printDocumentsTSV(docs []Document) error {
	w := csv.NewWriter(os.Stdout)
	w.Comma = '\t'
	if err := w.Write(documentHeaders); err != nil {
		return err
	}
	for _, d := range docs {
		if err := w.Write(documentRow(d)); err != nil {
			return err
		}
	}
	w.Flush()
	return w.Error()
}

func printDocumentsMarkdown(docs []Document) {
	fmt.Println("| " + strings.Join(documentHeaders, " | ") + " |")
	fmt.Println("|" + strings.Repeat(" --- |", len(documentHeaders)))
	for _, d := range docs {
		row := documentRow(d)
		fmt.Println("| " + strings.Join(row, " | ") + " |")
	}
}

func main() {
	format := flag.String("format", "json", "output format: json, csv, tsv, markdown")
	resource := flag.String("resource", "vulnerabilities", "resource to fetch: vulnerabilities, policies, documents")
	flag.Parse()

	client, err := NewClient(os.Getenv("VANTA_CLIENT_ID"), os.Getenv("VANTA_CLIENT_SECRET"))
	if err != nil {
		log.Fatal(err)
	}

	switch *resource {
	case "vulnerabilities":
		vulns, err := client.Vulnerabilities()
		if err != nil {
			log.Fatal(err)
		}
		switch *format {
		case "json":
			if err := printJSON(vulns); err != nil {
				log.Fatal(err)
			}
		case "csv":
			if err := printCSV(vulns); err != nil {
				log.Fatal(err)
			}
		case "tsv":
			if err := printTSV(vulns); err != nil {
				log.Fatal(err)
			}
		case "markdown":
			printMarkdown(vulns)
		default:
			log.Fatalf("unknown format %q: must be json, csv, tsv, or markdown", *format)
		}
	case "policies":
		policies, err := client.Policies()
		if err != nil {
			log.Fatal(err)
		}
		switch *format {
		case "json":
			if err := printJSON(policies); err != nil {
				log.Fatal(err)
			}
		case "csv":
			if err := printPoliciesCSV(policies); err != nil {
				log.Fatal(err)
			}
		case "tsv":
			if err := printPoliciesTSV(policies); err != nil {
				log.Fatal(err)
			}
		case "markdown":
			printPoliciesMarkdown(policies)
		default:
			log.Fatalf("unknown format %q: must be json, csv, tsv, or markdown", *format)
		}
	case "documents":
		docs, err := client.Documents()
		if err != nil {
			log.Fatal(err)
		}
		switch *format {
		case "json":
			if err := printJSON(docs); err != nil {
				log.Fatal(err)
			}
		case "csv":
			if err := printDocumentsCSV(docs); err != nil {
				log.Fatal(err)
			}
		case "tsv":
			if err := printDocumentsTSV(docs); err != nil {
				log.Fatal(err)
			}
		case "markdown":
			printDocumentsMarkdown(docs)
		default:
			log.Fatalf("unknown format %q: must be json, csv, tsv, or markdown", *format)
		}
	default:
		log.Fatalf("unknown resource %q: must be vulnerabilities, policies, or documents", *resource)
	}
}
