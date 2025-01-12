package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"strings"
	"sync/atomic"
	"time"

	"github.com/genc-murat/groq-client/pkg/groq"
)

func main() {
	apiKey := os.Getenv("GROQ_API_KEY")
	if apiKey == "" {
		log.Fatal("GROQ_API_KEY environment variable is required")
	}

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Minute)
	defer cancel()

	client := groq.NewClient(
		apiKey,
		groq.WithTimeout(30*time.Second),
		groq.WithRetryConfig(3, time.Second),
	)

	model := groq.ModelMixtral8x7b32768
	if !model.IsValid() {
		log.Fatal("Invalid model selected")
	}

	printModelInfo(model)

	req := &groq.ChatCompletionRequest{
		Model: model,
		Messages: []groq.ChatMessage{
			{
				Role:    "system",
				Content: "You are a helpful assistant that provides accurate and detailed information about Turkey.",
			},
			{
				Role:    "user",
				Content: "Türkiye'nin başkenti neresidir? Detaylı anlat.",
			},
		},
		Temperature: 0.7,
	}

	var (
		totalTokens uint64
		totalChars  uint64
		words       = make(map[string]int)
	)

	handler := createStreamHandler(&totalTokens, &totalChars, words)

	fmt.Println("\nStreaming response:\n" + strings.Repeat("-", 50) + "\n")

	startTime := time.Now()
	if err := client.CreateChatCompletionStream(ctx, req, handler); err != nil {
		if err == context.DeadlineExceeded {
			log.Fatal("Stream timeout exceeded")
		}
		log.Fatalf("Stream error: %v", err)
	}

	printStreamStats(startTime, totalTokens, totalChars, words)
}

func printModelInfo(model groq.ModelType) {
	info := model.GetInfo()
	fmt.Println("\nModel Information:")
	fmt.Printf("- Name: %s\n", model)
	fmt.Printf("- Developer: %s\n", info.Developer)
	fmt.Printf("- Context Window: %d tokens\n", info.ContextWindow)
	if info.MaxOutput > 0 {
		fmt.Printf("- Max Output: %d tokens\n", info.MaxOutput)
	}
	if info.IsPreview {
		fmt.Println("⚠️ This is a preview model")
	}
	fmt.Printf("- File Size Limit: %s\n", info.MaxFileSize)
}

func createStreamHandler(totalTokens *uint64, totalChars *uint64, words map[string]int) groq.StreamHandler {
	return func(chunk *groq.ChatCompletionChunk) error {
		if len(chunk.Choices) > 0 {
			content := chunk.Choices[0].Delta.Content

			fmt.Print(content)

			atomic.AddUint64(totalTokens, 1)
			atomic.AddUint64(totalChars, uint64(len(content)))

			for _, word := range strings.Fields(content) {
				words[strings.ToLower(word)]++
			}
		}
		return nil
	}
}

func printStreamStats(startTime time.Time, totalTokens, totalChars uint64, words map[string]int) {
	duration := time.Since(startTime)

	fmt.Printf("\n\n%s\n", strings.Repeat("-", 50))
	fmt.Println("Stream Statistics:")
	fmt.Printf("- Duration: %v\n", duration)
	fmt.Printf("- Total Chunks: %d\n", totalTokens)
	fmt.Printf("- Total Characters: %d\n", totalChars)
	fmt.Printf("- Average Speed: %.2f tokens/second\n",
		float64(totalTokens)/duration.Seconds())

	fmt.Printf("- Unique Words: %d\n", len(words))

	if len(words) > 0 {
		fmt.Println("\nMost Common Words:")
		topWords := findTopWords(words, 5)
		for i, w := range topWords {
			fmt.Printf("%d. %q (%d times)\n", i+1, w.word, w.count)
		}
	}
}

type wordCount struct {
	word  string
	count int
}

func findTopWords(words map[string]int, n int) []wordCount {
	stopWords := map[string]bool{
		"ve": true, "bir": true, "bu": true, "da": true, "de": true,
	}

	var wc []wordCount
	for word, count := range words {
		if !stopWords[word] && len(word) > 2 {
			wc = append(wc, wordCount{word, count})
		}
	}

	for i := 0; i < len(wc)-1; i++ {
		for j := i + 1; j < len(wc); j++ {
			if wc[j].count > wc[i].count {
				wc[i], wc[j] = wc[j], wc[i]
			}
		}
	}

	if len(wc) > n {
		wc = wc[:n]
	}
	return wc
}
