package groq

import "fmt"

type ModelType string

// Available Models
const (
	// Stable Models
	ModelDistilWhisperLargeV3En ModelType = "distil-whisper-large-v3-en"
	ModelGemma29bIt             ModelType = "gemma2-9b-it"
	ModelLlama33_70bVersatile   ModelType = "llama-3.3-70b-versatile"
	ModelLlama31_8bInstant      ModelType = "llama-3.1-8b-instant"
	ModelLlamaGuard3_8b         ModelType = "llama-guard-3-8b"
	ModelLlama3_70b_8192        ModelType = "llama3-70b-8192"
	ModelLlama3_8b_8192         ModelType = "llama3-8b-8192"
	ModelMixtral8x7b32768       ModelType = "mixtral-8x7b-32768"
	ModelWhisperLargeV3         ModelType = "whisper-large-v3"
	ModelWhisperLargeV3Turbo    ModelType = "whisper-large-v3-turbo"

	// Preview Models
	ModelLlama33_70bSpecdec ModelType = "llama-3.3-70b-specdec"
	ModelLlama32_1bPreview  ModelType = "llama-3.2-1b-preview"
	ModelLlama32_3bPreview  ModelType = "llama-3.2-3b-preview"
	ModelLlama32_11bVision  ModelType = "llama-3.2-11b-vision-preview"
	ModelLlama32_90bVision  ModelType = "llama-3.2-90b-vision-preview"
)

type ModelInfo struct {
	ContextWindow int    // Maximum context window in tokens, 0 means unspecified
	MaxOutput     int    // Maximum output tokens, 0 means unspecified
	MaxFileSize   string // Maximum file size, empty means unspecified
	IsPreview     bool   // Whether this is a preview model
	Developer     string // Model developer/organization
}

type ChatMessage struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

type ChatCompletionRequest struct {
	Model       ModelType     `json:"model"`
	Messages    []ChatMessage `json:"messages"`
	MaxTokens   int           `json:"max_tokens,omitempty"`
	Temperature float64       `json:"temperature,omitempty"`
	Stream      bool          `json:"stream,omitempty"`
}

type ChatCompletionResponse struct {
	ID      string    `json:"id"`
	Object  string    `json:"object"`
	Created int64     `json:"created"`
	Model   ModelType `json:"model"`
	Usage   struct {
		PromptTokens     int `json:"prompt_tokens"`
		CompletionTokens int `json:"completion_tokens"`
		TotalTokens      int `json:"total_tokens"`
	} `json:"usage"`
	Choices []struct {
		Message      ChatMessage `json:"message"`
		FinishReason string      `json:"finish_reason"`
	} `json:"choices"`
}

type ChatCompletionChunk struct {
	ID      string    `json:"id"`
	Object  string    `json:"object"`
	Created int64     `json:"created"`
	Model   ModelType `json:"model"`
	Choices []struct {
		Delta struct {
			Content string `json:"content"`
			Role    string `json:"role,omitempty"`
		} `json:"delta"`
		FinishReason string `json:"finish_reason"`
	} `json:"choices"`
}

type StreamHandler func(*ChatCompletionChunk) error

// String returns the string representation of the ModelType.
func (m ModelType) String() string {
	return string(m)
}

// IsValid checks if the ModelType exists in the modelInfoMap.
// It returns true if the ModelType exists, otherwise false.
func (m ModelType) IsValid() bool {
	_, exists := modelInfoMap[m]
	return exists
}

// GetInfo retrieves the ModelInfo associated with the ModelType.
// If the ModelType does not exist in the modelInfoMap, it returns an empty ModelInfo struct.
//
// Returns:
//   - ModelInfo: The information associated with the ModelType, or an empty ModelInfo if the ModelType does not exist.
func (m ModelType) GetInfo() ModelInfo {
	info, exists := modelInfoMap[m]
	if !exists {
		return ModelInfo{}
	}
	return info
}

