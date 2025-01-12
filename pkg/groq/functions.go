package groq

import (
	"context"
	"encoding/json"
	"fmt"
)

type Function struct {
	Name        string     `json:"name"`
	Description string     `json:"description"`
	Parameters  Parameters `json:"parameters"`
}

type Parameters struct {
	Type       string              `json:"type"`
	Properties map[string]Property `json:"properties"`
	Required   []string            `json:"required,omitempty"`
}

type Property struct {
	Type        string   `json:"type"`
	Description string   `json:"description"`
	Enum        []string `json:"enum,omitempty"`
}

type FunctionCall struct {
	Name      string          `json:"name"`
	Arguments json.RawMessage `json:"arguments"`
}

// WeatherFunction defines a function to get the current weather for a specified location.
//
// Name: get_weather
// Description: Get the current weather for a location
//
// Parameters:
// - location (string): City name or coordinates (required)
// - unit (string): Temperature unit, can be either "celsius" or "fahrenheit"
var WeatherFunction = Function{
	Name:        "get_weather",
	Description: "Get the current weather for a location",
	Parameters: Parameters{
		Type: "object",
		Properties: map[string]Property{
			"location": {
				Type:        "string",
				Description: "City name or coordinates",
			},
			"unit": {
				Type:        "string",
				Description: "Temperature unit",
				Enum:        []string{"celsius", "fahrenheit"},
			},
		},
		Required: []string{"location"},
	},
}

// CalendarFunction represents a function to schedule a meeting in the calendar.
//
// Name: schedule_meeting
// Description: Schedule a meeting in the calendar
//
// Parameters:
// - title (string): Meeting title (required)
// - duration (string): Meeting duration (e.g., '30m', '1h') (required)
// - attendees (string): Comma-separated list of attendee emails
var CalendarFunction = Function{
	Name:        "schedule_meeting",
	Description: "Schedule a meeting in the calendar",
	Parameters: Parameters{
		Type: "object",
		Properties: map[string]Property{
			"title": {
				Type:        "string",
				Description: "Meeting title",
			},
			"duration": {
				Type:        "string",
				Description: "Meeting duration (e.g., '30m', '1h')",
			},
			"attendees": {
				Type:        "string",
				Description: "Comma-separated list of attendee emails",
			},
		},
		Required: []string{"title", "duration"},
	},
}

type FunctionCallChatRequest struct {
	*ChatCompletionRequest
	Functions []Function `json:"functions,omitempty"`
}

// ParseArguments unmarshals the JSON-encoded arguments of the FunctionCall
// into the provided interface{}.
//
// Parameters:
//
//	v - A pointer to the variable where the unmarshaled data will be stored.
//
// Returns:
//
//	An error if the unmarshaling fails, otherwise nil.
func (f *FunctionCall) ParseArguments(v interface{}) error {
	return json.Unmarshal(f.Arguments, v)
}

type WeatherArgs struct {
	Location string `json:"location"`
	Unit     string `json:"unit,omitempty"`
}

type CalendarArgs struct {
	Title     string `json:"title"`
	Duration  string `json:"duration"`
	Attendees string `json:"attendees,omitempty"`
}

// CreateFunctionCall creates a chat completion based on the provided FunctionCallChatRequest.
// It validates the request and ensures that at least one function is provided before proceeding.
//
// Parameters:
//   - ctx: The context for the request, used for cancellation and timeouts.
//   - req: The FunctionCallChatRequest containing the details for the chat completion.
//
// Returns:
//   - *ChatCompletionResponse: The response from the chat completion.
//   - error: An error if the request is invalid or if the chat completion fails.
func (c *Client) CreateFunctionCall(ctx context.Context, req *FunctionCallChatRequest) (*ChatCompletionResponse, error) {
	if err := req.Validate(); err != nil {
		return nil, fmt.Errorf("invalid request: %w", err)
	}

	if len(req.Functions) == 0 {
		return nil, fmt.Errorf("at least one function must be provided")
	}

	return c.CreateChatCompletion(ctx, req.ChatCompletionRequest)
}
