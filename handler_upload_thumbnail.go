package main

import (
	"crypto/rand"
	"encoding/base64"
	"fmt"
	"io"
	"mime"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"github.com/bootdotdev/learn-file-storage-s3-golang-starter/internal/auth"
	"github.com/google/uuid"
)

func (cfg *apiConfig) handlerUploadThumbnail(w http.ResponseWriter, r *http.Request) {
	videoIDString := r.PathValue("videoID")
	videoID, err := uuid.Parse(videoIDString)
	if err != nil {
		respondWithError(w, http.StatusBadRequest, "Invalid ID", err)
		return
	}

	token, err := auth.GetBearerToken(r.Header)
	if err != nil {
		respondWithError(w, http.StatusUnauthorized, "Couldn't find JWT", err)
		return
	}

	userID, err := auth.ValidateJWT(token, cfg.jwtSecret)
	if err != nil {
		respondWithError(w, http.StatusUnauthorized, "Couldn't validate JWT", err)
		return
	}

	fmt.Println("uploading thumbnail for video", videoID, "by user", userID)

	const maxMemory = 10 << 20
	r.ParseMultipartForm(maxMemory)

	file, header, err := r.FormFile("thumbnail")
	if err != nil {
		respondWithError(w, http.StatusBadRequest, "Unable to parse form file", err)
	}

	contentType := header.Header.Get("Content-Type")
	mediaType, _, err := mime.ParseMediaType(contentType)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Unable to process mime-type", err)
	}
	if mediaType != "image/jpeg" && mediaType != "image/png" {
		respondWithError(w, http.StatusUnsupportedMediaType, "Unsupported MIME type", nil)
	}

	contentTypeParts := strings.Split(mediaType, "/")
	fileExtension := contentTypeParts[len(contentTypeParts)-1]

	randomName := make([]byte, 32)
	rand.Read(randomName)
	randomNameURL := make([]byte, base64.RawURLEncoding.EncodedLen(len(randomName)))
	base64.RawURLEncoding.Encode(randomNameURL, randomName)
	randomNameString := string(randomNameURL)

	filename := fmt.Sprintf("%s.%s", randomNameString, fileExtension)
	filePath := filepath.Join(cfg.assetsRoot, filename)

	mediaFile, err := os.Create(filePath)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Unable to create file", err)
	}

	metaData, err := cfg.db.GetVideo(videoID)
	if err != nil {
		respondWithError(w, http.StatusNotFound, "Unable to retrieve video from DB", err)
	}

	if metaData.UserID != userID {
		respondWithError(w, http.StatusUnauthorized, "Not authorized", err)
	}

	_, err = io.Copy(mediaFile, file)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Unable to write content to file", err)
	}

	thumbURL := fmt.Sprintf("http://localhost:%s/assets/%s", cfg.port, filename)
	metaData.ThumbnailURL = &thumbURL

	err = cfg.db.UpdateVideo(metaData)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Unable to update video metadata in DB", err)
	}

	respondWithJSON(w, http.StatusOK, metaData)
}
