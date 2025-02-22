package llm_service

import (
	"github.com/sashabaranov/go-openai"
	"github.com/sashabaranov/go-openai/jsonschema"
)

// Existing package details schema (for reference)
var GetPackageDetailsSchema = &openai.FunctionDefinition{
	Name:        "get_package_details",
	Description: "Get the details of a travel package by its ID",
	Parameters: jsonschema.Definition{
		Type: jsonschema.Object,
		Properties: map[string]jsonschema.Definition{
			"package_id": {
				Type:        jsonschema.Integer,
				Description: "The unique identifier for the travel package",
			},
		},
		Required: []string{"package_id"},
	},
}

// Define the new schema for the initial user query
var CreateUserInitialQuerySchema = &openai.FunctionDefinition{
	Name:        "create_user_initial_query",
	Description: "Create the initial query for the user, asking for travel details",
	Parameters: jsonschema.Definition{
		Type: jsonschema.Object,
		Properties: map[string]jsonschema.Definition{
			"no_of_people": {
				Type:        jsonschema.Integer,
				Description: "The number of people for the trip",
			},
			"preferred_destination": {
				Type:        jsonschema.String,
				Description: "The preferred destination for the trip",
			},
			"preferred_date": {
				Type:        jsonschema.String,
				Description: "The preferred date for the trip",
			},
		},
		Required: []string{"no_of_people", "preferred_destination", "preferred_date"},
	},
}

// Define the schema for the final booking query
var CreateUserFinalBookingSchema = &openai.FunctionDefinition{
	Name:        "create_user_final_booking",
	Description: "Create the final booking for the user, asking for the trip ID",
	Parameters: jsonschema.Definition{
		Type: jsonschema.Object,
		Properties: map[string]jsonschema.Definition{
			"trip_id": {
				Type:        jsonschema.Integer,
				Description: "The unique identifier for the trip",
			},
		},
		Required: []string{"trip_id"},
	},
}

// Define the schema for the upcoming trips query
var FetchUpcomingTripsSchema = &openai.FunctionDefinition{
	Name:        "fetch_upcoming_trips",
	Description: "Fetch the upcoming trips for a specific package by its ID",
	Parameters: jsonschema.Definition{
		Type: jsonschema.Object,
		Properties: map[string]jsonschema.Definition{
			"package_id": {
				Type:        jsonschema.Integer,
				Description: "The unique identifier for the package to fetch upcoming trips",
			},
		},
		Required: []string{"package_id"},
	},
}
