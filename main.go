package main

import (
	"encoding/json"
	"fmt"
	"image"
	"image/jpeg"
	"image/png"
	"io"
	"net/http"
	"os"
	"strconv"
	"time"

	"github.com/nfnt/resize"
)

type Response struct {
	Status  int64  `json:"status"`
	Message string `json:"message"`
	Error   string `json:"error"`
	Path    string `json:"path"`
}

func Upload(w http.ResponseWriter, r *http.Request) {
	var response Response
	// Maximum upload of 10 MB files
	r.ParseMultipartForm(10 << 20)
	timestamp := strconv.FormatInt(time.Now().UTC().UnixNano(), 10)
	// Get handler for filename, size and headers
	file, handler, err := r.FormFile("image")
	if err != nil {
		response.Status = 3
		response.Message = "Error Uploaded File"
		response.Error = err.Error()
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(response)
	}

	defer file.Close()
	fmt.Printf("Uploaded File: %+v\n", handler.Filename)
	fmt.Printf("File Size: %+v\n", handler.Size)
	fmt.Printf("MIME Header: %+v\n", handler.Header)

	// Create file

	images, _, err := image.Decode(file)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		response.Status = 4
		response.Message = "Error Created Thumbnail"
		response.Error = err.Error()
	}

	defer file.Close()

	newImages := resize.Thumbnail(192, 192, images, resize.Lanczos3)

	dst, err := os.Create("/var/www/html/images/" + timestamp + ".png")
	defer dst.Close()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		response.Status = 2
		response.Message = "Error Uploaded File"
		response.Error = err.Error()
	}

	jpeg.Encode(dst, newImages, nil)
	png.Encode(dst, newImages)

	// Copy the uploaded file to the created file on the filesystem
	if _, err := io.Copy(dst, file); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		response.Status = 1
		response.Message = "Error Uploaded File"
		response.Error = err.Error()
	} else {
		response.Status = 0
		response.Message = "Successfully Uploaded File"
		response.Path = os.Getenv("IMAGE_URL") + "/images/" + timestamp + ".png"
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

func setRouter() {
	http.HandleFunc("/api/v2/image", Upload)
	http.ListenAndServe(":8010", nil)
}

func main() {
	fmt.Println("run main")
	setRouter()
}
