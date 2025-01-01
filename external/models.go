package external

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

// PackageDetails struct to match the JSON structure of the package details API response
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

type Costings struct {
	QuadSharingCost   float64 `json:"quad_sharing_cost"`
	TripleSharingCost float64 `json:"triple_sharing_cost"`
	DoubleSharingCost float64 `json:"double_sharing_cost"`
}
