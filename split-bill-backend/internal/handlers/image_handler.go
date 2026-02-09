package handlers

import (
	"encoding/base64"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/splitbill/backend/internal/utils"

	"github.com/gin-gonic/gin"
)

type ImageHandler struct {
	uploadDir string
	baseURL   string
}

func NewImageHandler(uploadDir string, baseURL string) *ImageHandler {
	// Create upload directory if it doesn't exist
	_ = os.MkdirAll(uploadDir, os.ModePerm)
	return &ImageHandler{
		uploadDir: uploadDir,
		baseURL:   baseURL,
	}
}

// UploadImage godoc
// @Summary      Upload image via multipart form
// @Description  Uploads an image file (JPEG, PNG, WebP) with max 10MB size limit
// @Tags         Upload
// @Accept       multipart/form-data
// @Produce      json
// @Param        image  formData  file  true  "Image file to upload"
// @Success      200    {object}  utils.APIResponse
// @Failure      400    {object}  utils.APIResponse
// @Failure      500    {object}  utils.APIResponse
// @Security     BearerAuth
// @Router       /upload/image [post]
func (h *ImageHandler) UploadImage(c *gin.Context) {
	file, header, err := c.Request.FormFile("image")
	if err != nil {
		utils.RespondError(c, http.StatusBadRequest, "No image file provided")
		return
	}
	defer file.Close()

	// Validate file type
	contentType := header.Header.Get("Content-Type")
	if !isValidImageType(contentType) {
		utils.RespondError(c, http.StatusBadRequest, "Invalid image type. Supported: JPEG, PNG, WebP")
		return
	}

	// Validate file size (max 10MB)
	if header.Size > 10*1024*1024 {
		utils.RespondError(c, http.StatusBadRequest, "Image size exceeds 10MB limit")
		return
	}

	// Generate unique filename
	ext := filepath.Ext(header.Filename)
	if ext == "" {
		ext = getExtensionFromContentType(contentType)
	}
	filename := fmt.Sprintf("%d_%s%s", time.Now().UnixNano(), generateID(), ext)

	// Create date-based subdirectory
	dateDir := time.Now().Format("2006/01/02")
	fullDir := filepath.Join(h.uploadDir, dateDir)
	_ = os.MkdirAll(fullDir, os.ModePerm)

	// Save file
	filePath := filepath.Join(fullDir, filename)
	out, err := os.Create(filePath)
	if err != nil {
		utils.RespondError(c, http.StatusInternalServerError, "Failed to save image")
		return
	}
	defer out.Close()

	if _, err := io.Copy(out, file); err != nil {
		utils.RespondError(c, http.StatusInternalServerError, "Failed to write image")
		return
	}

	// Build URL
	imageURL := fmt.Sprintf("%s/uploads/%s/%s", h.baseURL, dateDir, filename)

	utils.RespondSuccess(c, http.StatusOK, "Image uploaded successfully", gin.H{
		"url":       imageURL,
		"file_name": filename,
		"size":      header.Size,
	})
}

// UploadBase64Image godoc
// @Summary      Upload image via base64
// @Description  Uploads a base64-encoded image (supports data URI format) with max 10MB size limit
// @Tags         Upload
// @Accept       json
// @Produce      json
// @Param        request  body      object  true  "Base64 image data"
// @Success      200      {object}  utils.APIResponse
// @Failure      400      {object}  utils.APIResponse
// @Failure      500      {object}  utils.APIResponse
// @Security     BearerAuth
// @Router       /upload/image-base64 [post]
func (h *ImageHandler) UploadBase64Image(c *gin.Context) {
	var req struct {
		Image    string `json:"image" binding:"required"`
		FileName string `json:"file_name"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		utils.RespondError(c, http.StatusBadRequest, "Invalid request: "+err.Error())
		return
	}

	// Parse base64 data
	imageData := req.Image
	var ext string

	// Handle data URI format: data:image/jpeg;base64,/9j/4AAQ...
	if strings.HasPrefix(imageData, "data:") {
		parts := strings.SplitN(imageData, ",", 2)
		if len(parts) != 2 {
			utils.RespondError(c, http.StatusBadRequest, "Invalid base64 image format")
			return
		}
		// Extract content type
		header := parts[0] // e.g., "data:image/jpeg;base64"
		if strings.Contains(header, "image/jpeg") || strings.Contains(header, "image/jpg") {
			ext = ".jpg"
		} else if strings.Contains(header, "image/png") {
			ext = ".png"
		} else if strings.Contains(header, "image/webp") {
			ext = ".webp"
		} else {
			ext = ".jpg"
		}
		imageData = parts[1]
	} else {
		ext = ".jpg" // Default
	}

	// Decode base64
	decoded, err := base64.StdEncoding.DecodeString(imageData)
	if err != nil {
		utils.RespondError(c, http.StatusBadRequest, "Invalid base64 encoding")
		return
	}

	// Validate size (max 10MB)
	if len(decoded) > 10*1024*1024 {
		utils.RespondError(c, http.StatusBadRequest, "Image size exceeds 10MB limit")
		return
	}

	// Generate filename
	filename := fmt.Sprintf("%d_%s%s", time.Now().UnixNano(), generateID(), ext)

	// Create date-based subdirectory
	dateDir := time.Now().Format("2006/01/02")
	fullDir := filepath.Join(h.uploadDir, dateDir)
	_ = os.MkdirAll(fullDir, os.ModePerm)

	// Save file
	filePath := filepath.Join(fullDir, filename)
	if err := os.WriteFile(filePath, decoded, 0644); err != nil {
		utils.RespondError(c, http.StatusInternalServerError, "Failed to save image")
		return
	}

	imageURL := fmt.Sprintf("%s/uploads/%s/%s", h.baseURL, dateDir, filename)

	utils.RespondSuccess(c, http.StatusOK, "Image uploaded successfully", gin.H{
		"url":       imageURL,
		"file_name": filename,
		"size":      len(decoded),
	})
}

func isValidImageType(contentType string) bool {
	validTypes := []string{
		"image/jpeg",
		"image/jpg",
		"image/png",
		"image/webp",
	}
	for _, t := range validTypes {
		if contentType == t {
			return true
		}
	}
	return false
}

func getExtensionFromContentType(contentType string) string {
	switch contentType {
	case "image/jpeg", "image/jpg":
		return ".jpg"
	case "image/png":
		return ".png"
	case "image/webp":
		return ".webp"
	default:
		return ".jpg"
	}
}

func generateID() string {
	return fmt.Sprintf("%x", time.Now().UnixNano()%0xFFFFFF)
}
