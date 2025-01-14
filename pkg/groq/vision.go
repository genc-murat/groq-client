package groq

import (
	"encoding/base64"
	"fmt"
	"io"
	"net/http"
)

const (
	MaxURLImageSize    = 20 * 1024 * 1024 // 20MB
	MaxBase64ImageSize = 4 * 1024 * 1024  // 4MB
)

// Content types for multimodal messages
type ContentType struct {
	Type     string    `json:"type"`
	Text     string    `json:"text,omitempty"`
	ImageURL *ImageURL `json:"image_url,omitempty"`
}

// ImageURL represents an image URL in the request
type ImageURL struct {
	URL string `json:"url"`
}

// NewTextContent creates a new ContentType with type "text" and the given text value.
// It is used for creating text content for the vision model.
//
// Parameters:
//   - text: The text content to be included in the ContentType
//
// Returns:
//   - ContentType: A new ContentType struct initialized with "text" type and the provided text
func NewTextContent(text string) ContentType {
	return ContentType{
		Type: "text",
		Text: text,
	}
}

// NewImageURLContent creates a new ContentType struct with type "image_url" and the provided URL.
// It takes a URL string as input and returns a ContentType struct initialized with the image URL.
// This helper function is used to create proper image URL content for API requests.
// Example:
//
//	content := NewImageURLContent("https://example.com/image.jpg")
func NewImageURLContent(url string) ContentType {
	return ContentType{
		Type: "image_url",
		ImageURL: &ImageURL{
			URL: url,
		},
	}
}

// ImageToBase64 converts an image from an io.Reader into a base64 encoded string with data URI prefix.
// The function reads the entire image data and encodes it to base64, prepending the data URI scheme
// for JPEG images. It enforces a maximum size limit defined by MaxBase64ImageSize.
//
// Parameters:
//   - reader: An io.Reader interface providing the image data to be encoded
//
// Returns:
//   - string: A base64 encoded string with "data:image/jpeg;base64," prefix
//   - error: An error if reading fails or if the image size exceeds MaxBase64ImageSize
func ImageToBase64(reader io.Reader) (string, error) {
	data, err := io.ReadAll(reader)
	if err != nil {
		return "", fmt.Errorf("error reading image: %w", err)
	}

	if len(data) > MaxBase64ImageSize {
		return "", fmt.Errorf("image size exceeds limit of %d bytes", MaxBase64ImageSize)
	}

	return fmt.Sprintf("data:image/jpeg;base64,%s", base64.StdEncoding.EncodeToString(data)), nil
}

// ValidateImageURL performs validation checks on a provided image URL.
// It verifies that:
// - The URL is accessible and returns a successful status code
// - The image size doesn't exceed MaxURLImageSize
// - The content type is a valid image format
//
// Parameters:
//   - url: The URL string to validate
//
// Returns:
//   - error: nil if validation passes, otherwise an error describing the validation failure
//
// Possible errors:
//   - Network errors when accessing the URL
//   - Invalid status code responses
//   - Image size exceeding MaxURLImageSize
//   - Invalid image content types
func ValidateImageURL(url string) error {
	resp, err := http.Head(url)
	if err != nil {
		return fmt.Errorf("error checking image URL: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("invalid image URL, status code: %d", resp.StatusCode)
	}

	size := resp.ContentLength
	if size > MaxURLImageSize {
		return fmt.Errorf("image size (%d bytes) exceeds limit of %d bytes", size, MaxURLImageSize)
	}

	contentType := resp.Header.Get("Content-Type")
	if !isValidImageType(contentType) {
		return fmt.Errorf("invalid image type: %s", contentType)
	}

	return nil
}

// isValidImageType checks if the provided MIME content type represents a supported image format.
// Currently supports JPEG, PNG, GIF, and WebP formats.
//
// Parameters:
//   - contentType: string - The MIME content type to validate
//
// Returns:
//   - bool - true if the content type is supported, false otherwise
func isValidImageType(contentType string) bool {
	validTypes := map[string]bool{
		"image/jpeg": true,
		"image/png":  true,
		"image/gif":  true,
		"image/webp": true,
	}
	return validTypes[contentType]
}

// init initializes the modelInfoMap with configuration details for vision-capable models.
// It sets up specifications for Llama-32 90B Vision and Llama-32 11B Vision models,
// including their context windows, maximum output sizes, developer information,
// preview status, maximum image size limitations, and supported features like
// vision capabilities, tool usage, and JSON mode operation.
func init() {
	modelInfoMap[ModelLlama32_90bVision] = ModelInfo{
		ContextWindow: 8192,
		MaxOutput:     8192,
		Developer:     "Meta",
		IsPreview:     true,
		MaxImageSize:  "20MB",
		Features:      []string{"vision", "tool-use", "json-mode"},
	}

	modelInfoMap[ModelLlama32_11bVision] = ModelInfo{
		ContextWindow: 8192,
		MaxOutput:     8192,
		Developer:     "Meta",
		IsPreview:     true,
		MaxImageSize:  "20MB",
		Features:      []string{"vision", "tool-use", "json-mode"},
	}
}

// CreateVisionRequest generates a chat completion request structured for vision tasks.
// It takes a model type, an image URL, and a question about the image as input.
// The function returns a properly formatted ChatCompletionRequest that can be used
// to make vision-based queries to the Groq API.
//
// Parameters:
//   - model: The ModelType to be used for the vision task
//   - imageURL: URL string pointing to the image to be analyzed
//   - question: String containing the question or prompt about the image
//
// Returns:
//   - *ChatCompletionRequest: A pointer to a new chat completion request configured
//     for vision tasks with the specified model, image, and question
func CreateVisionRequest(model ModelType, imageURL, question string) *ChatCompletionRequest {
	return &ChatCompletionRequest{
		Model: model,
		Messages: []ChatMessage{
			{
				Role: "user",
				Content: []ContentType{
					NewTextContent(question),
					NewImageURLContent(imageURL),
				},
			},
		},
	}
}

// validateVision checks if the ChatCompletionRequest is valid for vision-based tasks.
// It verifies that:
// 1. The selected model supports vision features
// 2. All image URLs in the message content are valid
//
// Returns an error if:
// - The model does not support vision features
// - Any image URL in the messages is invalid
func (r *ChatCompletionRequest) validateVision() error {
	info := r.Model.GetInfo()
	if !containsString(info.Features, "vision") {
		return fmt.Errorf("model %s does not support vision features", r.Model)
	}

	for _, msg := range r.Messages {
		if content, ok := msg.Content.([]ContentType); ok {
			for _, c := range content {
				if c.ImageURL != nil {
					if err := ValidateImageURL(c.ImageURL.URL); err != nil {
						return fmt.Errorf("invalid image URL: %w", err)
					}
				}
			}
		}
	}

	return nil
}

// containsString checks if a given string exists in a slice of strings.
// It returns true if the string is found, false otherwise.
//
// Parameters:
//   - slice: A slice of strings to search through
//   - str: The string to search for
//
// Returns:
//   - bool: true if str is found in slice, false otherwise
func containsString(slice []string, str string) bool {
	for _, s := range slice {
		if s == str {
			return true
		}
	}
	return false
}
