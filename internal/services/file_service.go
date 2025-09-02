package services

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/Run-Panel/VerTree/internal/database"
	"gorm.io/gorm"
)

// FileService handles file operations for version management
type FileService struct {
	db       *gorm.DB
	cacheDir string
	baseURL  string
	maxSize  int64 // Maximum file size in bytes
}

// CachedFile represents a cached file with metadata
type CachedFile struct {
	LocalPath string `json:"local_path"`
	URL       string `json:"url"`
	Size      int64  `json:"size"`
	Checksum  string `json:"checksum"`
}

// FileCache represents cached file metadata in database
type FileCache struct {
	ID           uint      `json:"id" gorm:"primaryKey"`
	AppID        string    `json:"app_id" gorm:"size:32;not null;index"`
	Version      string    `json:"version" gorm:"size:100;not null"`
	OriginalURL  string    `json:"original_url" gorm:"size:500;not null"`
	LocalPath    string    `json:"local_path" gorm:"size:500;not null"`
	FileSize     int64     `json:"file_size" gorm:"not null"`
	FileChecksum string    `json:"file_checksum" gorm:"size:128;not null"`
	ContentType  string    `json:"content_type" gorm:"size:100"`
	DownloadedAt time.Time `json:"downloaded_at"`
	LastAccessed time.Time `json:"last_accessed"`
	AccessCount  int64     `json:"access_count" gorm:"default:0"`
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
}

// TableName returns the table name for FileCache model
func (FileCache) TableName() string {
	return "file_cache"
}

// NewFileService creates a new file service instance
func NewFileService() *FileService {
	// Default configuration - should be loaded from config
	cacheDir := filepath.Join("uploads", "cache")
	baseURL := "/api/v1/download"            // Base URL for serving cached files
	maxSize := int64(5 * 1024 * 1024 * 1024) // 5GB max file size

	// Create cache directory if it doesn't exist
	if err := os.MkdirAll(cacheDir, 0755); err != nil {
		fmt.Printf("Warning: failed to create cache directory: %v\n", err)
	}

	return &FileService{
		db:       database.DB,
		cacheDir: cacheDir,
		baseURL:  baseURL,
		maxSize:  maxSize,
	}
}

