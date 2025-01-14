# Groq API Go Client

A high-performance Go client for the Groq API, featuring FastHTTP, semantic caching, streaming support, parallel request processing, audio, and vision capabilities.

[![Go Reference](https://pkg.go.dev/badge/github.com/genc-murat/groq-client.svg)](https://pkg.go.dev/github.com/genc-murat/groq-client)
[![Go Report Card](https://goreportcard.com/badge/github.com/genc-murat/groq-client)](https://goreportcard.com/report/github.com/genc-murat/groq-client)
[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](https://opensource.org/licenses/MIT)

## Features

- 🚀 High-performance FastHTTP-based client
- 🧠 Semantic caching with vector embeddings
- 📡 Streaming support with backpressure handling
- ⚡ Parallel request processing
- 🔄 Automatic retries with exponential backoff
- ⌛ Rate limiting with token bucket algorithm
- 🔒 Type-safe model selection
- 💾 Persistent cache storage
- 📊 Detailed metrics and monitoring
- 🎙️ Audio transcription and translation support
- 🌐 Multi-language audio processing
- 👁️ Vision and multimodal support
- 🖼️ Image analysis and understanding
- 📝 Flexible response formats
- 🔍 Visual question answering

## Table of Contents

- [Installation](#installation)
- [Quick Start](#quick-start)
- [Text Processing](#text-processing)
- [Streaming Support](#streaming-support)
- [Audio Processing](#audio-processing)
- [Vision Features](#vision-features)
- [Semantic Cache](#semantic-cache)
- [Parallel Processing](#parallel-processing)
- [Models](#available-models)
- [Configuration](#configuration-options)
- [Error Handling](#error-handling)
- [Best Practices](#best-practices)
- [Monitoring](#monitoring--metrics)

## Installation

```bash
go get github.com/genc-murat/groq-client
```

## Quick Start

```go
package main

import (
    "context"
    "fmt"
    "github.com/genc-murat/groq-client/pkg/groq"
    "github.com/genc-murat/groq-client/pkg/groq/semantic_cache"
)

func main() {
    // Initialize cache
    cacheConfig := semantic_cache.DefaultConfig()
    cache := semantic_cache.NewSemanticCache(cacheConfig)

    // Create client
    client := groq.NewClient(
        "your-api-key",
        groq.WithCache(cache),
        groq.WithTimeout(30*time.Second),
        groq.WithRetryConfig(3, time.Second),
    )

    // Simple text request
    resp, err := client.CreateChatCompletion(
        context.Background(),
        &groq.ChatCompletionRequest{
            Model: groq.ModelMixtral8x7b32768,
            Messages: []groq.ChatMessage{
                {
                    Role:    "user",
                    Content: "What is the capital of Turkey?",
                },
            },
        },
    )
    if err != nil {
        log.Fatal(err)
    }

    fmt.Println(resp.Choices[0].Message.Content)
}
```

## Text Processing

### Basic Chat

```go
req := &groq.ChatCompletionRequest{
    Model: groq.ModelMixtral8x7b32768,
    Messages: []groq.ChatMessage{
        {Role: "system", Content: "You are a helpful assistant."},
        {Role: "user", Content: "Hello!"},
    },
    Temperature: 0.7,
}
```

### Streaming Support

```go
handler := func(chunk *groq.ChatCompletionChunk) error {
    fmt.Print(chunk.Choices[0].Delta.Content)
    return nil
}

err := client.CreateChatCompletionStream(context.Background(), req, handler)
```

## Audio Processing

### Transcription

```go
file, _ := os.Open("audio.mp3")
defer file.Close()

req := &groq.TranscriptionRequest{
    File:           file,
    FileName:       "audio.mp3",
    Model:          groq.ModelWhisperLargeV3,
    Language:       "tr",
    ResponseFormat: "verbose_json",
    Temperature:    0.3,
    Prompt:        "Technical discussion",
}

resp, err := client.CreateTranscription(context.Background(), req)
```

### Translation

```go
req := &groq.TranslationRequest{
    File:           file,
    FileName:       "speech.mp3",
    Model:          groq.ModelWhisperLargeV3,
    ResponseFormat: "json",
    Temperature:    0.3,
}

resp, err := client.CreateTranslation(context.Background(), req)
```

### Supported Audio Formats
- mp3, mp4, mpeg, mpga, m4a, wav
- webm, ogg, flac

## Vision Features

### URL-based Analysis

```go
req := groq.CreateVisionRequest(
    groq.ModelLlama32_90bVision,
    "https://example.com/image.jpg",
    "What's in this image?",
)

resp, err := client.CreateChatCompletion(context.Background(), req)
```

### Local Image Processing

```go
// Read and encode image
file, _ := os.Open("image.jpg")
base64Image, _ := groq.ImageToBase64(file)

// Create request
req := &groq.ChatCompletionRequest{
    Model: groq.ModelLlama32_90bVision,
    Messages: []groq.ChatMessage{
        {
            Role: "user",
            Content: []groq.ContentType{
                groq.NewTextContent("Analyze this image"),
                groq.NewImageURLContent(base64Image),
            },
        },
    },
}
```

### Multi-turn Visual Dialog

```go
req := &groq.ChatCompletionRequest{
    Model: groq.ModelLlama32_90bVision,
    Messages: []groq.ChatMessage{
        {
            Role: "user",
            Content: []groq.ContentType{
                groq.NewTextContent("What's in this image?"),
                groq.NewImageURLContent(imageURL),
            },
        },
        {
            Role: "assistant",
            Content: "I see a cityscape...",
        },
        {
            Role: "user",
            Content: "List all landmarks visible.",
        },
    },
}
```

## Semantic Cache

```go
config := semantic_cache.DefaultConfig()
config.SimilarityThreshold = 0.85
config.TTL = time.Hour
config.PersistPath = "cache.json"
cache := semantic_cache.NewSemanticCache(config)
```

## Parallel Processing

```go
requests := []*groq.ChatCompletionRequest{
    {
        Model: groq.ModelMixtral8x7b32768,
        Messages: []groq.ChatMessage{{Role: "user", Content: "Q1"}},
    },
    {
        Model: groq.ModelMixtral8x7b32768,
        Messages: []groq.ChatMessage{{Role: "user", Content: "Q2"}},
    },
}

processor := client.NewBatchProcessor(10, 5)
responses := processor.ProcessBatch(context.Background(), requests)
```

## Available Models

```go
// Chat Models
groq.ModelMixtral8x7b32768       // General purpose
groq.ModelLlama33_70bVersatile   // Versatile
groq.ModelGemma29bIt             // Efficient

// Vision Models
groq.ModelLlama32_90bVision      // High capability
groq.ModelLlama32_11bVision      // Fast response

// Audio Models
groq.ModelWhisperLargeV3         // Transcription/Translation
```

## Configuration Options

```go
client := groq.NewClient(
    apiKey,
    groq.WithTimeout(30*time.Second),
    groq.WithRetryConfig(3, time.Second),
    groq.WithRateLimit(60),
    groq.WithCache(cache),
)
```

## Best Practices

### Text Processing
```go
// Temperature control
req.Temperature = 0.7  // Creative
req.Temperature = 0.2  // Focused

// Context management
messages := []groq.ChatMessage{
    {Role: "system", Content: "You are an expert."},
    {Role: "user", Content: "Question"},
}
```

### Audio Processing
```go
// Large file handling
client := groq.NewClient(
    apiKey,
    groq.WithTimeout(5*time.Minute),
)

// Accuracy improvement
req.Language = "tr"
req.Prompt = "Technical terms"
```

### Vision Processing
```go
// Size validation
if err := groq.ValidateImageURL(imageURL); err != nil {
    log.Fatal(err)
}

// Detailed instructions
content := []groq.ContentType{
    groq.NewTextContent("Analyze: objects, text, colors"),
    groq.NewImageURLContent(imageURL),
}
```

## Monitoring & Metrics

```go
stats := client.GetCacheStats()
fmt.Printf("Hits: %d\n", stats.Hits)
fmt.Printf("Misses: %d\n", stats.Misses)
fmt.Printf("Size: %d bytes\n", stats.Size)
```

## Documentation

For detailed API documentation, visit [Go Package Documentation](https://pkg.go.dev/github.com/genc-murat/groq-client).

## Contributing

Contributions are welcome! Please feel free to submit a Pull Request.

## License

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.