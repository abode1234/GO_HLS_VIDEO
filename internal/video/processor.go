package video

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
)

type Processor struct {
}

// NewProcessor ينشئ معالج فيديو جديد
func NewProcessor() *Processor {
	return &Processor{}
}

// ProcessVideo يعالج الفيديو
func (p *Processor) ProcessVideo(inputPath, outputDir string) error {
	err := os.MkdirAll(outputDir, os.ModePerm)
	if err != nil {
		return fmt.Errorf("failed to create output directory: %v", err)
	}

	outputPath := filepath.Join(outputDir, "playlist.m3u8")

	cmd := exec.Command(
		"ffmpeg",
		"-i", inputPath,
		"-profile:v", "baseline",
		"-level", "3.0",
		"-start_number", "0",
		"-hls_time", "10",
		"-hls_list_size", "0",
		"-f", "hls",
		outputPath,
	)

	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("ffmpeg command failed: %v\nOutput: %s", err, string(output))
	}

	return nil
}
