package server

import (
	"encoding/json"
	"io"
	"log"
	"net/http"
	"os"
)

type Server struct {
	Port string
}

func metaDataHandler(w http.ResponseWriter, r *http.Request) {
	idParam := r.URL.Query().Get("id")
	response := map[string]interface{}{
		"imageId": idParam,
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	err := json.NewEncoder(w).Encode(response)
	if err != nil {
		http.Error(w, "Failed to send response", http.StatusInternalServerError)
		return
	}
}

func imageHandler(w http.ResponseWriter, r *http.Request) {
	file, err := os.Open("img.png")
	if err != nil {
		http.Error(w, "Image not found", http.StatusNotFound)
		log.Printf("Failed to open image: %v", err)
		return
	}
	defer file.Close()

	w.Header().Set("Content-Type", "image/png")
	w.Header().Set("Cache-Control", "max-age=3600")

	// Write the image file to the response
	_, err = io.Copy(w, file)
	if err != nil {
		http.Error(w, "Failed to send image", http.StatusInternalServerError)
		log.Printf("Failed to copy image to response: %v", err)
	}
}

func (s *Server) Start() {
	http.HandleFunc("/image", imageHandler)
	http.HandleFunc("/meta", metaDataHandler)

	log.Printf("Starting server on %s", s.Port)

	go func() {
		if err := http.ListenAndServe(s.Port, nil); err != nil {
			log.Fatalf("Server failed: %v", err)
		}
	}()
}
