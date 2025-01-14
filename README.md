# Groq API Go Client

A high-performance Go client for the Groq API, featuring FastHTTP, semantic caching, streaming support, and parallel request processing.

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
- 📝 Flexible response formats
- 👁️ Vision and multimodal support
- 🖼️ Image analysis and understanding
- 🔍 Visual question answering

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
    // Initialize semantic cache
    cacheConfig := semantic_cache.DefaultConfig()
    cache := semantic_cache.NewSemanticCache(cacheConfig)

    // Initialize client with cache
    client := groq.NewClient(
        "your-api-key",
        groq.WithCache(cache),
        groq.WithTimeout(30*time.Second),
        groq.WithRetryConfig(3, time.Second),
    )

    // Create request
    req := &groq.ChatCompletionRequest{
        Model: groq.ModelMixtral8x7b32768,  // Type-safe model selection
        Messages: []groq.ChatMessage{
            {
                Role:    "user",
                Content: "What is the capital of Turkey?",
            },
        },
    }

    // Send request
    resp, err := client.CreateChatCompletion(context.Background(), req)
    if err != nil {
        log.Fatal(err)
    }

    fmt.Println(resp.Choices[0].Message.Content)

    // Check cache stats
    stats := client.GetCacheStats()
    fmt.Printf("Cache Stats: Hits=%d, Misses=%d\n", stats.Hits, stats.Misses)
}
```

## Streaming Support

Stream responses for real-time processing:

```go
handler := func(chunk *groq.ChatCompletionChunk) error {
    fmt.Print(chunk.Choices[0].Delta.Content)
    return nil
}

err := client.CreateChatCompletionStream(context.Background(), req, handler)
if err != nil {
    log.Fatal(err)
}
```

## Semantic Cache

The client includes a sophisticated semantic caching system that uses vector embeddings to find similar queries:

```go
config := semantic_cache.DefaultConfig()
config.SimilarityThreshold = 0.85
config.TTL = time.Hour
config.PersistPath = "cache.json"
config.MaxEntries = 1000

cache := semantic_cache.NewSemanticCache(config)

client := groq.NewClient(
    "your-api-key",
    groq.WithCache(cache),
)
```

### Cache Features
- Vector embedding-based similarity search
- Time-To-Live (TTL) support
- Automatic pruning of old entries
- Persistent storage
- Memory usage limits
- Access frequency tracking

## Parallel Processing

Process multiple requests in parallel with automatic rate limiting:

```go
requests := []*groq.ChatCompletionRequest{
    {
        Model: groq.ModelMixtral8x7b32768,
        Messages: []groq.ChatMessage{{Role: "user", Content: "Question 1"}},
    },
    {
        Model: groq.ModelMixtral8x7b32768,
        Messages: []groq.ChatMessage{{Role: "user", Content: "Question 2"}},
    },
}

// Process in batches of 10 with max 5 parallel requests
processor := client.NewBatchProcessor(10, 5)
responses := processor.ProcessBatch(context.Background(), requests)

for _, resp := range responses {
    if resp.Error != nil {
        log.Printf("Error at index %d: %v", resp.Index, resp.Error)
        continue
    }
    fmt.Printf("Response %d: %s\n", resp.Index, resp.Response.Choices[0].Message.Content)
}
```

## Audio Processing

### Transcription

Convert audio files to text in their original language:

```go
file, err := os.Open("audio.mp3")
if err != nil {
    log.Fatal(err)
}
defer file.Close()

// Create transcription request
req := &groq.TranscriptionRequest{
    File:           file,
    FileName:       "audio.mp3",
    Model:          groq.ModelWhisperLargeV3,
    Language:       "tr",            // Optional: specify language for better accuracy
    ResponseFormat: "verbose_json",  // Options: "json", "text", "verbose_json"
    Temperature:    0.3,             // Control output randomness
    Prompt:        "Technology discussion in Turkish", // Optional context
}

// Send request
resp, err := client.CreateTranscription(context.Background(), req)
if err != nil {
    log.Fatal(err)
}

fmt.Printf("Transcription: %s\n", resp.Text)
fmt.Printf("Request ID: %s\n", resp.XGroq.ID)
```

### Translation

Translate audio directly to English:

```go
req := &groq.TranslationRequest{
    File:           file,
    FileName:       "foreign_speech.mp3",
    Model:          groq.ModelWhisperLargeV3,
    ResponseFormat: "json",
    Temperature:    0.3,
    Prompt:        "Business meeting discussion", // Optional context
}

resp, err := client.CreateTranslation(context.Background(), req)
if err != nil {
    log.Fatal(err)
}

fmt.Printf("English Translation: %s\n", resp.Text)
```

### Supported Audio Formats

- mp3
- mp4
- mpeg
- mpga
- m4a
- wav
- webm
- ogg
- flac

## Vision Support

### Basic Image Analysis

```go
// Using image URL
urlRequest := groq.CreateVisionRequest(
    groq.ModelLlama32_90bVision,
    "https://example.com/image.jpg",
    "What's in this image?",
)

resp, err := client.CreateChatCompletion(context.Background(), urlRequest)
if err != nil {
    log.Fatal(err)
}
fmt.Printf("Analysis: %s\n", resp.Choices[0].Message.Content)
```

### Local Image Processing

```go
// Open local image
file, err := os.Open("image.jpg")
if err != nil {
    log.Fatal(err)
}
defer file.Close()

