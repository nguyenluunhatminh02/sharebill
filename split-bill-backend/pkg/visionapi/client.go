package visionapi

import (
	"context"
	"encoding/base64"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	vision "cloud.google.com/go/vision/v2/apiv1"
	"cloud.google.com/go/vision/v2/apiv1/visionpb"
	"go.uber.org/zap"
)

// Client wraps Google Cloud Vision API
type Client struct {
	client *vision.ImageAnnotatorClient
	logger *zap.Logger
	apiKey string // Fallback: use REST API with key if client is nil
}

// NewClient creates a new Vision API client
// If credentials are available, uses gRPC client; otherwise falls back to API key
func NewClient(logger *zap.Logger, credentialsFile string, apiKey string) (*Client, error) {
	c := &Client{
		logger: logger,
		apiKey: apiKey,
	}

	if credentialsFile != "" {
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()

		client, err := vision.NewImageAnnotatorClient(ctx)
		if err != nil {
			logger.Warn("Failed to create Vision gRPC client, will use API key fallback", zap.Error(err))
		} else {
			c.client = client
			logger.Info("Google Vision API client initialized (gRPC)")
		}
	}

	if c.client == nil && apiKey == "" {
		logger.Warn("No Vision API credentials configured - OCR will use mock/demo mode")
	}

	return c, nil
}

// DetectText performs OCR text detection on an image
func (c *Client) DetectText(ctx context.Context, imageSource string) (string, error) {
	startTime := time.Now()

	var text string
	var err error

	if c.client != nil {
		text, err = c.detectTextGRPC(ctx, imageSource)
	} else if c.apiKey != "" {
		text, err = c.detectTextREST(ctx, imageSource)
	} else {
		// Demo mode - return mock data for development
		text = c.getMockReceiptText()
		err = nil
	}

	elapsed := time.Since(startTime)
	c.logger.Info("OCR detection completed",
		zap.Duration("elapsed", elapsed),
		zap.Int("text_length", len(text)),
		zap.Error(err),
	)

	return text, err
}

// detectTextGRPC uses the gRPC Vision API client
func (c *Client) detectTextGRPC(ctx context.Context, imageSource string) (string, error) {
	var image *visionpb.Image

	if strings.HasPrefix(imageSource, "http://") || strings.HasPrefix(imageSource, "https://") {
		// URL-based image
		image = &visionpb.Image{
			Source: &visionpb.ImageSource{
				ImageUri: imageSource,
			},
		}
	} else if strings.HasPrefix(imageSource, "data:") || isBase64(imageSource) {
		// Base64 encoded image
		b64Data := imageSource
		if idx := strings.Index(b64Data, ","); idx != -1 {
			b64Data = b64Data[idx+1:]
		}
		decoded, err := base64.StdEncoding.DecodeString(b64Data)
		if err != nil {
			return "", fmt.Errorf("failed to decode base64 image: %w", err)
		}
		image = &visionpb.Image{
			Content: decoded,
		}
	} else {
		return "", fmt.Errorf("unsupported image source format")
	}

	// Use BatchAnnotateImages with TEXT_DETECTION feature
	req := &visionpb.BatchAnnotateImagesRequest{
		Requests: []*visionpb.AnnotateImageRequest{
			{
				Image: image,
				Features: []*visionpb.Feature{
					{
						Type:       visionpb.Feature_TEXT_DETECTION,
						MaxResults: 1,
					},
				},
			},
		},
	}

	resp, err := c.client.BatchAnnotateImages(ctx, req)
	if err != nil {
		return "", fmt.Errorf("vision API error: %w", err)
	}

	if len(resp.Responses) == 0 {
		return "", fmt.Errorf("no response from Vision API")
	}

	response := resp.Responses[0]
	if response.Error != nil {
		return "", fmt.Errorf("vision API response error: %s", response.Error.Message)
	}

	if len(response.TextAnnotations) == 0 {
		return "", fmt.Errorf("no text detected in image")
	}

	// First annotation contains the full text
	return response.TextAnnotations[0].GetDescription(), nil
}