// AllModels returns a slice of all ModelType values present in the modelInfoMap.
// It initializes a slice with a capacity equal to the length of modelInfoMap,
// iterates over the map, and appends each model to the slice.
func AllModels() []ModelType {
	models := make([]ModelType, 0, len(modelInfoMap))
	for model := range modelInfoMap {
		models = append(models, model)
	}
	return models
}

// StableModels returns a slice of ModelType containing all models that are not in preview.
// It iterates over the modelInfoMap and appends models to the slice if their info indicates they are not in preview.
func StableModels() []ModelType {
	models := make([]ModelType, 0)
	for model, info := range modelInfoMap {
		if !info.IsPreview {
			models = append(models, model)
		}
	}
	return models
}

// PreviewModels returns a slice of ModelType containing all models
// that are marked as preview in the modelInfoMap.
func PreviewModels() []ModelType {
	models := make([]ModelType, 0)
	for model, info := range modelInfoMap {
		if info.IsPreview {
			models = append(models, model)
		}
	}
	return models
}

var modelInfoMap = map[ModelType]ModelInfo{
	ModelDistilWhisperLargeV3En: {
		MaxFileSize: "25 MB",
		Developer:   "HuggingFace",
	},
	ModelGemma29bIt: {
		ContextWindow: 8192,
		Developer:     "Google",
	},
	ModelLlama33_70bVersatile: {
		ContextWindow: 128000,
		MaxOutput:     32768,
		Developer:     "Meta",
	},
	ModelLlama31_8bInstant: {
		ContextWindow: 128000,
		MaxOutput:     8192,
		Developer:     "Meta",
	},
	ModelLlamaGuard3_8b: {
		ContextWindow: 8192,
		Developer:     "Meta",
	},
	ModelLlama3_70b_8192: {
		ContextWindow: 8192,
		Developer:     "Meta",
	},
	ModelLlama3_8b_8192: {
		ContextWindow: 8192,
		Developer:     "Meta",
	},
	ModelMixtral8x7b32768: {
		ContextWindow: 32768,
		Developer:     "Mistral",
	},
	ModelWhisperLargeV3: {
		MaxFileSize: "25 MB",
		Developer:   "OpenAI",
	},
	ModelWhisperLargeV3Turbo: {
		MaxFileSize: "25 MB",
		Developer:   "OpenAI",
	},

	// Preview Models
	ModelLlama33_70bSpecdec: {
		ContextWindow: 8192,
		Developer:     "Meta",
		IsPreview:     true,
	},
	ModelLlama32_1bPreview: {
		ContextWindow: 128000,
		MaxOutput:     8192,
		Developer:     "Meta",
		IsPreview:     true,
	},
	ModelLlama32_3bPreview: {
		ContextWindow: 128000,
		MaxOutput:     8192,
		Developer:     "Meta",
		IsPreview:     true,
	},
	ModelLlama32_11bVision: {
		ContextWindow: 128000,
		MaxOutput:     8192,
		Developer:     "Meta",
		IsPreview:     true,
	},
	ModelLlama32_90bVision: {
		ContextWindow: 128000,
		MaxOutput:     8192,
		Developer:     "Meta",
		IsPreview:     true,
	},
}

// Validate checks the ChatCompletionRequest for validity.
// It ensures that the model is valid, there is at least one message,
// and the max_tokens value does not exceed the model's maximum output limit.
// Returns an error if any of these conditions are not met.
func (r *ChatCompletionRequest) Validate() error {
	if !r.Model.IsValid() {
		return fmt.Errorf("invalid model: %s", r.Model)
	}
	if len(r.Messages) == 0 {
		return fmt.Errorf("at least one message is required")
	}

	info := r.Model.GetInfo()
	if info.MaxOutput > 0 && r.MaxTokens > info.MaxOutput {
		return fmt.Errorf("max_tokens exceeds model limit of %d", info.MaxOutput)
	}

	return nil
}
