package utils

import (
	"math"
	"regexp"
	"strconv"
	"strings"

	"github.com/splitbill/backend/internal/models"
)

// ReceiptParser parses raw OCR text into structured receipt items
type ReceiptParser struct {
	// Regex patterns for Vietnamese receipts
	itemPatterns    []*regexp.Regexp
	totalPatterns   []*regexp.Regexp
	taxPatterns     []*regexp.Regexp
	servicePatterns []*regexp.Regexp
	discountPatterns []*regexp.Regexp
	numberPattern   *regexp.Regexp
}

// NewReceiptParser creates a new receipt parser with Vietnamese receipt patterns
func NewReceiptParser() *ReceiptParser {
	return &ReceiptParser{
		itemPatterns: []*regexp.Regexp{
			// Pattern: "Item name    2 x 50,000 = 100,000"
			regexp.MustCompile(`(?i)(.+?)\s+(\d+)\s*[xX×]\s*([\d.,]+)\s*=?\s*([\d.,]+)`),
			// Pattern: "Item name    100,000"
			regexp.MustCompile(`(?i)^(.+?)\s{2,}([\d.,]+)\s*$`),
			// Pattern: "1. Item name    100,000"
			regexp.MustCompile(`(?i)^\d+[.)]\s*(.+?)\s{2,}([\d.,]+)\s*$`),
			// Pattern: "Item name x2    100,000"
			regexp.MustCompile(`(?i)(.+?)\s*[xX×](\d+)\s+([\d.,]+)\s*$`),
		},
		totalPatterns: []*regexp.Regexp{
			regexp.MustCompile(`(?i)(t[oô]ng|total|th[àa]nh\s*ti[eề]n|c[oộ]ng|sum|amount)\s*:?\s*([\d.,]+)`),
			regexp.MustCompile(`(?i)(t[oô]ng\s*c[oộ]ng)\s*:?\s*([\d.,]+)`),
			regexp.MustCompile(`(?i)(TOTAL|TỔNG)\s*:?\s*([\d.,]+)`),
		},
		taxPatterns: []*regexp.Regexp{
			regexp.MustCompile(`(?i)(vat|tax|thu[eế])\s*:?\s*(\d+%?)?\s*:?\s*([\d.,]+)`),
			regexp.MustCompile(`(?i)(VAT|Thu[eế])\s*\(?(\d+)%?\)?\s*:?\s*([\d.,]+)`),
		},
		servicePatterns: []*regexp.Regexp{
			regexp.MustCompile(`(?i)(service|ph[íi]\s*ph[uụ]c\s*v[uụ]|ph[íi]\s*d[iị]ch\s*v[uụ])\s*:?\s*(\d+%?)?\s*:?\s*([\d.,]+)`),
			regexp.MustCompile(`(?i)(Service\s*charge|Ph[íi]\s*PV)\s*:?\s*([\d.,]+)`),
		},
		discountPatterns: []*regexp.Regexp{
			regexp.MustCompile(`(?i)(discount|gi[aả]m\s*gi[aá]|khuy[eế]n\s*m[aã]i|ưu\s*đãi)\s*:?\s*-?\s*([\d.,]+)`),
			regexp.MustCompile(`(?i)(CK|Giảm)\s*:?\s*-?\s*([\d.,]+)`),
		},
		numberPattern: regexp.MustCompile(`[\d.,]+`),
	}
}

// ParseResult represents the full parsed receipt
type ParseResult struct {
	Items       []models.ParsedItem
	Total       float64
	Tax         float64
	ServiceFee  float64
	Discount    float64
	Confidence  float64
}

