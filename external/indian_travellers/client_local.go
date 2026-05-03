package indian_travellers

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"
)

const (
	maxResponseBodyBytes = 2 << 20 // 2 MiB

	// Short cache TTLs reduce repeated same-host calls but keep data fresh.
	cacheTTLPackageList    = 30 * time.Second
	cacheTTLPackageDetails = 30 * time.Second
	cacheTTLWorkflow       = 30 * time.Second
)

type cacheEntry struct {
	value     interface{}
	expiresAt time.Time
}

func (c *Client) newRequest(method, path string, body []byte) (*http.Request, error) {
	fullURL := fmt.Sprintf("%s%s", strings.TrimRight(c.baseURL, "/"), path)

	req, err := http.NewRequestWithContext(context.Background(), method, fullURL, bytes.NewReader(body))
	if err != nil {
		return nil, err
	}

	req.Header.Set("Accept", "application/json")
	if len(body) > 0 {
		req.Header.Set("Content-Type", "application/json")
	}

	return req, nil
}

func (c *Client) doJSON(req *http.Request, expectedStatus int, out interface{}) error {
	ctx, cancel := context.WithTimeout(req.Context(), c.timeout)
	defer cancel()

	resp, err := c.httpClient.Do(req.WithContext(ctx))
	if err != nil {
		return fmt.Errorf("request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != expectedStatus {
		return fmt.Errorf("unexpected status: %s", resp.Status)
	}

	limitedBody := io.LimitReader(resp.Body, maxResponseBodyBytes)
	if err := json.NewDecoder(limitedBody).Decode(out); err != nil {
		return fmt.Errorf("failed to decode response: %w", err)
	}

	return nil
}

func (c *Client) getCached(key string) (interface{}, bool) {
	if c.cache == nil {
		return nil, false
	}

	c.cacheMu.RLock()
	entry, ok := c.cache[key]
	c.cacheMu.RUnlock()
	if !ok || time.Now().After(entry.expiresAt) {
		if ok {
			c.cacheMu.Lock()
			delete(c.cache, key)
			c.cacheMu.Unlock()
		}
		return nil, false
	}
	return entry.value, true
}

func (c *Client) setCached(key string, value interface{}, ttl time.Duration) {
	if c.cache == nil {
		return
	}

	c.cacheMu.Lock()
	c.cache[key] = cacheEntry{value: value, expiresAt: time.Now().Add(ttl)}
	c.cacheMu.Unlock()
}

func (c *Client) getPackageListLocal() ([]Package, error) {
	if cached, ok := c.getCached("packages:list"); ok {
		if packages, valid := cached.([]Package); valid {
			return packages, nil
		}
	}

	req, err := c.newRequest(http.MethodGet, "/api/packages/", nil)
	if err != nil {
		return nil, fmt.Errorf("failed to build packages request: %w", err)
	}

	var packages []Package
	if err := c.doJSON(req, http.StatusOK, &packages); err != nil {
		return nil, fmt.Errorf("error fetching packages: %w", err)
	}

	c.setCached("packages:list", packages, cacheTTLPackageList)

	return packages, nil
}

func (c *Client) getPackageDetailsLocal(packageID int) (*PackageDetails, error) {
	if packageID <= 0 {
		return nil, fmt.Errorf("invalid packageID: %d", packageID)
	}

	cacheKey := fmt.Sprintf("packages:details:%d", packageID)
	if cached, ok := c.getCached(cacheKey); ok {
		if details, valid := cached.(*PackageDetails); valid {
			return details, nil
		}
	}

	req, err := c.newRequest(http.MethodGet, fmt.Sprintf("/api/packages/%d", packageID), nil)
	if err != nil {
		return nil, fmt.Errorf("failed to build package details request: %w", err)
	}

	var packageDetails PackageDetails
	if err := c.doJSON(req, http.StatusOK, &packageDetails); err != nil {
		return nil, fmt.Errorf("error fetching package details: %w", err)
	}

	c.setCached(cacheKey, &packageDetails, cacheTTLPackageDetails)

	return &packageDetails, nil
}

func (c *Client) createUserInitialQueryLocal(threadID string, mobile string, noOfPeople int, preferredDestination string, preferredDate string) (*ToolResponse, error) {
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

	req, err := c.newRequest(http.MethodPost, "/api/agent/function/create-user-initial-query/", requestBody)
	if err != nil {
		return nil, fmt.Errorf("failed to build request: %w", err)
	}

	var response ToolResponse
	if err := c.doJSON(req, http.StatusCreated, &response); err != nil {
		return nil, fmt.Errorf("error creating initial query: %w", err)
	}

	return &response, nil
}

func (c *Client) createUserFinalBookingLocal(threadID string, tripID int) (*ToolResponse, error) {
	if tripID <= 0 {
		return nil, fmt.Errorf("invalid tripID: %d", tripID)
	}

	requestPayload := CreateUserFinalBookingRequest{
		ThreadID: threadID,
		TripID:   tripID,
	}

	requestBody, err := json.Marshal(requestPayload)
	if err != nil {
		return nil, fmt.Errorf("error marshalling request payload: %v", err)
	}

	req, err := c.newRequest(http.MethodPost, "/api/agent/function/create-user-final-booking/", requestBody)
	if err != nil {
		return nil, fmt.Errorf("failed to build request: %w", err)
	}

	var response ToolResponse
	if err := c.doJSON(req, http.StatusCreated, &response); err != nil {
		return nil, fmt.Errorf("error creating final booking: %w", err)
	}

	return &response, nil
}

func (c *Client) getUpcomingTripsLocal(packageID int) (*UpcomingTripsResponseInternal, error) {
	if packageID <= 0 {
		return nil, fmt.Errorf("invalid packageID: %d", packageID)
	}

	req, err := c.newRequest(http.MethodGet, fmt.Sprintf("/api/v1/web/upcoming-trips/%d/", packageID), nil)
	if err != nil {
		return nil, fmt.Errorf("failed to build upcoming trips request: %w", err)
	}

	var upcomingTrips UpcomingTripsResponse
	if err := c.doJSON(req, http.StatusOK, &upcomingTrips); err != nil {
		return nil, fmt.Errorf("error fetching upcoming trips: %w", err)
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

func (c *Client) getWorkflowLocal(workflowID int) (*WorkflowResponse, error) {
	if workflowID <= 0 {
		return nil, fmt.Errorf("invalid workflowID: %d", workflowID)
	}

	cacheKey := fmt.Sprintf("workflow:%d", workflowID)
	if cached, ok := c.getCached(cacheKey); ok {
		if workflow, valid := cached.(*WorkflowResponse); valid {
			return workflow, nil
		}
	}

	req, err := c.newRequest(http.MethodGet, fmt.Sprintf("/api/agent/workflow/%d/", workflowID), nil)
	if err != nil {
		return nil, fmt.Errorf("failed to build workflow request: %w", err)
	}

	var workflowResponse WorkflowResponse
	if err := c.doJSON(req, http.StatusOK, &workflowResponse); err != nil {
		return nil, fmt.Errorf("error fetching workflow: %w", err)
	}

	flow, ok := workflowResponse.Flow.(map[string]interface{})
	if !ok {
		return nil, fmt.Errorf("error asserting Flow to correct type")
	}
	workflowResponse.Flow = flow

	c.setCached(cacheKey, &workflowResponse, cacheTTLWorkflow)

	return &workflowResponse, nil
}
