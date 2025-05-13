package main

import (
	"bytes"
	"encoding/json"
	"log"
	"math"
	"os/exec"
)

func getVideoAspectRatio(filePath string) (string, error) {
	var output struct {
		Streams []struct {
			Width  int `json:"width"`
			Height int `json:"height"`
		} `json:"streams"`
	}
	cmd := exec.Command("ffprobe", "-v", "error", "-print_format", "json", "-show_streams", filePath)
	var out bytes.Buffer
	cmd.Stdout = &out
	err := cmd.Run()
	if err != nil {
		log.Fatalf("Command execution failed: %v", err)
	}

	err = json.Unmarshal(out.Bytes(), &output)
	if err != nil {
		log.Fatalf("Unable to get video properties: %v", err)
	}

	width := output.Streams[0].Width
	height := output.Streams[0].Height

	var aspectRatio string
	switch math.Round(float64(width)/float64(height)*100) / 100 {
	case math.Round(float64(16)/float64(9)*100) / 100:
		aspectRatio = "16:9"
	case math.Round(float64(9)/float64(16)*100) / 100:
		aspectRatio = "9:16"
	default:
		aspectRatio = "other"
	}

	return aspectRatio, nil
}

func processVideoForFastStart(filePath string) (string, error) {
	outFile := filePath + ".processing"
	cmd := exec.Command("ffmpeg", "-i", filePath, "-c", "copy", "-movflags", "faststart", "-f", "mp4", outFile)
	err := cmd.Run()
	if err != nil {
		log.Fatalf("Unable to process video file: %v", err)
	}
	return outFile, nil
}
