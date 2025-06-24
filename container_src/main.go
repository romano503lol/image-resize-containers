package main

import (
	"fmt"
	"image"
	_ "image/jpeg"
	_ "image/png"
	"log"
	"net/http"
	"strconv"

	"github.com/disintegration/imaging"
)

// Simple handler for root path
func handler(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintf(w, "Image Resize Server - POST /resize with image file")
}

// Resize handler processes image resize requests

func resizeHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Parse width and height parameters (optional)
	widthStr := r.URL.Query().Get("width")
	heightStr := r.URL.Query().Get("height")

	var width, height int
	var err error

	if widthStr != "" {
		width, err = strconv.Atoi(widthStr)
		if err != nil || width <= 0 {
			http.Error(w, "Invalid width parameter", http.StatusBadRequest)
			return
		}
	}

	if heightStr != "" {
		height, err = strconv.Atoi(heightStr)
		if err != nil || height <= 0 {
			http.Error(w, "Invalid height parameter", http.StatusBadRequest)
			return
		}
	}

	// Get uploaded file
	file, _, err := r.FormFile("image")
	if err != nil {
		http.Error(w, "Failed to get image file", http.StatusBadRequest)
		return
	}
	defer file.Close()

	// Decode and detect format
	img, format, err := image.Decode(file)
	if err != nil {
		http.Error(w, "Failed to decode image", http.StatusBadRequest)
		return
	}

	// Resize with aspect ratio preserved, or return original if no parameters
	var resized image.Image
	if widthStr == "" && heightStr == "" {
		// No parameters - return original image
		resized = img
	} else if width > 0 && height > 0 {
		// Both dimensions specified - fit within bounds
		resized = imaging.Fit(img, width, height, imaging.Lanczos)
	} else if width > 0 {
		// Only width specified - resize to width, keep aspect ratio
		resized = imaging.Resize(img, width, 0, imaging.Lanczos)
	} else {
		// Only height specified - resize to height, keep aspect ratio
		resized = imaging.Resize(img, 0, height, imaging.Lanczos)
	}

	// Output same format as input
	if format == "png" {
		w.Header().Set("Content-Type", "image/png")
		if err := imaging.Encode(w, resized, imaging.PNG); err != nil {
			http.Error(w, "Failed to encode image", http.StatusInternalServerError)
			return
		}
	} else {
		// Default to JPEG for all other formats (jpeg, gif, etc.)
		w.Header().Set("Content-Type", "image/jpeg")
		if err := imaging.Encode(w, resized, imaging.JPEG); err != nil {
			http.Error(w, "Failed to encode image", http.StatusInternalServerError)
			return
		}
	}
}

func main() {
	http.HandleFunc("/", handler)
	http.HandleFunc("/resize", resizeHandler)
	log.Fatal(http.ListenAndServe(":8080", nil))
}
