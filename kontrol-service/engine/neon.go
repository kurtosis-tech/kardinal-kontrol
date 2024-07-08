package engine

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
)

// createNeonBranch - creates a branch on neon for a given project & parent branch id
// returns a hostname that needs to be replaced in the original postgresql string
func createNeonBranch(neonApiKey, projectID, parentBranchID string) (string, error) {
	url := fmt.Sprintf("https://console.neon.tech/api/v2/projects/%s/branches", projectID)

	jsonPayload := []byte(`{"endpoints": [{"type": "read_write"}], "branch": {"parent_id": "` + parentBranchID + `"}}`)

	req, err := http.NewRequest("POST", url, bytes.NewBuffer(jsonPayload))
	if err != nil {
		return "", fmt.Errorf("error creating request: %v", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+neonApiKey)

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return "", fmt.Errorf("error sending request: %v", err)
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("error reading response body: %v", err)
	}

	if resp.StatusCode != http.StatusCreated {
		return "", fmt.Errorf("unexpected status code: %d, body: %s", resp.StatusCode, string(body))
	}

	var result map[string]interface{}
	err = json.Unmarshal(body, &result)
	if err != nil {
		return "", fmt.Errorf("error unmarshaling response: %v", err)
	}

	endpoints, ok := result["endpoints"].([]interface{})
	if !ok || len(endpoints) == 0 {
		return "", fmt.Errorf("endpoints not found in response")
	}

	firstEndpoint, ok := endpoints[0].(map[string]interface{})
	if !ok {
		return "", fmt.Errorf("invalid endpoint format in response")
	}

	host, ok := firstEndpoint["host"].(string)
	if !ok {
		return "", fmt.Errorf("host not found in response")
	}

	return host, nil
}

func updateConnectionString(originalString, newHost string) (string, error) {
	// Parse the original connection string
	u, err := url.Parse(originalString)
	if err != nil {
		return "", fmt.Errorf("error parsing connection string: %v", err)
	}

	// Update the host
	u.Host = newHost

	// Reconstruct the connection string
	return u.String(), nil
}
