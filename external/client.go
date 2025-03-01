package external

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"time"
)

// var baseURL = "https://indiantravellersteam.in/api"

var baseURL = "http://127.0.0.1:8000/api"

// httpClient initializes a new HTTP client with a timeout
var httpClient = &http.Client{
	Timeout: time.Second * 60, // Set a reasonable timeout for the API call
}

// GetPackageList fetches the list of packages from the external API
func GetPackageList() ([]Package, error) {
	url := fmt.Sprintf("%s/packages", baseURL)
	resp, err := httpClient.Get(url)
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
func CreateUserInitialQuery(threadID string, mobile string, noOfPeople int, preferredDestination string, preferredDate string) (*ToolResponse, error) {
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
	if resp.StatusCode != http.StatusCreated {
		return nil, fmt.Errorf("error in API response: status %v", resp.Status)
	}

	// Decode the response body
	var response ToolResponse
	if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
		return nil, fmt.Errorf("error decoding API response: %v", err)
	}

	return &response, nil
}

// CreateUserFinalBooking makes a POST request to the external API to create the user final booking
func CreateUserFinalBooking(threadID string, tripID int) (*ToolResponse, error) {
	// Construct the request payload
	requestPayload := CreateUserFinalBookingRequest{
		ThreadID: threadID,
		TripID:   tripID,
	}

	// Marshal the payload into JSON
	requestBody, err := json.Marshal(requestPayload)
	if err != nil {
		return nil, fmt.Errorf("error marshalling request payload: %v", err)
	}

	// Send POST request to the external API
	apiURL := fmt.Sprintf("%s/agent/function/create-user-final-booking/", baseURL)
	resp, err := httpClient.Post(apiURL, "application/json", bytes.NewBuffer(requestBody))
	if err != nil {
		return nil, fmt.Errorf("error sending request to API: %v", err)
	}
	defer resp.Body.Close()

	// Check if the response status is OK
	if resp.StatusCode != http.StatusCreated {
		return nil, fmt.Errorf("error in API response: status %v", resp.Status)
	}

	// Decode the response body
	var response ToolResponse
	if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
		return nil, fmt.Errorf("error decoding API response: %v", err)
	}

	return &response, nil
}

// GetUpcomingTrips fetches the upcoming trips for a specific package by its ID
func GetUpcomingTrips(packageID int) (*UpcomingTripsResponseInternal, error) {
	// Send GET request to fetch the upcoming trips for the given package ID
	apiURL := fmt.Sprintf("%s/v1/web/upcoming-trips/%d/", baseURL, packageID)
	resp, err := httpClient.Get(apiURL)
	if err != nil {
		return nil, fmt.Errorf("error sending GET request to API: %v", err)
	}
	defer resp.Body.Close()

	// Check if the response status is OK
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("error in API response: status %v", resp.Status)
	}

	// Decode the response body into a slice of UpcomingTripsResponse (external version)
	var upcomingTrips UpcomingTripsResponse
	if err := json.NewDecoder(resp.Body).Decode(&upcomingTrips); err != nil {
		return nil, fmt.Errorf("error decoding API response: %v", err)
	}

	// Map 'id' to 'trip_id' and convert the response
	var internalTrips UpcomingTripsResponseInternal
	for _, trip := range upcomingTrips {
		internalTrips = append(internalTrips, UpcomingTripInternal{
			TripID:         trip.ID,
			Package:        trip.Package,
			StartDate:      trip.StartDate,
			EndDate:        trip.EndDate,
			TotalDays:      trip.TotalDays,
			AdvancePayment: trip.AdvancePayment,
			Discount:       trip.Discount,
		})
	}

	// Return the internal response
	return &internalTrips, nil
}

// GetWorkflow fetches the workflow details for a given workflow ID
func GetWorkflow(workflowID int) (*WorkflowResponse, error) {
	// Send GET request to fetch the workflow details for the given workflow ID
	apiURL := fmt.Sprintf("%s/agent/workflow/%d/", baseURL, workflowID)
	resp, err := httpClient.Get(apiURL)
	if err != nil {
		return nil, fmt.Errorf("error sending GET request to API: %v", err)
	}
	defer resp.Body.Close()

	// Check if the response status is OK
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("error in API response: status %v", resp.Status)
	}

	// Decode the response body into a WorkflowResponse
	var workflowResponse WorkflowResponse
	if err := json.NewDecoder(resp.Body).Decode(&workflowResponse); err != nil {
		return nil, fmt.Errorf("error decoding API response: %v", err)
	}

	// Type assert flow to the correct type
	flow, ok := workflowResponse.Flow.(map[string]interface{})
	if !ok {
		return nil, fmt.Errorf("error asserting Flow to correct type")
	}

	// Re-assign the fields from flow to the WorkflowResponse struct
	workflowResponse.Flow = flow

	return &workflowResponse, nil
}
