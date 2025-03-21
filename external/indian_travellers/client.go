package indian_travellers

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"smart-chat/config"
)

// Client wraps the HTTP client and the base URL for the Indian Travellers API.
type Client struct {
	baseURL    string
	httpClient *http.Client
}

// NewClient initializes a new Indian Travellers API client using the provided configuration.
func NewClient(cfg *config.Config) *Client {
	return &Client{
		baseURL: cfg.IndianTeavellersURL,
		httpClient: &http.Client{
			Timeout: time.Second * 60,
		},
	}
}

// GetPackageList fetches the list of packages from the external API.
func (c *Client) GetPackageList() ([]Package, error) {
	url := fmt.Sprintf("%s/api/packages", c.baseURL)
	resp, err := c.httpClient.Get(url)
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

// GetPackageDetails fetches the details of a specific package by ID.
func (c *Client) GetPackageDetails(packageID int) (*PackageDetails, error) {
	url := fmt.Sprintf("%s/api/packages/%d", c.baseURL, packageID)
	resp, err := c.httpClient.Get(url)
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

// CreateUserInitialQuery makes a POST request to create a user initial query.
func (c *Client) CreateUserInitialQuery(threadID string, mobile string, noOfPeople int, preferredDestination string, preferredDate string) (*ToolResponse, error) {
	requestPayload := UserInitialQueryRequest{
		ThreadID:             threadID,
		Mobile:               mobile,
		NoOfPeople:           noOfPeople,
		PreferredDestination: preferredDestination,
		PreferredDate:        preferredDate,
	}

	requestBody, err := json.Marshal(requestPayload)
	if err != nil {
		return nil, fmt.Errorf("error marshalling request payload: %v", err)
	}

	apiURL := fmt.Sprintf("%s/agent/function/create-user-initial-query/", c.baseURL)
	resp, err := c.httpClient.Post(apiURL, "application/json", bytes.NewBuffer(requestBody))
	if err != nil {
		return nil, fmt.Errorf("error sending request to API: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusCreated {
		return nil, fmt.Errorf("error in API response: status %v", resp.Status)
	}

	var response ToolResponse
	if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
		return nil, fmt.Errorf("error decoding API response: %v", err)
	}

	return &response, nil
}

// CreateUserFinalBooking makes a POST request to create the user final booking.
func (c *Client) CreateUserFinalBooking(threadID string, tripID int) (*ToolResponse, error) {
	requestPayload := CreateUserFinalBookingRequest{
		ThreadID: threadID,
		TripID:   tripID,
	}

	requestBody, err := json.Marshal(requestPayload)
	if err != nil {
		return nil, fmt.Errorf("error marshalling request payload: %v", err)
	}

	apiURL := fmt.Sprintf("%s/agent/function/create-user-final-booking/", c.baseURL)
	resp, err := c.httpClient.Post(apiURL, "application/json", bytes.NewBuffer(requestBody))
	if err != nil {
		return nil, fmt.Errorf("error sending request to API: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusCreated {
		return nil, fmt.Errorf("error in API response: status %v", resp.Status)
	}

	var response ToolResponse
	if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
		return nil, fmt.Errorf("error decoding API response: %v", err)
	}

	return &response, nil
}

// GetUpcomingTrips fetches the upcoming trips for a specific package by its ID.
func (c *Client) GetUpcomingTrips(packageID int) (*UpcomingTripsResponseInternal, error) {
	apiURL := fmt.Sprintf("%s/v1/web/upcoming-trips/%d/", c.baseURL, packageID)
	resp, err := c.httpClient.Get(apiURL)
	if err != nil {
		return nil, fmt.Errorf("error sending GET request to API: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("error in API response: status %v", resp.Status)
	}

	var upcomingTrips UpcomingTripsResponse
	if err := json.NewDecoder(resp.Body).Decode(&upcomingTrips); err != nil {
		return nil, fmt.Errorf("error decoding API response: %v", err)
	}

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

	return &internalTrips, nil
}

// GetWorkflow fetches the workflow details for a given workflow ID.
func (c *Client) GetWorkflow(workflowID int) (*WorkflowResponse, error) {
	apiURL := fmt.Sprintf("%s/agent/workflow/%d/", c.baseURL, workflowID)
	resp, err := c.httpClient.Get(apiURL)
	if err != nil {
		return nil, fmt.Errorf("error sending GET request to API: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("error in API response: status %v", resp.Status)
	}

	var workflowResponse WorkflowResponse
	if err := json.NewDecoder(resp.Body).Decode(&workflowResponse); err != nil {
		return nil, fmt.Errorf("error decoding API response: %v", err)
	}

	flow, ok := workflowResponse.Flow.(map[string]interface{})
	if !ok {
		return nil, fmt.Errorf("error asserting Flow to correct type")
	}
	workflowResponse.Flow = flow

	return &workflowResponse, nil
}
