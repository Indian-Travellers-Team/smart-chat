package external

import (
	"encoding/json"
	"fmt"
	"net/http"
	"time"
)

var baseURL = "https://indiantravellersteam.in/api"

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
