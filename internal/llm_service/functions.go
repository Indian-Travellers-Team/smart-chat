package llm_service

import (
	openai "github.com/sashabaranov/go-openai"
	"github.com/sashabaranov/go-openai/jsonschema"
)

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
