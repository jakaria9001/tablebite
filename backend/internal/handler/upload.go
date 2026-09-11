package handler

import (
	"bytes"
	"crypto/sha1"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
	"time"
)

const (
	maxUploadSizeBytes = 5 * 1024 * 1024
	cloudinaryCloud    = "vqdfudcx"
	cloudinaryFolder   = "indian_restaurant"
)

func (h *MenuHandler) UploadImage(w http.ResponseWriter, r *http.Request) {
	if err := r.ParseMultipartForm(maxUploadSizeBytes); err != nil {
		http.Error(w, "image file is too large or invalid", http.StatusBadRequest)
		return
	}

	file, header, err := r.FormFile("file")
	if err != nil {
		http.Error(w, "image file is required", http.StatusBadRequest)
		return
	}
	defer file.Close()

	if err := validateUploadFile(header); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	buffer, err := io.ReadAll(io.LimitReader(file, maxUploadSizeBytes+1))
	if err != nil {
		http.Error(w, "unable to read image file", http.StatusBadRequest)
		return
	}
	if len(buffer) > maxUploadSizeBytes {
		http.Error(w, "image file exceeds 5 MB limit", http.StatusBadRequest)
		return
	}

	publicURL, err := uploadToCloudinary(buffer, header.Filename)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(map[string]any{"url": publicURL})
}

func validateUploadFile(header *multipart.FileHeader) error {
	allowed := map[string]bool{"image/jpeg": true, "image/png": true, "image/webp": true}
	if !allowed[header.Header.Get("Content-Type")] {
		return fmt.Errorf("unsupported image type")
	}

	if header.Size > maxUploadSizeBytes {
		return fmt.Errorf("image must be 5 MB or smaller")
	}

	extension := strings.ToLower(filepath.Ext(header.Filename))
	if extension != ".jpg" && extension != ".jpeg" && extension != ".png" && extension != ".webp" {
		return fmt.Errorf("unsupported image extension")
	}

	return nil
}

func uploadToCloudinary(buffer []byte, filename string) (string, error) {
	cloudName := os.Getenv("CLOUDINARY_CLOUD_NAME")
	apiKey := os.Getenv("CLOUDINARY_API_KEY")
	apiSecret := os.Getenv("CLOUDINARY_API_SECRET")
	folder := os.Getenv("CLOUDINARY_FOLDER")
	if folder == "" {
		folder = cloudinaryFolder
	}
	if cloudName == "" || apiKey == "" || apiSecret == "" {
		return "", fmt.Errorf("cloudinary credentials are not configured")
	}

	if cloudName != cloudinaryCloud {
		return "", fmt.Errorf("cloudinary cloud name is invalid")
	}

	publicID := strings.TrimSuffix(filepath.Base(filename), filepath.Ext(filename))
	timestamp := strconv.FormatInt(time.Now().Unix(), 10)
	params := map[string]string{
		"folder":     folder,
		"public_id":  publicID,
		"timestamp":  timestamp,
	}

	pairs := make([]string, 0, len(params))
	for key, value := range params {
		pairs = append(pairs, fmt.Sprintf("%s=%s", key, value))
	}
	sort.Strings(pairs)
	signaturePayload := strings.Join(pairs, "&") + apiSecret
	signature := sha1.Sum([]byte(signaturePayload))
	signatureHex := hex.EncodeToString(signature[:])

	var body bytes.Buffer
	writer := multipart.NewWriter(&body)
	part, err := writer.CreateFormFile("file", filename)
	if err != nil {
		return "", err
	}
	if _, err := part.Write(buffer); err != nil {
		return "", err
	}
	if err := writer.WriteField("api_key", apiKey); err != nil {
		return "", err
	}
	if err := writer.WriteField("timestamp", timestamp); err != nil {
		return "", err
	}
	if err := writer.WriteField("folder", folder); err != nil {
		return "", err
	}
	if err := writer.WriteField("public_id", publicID); err != nil {
		return "", err
	}
	if err := writer.WriteField("signature", signatureHex); err != nil {
		return "", err
	}
	if err := writer.Close(); err != nil {
		return "", err
	}

	req, err := http.NewRequest(http.MethodPost, fmt.Sprintf("https://api.cloudinary.com/v1_1/%s/image/upload", cloudName), &body)
	if err != nil {
		return "", err
	}
	req.Header.Set("Content-Type", writer.FormDataContentType())
	req.SetBasicAuth(apiKey, apiSecret)

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 400 {
		responseBody, _ := io.ReadAll(resp.Body)
		return "", fmt.Errorf("cloudinary upload failed: %s", strings.TrimSpace(string(responseBody)))
	}

	var payload struct {
		SecureURL string `json:"secure_url"`
		URL       string `json:"url"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&payload); err != nil {
		return "", err
	}

	if payload.SecureURL != "" {
		return payload.SecureURL, nil
	}
	if payload.URL != "" {
		return payload.URL, nil
	}
	return "", fmt.Errorf("cloudinary returned no public url")
}
