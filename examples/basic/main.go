package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/genc-murat/groq-client/pkg/groq"
	"github.com/genc-murat/groq-client/pkg/groq/semantic_cache"
)

func main() {
	apiKey := os.Getenv("GROQ_API_KEY")
	if apiKey == "" {
		log.Fatal("GROQ_API_KEY environment variable is required")
	}

	cacheConfig := semantic_cache.DefaultConfig()
	cacheConfig.TTL = 1 * time.Hour
	cache := semantic_cache.NewSemanticCache(cacheConfig)

	client := groq.NewClient(
		apiKey,
		groq.WithTimeout(30*time.Second),
		groq.WithRetryConfig(3, time.Second),
		groq.WithRateLimit(60),
		groq.WithCache(cache),
	)

	model := groq.ModelLlama33_70bVersatile
	if !model.IsValid() {
		log.Fatal("Invalid model selected")
	}

	modelInfo := model.GetInfo()
	printModelInfo(modelInfo, model)

	req := &groq.ChatCompletionRequest{
		Model:       model,
		Messages:    buildMessages("Adana hangi bölgede?"),
		Temperature: 0.7,
	}

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	resp, err := sendRequest(ctx, client, req)
	if err != nil {
		log.Fatal(err)
	}

	printResults(resp)

	if stats := client.GetCacheStats(); stats != nil {
		printCacheStats(stats)
	}
}

func printModelInfo(info groq.ModelInfo, model groq.ModelType) {
	fmt.Println("\nModel Information:")
	fmt.Printf("- Model: %s\n", model)
	fmt.Printf("- Developer: %s\n", info.Developer)
	fmt.Printf("- Context Window: %d tokens\n", info.ContextWindow)
	fmt.Printf("- Max Output: %d tokens\n", info.MaxOutput)
	if info.IsPreview {
		fmt.Println("⚠️ This is a preview model")
	}
	fmt.Println()
}

func buildMessages(query string) []groq.ChatMessage {
	return []groq.ChatMessage{
		{
			Role:    "system",
			Content: "You are a helpful assistant that provides accurate and concise information about Turkey's geography.",
		},
		{
			Role:    "user",
			Content: query,
		},
	}
}

func sendRequest(ctx context.Context, client *groq.Client, req *groq.ChatCompletionRequest) (*groq.ChatCompletionResponse, error) {
	startTime := time.Now()
	resp, err := client.CreateChatCompletion(ctx, req)
	if err != nil {
		return nil, fmt.Errorf("request failed: %w", err)
	}

	fmt.Printf("Request completed in: %v\n\n", time.Since(startTime))
	return resp, nil
}

func printResults(resp *groq.ChatCompletionResponse) {
	fmt.Println("Token Usage:")
	fmt.Printf("- Prompt Tokens: %d\n", resp.Usage.PromptTokens)
	fmt.Printf("- Completion Tokens: %d\n", resp.Usage.CompletionTokens)
	fmt.Printf("- Total Tokens: %d\n", resp.Usage.TotalTokens)

	fmt.Printf("\nResponse: %s\n", resp.Choices[0].Message.Content)
}

func printCacheStats(stats *groq.CacheStats) {
	fmt.Println("\nCache Statistics:")
	fmt.Printf("- Hits: %d\n", stats.Hits)
	fmt.Printf("- Misses: %d\n", stats.Misses)
	fmt.Printf("- Items: %d\n", stats.ItemCount)
	fmt.Printf("- Size: %.2f MB\n", float64(stats.Size)/(1024*1024))
}
