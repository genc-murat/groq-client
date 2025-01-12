package groq

import "testing"

func TestModelType_String(t *testing.T) {
	tests := []struct {
		name     string
		model    ModelType
		expected string
	}{
		{
			name:     "Stable model",
			model:    ModelLlama31_8bInstant,
			expected: "llama-3.1-8b-instant",
		},
		{
			name:     "Preview model",
			model:    ModelLlama32_1bPreview,
			expected: "llama-3.2-1b-preview",
		},
		{
			name:     "Audio model",
			model:    ModelWhisperLargeV3,
			expected: "whisper-large-v3",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.model.String(); got != tt.expected {
				t.Errorf("ModelType.String() = %v, want %v", got, tt.expected)
			}
		})
	}
}
func TestModelType_IsValid(t *testing.T) {
	tests := []struct {
		name     string
		model    ModelType
		expected bool
	}{
		{
			name:     "Valid stable model",
			model:    ModelLlama31_8bInstant,
			expected: true,
		},
		{
			name:     "Valid preview model",
			model:    ModelLlama32_1bPreview,
			expected: true,
		},
		{
			name:     "Invalid model",
			model:    "invalid-model",
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.model.IsValid(); got != tt.expected {
				t.Errorf("ModelType.IsValid() = %v, want %v", got, tt.expected)
			}
		})
	}
}
func TestModelType_GetInfo(t *testing.T) {
	tests := []struct {
		name     string
		model    ModelType
		expected ModelInfo
	}{
		{
			name:  "Valid stable model",
			model: ModelLlama31_8bInstant,
			expected: ModelInfo{
				ContextWindow: 128000,
				MaxOutput:     8192,
				Developer:     "Meta",
				IsPreview:     false,
			},
		},
		{
			name:  "Valid preview model",
			model: ModelLlama32_1bPreview,
			expected: ModelInfo{
				ContextWindow: 128000,
				MaxOutput:     8192,
				Developer:     "Meta",
				IsPreview:     true,
			},
		},
		{
			name:     "Invalid model",
			model:    "invalid-model",
			expected: ModelInfo{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.model.GetInfo()
			if got != tt.expected {
				t.Errorf("ModelType.GetInfo() = %v, want %v", got, tt.expected)
			}
		})
	}
}
func TestAllModels(t *testing.T) {
	got := AllModels()

	// Check that we got all models from modelInfoMap
	if len(got) != len(modelInfoMap) {
		t.Errorf("AllModels() returned %d models, want %d", len(got), len(modelInfoMap))
	}

	// Verify each model in the result exists in modelInfoMap
	for _, model := range got {
		if _, exists := modelInfoMap[model]; !exists {
			t.Errorf("AllModels() returned model %s that doesn't exist in modelInfoMap", model)
		}
	}

	// Verify no duplicates
	seen := make(map[ModelType]bool)
	for _, model := range got {
		if seen[model] {
			t.Errorf("AllModels() returned duplicate model: %s", model)
		}
		seen[model] = true
	}
}
func TestStableModels(t *testing.T) {
	got := StableModels()

	// Check that all returned models exist and are stable
	for _, model := range got {
		info, exists := modelInfoMap[model]
		if !exists {
			t.Errorf("StableModels() returned model %s that doesn't exist in modelInfoMap", model)
		}
		if info.IsPreview {
			t.Errorf("StableModels() returned preview model: %s", model)
		}
	}

	// Count stable models in modelInfoMap to verify we got them all
	stableCount := 0
	for _, info := range modelInfoMap {
		if !info.IsPreview {
			stableCount++
		}
	}

	if len(got) != stableCount {
		t.Errorf("StableModels() returned %d models, want %d", len(got), stableCount)
	}

	// Verify no duplicates
	seen := make(map[ModelType]bool)
	for _, model := range got {
		if seen[model] {
			t.Errorf("StableModels() returned duplicate model: %s", model)
		}
		seen[model] = true
	}
}
func TestPreviewModels(t *testing.T) {
	got := PreviewModels()

	// Check that all returned models exist and are preview models
	for _, model := range got {
		info, exists := modelInfoMap[model]
		if !exists {
			t.Errorf("PreviewModels() returned model %s that doesn't exist in modelInfoMap", model)
		}
		if !info.IsPreview {
			t.Errorf("PreviewModels() returned non-preview model: %s", model)
		}
	}

	// Count preview models in modelInfoMap to verify we got them all
	previewCount := 0
	for _, info := range modelInfoMap {
		if info.IsPreview {
			previewCount++
		}
	}

	if len(got) != previewCount {
		t.Errorf("PreviewModels() returned %d models, want %d", len(got), previewCount)
	}

	// Verify no duplicates
	seen := make(map[ModelType]bool)
	for _, model := range got {
		if seen[model] {
			t.Errorf("PreviewModels() returned duplicate model: %s", model)
		}
		seen[model] = true
	}
}
func TestChatCompletionRequest_Validate(t *testing.T) {
	tests := []struct {
		name    string
		request ChatCompletionRequest
		wantErr bool
		errMsg  string
	}{
		{
			name: "Valid request",
			request: ChatCompletionRequest{
				Model: ModelLlama31_8bInstant,
				Messages: []ChatMessage{
					{Role: "user", Content: "Hello"},
				},
				MaxTokens: 100,
			},
			wantErr: false,
		},
		{
			name: "Invalid model",
			request: ChatCompletionRequest{
				Model: "invalid-model",
				Messages: []ChatMessage{
					{Role: "user", Content: "Hello"},
				},
			},
			wantErr: true,
			errMsg:  "invalid model: invalid-model",
		},
		{
			name: "Empty messages",
			request: ChatCompletionRequest{
				Model:    ModelLlama31_8bInstant,
				Messages: []ChatMessage{},
			},
			wantErr: true,
			errMsg:  "at least one message is required",
		},
		{
			name: "MaxTokens exceeds limit",
			request: ChatCompletionRequest{
				Model: ModelLlama31_8bInstant,
				Messages: []ChatMessage{
					{Role: "user", Content: "Hello"},
				},
				MaxTokens: 10000, // ModelLlama31_8bInstant has MaxOutput of 8192
			},
			wantErr: true,
			errMsg:  "max_tokens exceeds model limit of 8192",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.request.Validate()
			if (err != nil) != tt.wantErr {
				t.Errorf("ChatCompletionRequest.Validate() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if err != nil && err.Error() != tt.errMsg {
				t.Errorf("ChatCompletionRequest.Validate() error message = %v, want %v", err.Error(), tt.errMsg)
			}
		})
	}
}
