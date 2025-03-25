package dns

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/miles-w-3/ddns/pkg/models"
)

type RecordPayload struct {
}

type CloudflareClient struct {
	baseURL    string
	authToken  string
	zoneID     string
	recordID   string
	httpClient *http.Client
}

func NewCloudflareClient(baseURL string) (*CloudflareClient, error) {
	authToken := os.Getenv("CLOUDFLARE_TOKEN")

	if authToken == "" {
		return nil, fmt.Errorf("Failed to initialize client - no auth token")
	}

	client := &CloudflareClient{
		baseURL:   baseURL,
		authToken: authToken,
		httpClient: &http.Client{
			Timeout: 10 * time.Second,
		},
	}

	err := client.validateAuth()
	if err != nil {
		return nil, err
	}

	err = client.fetchZoneID()
	if err != nil {
		return nil, err
	}

	err = client.fetchRecordID()
	if err != nil {
		return nil, err
	}

	return client, nil
}

func (c *CloudflareClient) validateAuth() error {
	result, err := c.Request("GET", "/v4/user/tokens/verify", nil)
	if err != nil {
		return err
	}
	if result.StatusCode != 200 {
		return fmt.Errorf("Token validation failed")
	}

	fmt.Println("Successfully validated auth token")
	return nil
}

func (c *CloudflareClient) fetchZoneID() error {
	zoneName := os.Getenv("ZONE_NAME")

	if zoneName == "" {
		return fmt.Errorf("No ZONE_NAME specified")
	}

	response, err := c.Request("GET", "/v4/zones?name="+zoneName, nil)
	if err != nil {
		return err
	}

	body, err := io.ReadAll(response.Body)
	if err != nil {
		return fmt.Errorf("Error reading response body: %v\n", err)
	}

	var zoneResp models.CloudflareZonesResponse
	if err := json.Unmarshal(body, &zoneResp); err != nil {
		return fmt.Errorf("Error parsing response body: %v\n", err)
	}

	if !zoneResp.Success {
		return fmt.Errorf("Unsuccessful request: %+v", zoneResp)
	}

	if len(zoneResp.Result) == 0 {
		return fmt.Errorf("Zone ID not found for %s", zoneName)
	}

	zoneID := zoneResp.Result[0].ID
	c.zoneID = zoneID

	fmt.Printf("Found Zone ID %s for name %s\n", zoneID, zoneName)
	return nil
}

func (c *CloudflareClient) fetchRecordID() error {
	recordName := os.Getenv("RECORD_NAME")
	zoneName := os.Getenv("ZONE_NAME")

	if zoneName == "" || recordName == "" {
		return fmt.Errorf("RECORD_NAME and ZONE_NAME must be specified")
	}

	fullRecordName := recordName + "." + zoneName

	response, err := c.Request("GET", "/v4/zones/"+c.zoneID+"/dns_records?type=A&name="+fullRecordName, nil)
	if err != nil {
		return err
	}

	body, err := io.ReadAll(response.Body)
	if err != nil {
		return fmt.Errorf("Error reading response body: %v\n", err)
	}

	var recordResp models.CloudflareRecordsResponse
	if err := json.Unmarshal(body, &recordResp); err != nil {
		return fmt.Errorf("Error parsing response body: %v\n", err)
	}

	if !recordResp.Success {
		return fmt.Errorf("Unsuccessful request: %+v", recordResp)
	}

	if len(recordResp.Result) == 0 {
		return fmt.Errorf("Record ID not found for %s", zoneName)
	}

	recordID := recordResp.Result[0].ID
	c.recordID = recordID

	fmt.Printf("Found Record ID %s for host %s\n", recordID, fullRecordName)
	return nil
}

func (c *CloudflareClient) Request(method, path string, body io.Reader) (*http.Response, error) {
	url := c.baseURL + path

	req, err := http.NewRequest(method, url, body)

	if err != nil {
		return nil, fmt.Errorf("Error creating request: %v", err)
	}

	// Add the auth token to every request
	req.Header.Add("Authorization", "Bearer "+c.authToken)

	// Add common headers
	req.Header.Add("Content-Type", "application/json")
	// req.Header.Add("Accept", "application/json")

	result, err := c.httpClient.Do(req)

	if err != nil {
		return nil, fmt.Errorf("Failed to complete request - %s", err.Error())
	}

	return result, nil
}

func (c *CloudflareClient) GetCurrentIP() (string, error) {
	// TODO: zone and record variables, logic to retrieve zone and record ids from names
	response, err := c.Request("GET", "/v4/zones/"+c.zoneID+"/dns_records/"+c.recordID, nil)
	if err != nil {
		return "", err
	}

	defer response.Body.Close()

	body, err := io.ReadAll(response.Body)
	if err != nil {
		return "", fmt.Errorf("Error reading response body: %v", err)
	}

	var result map[string]interface{}

	if err := json.Unmarshal(body, &result); err != nil {
		return "", fmt.Errorf("Failed to parse ip details as JSON, %v", err)
	}

	if resultContent, ok := result["result"].(map[string]interface{}); ok {
		if recordContent, ok := resultContent["content"]; ok {
			return recordContent.(string), nil
		}
	}
	return "", fmt.Errorf("Failed to parse retrieve ip content from DNS record")
}

func (c *CloudflareClient) UpdateIP(newIP string) error {
	payload := map[string]string{
		"type":    "A",
		"content": newIP,
		"comment": "DDNS Managed",
	}
	jsonPayload, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("Error preparing payload: %v", err)
	}

	response, err := c.Request("PATCH", "/v4/zones/"+c.zoneID+"/dns_records/"+c.recordID, bytes.NewBuffer(jsonPayload))
	if err != nil {
		return fmt.Errorf("Error preparing request: %v", err)
	}
	if response.StatusCode != 200 {
		return fmt.Errorf("Response %s", response.Status)
	}
	log.Println("Successfully updated IP")
	return nil
}
