package handlers

import (
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"text/template"
	"time"
	"video-hls/internal/database"
	"video-hls/internal/video"
	"go.mongodb.org/mongo-driver/bson"
)

type VideoHandler struct {
	videoProcessor *video.Processor
	outputDir      string
	mongodb        *database.MongoDB
}

// NewVideoHandler ينشئ معالج فيديو جديد
func NewVideoHandler(videoProcessor *video.Processor, outputDir string, mongodb *database.MongoDB) *VideoHandler {
	return &VideoHandler{
		videoProcessor: videoProcessor,
		outputDir:      outputDir,
		mongodb:        mongodb,
	}
}

// HandleVideoUpload يعالج طلبات رفع الفيديو
func (h *VideoHandler) HandleVideoUpload(w http.ResponseWriter, r *http.Request) {
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

	// حفظ الفيديو في المسار المحدد
	videoPath := filepath.Join(h.outputDir, header.Filename)
	outFile, err := os.Create(videoPath)
	if err != nil {
		http.Error(w, "Error saving video file: "+err.Error(), http.StatusInternalServerError)
		return
	}
	defer outFile.Close()

	_, err = io.Copy(outFile, file)
	if err != nil {
		http.Error(w, "Error saving video content: "+err.Error(), http.StatusInternalServerError)
		return
	}

	// معالجة الفيديو باستخدام FFmpeg
	err = h.videoProcessor.ProcessVideo(videoPath, h.outputDir)
	if err != nil {
		http.Error(w, "Error processing video: "+err.Error(), http.StatusInternalServerError)
		return
	}

	// حفظ بيانات الفيديو في MongoDB
	videoMetadata := map[string]interface{}{
		"filename":   header.Filename,
		"uploadTime": time.Now(),
		"outputPath": filepath.Join(h.outputDir, "playlist.m3u8"),
	}


	err = h.mongodb.InsertDocument("video_uploads", videoMetadata)
	if err != nil {
		log.Printf("Error saving video metadata: %v", err) // إضافة سجل خطأ
		http.Error(w, "Error saving video metadata: "+err.Error(), http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("Video uploaded and processed successfully"))
}

// ServeHLS يعرض فيديو HLS
func (h *VideoHandler) ServeHLS(w http.ResponseWriter, r *http.Request) {
	videoFile := r.URL.Path[len("/hls/"):]
	http.ServeFile(w, r, filepath.Join(h.outputDir, videoFile))
}

// ServeVideoPlayer يعرض مشغل الفيديو
func (h *VideoHandler) ServeVideoPlayer(w http.ResponseWriter, r *http.Request) {
	tmpl, err := template.ParseFiles("web/templates/video-player.html")
	if err != nil {
		http.Error(w, "Error loading template: "+err.Error(), http.StatusInternalServerError)
		return
	}

	videoDocs, err := h.mongodb.FindDocuments("video_uploads", bson.M{})
	if err != nil {
		log.Printf("Error retrieving video list: %v", err)
		http.Error(w, "Error retrieving video list", http.StatusInternalServerError)
		return
	}

	if videoDocs == nil || len(videoDocs) == 0 {
		http.Error(w, "No videos found", http.StatusNotFound)
		return
	}

	err = tmpl.Execute(w, videoDocs)
	if err != nil {
		http.Error(w, "Error executing template: "+err.Error(), http.StatusInternalServerError)
	}
}

// ServeVideoList يعرض قائمة الفيديوهات
func (h *VideoHandler) ServeVideoList(w http.ResponseWriter, r *http.Request) {
	videoFiles, err := h.mongodb.FindDocuments("video_uploads", nil)
	if err != nil {
		http.Error(w, "Error retrieving video list: "+err.Error(), http.StatusInternalServerError)
		return
	}

	// عرض قائمة الفيديوهات
	for _, video := range videoFiles {
		fmt.Fprintf(w, "<a href='/hls/%s'>%s</a><br>", video["outputPath"], video["filename"])
	}
}

// LoadVideo يقوم بتحميل الفيديو
func (h *VideoHandler) LoadVideo(w http.ResponseWriter, r *http.Request) {
	videoFile := r.URL.Query().Get("file")
	http.ServeFile(w, r, filepath.Join(h.outputDir, videoFile))
}
