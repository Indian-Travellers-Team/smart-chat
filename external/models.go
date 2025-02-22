package external

// Struct for the request payload for the "create-user-initial-query" API
type UserInitialQueryRequest struct {
	ThreadID             string `json:"thread_id"`
	Mobile               string `json:"mobile"`
	NoOfPeople           int    `json:"no_of_people"`
	PreferredDestination string `json:"preferred_destination"`
	PreferredDate        string `json:"preferred_date"`
}

// Struct to handle the response from the "create-user-initial-query" API
type ToolResponse struct {
	Message string `json:"message"`
	Status  string `json:"status"`
}

// Struct for the request payload for the "create-user-final-booking" API
type CreateUserFinalBookingRequest struct {
	ThreadID string `json:"thread_id"` // The conversation ID (thread ID)
	TripID   int    `json:"trip"`      // The trip ID
}

// Struct to represent a package
type Package struct {
	ID                 int      `json:"id"`
	Name               string   `json:"name"`
	Duration           string   `json:"duration"`
	PackageLink        string   `json:"package_link"`
	QuadSharingPrice   float64  `json:"quad_sharing_price"`
	TripleSharingPrice float64  `json:"triple_sharing_price"`
	DoubleSharingPrice float64  `json:"double_sharing_price"`
	UpcomingTripDates  []string `json:"upcoming_trip_dates"`
}

// Struct to match the JSON structure of the package details API response
type PackageDetails struct {
	ID          int      `json:"id"`
	Name        string   `json:"name"`
	Location    string   `json:"location"`
	Days        int      `json:"days"`
	Nights      int      `json:"nights"`
	Itinerary   string   `json:"itinerary"`
	Inclusion   string   `json:"inclusion"`
	Exclusion   string   `json:"exclusion"`
	Costings    Costings `json:"costings"`
	PackageLink string   `json:"package_link"`
}

// Struct for costings inside a package details
type Costings struct {
	QuadSharingCost   float64 `json:"quad_sharing_cost"`
	TripleSharingCost float64 `json:"triple_sharing_cost"`
	DoubleSharingCost float64 `json:"double_sharing_cost"`
}

// UpcomingTrip represents the structure of a single upcoming trip for a package
type UpcomingTrip struct {
	ID             int     `json:"id"`
	PackageID      int     `json:"package"`
	StartDate      string  `json:"start_date"`
	EndDate        string  `json:"end_date"`
	TotalDays      int     `json:"total_days"`
	AdvancePayment float64 `json:"advance_payment"`
	Discount       float64 `json:"discount"`
}

// UpcomingTripsResponse represents the response structure from the upcoming trips API
type UpcomingTripsResponse []UpcomingTrip
