package main

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"os"
)

const groqAPIURL = "https://api.groq.com/openai/v1/audio/transcriptions"

// TranscribeAudio sends the audio data to Groq's API for transcription using whisper-large-v3.
// It assumes that `base64Data` is a raw base64-encoded audio string (without any data URI prefix).
func TranscribeAudio(origin, name, base64Data string) (string, error) {
	apiKey := os.Getenv("GROQ_API_KEY")
	if apiKey == "" {
		return "", fmt.Errorf("GROQ_API_KEY environment variable is not set")
	}

	// Decode the base64-encoded audio data into raw bytes.
	audioBytes, err := base64.StdEncoding.DecodeString(base64Data)
	if err != nil {
		return "", fmt.Errorf("failed to decode base64 audio: %w", err)
	}

	// Create a buffer and a multipart writer.
	var buf bytes.Buffer
	writer := multipart.NewWriter(&buf)

	// Create the file field. Field name "file" is standard for file uploads.
	filePart, err := writer.CreateFormFile("file", name)
	if err != nil {
		return "", fmt.Errorf("failed to create form file: %w", err)
	}
	if _, err := filePart.Write(audioBytes); err != nil {
		return "", fmt.Errorf("failed to write audio data: %w", err)
	}

	// Write the model field.
	if err := writer.WriteField("model", "whisper-large-v3"); err != nil {
		return "", fmt.Errorf("failed to write field 'model': %w", err)
	}

	// Note: We are no longer writing a metadata field since it isn't accepted.
	// If required by your service, add any other accepted fields here.

	// Close the writer to flush the data.
	if err := writer.Close(); err != nil {
		return "", fmt.Errorf("failed to close multipart writer: %w", err)
	}

	// Create a new HTTP POST request.
	req, err := http.NewRequest("POST", groqAPIURL, &buf)
	if err != nil {
		return "", fmt.Errorf("failed to create request: %w", err)
	}
	req.Header.Set("Authorization", "Bearer "+apiKey)
	req.Header.Set("Content-Type", writer.FormDataContentType())

	// Execute the request.
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return "", fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	// Read the entire response body.
	responseBody, _ := io.ReadAll(resp.Body)

	// If status code is not OK, return an error with the response body.
	if resp.StatusCode >= 300 {
		fmt.Println(string(responseBody))
		return "", fmt.Errorf("transcription failed: %s", string(responseBody))
	}

	// Attempt to decode the response as JSON.
	var result struct {
		Transcription string `json:"transcription"`
		Error         string `json:"error,omitempty"`
	}
	if err := json.Unmarshal(responseBody, &result); err != nil || result.Transcription == "" {
		// Fallback: assume the response body is plain text transcription.
		return string(responseBody), nil
	}

	return result.Transcription, nil
}
