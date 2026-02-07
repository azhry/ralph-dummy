package services

import (
	"bytes"
	"context"
	"encoding/hex"
	"fmt"
	"io"
	"mime/multipart"
	"path/filepath"
	"strings"
)

// FileValidator validates uploaded files
type FileValidator interface {
	Validate(ctx context.Context, file io.Reader, header *multipart.FileHeader) (*ValidationResult, error)
}

// ValidationResult contains the result of file validation
type ValidationResult struct {
	MimeType  string
	Extension string
	IsValid   bool
}

type fileValidator struct {
	allowedTypes []string
	maxSize      int64
	magicNumbers map[string][]byte
}

// NewFileValidator creates a new file validator
func NewFileValidator(allowedTypes []string, maxSize int64) FileValidator {
	return &fileValidator{
		allowedTypes: allowedTypes,
		maxSize:      maxSize,
		magicNumbers: map[string][]byte{
			"image/jpeg": {0xFF, 0xD8, 0xFF},
			"image/png":  {0x89, 0x50, 0x4E, 0x47, 0x0D, 0x0A, 0x1A, 0x0A},
			"image/webp": {0x52, 0x49, 0x46, 0x46},
		},
	}
}

// Validate validates a file based on its content and metadata
func (v *fileValidator) Validate(ctx context.Context, file io.Reader, header *multipart.FileHeader) (*ValidationResult, error) {
	// Check file extension
	ext := strings.ToLower(filepath.Ext(header.Filename))
	if ext == "" {
		return nil, fmt.Errorf("file must have an extension")
	}
	// Remove the leading dot from extension
	ext = strings.TrimPrefix(ext, ".")

	// Map extension to MIME type
	mimeType := v.extensionToMimeType(ext)
	if mimeType == "" {
		return nil, fmt.Errorf("unsupported file extension: %s", ext)
	}

	// Validate MIME type is allowed
	if !v.isAllowedType(mimeType) {
		return nil, fmt.Errorf("file type not allowed: %s", mimeType)
	}

	// Read first 512 bytes for magic number validation
	buf := make([]byte, 512)
	n, err := io.ReadFull(file, buf)
	if err != nil && err != io.EOF && err != io.ErrUnexpectedEOF {
		return nil, fmt.Errorf("failed to read file: %w", err)
	}
	buf = buf[:n]

	// Validate magic number
	if !v.validateMagicNumber(buf, mimeType) {
		return nil, fmt.Errorf("file content does not match extension: invalid magic number")
	}

	// Additional validation for WebP (check WEBP signature after RIFF)
	if mimeType == "image/webp" && len(buf) >= 12 {
		if !bytes.Equal(buf[8:12], []byte("WEBP")) {
			return nil, fmt.Errorf("invalid WebP file format")
		}
	}

	return &ValidationResult{
		MimeType:  mimeType,
		Extension: ext,
		IsValid:   true,
	}, nil
}

// extensionToMimeType maps file extensions to MIME types
func (v *fileValidator) extensionToMimeType(ext string) string {
	switch ext {
	case "jpg", "jpeg":
		return "image/jpeg"
	case "png":
		return "image/png"
	case "webp":
		return "image/webp"
	default:
		return ""
	}
}

// isAllowedType checks if a MIME type is in the allowed list
func (v *fileValidator) isAllowedType(mimeType string) bool {
	for _, allowed := range v.allowedTypes {
		if allowed == mimeType {
			return true
		}
	}
	return false
}

// validateMagicNumber validates file signature
func (v *fileValidator) validateMagicNumber(data []byte, expectedMime string) bool {
	magic, ok := v.magicNumbers[expectedMime]
	if !ok {
		return false
	}

	if len(data) < len(magic) {
		return false
	}

	return bytes.Equal(data[:len(magic)], magic)
}

// MagicNumberInfo provides detailed magic number information for debugging
func (v *fileValidator) MagicNumberInfo(data []byte) string {
	if len(data) == 0 {
		return "empty file"
	}

	limit := 16
	if len(data) < limit {
		limit = len(data)
	}

	return hex.EncodeToString(data[:limit])
}
