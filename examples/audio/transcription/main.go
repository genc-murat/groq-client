package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/genc-murat/groq-client/pkg/groq"
)

func main() {
	apiKey := os.Getenv("GROQ_API_KEY")
	if apiKey == "" {
		log.Fatal("GROQ_API_KEY environment variable is required")
	}

	client := groq.NewClient(
		apiKey,
		groq.WithTimeout(2*time.Minute),
		groq.WithRetryConfig(3, time.Second),
	)

	// Transcription örneği
	file, err := os.Open("Recording.m4a")
	if err != nil {
		log.Fatalf("Error opening file: %v", err)
	}
	defer file.Close()

	// Transcription isteği
	transcriptionReq := &groq.TranscriptionRequest{
		File:           file,
		FileName:       "Recording.m4a",
		Model:          groq.ModelWhisperLargeV3,
		Language:       "tr",
		ResponseFormat: "json",
		Temperature:    0.3,
	}

	// İsteği gönder
	resp, err := client.CreateTranscription(context.Background(), transcriptionReq)
	if err != nil {
		log.Fatalf("Transcription error: %v", err)
	}

	fmt.Printf("Transcription Result:\n%s\n", resp.Text)
	fmt.Printf("Request ID: %s\n", resp.XGroq.ID)
}
