package semantic_cache

import (
	"context"
	"math"
	"testing"
)

func TestNewEmbeddingService(t *testing.T) {
	tests := []struct {
		name      string
		model     string
		wantModel string
		wantDim   int
	}{
		{
			name:      "creates service with default dimension",
			model:     "test-model",
			wantModel: "test-model",
			wantDim:   128,
		},
		{
			name:      "creates service with empty model string",
			model:     "",
			wantModel: "",
			wantDim:   128,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := NewEmbeddingService(tt.model)

			if got.model != tt.wantModel {
				t.Errorf("NewEmbeddingService().model = %v, want %v", got.model, tt.wantModel)
			}

			if got.dimension != tt.wantDim {
				t.Errorf("NewEmbeddingService().dimension = %v, want %v", got.dimension, tt.wantDim)
			}
		})
	}
}
func TestGetEmbedding(t *testing.T) {
	tests := []struct {
		name    string
		text    string
		dim     int
		wantErr bool
	}{
		{
			name:    "gets embedding with default dimension",
			text:    "test text",
			dim:     128,
			wantErr: false,
		},
		{
			name:    "gets embedding with custom dimension",
			text:    "test text",
			dim:     256,
			wantErr: false,
		},
		{
			name:    "gets embedding for empty text",
			text:    "",
			dim:     128,
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := context.Background()
			es := NewEmbeddingService("test-model")
			es.SetDimension(tt.dim)

			got, err := es.GetEmbedding(ctx, tt.text)
			if (err != nil) != tt.wantErr {
				t.Errorf("GetEmbedding() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if len(got) != tt.dim {
				t.Errorf("GetEmbedding() vector length = %v, want %v", len(got), tt.dim)
			}

			// Check if vector is normalized
			var sum float32
			for _, v := range got {
				sum += v * v
			}
			magnitude := float32(math.Sqrt(float64(sum)))
			if !almostEqual(magnitude, 1.0) {
				t.Errorf("GetEmbedding() vector magnitude = %v, want 1.0", magnitude)
			}
		})
	}
}

func TestSetDimension(t *testing.T) {
	tests := []struct {
		name       string
		dimension  int
		wantChange bool
	}{
		{
			name:       "set positive dimension",
			dimension:  256,
			wantChange: true,
		},
		{
			name:       "set zero dimension",
			dimension:  0,
			wantChange: false,
		},
		{
			name:       "set negative dimension",
			dimension:  -1,
			wantChange: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			es := NewEmbeddingService("test-model")
			originalDim := es.GetDimension()

			es.SetDimension(tt.dimension)

			if tt.wantChange && es.GetDimension() != tt.dimension {
				t.Errorf("SetDimension(%v) = %v, want %v", tt.dimension, es.GetDimension(), tt.dimension)
			}
			if !tt.wantChange && es.GetDimension() != originalDim {
				t.Errorf("SetDimension(%v) changed dimension to %v, should remain %v", tt.dimension, es.GetDimension(), originalDim)
			}
		})
	}
}

func TestGetDimension(t *testing.T) {
	tests := []struct {
		name       string
		dimension  int
		wantResult int
	}{
		{
			name:       "get default dimension",
			dimension:  128,
			wantResult: 128,
		},
		{
			name:       "get custom dimension",
			dimension:  256,
			wantResult: 256,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			es := NewEmbeddingService("test-model")
			es.SetDimension(tt.dimension)

			got := es.GetDimension()
			if got != tt.wantResult {
				t.Errorf("GetDimension() = %v, want %v", got, tt.wantResult)
			}
		})
	}
}

func almostEqual(a, b float32) bool {
	return math.Abs(float64(a-b)) < 1e-6
}
