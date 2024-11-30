package main

import (
	"log"
	"net/http"
	"os"

	"video-hls/internal/database"
	"video-hls/internal/handlers"
	"video-hls/internal/storage"
	"video-hls/internal/video"

	"github.com/joho/godotenv"
)

func main() {
	if err := godotenv.Load(); err != nil {
		log.Println("Warning: Error loading .env file")
	}

	// Check required environment variables
	accessKey := os.Getenv("AWS_ACCESS_KEY")
	secretKey := os.Getenv("AWS_SECRET_KEY")
	bucket := os.Getenv("AWS_BUCKET")
	region := os.Getenv("AWS_REGION")
	mongoURI := os.Getenv("MONGO_URL")
	videoOutputDir := os.Getenv("VIDEO_OUTPUT_DIR")

	if accessKey == "" || secretKey == "" || bucket == "" || mongoURI == "" || videoOutputDir == "" {
		log.Fatal("Missing required environment variables")
	}

	if region == "" {
		log.Println("AWS_REGION not set, defaulting to eu-north-1")
		region = "eu-north-1"
	}

	// Initialize S3 client
	s3Client, err := storage.NewS3Client(accessKey, secretKey, bucket, region)
	if err != nil {
		log.Fatalf("Failed to initialize S3 client: %v", err)
	}

	// Initialize MongoDB client
	mongodb, err := database.NewMongoDB(mongoURI, "yourDatabaseName", "file_uploads")
	if err != nil {
		log.Fatalf("Failed to initialize MongoDB client: %v", err)
	}
	defer mongodb.Close()

	// Initialize video processor
	videoProcessor := video.NewProcessor()

	// Initialize handlers
	uploadHandler := handlers.News3UploadHandler(s3Client, mongodb)
	videoHandler := handlers.NewVideoHandler(videoProcessor, videoOutputDir, mongodb)

	// Set up routes
	http.HandleFunc("/", serveHome)
	http.HandleFunc("/upload", uploadHandler.HandleUpload)
	http.HandleFunc("/upload-video", videoHandler.HandleVideoUpload) // تأكد من وجود هذه الدالة في VideoHandler
	http.HandleFunc("/hls/", videoHandler.ServeHLS) // تأكد من وجود هذه الدالة في VideoHandler
	http.HandleFunc("/video-player", videoHandler.ServeVideoPlayer) // تأكد من وجود هذه الدالة في VideoHandler
	http.HandleFunc("/video-list", videoHandler.ServeVideoList) // تأكد من وجود هذه الدالة في VideoHandler
	http.HandleFunc("/load-video", videoHandler.LoadVideo) // تأكد من وجود هذه الدالة في VideoHandler

	// Serve static files (including video segments)
	fs := http.FileServer(http.Dir(videoOutputDir))
	http.Handle("/segments/", http.StripPrefix("/segments/", fs))

	// Start server
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	log.Printf("Server starting on port %s...\n", port)
	log.Printf("Using AWS region: %s and bucket: %s\n", region, bucket)

	if err := http.ListenAndServe(":"+port, nil); err != nil {
		log.Fatalf("Server error: %v", err)
	}
}

func serveHome(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != "/" {
		http.NotFound(w, r)
		return
	}

	http.ServeFile(w, r, "web/templates/upload.html")
}
