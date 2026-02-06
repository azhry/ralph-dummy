package services

import (
	"bytes"
	"context"
	"fmt"
	"image"
	"image/jpeg"
	"image/png"
	"io"
	"time"

	"github.com/chai2010/webp"
	"github.com/disintegration/imaging"
	"github.com/rwcarreira/goexif/exif"
)

// ImageProcessor processes images and generates thumbnails
type ImageProcessor interface {
	Process(ctx context.Context, reader io.Reader, mimeType string) (*ProcessedImage, error)
	GenerateThumbnail(data []byte, width, height int, format string) ([]byte, error)
	ExtractEXIF(data []byte) (map[string]interface{}, error)
	ConvertToWebP(data []byte, quality float32) ([]byte, error)
}

// ProcessedImage contains processed image data and metadata
type ProcessedImage struct {
	OriginalData []byte
	Thumbnails   map[string][]byte
	Metadata     *ImageMetadata
}

// ImageMetadata contains image information
type ImageMetadata struct {
	Width  int
	Height int
	Format string
	EXIF   map[string]interface{}
}

// ThumbnailSize defines thumbnail dimensions
type ThumbnailSize struct {
	Name   string
	Width  int
	Height int
}

type imageProcessor struct {
	thumbnailSizes []ThumbnailSize
	enableWebP     bool
}

// NewImageProcessor creates a new image processor
func NewImageProcessor(sizes []ThumbnailSize, enableWebP bool) ImageProcessor {
	return &imageProcessor{
		thumbnailSizes: sizes,
		enableWebP:     enableWebP,
	}
}

// Process processes an image and generates thumbnails
func (p *imageProcessor) Process(ctx context.Context, reader io.Reader, mimeType string) (*ProcessedImage, error) {
	// Read all data
	data, err := io.ReadAll(reader)
	if err != nil {
		return nil, fmt.Errorf("failed to read image data: %w", err)
	}

	// Decode image to get dimensions
	img, format, err := image.Decode(bytes.NewReader(data))
	if err != nil {
		return nil, fmt.Errorf("failed to decode image: %w", err)
	}

	bounds := img.Bounds()
	metadata := &ImageMetadata{
		Width:  bounds.Dx(),
		Height: bounds.Dy(),
		Format: format,
	}

	// Extract EXIF data
	exifData, err := p.ExtractEXIF(data)
	if err == nil {
		metadata.EXIF = exifData
	}

	// Generate thumbnails
	thumbnails := make(map[string][]byte)
	for _, size := range p.thumbnailSizes {
		thumb, err := p.GenerateThumbnail(data, size.Width, size.Height, format)
		if err != nil {
			continue // Log error but continue with other sizes
		}
		thumbnails[size.Name] = thumb
	}

	// Optionally convert to WebP
	if p.enableWebP && format != "webp" {
		webpData, err := p.ConvertToWebP(data, 85.0)
		if err == nil {
			// Use WebP as primary format if conversion succeeds and is smaller
			if len(webpData) < len(data) {
				data = webpData
				metadata.Format = "webp"
			}
		}
	}

	return &ProcessedImage{
		OriginalData: data,
		Thumbnails:   thumbnails,
		Metadata:     metadata,
	}, nil
}

// GenerateThumbnail generates a thumbnail with specified dimensions
func (p *imageProcessor) GenerateThumbnail(data []byte, width, height int, format string) ([]byte, error) {
	// Decode image
	img, _, err := image.Decode(bytes.NewReader(data))
	if err != nil {
		return nil, fmt.Errorf("failed to decode image: %w", err)
	}

	// Resize using Lanczos resampling (high quality)
	thumb := imaging.Thumbnail(img, width, height, imaging.Lanczos)

	// Encode based on format
	buf := new(bytes.Buffer)
	switch format {
	case "jpeg", "jpg":
		err = jpeg.Encode(buf, thumb, &jpeg.Options{Quality: 85})
	case "png":
		err = png.Encode(buf, thumb)
	case "webp":
		err = webp.Encode(buf, thumb, &webp.Options{Quality: 85})
	default:
		err = jpeg.Encode(buf, thumb, &jpeg.Options{Quality: 85})
	}

	if err != nil {
		return nil, fmt.Errorf("failed to encode thumbnail: %w", err)
	}

	return buf.Bytes(), nil
}

// ExtractEXIF extracts EXIF data from an image
func (p *imageProcessor) ExtractEXIF(data []byte) (map[string]interface{}, error) {
	exifData, err := exif.Decode(bytes.NewReader(data))
	if err != nil {
		return nil, fmt.Errorf("no EXIF data: %w", err)
	}

	result := make(map[string]interface{})

	// Extract common fields
	if dateTime, err := exifData.DateTime(); err == nil {
		result["dateTime"] = dateTime.Format("2006-01-02 15:04:05")
	}

	if lat, long, err := exifData.LatLong(); err == nil {
		result["gpsLatitude"] = lat
		result["gpsLongitude"] = long
	}

	if make, err := exifData.Get(exif.Make); err == nil {
		result["make"] = make.String()
	}

	if model, err := exifData.Get(exif.Model); err == nil {
		result["model"] = model.String()
	}

	if orientation, err := exifData.Get(exif.Orientation); err == nil {
		result["orientation"] = orientation.String()
	}

	return result, nil
}

// ConvertToWebP converts an image to WebP format
func (p *imageProcessor) ConvertToWebP(data []byte, quality float32) ([]byte, error) {
	img, _, err := image.Decode(bytes.NewReader(data))
	if err != nil {
		return nil, fmt.Errorf("failed to decode image: %w", err)
	}

	buf := new(bytes.Buffer)
	if err := webp.Encode(buf, img, &webp.Options{Quality: quality}); err != nil {
		return nil, fmt.Errorf("failed to encode WebP: %w", err)
	}

	return buf.Bytes(), nil
}

// GetImageDimensions returns the dimensions of an image without fully decoding it
func GetImageDimensions(data []byte) (width, height int, format string, err error) {
	config, format, err := image.DecodeConfig(bytes.NewReader(data))
	if err != nil {
		return 0, 0, "", fmt.Errorf("failed to decode image config: %w", err)
	}
	return config.Width, config.Height, format, nil
}