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
type UserInitialQueryResponse struct {
	Message string `json:"message"`
	Status  string `json:"status"`
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