// CacheGitHubAsset downloads and caches a GitHub release asset
func (s *FileService) CacheGitHubAsset(downloadURL, appID, version string) (*CachedFile, error) {
	// Check if file is already cached
	var existingCache FileCache
	if err := s.db.Where("app_id = ? AND version = ? AND original_url = ?",
		appID, version, downloadURL).First(&existingCache).Error; err == nil {
		// Update access information
		existingCache.LastAccessed = time.Now()
		existingCache.AccessCount++
		s.db.Save(&existingCache)

		// Verify file still exists on disk
		if _, err := os.Stat(existingCache.LocalPath); err == nil {
			return &CachedFile{
				LocalPath: existingCache.LocalPath,
				URL:       s.buildFileURL(existingCache.ID),
				Size:      existingCache.FileSize,
				Checksum:  existingCache.FileChecksum,
			}, nil
		}
		// File missing from disk, re-download
		s.db.Delete(&existingCache)
	}

	// Download the file
	resp, err := http.Get(downloadURL)
	if err != nil {
		return nil, fmt.Errorf("failed to download file: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("failed to download file: HTTP %d", resp.StatusCode)
	}

	// Check file size
	contentLength := resp.Header.Get("Content-Length")
	if contentLength != "" {
		size, err := strconv.ParseInt(contentLength, 10, 64)
		if err == nil && size > s.maxSize {
			return nil, fmt.Errorf("file too large: %d bytes (max: %d bytes)", size, s.maxSize)
		}
	}

	// Generate local file path
	fileName := s.generateFileName(downloadURL, appID, version)
	localPath := filepath.Join(s.cacheDir, appID, fileName)

	// Create directory structure
	if err := os.MkdirAll(filepath.Dir(localPath), 0755); err != nil {
		return nil, fmt.Errorf("failed to create cache directory: %w", err)
	}

	// Create temporary file first
	tempPath := localPath + ".tmp"
	file, err := os.Create(tempPath)
	if err != nil {
		return nil, fmt.Errorf("failed to create temporary file: %w", err)
	}
	defer os.Remove(tempPath) // Clean up temp file on error

	// Download with size limit and checksum calculation
	hasher := sha256.New()
	limitedReader := io.LimitReader(resp.Body, s.maxSize+1)
	multiWriter := io.MultiWriter(file, hasher)

	written, err := io.Copy(multiWriter, limitedReader)
	file.Close()

	if err != nil {
		return nil, fmt.Errorf("failed to download file: %w", err)
	}

	if written > s.maxSize {
		return nil, fmt.Errorf("file too large: %d bytes (max: %d bytes)", written, s.maxSize)
	}

	// Calculate checksum
	checksum := hex.EncodeToString(hasher.Sum(nil))

	// Move temp file to final location
	if err := os.Rename(tempPath, localPath); err != nil {
		return nil, fmt.Errorf("failed to move file to final location: %w", err)
	}

	// Store cache metadata in database
	cache := &FileCache{
		AppID:        appID,
		Version:      version,
		OriginalURL:  downloadURL,
		LocalPath:    localPath,
		FileSize:     written,
		FileChecksum: checksum,
		ContentType:  resp.Header.Get("Content-Type"),
		DownloadedAt: time.Now(),
		LastAccessed: time.Now(),
		AccessCount:  1,
	}

	if err := s.db.Create(cache).Error; err != nil {
		// Clean up file if database insert fails
		os.Remove(localPath)
		return nil, fmt.Errorf("failed to store cache metadata: %w", err)
	}

	return &CachedFile{
		LocalPath: localPath,
		URL:       s.buildFileURL(cache.ID),
		Size:      written,
		Checksum:  checksum,
	}, nil
}

// GetCachedFile retrieves cached file information by ID
func (s *FileService) GetCachedFile(id uint) (*FileCache, error) {
	var cache FileCache
	if err := s.db.First(&cache, id).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, fmt.Errorf("cached file not found")
		}
		return nil, fmt.Errorf("failed to get cached file: %w", err)
	}

	// Update access information
	cache.LastAccessed = time.Now()
	cache.AccessCount++
	s.db.Save(&cache)

	return &cache, nil
}

// ServeCachedFile serves a cached file for download
func (s *FileService) ServeCachedFile(id uint, w http.ResponseWriter, r *http.Request) error {
	cache, err := s.GetCachedFile(id)
	if err != nil {
		return err
	}

	// Check if file exists on disk
	fileInfo, err := os.Stat(cache.LocalPath)
	if err != nil {
		if os.IsNotExist(err) {
			// File was deleted, remove from cache
			s.db.Delete(cache)
			return fmt.Errorf("cached file no longer exists")
		}
		return fmt.Errorf("failed to access cached file: %w", err)
	}

	// Open file for reading
	file, err := os.Open(cache.LocalPath)
	if err != nil {
		return fmt.Errorf("failed to open cached file: %w", err)
	}
	defer file.Close()

	// Set headers
	w.Header().Set("Content-Type", cache.ContentType)
	w.Header().Set("Content-Length", strconv.FormatInt(cache.FileSize, 10))
	w.Header().Set("Content-Disposition", fmt.Sprintf("attachment; filename=\"%s\"",
		filepath.Base(cache.LocalPath)))
	w.Header().Set("Cache-Control", "public, max-age=86400") // 24 hour cache
	w.Header().Set("ETag", fmt.Sprintf("\"%s\"", cache.FileChecksum))

	// Check for conditional requests
	if match := r.Header.Get("If-None-Match"); match != "" {
		if strings.Contains(match, cache.FileChecksum) {
			w.WriteHeader(http.StatusNotModified)
			return nil
		}
	}

	// Set last modified header
	w.Header().Set("Last-Modified", fileInfo.ModTime().UTC().Format(http.TimeFormat))

	// Support range requests for large files
	if r.Header.Get("Range") != "" {
		s.serveFileRange(file, cache.FileSize, w, r)
		return nil
	}

	// Serve entire file
	w.WriteHeader(http.StatusOK)
	_, err = io.Copy(w, file)
	return err
}