// Parse parses raw OCR text into structured receipt data
func (p *ReceiptParser) Parse(rawText string) *ParseResult {
	result := &ParseResult{}
	lines := strings.Split(rawText, "\n")

	// Track which lines are used for items vs metadata
	usedLines := make(map[int]bool)

	// First pass: extract total, tax, service fee, discount
	for i, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}

		// Check total
		if result.Total == 0 {
			for _, pattern := range p.totalPatterns {
				if matches := pattern.FindStringSubmatch(line); len(matches) > 0 {
					result.Total = parseAmount(matches[len(matches)-1])
					usedLines[i] = true
					break
				}
			}
		}

		// Check tax
		if result.Tax == 0 {
			for _, pattern := range p.taxPatterns {
				if matches := pattern.FindStringSubmatch(line); len(matches) > 0 {
					result.Tax = parseAmount(matches[len(matches)-1])
					usedLines[i] = true
					break
				}
			}
		}

		// Check service fee
		if result.ServiceFee == 0 {
			for _, pattern := range p.servicePatterns {
				if matches := pattern.FindStringSubmatch(line); len(matches) > 0 {
					result.ServiceFee = parseAmount(matches[len(matches)-1])
					usedLines[i] = true
					break
				}
			}
		}

		// Check discount
		if result.Discount == 0 {
			for _, pattern := range p.discountPatterns {
				if matches := pattern.FindStringSubmatch(line); len(matches) > 0 {
					result.Discount = parseAmount(matches[len(matches)-1])
					usedLines[i] = true
					break
				}
			}
		}
	}

	// Second pass: extract items from remaining lines
	for i, line := range lines {
		if usedLines[i] {
			continue
		}
		line = strings.TrimSpace(line)
		if line == "" || len(line) < 3 {
			continue
		}

		// Skip header/footer lines
		if isHeaderOrFooter(line) {
			continue
		}

		item := p.parseItemLine(line)
		if item != nil {
			result.Items = append(result.Items, *item)
		}
	}

	// Calculate confidence score
	result.Confidence = p.calculateConfidence(result)

	return result
}

// parseItemLine attempts to parse a single line as a receipt item
func (p *ReceiptParser) parseItemLine(line string) *models.ParsedItem {
	// Try each item pattern
	for _, pattern := range p.itemPatterns {
		matches := pattern.FindStringSubmatch(line)
		if len(matches) < 3 {
			continue
		}

		item := &models.ParsedItem{}

		switch len(matches) {
		case 5: // name, qty, unit_price, total
			item.Name = cleanItemName(matches[1])
			item.Quantity = parseInt(matches[2])
			item.UnitPrice = parseAmount(matches[3])
			item.TotalPrice = parseAmount(matches[4])
		case 4: // name x qty, total  OR  name, qty, price
			item.Name = cleanItemName(matches[1])
			qty := parseInt(matches[2])
			amount := parseAmount(matches[3])
			if qty > 0 && amount > 0 {
				item.Quantity = qty
				item.UnitPrice = amount / float64(qty)
				item.TotalPrice = amount
			}
		case 3: // name, amount
			item.Name = cleanItemName(matches[1])
			item.Quantity = 1
			item.TotalPrice = parseAmount(matches[2])
			item.UnitPrice = item.TotalPrice
		}

		if item.Name != "" && item.TotalPrice > 0 {
			if item.Quantity == 0 {
				item.Quantity = 1
			}
			if item.UnitPrice == 0 {
				item.UnitPrice = item.TotalPrice / float64(item.Quantity)
			}
			item.Confidence = calculateItemConfidence(item)
			return item
		}
	}

	return nil
}

// calculateConfidence calculates overall confidence of the parsed receipt
func (p *ReceiptParser) calculateConfidence(result *ParseResult) float64 {
	if len(result.Items) == 0 {
		return 0
	}

	// Factor 1: Number of items parsed (more items = higher confidence)
	itemScore := math.Min(float64(len(result.Items))/3.0, 1.0)

	// Factor 2: Total matches sum of items
	itemSum := 0.0
	for _, item := range result.Items {
		itemSum += item.TotalPrice
	}

	totalScore := 0.0
	if result.Total > 0 {
		// Check if items sum approximately matches total (with tax/service)
		expectedTotal := itemSum + result.Tax + result.ServiceFee - result.Discount
		ratio := expectedTotal / result.Total
		if ratio > 0.8 && ratio < 1.2 {
			totalScore = 1.0 - math.Abs(1.0-ratio)
		}
	} else {
		totalScore = 0.3 // No total found, low confidence
	}

	// Factor 3: Average item confidence
	avgItemConf := 0.0
	for _, item := range result.Items {
		avgItemConf += item.Confidence
	}
	avgItemConf /= float64(len(result.Items))

	// Weighted average
	confidence := itemScore*0.3 + totalScore*0.4 + avgItemConf*0.3
	return math.Round(confidence*100) / 100
}

