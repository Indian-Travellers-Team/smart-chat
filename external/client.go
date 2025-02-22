package external

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"time"
)

// var baseURL = "https://indiantravellersteam.in/api"

var baseURL = "http://localhost:8000/api"

// httpClient initializes a new HTTP client with a timeout
var httpClient = &http.Client{
	Timeout: time.Second * 60, // Set a reasonable timeout for the API call
}

// GetPackageList fetches the list of packages from the external API
func GetPackageList() ([]Package, error) {
	resp, err := httpClient.Get(fmt.Sprintf("%s/packages", baseURL))
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("error fetching packages: status %d", resp.StatusCode)
	}

	var packages []Package
	if err := json.NewDecoder(resp.Body).Decode(&packages); err != nil {
		return nil, err
	}

	return packages, nil
}

// GetPackageDetails fetches the details of a specific package by ID
func GetPackageDetails(packageID int) (*PackageDetails, error) {
	resp, err := httpClient.Get(fmt.Sprintf("%s/packages/%d", baseURL, packageID))
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("error fetching package details: status %d", resp.StatusCode)
	}

	var packageDetails PackageDetails
	if err := json.NewDecoder(resp.Body).Decode(&packageDetails); err != nil {
		return nil, err
	}

	return &packageDetails, nil
}

// CreateUserInitialQuery makes a POST request to the external API to create a user initial query
func CreateUserInitialQuery(threadID string, mobile string, noOfPeople int, preferredDestination string, preferredDate string) (*UserInitialQueryResponse, error) {
	// Construct the request payload
	requestPayload := UserInitialQueryRequest{
		ThreadID:             threadID,
		Mobile:               mobile,
		NoOfPeople:           noOfPeople,
		PreferredDestination: preferredDestination,
		PreferredDate:        preferredDate,
	}

	// Marshal the payload into JSON
	requestBody, err := json.Marshal(requestPayload)
	if err != nil {
		return nil, fmt.Errorf("error marshalling request payload: %v", err)
	}

	// Send POST request to the external API
	apiURL := fmt.Sprintf("%s/agent/function/create-user-initial-query/", baseURL)
	resp, err := httpClient.Post(apiURL, "application/json", bytes.NewBuffer(requestBody))
	if err != nil {
		return nil, fmt.Errorf("error sending request to API: %v", err)
	}
	defer resp.Body.Close()

	// Check if the response status is OK
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("error in API response: status %v", resp.Status)
	}

	// Decode the response body
	var response UserInitialQueryResponse
	if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
		return nil, fmt.Errorf("error decoding API response: %v", err)
	}

	return &response, nil
}
