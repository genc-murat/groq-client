package semantic_cache

import (
	"context"
	"crypto/sha256"
	"encoding/binary"
	"math"
)

type EmbeddingService struct {
	model     string
	dimension int
}

// NewEmbeddingService creates a new instance of EmbeddingService with the specified model.
// It initializes the dimension to 128.
//
// Parameters:
//   - model: A string representing the model to be used.
//
// Returns:
//   - A pointer to an EmbeddingService instance.
func NewEmbeddingService(model string) *EmbeddingService {
	return &EmbeddingService{
		model:     model,
		dimension: 128,
	}
}

// GetEmbedding retrieves the embedding vector for the given text.
// If the context is done before the embedding is retrieved, it returns an error.
//
// Parameters:
//   - ctx: The context for controlling cancellation and deadlines.
//   - text: The input text for which the embedding is to be generated.
//
// Returns:
//   - Vector: The embedding vector for the input text.
//   - error: An error if the context is done before the embedding is retrieved.
func (es *EmbeddingService) GetEmbedding(ctx context.Context, text string) (Vector, error) {
	select {
	case <-ctx.Done():
		return nil, ctx.Err()
	default:
		return mockEmbedding(text, es.dimension), nil
	}
}

// mockEmbedding generates a mock embedding vector for the given text.
// The embedding is created by hashing the text using SHA-256 and then
// converting the hash into a vector of the specified dimension. Each
// element of the vector is a float32 value derived from the hash.
//
// Parameters:
//   - text: The input text to be embedded.
//   - dimension: The dimension of the resulting embedding vector.
//
// Returns:
//
//	A Vector of the specified dimension containing the mock embedding.
func mockEmbedding(text string, dimension int) Vector {
	hash := sha256.Sum256([]byte(text))
	vector := make(Vector, dimension)

	for i := 0; i < dimension; i++ {
		hashIndex := (i * 4) % len(hash)
		bits := binary.BigEndian.Uint32(hash[hashIndex : hashIndex+4])

		float := float32(bits) / float32(math.MaxUint32)

		vector[i] = float
	}

	normalize(vector)

	return vector
}

// normalize scales the elements of the given vector so that the vector's magnitude becomes 1.
// If the magnitude of the vector is 0, the function returns without modifying the vector.
//
// Parameters:
//
//	v - the vector to be normalized
func normalize(v Vector) {
	var sum float32
	for _, x := range v {
		sum += x * x
	}
	magnitude := float32(math.Sqrt(float64(sum)))

	if magnitude == 0 {
		return
	}

	for i := range v {
		v[i] /= magnitude
	}
}

// SetDimension sets the dimension of the embedding service to the specified value
// if the provided dimension is greater than 0.
//
// Parameters:
//
//	dimension - the new dimension to set for the embedding service; must be greater than 0
func (es *EmbeddingService) SetDimension(dimension int) {
	if dimension > 0 {
		es.dimension = dimension
	}
}

// GetDimension returns the dimension of the embedding service.
// It retrieves the value of the dimension field from the EmbeddingService instance.
func (es *EmbeddingService) GetDimension() int {
	return es.dimension
}