// CleanupExpiredCache removes old cached files
func (s *FileService) CleanupExpiredCache(maxAge time.Duration) error {
	cutoff := time.Now().Add(-maxAge)

	var expiredCaches []FileCache
	if err := s.db.Where("last_accessed < ?", cutoff).Find(&expiredCaches).Error; err != nil {
		return fmt.Errorf("failed to find expired cache entries: %w", err)
	}

	for _, cache := range expiredCaches {
		// Remove file from disk
		if err := os.Remove(cache.LocalPath); err != nil && !os.IsNotExist(err) {
			fmt.Printf("Warning: failed to remove cached file %s: %v\n", cache.LocalPath, err)
		}

		// Remove from database
		if err := s.db.Delete(&cache).Error; err != nil {
			fmt.Printf("Warning: failed to remove cache entry %d: %v\n", cache.ID, err)
		}
	}

	return nil
}

// GetCacheStats returns cache statistics
func (s *FileService) GetCacheStats() (map[string]interface{}, error) {
	var totalFiles int64
	var totalSize int64

	s.db.Model(&FileCache{}).Count(&totalFiles)
	s.db.Model(&FileCache{}).Select("COALESCE(SUM(file_size), 0)").Row().Scan(&totalSize)

	// Calculate directory size
	var diskUsage int64
	filepath.Walk(s.cacheDir, func(path string, info os.FileInfo, err error) error {
		if err == nil && !info.IsDir() {
			diskUsage += info.Size()
		}
		return nil
	})

	return map[string]interface{}{
		"total_files":     totalFiles,
		"total_size":      totalSize,
		"disk_usage":      diskUsage,
		"cache_directory": s.cacheDir,
		"max_file_size":   s.maxSize,
	}, nil
}

// Private helper methods

func (s *FileService) generateFileName(url, appID, version string) string {
	// Extract filename from URL
	urlParts := strings.Split(url, "/")
	originalName := urlParts[len(urlParts)-1]

	// Remove query parameters
	if idx := strings.Index(originalName, "?"); idx != -1 {
		originalName = originalName[:idx]
	}

	// If no extension or filename, generate one
	if originalName == "" || !strings.Contains(originalName, ".") {
		originalName = fmt.Sprintf("%s-%s.bin", appID, version)
	}

	// Add version prefix for uniqueness
	return fmt.Sprintf("%s_%s", version, originalName)
}

func (s *FileService) buildFileURL(cacheID uint) string {
	return fmt.Sprintf("%s/cached/%d", s.baseURL, cacheID)
}

func (s *FileService) serveFileRange(file *os.File, fileSize int64, w http.ResponseWriter, r *http.Request) {
	// Parse Range header
	ranges := r.Header.Get("Range")
	if !strings.HasPrefix(ranges, "bytes=") {
		w.WriteHeader(http.StatusRequestedRangeNotSatisfiable)
		return
	}

	// Simple range parsing (only support single range for now)
	rangeSpec := strings.TrimPrefix(ranges, "bytes=")
	parts := strings.Split(rangeSpec, "-")

	if len(parts) != 2 {
		w.WriteHeader(http.StatusRequestedRangeNotSatisfiable)
		return
	}

	var start, end int64
	var err error

	if parts[0] != "" {
		start, err = strconv.ParseInt(parts[0], 10, 64)
		if err != nil {
			w.WriteHeader(http.StatusRequestedRangeNotSatisfiable)
			return
		}
	}

	if parts[1] != "" {
		end, err = strconv.ParseInt(parts[1], 10, 64)
		if err != nil {
			w.WriteHeader(http.StatusRequestedRangeNotSatisfiable)
			return
		}
	} else {
		end = fileSize - 1
	}

	if start < 0 || end >= fileSize || start > end {
		w.WriteHeader(http.StatusRequestedRangeNotSatisfiable)
		return
	}

	// Seek to start position
	if _, err := file.Seek(start, 0); err != nil {
		http.Error(w, "Failed to seek file", http.StatusInternalServerError)
		return
	}

	// Set headers for partial content
	contentLength := end - start + 1
	w.Header().Set("Content-Range", fmt.Sprintf("bytes %d-%d/%d", start, end, fileSize))
	w.Header().Set("Content-Length", strconv.FormatInt(contentLength, 10))
	w.WriteHeader(http.StatusPartialContent)

	// Copy range to response
	io.CopyN(w, file, contentLength)
}