// detectTextREST uses the REST API with API key
func (c *Client) detectTextREST(ctx context.Context, imageSource string) (string, error) {
	var requestBody string

	if strings.HasPrefix(imageSource, "http://") || strings.HasPrefix(imageSource, "https://") {
		requestBody = fmt.Sprintf(`{
			"requests": [{
				"image": {"source": {"imageUri": "%s"}},
				"features": [{"type": "TEXT_DETECTION"}]
			}]
		}`, imageSource)
	} else {
		b64Data := imageSource
		if idx := strings.Index(b64Data, ","); idx != -1 {
			b64Data = b64Data[idx+1:]
		}
		requestBody = fmt.Sprintf(`{
			"requests": [{
				"image": {"content": "%s"},
				"features": [{"type": "TEXT_DETECTION"}]
			}]
		}`, b64Data)
	}

	url := fmt.Sprintf("https://vision.googleapis.com/v1/images:annotate?key=%s", c.apiKey)

	req, err := http.NewRequestWithContext(ctx, "POST", url, strings.NewReader(requestBody))
	if err != nil {
		return "", err
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return "", fmt.Errorf("vision API request failed: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("failed to read response: %w", err)
	}

	if resp.StatusCode != 200 {
		return "", fmt.Errorf("vision API error (status %d): %s", resp.StatusCode, string(body))
	}

	// Simple extraction - in production use proper JSON parsing
	text := extractTextFromResponse(string(body))
	return text, nil
}

// getMockReceiptText returns mock receipt text for development/testing
func (c *Client) getMockReceiptText() string {
	return `NHÀ HÀNG PHỞ 24
Địa chỉ: 123 Nguyễn Huệ, Q.1, HCM
ĐT: 028-1234-5678
----------------------------
Bàn: 5     Ngày: 15/01/2025
----------------------------
Phở bò tái    2 x 65,000 = 130,000
Bún bò Huế    1 x 55,000 = 55,000
Gỏi cuốn      2 x 35,000 = 70,000
Chả giò       1 x 45,000 = 45,000
Nước ngọt     3 x 15,000 = 45,000
Trà đá        2 x 10,000 = 20,000
----------------------------
Tổng cộng:           365,000
VAT (10%):            36,500
Phí phục vụ (5%):     18,250
----------------------------
Thành tiền:          419,750
----------------------------
Cảm ơn quý khách!`
}

// extractTextFromResponse extracts text from Vision API JSON response
func extractTextFromResponse(body string) string {
	// Simple extraction - look for "description" field in first textAnnotation
	idx := strings.Index(body, `"description"`)
	if idx == -1 {
		return ""
	}

	// Find the value after "description":
	// Skip "description"
	remaining := body[idx+len(`"description"`):]
	// Skip : and whitespace
	remaining = strings.TrimLeft(remaining, `: "`)
	// Find closing quote (handle escaped quotes)
	end := findUnescapedQuote(remaining)
	if end == -1 {
		return remaining
	}

	text := remaining[:end]
	// Unescape
	text = strings.ReplaceAll(text, `\n`, "\n")
	text = strings.ReplaceAll(text, `\"`, `"`)
	return text
}

func findUnescapedQuote(s string) int {
	for i := 0; i < len(s); i++ {
		if s[i] == '"' && (i == 0 || s[i-1] != '\\') {
			return i
		}
	}
	return -1
}

func isBase64(s string) bool {
	if len(s) < 100 {
		return false
	}
	// Check if string contains only base64 characters
	for _, c := range s[:100] {
		if !((c >= 'A' && c <= 'Z') || (c >= 'a' && c <= 'z') || (c >= '0' && c <= '9') || c == '+' || c == '/' || c == '=') {
			return false
		}
	}
	return true
}

// Close closes the Vision API client
func (c *Client) Close() error {
	if c.client != nil {
		return c.client.Close()
	}
	return nil
}
