package groq

import (
	"io"
)

type TranscriptionRequest struct {
	File           io.Reader
	FileName       string
	Language       string
	Model          ModelType
	Prompt         string
	ResponseFormat string
	Temperature    float64
}

type TranslationRequest struct {
	File           io.Reader
	FileName       string
	Model          ModelType
	Prompt         string
	ResponseFormat string
	Temperature    float64
}

type TranscriptionResponse struct {
	Text  string `json:"text"`
	XGroq struct {
		ID string `json:"id"`
	} `json:"x_groq"`
}

type TranslationResponse struct {
	Text  string `json:"text"`
	XGroq struct {
		ID string `json:"id"`
	} `json:"x_groq"`
}
