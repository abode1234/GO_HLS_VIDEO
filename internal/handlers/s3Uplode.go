package handlers

import (
	"context"
	"net/http"
	"time"
	"video-hls/internal/database"
	"video-hls/internal/storage"
)

type s3UploadHandler struct {
	s3Client *storage.S3Client
	mongodb  *database.MongoDB
}

func News3UploadHandler(s3Client *storage.S3Client, mongodb *database.MongoDB) *s3UploadHandler {
	return &s3UploadHandler{
		s3Client: s3Client,
		mongodb:  mongodb,
	}
}

func (h *s3UploadHandler) HandleUpload(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	file, header, err := r.FormFile("file")
	if err != nil {
		http.Error(w, "Error getting file from form: "+err.Error(), http.StatusBadRequest)
		return
	}
	defer file.Close()

	ctx, cancel := context.WithTimeout(r.Context(), 30*time.Second)
	defer cancel()

	err = h.s3Client.UploadFile(ctx, header.Filename, file)
	if err != nil {
		http.Error(w, "Error uploading to S3: "+err.Error(), http.StatusInternalServerError)
		return
	}

	// Store file metadata in MongoDB
	fileMetadata := map[string]interface{}{
		"filename":    header.Filename,
		"size":        header.Size,
		"contentType": header.Header.Get("Content-Type"),
		"uploadTime":  time.Now(),
	}

	err = h.mongodb.InsertDocument("file_uploads", fileMetadata)
	if err != nil {
		http.Error(w, "Error storing metadata in MongoDB: "+err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	w.Write([]byte("File uploaded successfully and metadata stored"))
}
