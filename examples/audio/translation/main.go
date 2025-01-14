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
	// API anahtarını kontrol et
	apiKey := os.Getenv("GROQ_API_KEY")
	if apiKey == "" {
		log.Fatal("GROQ_API_KEY environment variable is required")
	}

	// Client'ı oluştur
	client := groq.NewClient(
		apiKey,
		groq.WithTimeout(2*time.Minute), // Dosya yükleme için uzun timeout
	)

	// Dosyayı aç
	file, err := os.Open("Recording.m4a")
	if err != nil {
		log.Fatalf("Error opening file: %v", err)
	}
	defer file.Close()

	// Translation isteğini hazırla
	req := &groq.TranslationRequest{
		File:           file,
		FileName:       "Recording.m4a",
		Model:          groq.ModelWhisperLargeV3,
		ResponseFormat: "verbose_json",                               // Detaylı çıktı için
		Temperature:    0.3,                                          // Daha tutarlı sonuçlar için
		Prompt:         "This is a Turkish speech about technology.", // İsteğe bağlı prompt
	}

	fmt.Println("Starting translation...")

	// Çeviri yap
	resp, err := client.CreateTranslation(context.Background(), req)
	if err != nil {
		log.Fatalf("Translation failed: %v", err)
	}

	// Sonuçları yazdır
	fmt.Printf("\nTranslation Results:\n")
	fmt.Printf("Text: %s\n", resp.Text)
	fmt.Printf("Request ID: %s\n", resp.XGroq.ID)
}
