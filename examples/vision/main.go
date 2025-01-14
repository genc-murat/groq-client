package main

import (
	"context"
	"fmt"
	"log"
	"os"

	"github.com/genc-murat/groq-client/pkg/groq"
)

func main() {
	client := groq.NewClient(os.Getenv("GROQ_API_KEY"))

	// Example 1: Using image URL
	urlRequest := groq.CreateVisionRequest(
		groq.ModelLlama32_90bVision,
		"https://i0.wp.com/picjumbo.com/wp-content/uploads/san-francisco-bay-area-beautiful-sunset-evening-cityscape-free-photo.jpg?w=2210&quality=70",
		"What's in this image?",
	)

	resp, err := client.CreateChatCompletion(context.Background(), urlRequest)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("URL Image Analysis: %s\n", resp.Choices[0].Message.Content)

	// Example 2: Using local image
	file, err := os.Open("local_image.jpg")
	if err != nil {
		log.Fatal(err)
	}
	defer file.Close()

	// Convert image to base64
	base64Image, err := groq.ImageToBase64(file)
	if err != nil {
		log.Fatal(err)
	}

	// Create request with local image
	localRequest := &groq.ChatCompletionRequest{
		Model: groq.ModelLlama32_90bVision,
		Messages: []groq.ChatMessage{
			{
				Role: "user",
				Content: []groq.ContentType{
					groq.NewTextContent("Describe this image in detail"),
					groq.NewImageURLContent(base64Image),
				},
			},
		},
	}

	// Send request
	resp, err = client.CreateChatCompletion(context.Background(), localRequest)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("Local Image Analysis: %s\n", resp.Choices[0].Message.Content)

	// Example 3: Multi-turn conversation with image
	multiTurnRequest := &groq.ChatCompletionRequest{
		Model: groq.ModelLlama32_90bVision,
		Messages: []groq.ChatMessage{
			{
				Role: "user",
				Content: []groq.ContentType{
					groq.NewTextContent("What's in this image?"),
					groq.NewImageURLContent("https://i0.wp.com/picjumbo.com/wp-content/uploads/san-francisco-bay-area-beautiful-sunset-evening-cityscape-free-photo.jpg?w=2210&quality=70"),
				},
			},
			{
				Role:    "assistant",
				Content: "The image shows a beautiful cityscape of San Francisco...",
			},
			{
				Role:    "user",
				Content: "What famous landmarks can you identify?",
			},
		},
	}

	resp, err = client.CreateChatCompletion(context.Background(), multiTurnRequest)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("Multi-turn Response: %s\n", resp.Choices[0].Message.Content)
}
