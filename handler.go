package main

import (
	"encoding/json"
	"fmt"
	"net/http"
)

func transcribeHandler(w http.ResponseWriter, r *http.Request) {
	var payload struct {
		Origin string `json:"origin"`
		Name   string `json:"name"`
		Data   string `json:"data"`
	}

	if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
	}

	transcribedText, err := TranscribeAudio(payload.Origin, payload.Name, payload.Data)
	if err != nil {
		http.Error(w, fmt.Sprintf("Transcription failed: %v", err), http.StatusInternalServerError)
		return
	}

	response := map[string]string{
		"transcription": transcribedText,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}
