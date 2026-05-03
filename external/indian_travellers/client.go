package indian_travellers

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"net"
	"net/http"
	"net/url"
	"strings"
	"sync"
	"time"

	"smart-chat/config"
)

const (
	defaultRequestTimeout = 60 * time.Second
	localRequestTimeout   = 10 * time.Second
)

// Client wraps the HTTP client and the base URL for the Indian Travellers API.
type Client struct {
	baseURL    string
	httpClient *http.Client
	timeout    time.Duration
	useLocal   bool

	cacheMu sync.RWMutex
	cache   map[string]cacheEntry
}

// NewClient initializes a new Indian Travellers API client using the provided configuration.
func NewClient(cfg *config.Config) *Client {
	baseURL := strings.TrimSpace(cfg.IndianTeavellersURL)
	if baseURL == "" {
		baseURL = "http://127.0.0.1:8000"
	}
	useLocal := cfg.EnableLocalIndianTravellers

	isLocal := isLoopbackBaseURL(baseURL)
	timeout := defaultRequestTimeout
	if useLocal && isLocal {
		timeout = localRequestTimeout
	}

	transport := http.DefaultTransport
	if useLocal {
		transport = &http.Transport{
			Proxy: http.ProxyFromEnvironment,
			DialContext: (&net.Dialer{
				Timeout:   500 * time.Millisecond,
				KeepAlive: 30 * time.Second,
			}).DialContext,
			ForceAttemptHTTP2:     true,
			MaxIdleConns:          100,
			MaxIdleConnsPerHost:   50,
			IdleConnTimeout:       90 * time.Second,
			TLSHandshakeTimeout:   2 * time.Second,
			ExpectContinueTimeout: 1 * time.Second,
			DisableCompression:    isLocal,
		}
	}

	cache := make(map[string]cacheEntry)
	if !useLocal {
		cache = nil
	}

	return &Client{
		baseURL: baseURL,
		httpClient: &http.Client{
			Timeout:   timeout,
			Transport: transport,
		},
		timeout:  timeout,
		useLocal: useLocal,
		cache:    cache,
	}
}

func isLoopbackBaseURL(rawURL string) bool {
	parsed, err := url.Parse(rawURL)
	if err != nil {
		return false
	}
	host := parsed.Hostname()
	if host == "" {
		return false
	}
	if strings.EqualFold(host, "localhost") {
		return true
	}
	ip := net.ParseIP(host)
	return ip != nil && ip.IsLoopback()
}

// GetPackageList fetches the list of packages from the external API.
func (c *Client) GetPackageList() ([]Package, error) {
	if c.useLocal {
		return c.getPackageListLocal()
	}
	return c.getPackageListLegacy()
}

// GetPackageDetails fetches the details of a specific package by ID.
func (c *Client) GetPackageDetails(packageID int) (*PackageDetails, error) {
	if c.useLocal {
		return c.getPackageDetailsLocal(packageID)
	}
	return c.getPackageDetailsLegacy(packageID)
}

// CreateUserInitialQuery makes a POST request to create a user initial query.
func (c *Client) CreateUserInitialQuery(threadID string, mobile string, noOfPeople int, preferredDestination string, preferredDate string) (*ToolResponse, error) {
	if c.useLocal {
		return c.createUserInitialQueryLocal(threadID, mobile, noOfPeople, preferredDestination, preferredDate)
	}
	return c.createUserInitialQueryLegacy(threadID, mobile, noOfPeople, preferredDestination, preferredDate)
}

// CreateUserFinalBooking makes a POST request to create the user final booking.
func (c *Client) CreateUserFinalBooking(threadID string, tripID int) (*ToolResponse, error) {
	if c.useLocal {
		return c.createUserFinalBookingLocal(threadID, tripID)
	}
	return c.createUserFinalBookingLegacy(threadID, tripID)
}

// GetUpcomingTrips fetches the upcoming trips for a specific package by its ID.
func (c *Client) GetUpcomingTrips(packageID int) (*UpcomingTripsResponseInternal, error) {
	if c.useLocal {
		return c.getUpcomingTripsLocal(packageID)
	}
	return c.getUpcomingTripsLegacy(packageID)
}

// GetWorkflow fetches the workflow details for a given workflow ID.
func (c *Client) GetWorkflow(workflowID int) (*WorkflowResponse, error) {
	if c.useLocal {
		return c.getWorkflowLocal(workflowID)
	}
	return c.getWorkflowLegacy(workflowID)
}

func (c *Client) getPackageListLegacy() ([]Package, error) {
	url := fmt.Sprintf("%s/api/packages/", c.baseURL)
	log.Println(url)
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

func (c *Client) getPackageDetailsLegacy(packageID int) (*PackageDetails, error) {
	url := fmt.Sprintf("%s/api/packages/%d", c.baseURL, packageID)
	log.Println(url)
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

func (c *Client) createUserInitialQueryLegacy(threadID string, mobile string, noOfPeople int, preferredDestination string, preferredDate string) (*ToolResponse, error) {
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

	apiURL := fmt.Sprintf("%s/api/agent/function/create-user-initial-query/", c.baseURL)
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

func (c *Client) createUserFinalBookingLegacy(threadID string, tripID int) (*ToolResponse, error) {
	requestPayload := CreateUserFinalBookingRequest{
		ThreadID: threadID,
		TripID:   tripID,
	}

	requestBody, err := json.Marshal(requestPayload)
	if err != nil {
		return nil, fmt.Errorf("error marshalling request payload: %v", err)
	}

	apiURL := fmt.Sprintf("%s/api/agent/function/create-user-final-booking/", c.baseURL)
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

func (c *Client) getUpcomingTripsLegacy(packageID int) (*UpcomingTripsResponseInternal, error) {
	apiURL := fmt.Sprintf("%s/api/v1/web/upcoming-trips/%d/", c.baseURL, packageID)
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

func (c *Client) getWorkflowLegacy(workflowID int) (*WorkflowResponse, error) {
	apiURL := fmt.Sprintf("%s/api/agent/workflow/%d/", c.baseURL, workflowID)
	log.Println(apiURL)
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
