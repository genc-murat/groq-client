package groq

import (
	"context"
	"sync"
)

type ParallelResponse struct {
	Response *ChatCompletionResponse
	Error    error
	Index    int
}

// CreateParallelCompletions sends multiple chat completion requests in parallel and returns their responses.
// It respects the rate limit configuration of the client.
//
// Parameters:
//   - ctx: The context to control cancellation and timeout.
//   - requests: A slice of pointers to ChatCompletionRequest, each representing a request to be sent.
//
// Returns:
//   - A slice of ParallelResponse, each containing the response, error (if any), and the index of the request.
func (c *Client) CreateParallelCompletions(ctx context.Context, requests []*ChatCompletionRequest) []ParallelResponse {
	responses := make([]ParallelResponse, len(requests))
	var wg sync.WaitGroup

	rateLimiter := make(chan struct{}, c.config.RateLimit.RequestsPerMinute)

	for i, req := range requests {
		wg.Add(1)
		go func(index int, request *ChatCompletionRequest) {
			defer wg.Done()

			if c.config.RateLimit.Enabled {
				rateLimiter <- struct{}{}
				defer func() { <-rateLimiter }()
			}

			resp, err := c.CreateChatCompletion(ctx, request)
			responses[index] = ParallelResponse{
				Response: resp,
				Error:    err,
				Index:    index,
			}
		}(i, req)
	}

	wg.Wait()
	return responses
}

type BatchProcessor struct {
	client       *Client
	batchSize    int
	maxParallel  int
	rateLimiting bool
}

// NewBatchProcessor creates a new BatchProcessor with the specified batch size and maximum parallelism.
// It initializes the BatchProcessor with rate limiting enabled.
//
// Parameters:
//   - batchSize: The size of each batch to be processed.
//   - maxParallel: The maximum number of parallel processes allowed.
//
// Returns:
//
//	A pointer to the newly created BatchProcessor.
func (c *Client) NewBatchProcessor(batchSize, maxParallel int) *BatchProcessor {
	return &BatchProcessor{
		client:       c,
		batchSize:    batchSize,
		maxParallel:  maxParallel,
		rateLimiting: true,
	}
}

// ProcessBatch processes a batch of ChatCompletionRequest objects in parallel.
// It divides the requests into smaller batches based on the batchSize of the BatchProcessor,
// sends them to the client for parallel processing, and collects the responses.
//
// Parameters:
//   - ctx: The context for controlling the request lifetime.
//   - requests: A slice of pointers to ChatCompletionRequest objects to be processed.
//
// Returns:
//
//	A slice of ParallelResponse objects containing the results of the processed requests.
func (bp *BatchProcessor) ProcessBatch(ctx context.Context, requests []*ChatCompletionRequest) []ParallelResponse {
	totalResponses := make([]ParallelResponse, 0, len(requests))

	for i := 0; i < len(requests); i += bp.batchSize {
		end := i + bp.batchSize
		if end > len(requests) {
			end = len(requests)
		}

		batch := requests[i:end]
		responses := bp.client.CreateParallelCompletions(ctx, batch)
		totalResponses = append(totalResponses, responses...)
	}

	return totalResponses
}