// calculateItemConfidence calculates confidence for a single parsed item
func calculateItemConfidence(item *models.ParsedItem) float64 {
	confidence := 0.5

	// Has name
	if len(item.Name) > 1 {
		confidence += 0.15
	}

	// Has valid quantity
	if item.Quantity > 0 && item.Quantity < 100 {
		confidence += 0.1
	}

	// Has unit price matching total
	if item.Quantity > 0 && item.UnitPrice > 0 {
		expected := item.UnitPrice * float64(item.Quantity)
		if math.Abs(expected-item.TotalPrice) < 1 {
			confidence += 0.15
		}
	}

	// Price is reasonable (not too small or too large for VND)
	if item.TotalPrice >= 1000 && item.TotalPrice <= 50000000 {
		confidence += 0.1
	}

	return math.Min(confidence, 1.0)
}

// parseAmount parses a number string (handles Vietnamese format: 100.000 or 100,000)
func parseAmount(s string) float64 {
	s = strings.TrimSpace(s)

	// Remove spaces
	s = strings.ReplaceAll(s, " ", "")

	// Handle Vietnamese format: dots as thousand separators, commas as decimal
	// If has both dots and commas, determine which is which
	hasDot := strings.Contains(s, ".")
	hasComma := strings.Contains(s, ",")

	if hasDot && hasComma {
		// Both present: determine format
		lastDot := strings.LastIndex(s, ".")
		lastComma := strings.LastIndex(s, ",")
		if lastDot > lastComma {
			// Dot is decimal separator (e.g., 1,000.50)
			s = strings.ReplaceAll(s, ",", "")
		} else {
			// Comma is decimal separator (e.g., 1.000,50)
			s = strings.ReplaceAll(s, ".", "")
			s = strings.ReplaceAll(s, ",", ".")
		}
	} else if hasDot {
		// Only dots: check if it's a thousand separator
		parts := strings.Split(s, ".")
		if len(parts) > 1 && len(parts[len(parts)-1]) == 3 {
			// Thousand separator (e.g., 100.000)
			s = strings.ReplaceAll(s, ".", "")
		}
		// Otherwise it's a decimal (e.g., 100.50)
	} else if hasComma {
		// Only commas: check if it's a thousand separator
		parts := strings.Split(s, ",")
		if len(parts) > 1 && len(parts[len(parts)-1]) == 3 {
			// Thousand separator (e.g., 100,000)
			s = strings.ReplaceAll(s, ",", "")
		} else {
			// Decimal separator
			s = strings.ReplaceAll(s, ",", ".")
		}
	}

	val, err := strconv.ParseFloat(s, 64)
	if err != nil {
		return 0
	}
	return val
}

// parseInt parses an integer string
func parseInt(s string) int {
	s = strings.TrimSpace(s)
	val, err := strconv.Atoi(s)
	if err != nil {
		return 0
	}
	return val
}

// cleanItemName cleans and normalizes an item name
func cleanItemName(name string) string {
	name = strings.TrimSpace(name)
	// Remove leading numbers/dots/dashes
	name = regexp.MustCompile(`^[\d.)\-\s]+`).ReplaceAllString(name, "")
	// Remove trailing dots
	name = strings.TrimRight(name, ".")
	// Remove excessive whitespace
	name = regexp.MustCompile(`\s+`).ReplaceAllString(name, " ")
	return strings.TrimSpace(name)
}

// isHeaderOrFooter checks if a line is likely a header or footer (not an item)
func isHeaderOrFooter(line string) bool {
	lowerLine := strings.ToLower(line)
	skipKeywords := []string{
		"hóa đơn", "invoice", "receipt", "bill",
		"nhà hàng", "restaurant", "cafe", "quán",
		"địa chỉ", "address", "đt:", "tel:", "phone",
		"mã hóa đơn", "invoice no", "table", "bàn",
		"ngày", "date", "time", "giờ",
		"nhân viên", "staff", "cashier", "thu ngân",
		"cảm ơn", "thank you", "thanks",
		"wifi", "password", "mật khẩu",
		"---", "===", "***",
		"stk", "số tài khoản", "bank",
	}

	for _, keyword := range skipKeywords {
		if strings.Contains(lowerLine, keyword) {
			return true
		}
	}

	// Skip lines that are only numbers or special characters
	if regexp.MustCompile(`^[\d\s\-=*#.]+$`).MatchString(line) {
		return true
	}

	return false
}