// Convert to base64
base64Image, err := groq.ImageToBase64(file)
if err != nil {
    log.Fatal(err)
}

// Create request
req := &groq.ChatCompletionRequest{
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

resp, err := client.CreateChatCompletion(context.Background(), req)
if err != nil {
    log.Fatal(err)
}
```

### Multi-turn Visual Conversations

```go
req := &groq.ChatCompletionRequest{
    Model: groq.ModelLlama32_90bVision,
    Messages: []groq.ChatMessage{
        {
            Role: "user",
            Content: []groq.ContentType{
                groq.NewTextContent("What's in this image?"),
                groq.NewImageURLContent("https://example.com/image.jpg"),
            },
        },
        {
            Role: "assistant",
            Content: "The image shows a cityscape...",
        },
        {
            Role: "user",
            Content: "What landmarks can you identify?",
        },
    },
}
```

### Vision Features

- **Multiple Input Methods**:
  - URL-based image analysis
  - Local file processing with base64 encoding
  - Support for popular image formats

- **Image Limitations**:
  - URL images: Up to 20MB
  - Base64 encoded: Up to 4MB
  - Supported formats: JPEG, PNG, GIF, WebP

- **Vision Models**:
  ```go
  // Available vision models
  groq.ModelLlama32_90bVision  // 90B parameter model
  groq.ModelLlama32_11bVision  // 11B parameter model
  
  // Get model capabilities
  modelInfo := groq.ModelLlama32_90bVision.GetInfo()
  ```

- **Advanced Features**:
  - Multi-turn visual conversations
  - Visual question answering
  - Detailed image analysis
  - OCR capabilities

### Best Practices

```go
// 1. Set appropriate timeouts for image processing
client := groq.NewClient(
    apiKey,
    groq.WithTimeout(60*time.Second),
)

// 2. Validate image sizes
if err := groq.ValidateImageURL(imageURL); err != nil {
    log.Fatal(err)
}

// 3. Use appropriate models for your use case
req.Model = groq.ModelLlama32_90bVision  // More capable
// or
req.Model = groq.ModelLlama32_11bVision  // Faster response

// 4. Provide clear instructions
content := []groq.ContentType{
    groq.NewTextContent("Analyze this image and identify:
        1. Main objects
        2. Colors
        3. Text if any
    "),
    groq.NewImageURLContent(imageURL),
}
```

## Available Models

The client provides type-safe model selection:

```go
// Stable Models
groq.ModelMixtral8x7b32768
groq.ModelLlama33_70bVersatile
groq.ModelGemma29bIt
groq.ModelWhisperLargeV3
// ... and more

// Get model information
modelInfo := groq.ModelMixtral8x7b32768.GetInfo()
fmt.Printf("Context Window: %d\n", modelInfo.ContextWindow)
fmt.Printf("Developer: %s\n", modelInfo.Developer)

// List models
allModels := groq.AllModels()
stableModels := groq.StableModels()
previewModels := groq.PreviewModels()
```

## Configuration Options

```go
client := groq.NewClient(
    "your-api-key",
    // Timeout settings
    groq.WithTimeout(30*time.Second),
    
    // Retry configuration
    groq.WithRetryConfig(3, time.Second),
    
    // Rate limiting
    groq.WithRateLimit(60),
    
    // Cache configuration
    groq.WithCache(cache),
    
    // Custom base URL
    groq.WithBaseURL("https://api.custom-groq.com/v1"),
)
```

## Error Handling

```go
resp, err := client.CreateChatCompletion(ctx, req)
if err != nil {
    switch {
    case errors.Is(err, groq.ErrInvalidRequest):
        // Handle invalid request
    case errors.Is(err, groq.ErrHTTPRequest):
        // Handle HTTP error
    default:
        // Handle other errors
    }
}
```

## Best Practices

### Audio Processing
```go
// 1. Set appropriate timeouts for large files
client := groq.NewClient(
    apiKey,
    groq.WithTimeout(5*time.Minute),
)

// 2. Use response formats based on needs
req.ResponseFormat = "verbose_json"  // For detailed output
// or
req.ResponseFormat = "text"         // For simple text output

// 3. Improve accuracy with language hints
req.Language = "tr"  // For Turkish audio

// 4. Use context prompts for better results
req.Prompt = "Medical terminology discussion"
```

### Chat Completions
```go
// 1. Use appropriate temperature for your use case
req.Temperature = 0.7  // More creative
// or
req.Temperature = 0.2  // More focused

// 2. Implement proper context management
messages := []groq.ChatMessage{
    {Role: "system", Content: "You are a helpful assistant."},
    {Role: "user", Content: "Hello!"},
}

// 3. Use streaming for long responses
handler := func(chunk *groq.ChatCompletionChunk) error {
    // Process chunk
    return nil
}
```

## Monitoring & Metrics

```go
// Get cache statistics
stats := client.GetCacheStats()
fmt.Printf("Cache Hits: %d\n", stats.Hits)
fmt.Printf("Cache Misses: %d\n", stats.Misses)
fmt.Printf("Cache Size: %d bytes\n", stats.Size)
fmt.Printf("Item Count: %d\n", stats.ItemCount)
```

## Documentation

For detailed documentation and examples, visit our [Go Package Documentation](https://pkg.go.dev/github.com/genc-murat/groq-client).

## Contributing

Contributions are welcome! Please feel free to submit a Pull Request.

## License

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.