package utils

import (
	"bufio"
	"crypto/md5"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"hash"
	"io"
	"mime"
	"os"
	"path/filepath"
	"strings"
	"time"
)

// File utilities - enhanced file manipulation functions

// FileInfo represents enhanced file information
type FileInfo struct {
	Path        string    `json:"path"`
	Name        string    `json:"name"`
	BaseName    string    `json:"baseName"`
	Extension   string    `json:"extension"`
	Size        int64     `json:"size"`
	SizeHuman   string    `json:"sizeHuman"`
	MimeType    string    `json:"mimeType"`
	IsDir       bool      `json:"isDir"`
	IsHidden    bool      `json:"isHidden"`
	Permissions string    `json:"permissions"`
	Owner       string    `json:"owner"`
	Group       string    `json:"group"`
	CreatedAt   time.Time `json:"createdAt"`
	ModifiedAt  time.Time `json:"modifiedAt"`
	AccessedAt  time.Time `json:"accessedAt"`
	MD5Hash     string    `json:"md5Hash"`
	SHA256Hash  string    `json:"sha256Hash"`
}

// GetDetailedFileInfo returns detailed file information
func GetDetailedFileInfo(filePath string) (*FileInfo, error) {
	info, err := os.Stat(filePath)
	if err != nil {
		return nil, err
	}

	fileInfo := &FileInfo{
		Path:       filePath,
		Name:       filepath.Base(filePath),
		BaseName:   GetFileNameWithoutExtension(filePath),
		Extension:  GetFileExtension(filePath),
		Size:       info.Size(),
		SizeHuman:  FormatBytes(info.Size()),
		IsDir:      info.IsDir(),
		ModifiedAt: info.ModTime(),
	}

	// Get MIME type
	if !fileInfo.IsDir {
		fileInfo.MimeType = GetMimeType(filePath)
	}

	// Get file hashes
	if !fileInfo.IsDir {
		if md5Hash, err := CalculateFileHash(filePath, "md5"); err == nil {
			fileInfo.MD5Hash = md5Hash
		}
		if sha256Hash, err := CalculateFileHash(filePath, "sha256"); err == nil {
			fileInfo.SHA256Hash = sha256Hash
		}
	}

	return fileInfo, nil
}

// GetMimeType returns the MIME type of a file
func GetMimeType(filePath string) string {
	ext := strings.ToLower(filepath.Ext(filePath))
	mimeType := mime.TypeByExtension(ext)
	if mimeType == "" {
		// Try to detect from file content
		if file, err := os.Open(filePath); err == nil {
			defer file.Close()

			buffer := make([]byte, 512)
			if _, err := file.Read(buffer); err == nil {
				// Simple MIME type detection based on file signatures
				mimeType = detectMimeTypeBySignature(buffer)
			}
		}
	}

	return mimeType
}

// detectMimeTypeBySignature detects MIME type based on file signature
func detectMimeTypeBySignature(buffer []byte) string {
	if len(buffer) < 4 {
		return "application/octet-stream"
	}

	// Check for common file signatures
	if buffer[0] == 0xFF && buffer[1] == 0xD8 && buffer[2] == 0xFF {
		return "image/jpeg"
	}
	if buffer[0] == 0x89 && buffer[1] == 0x50 && buffer[2] == 0x4E && buffer[3] == 0x47 {
		return "image/png"
	}
	if buffer[0] == 0x47 && buffer[1] == 0x49 && buffer[2] == 0x46 && buffer[3] == 0x38 {
		return "image/gif"
	}
	if buffer[0] == 0x25 && buffer[1] == 0x50 && buffer[2] == 0x44 && buffer[3] == 0x46 {
		return "application/pdf"
	}
	if buffer[0] == 0x50 && buffer[1] == 0x4B && buffer[2] == 0x03 && buffer[3] == 0x04 {
		return "application/zip"
	}

	// Check for text files
	for i := 0; i < len(buffer); i++ {
		if buffer[i] == 0 {
			return "application/octet-stream"
		}
	}

	return "text/plain"
}

// CalculateFileHash calculates the hash of a file
func CalculateFileHash(filePath string, algorithm string) (string, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return "", err
	}
	defer file.Close()

	var h hash.Hash
	switch algorithm {
	case "md5":
		h = md5.New()
	case "sha256":
		h = sha256.New()
	default:
		return "", fmt.Errorf("unsupported hash algorithm: %s", algorithm)
	}

	if _, err := io.Copy(h, file); err != nil {
		return "", err
	}

	return hex.EncodeToString(h.Sum(nil)), nil
}

// FormatBytes formats bytes to human readable string
func FormatBytes(bytes int64) string {
	const unit = 1024
	if bytes < unit {
		return fmt.Sprintf("%d B", bytes)
	}

	div, exp := int64(unit), 0
	for n := bytes / unit; n >= unit; n /= unit {
		div *= unit
		exp++
	}

	return fmt.Sprintf("%.1f %cB", float64(bytes)/float64(div), "KMGTPE"[exp])
}

// CopyFile copies a file from source to destination
func CopyFile(src, dst string) error {
	// Check if source file exists
	srcInfo, err := os.Stat(src)
	if err != nil {
		return err
	}

	if !srcInfo.Mode().IsRegular() {
		return fmt.Errorf("%s is not a regular file", src)
	}

	// Create destination directory if it doesn't exist
	dstDir := filepath.Dir(dst)
	if err := os.MkdirAll(dstDir, 0755); err != nil {
		return err
	}

	// Open source file
	source, err := os.Open(src)
	if err != nil {
		return err
	}
	defer source.Close()

	// Create destination file
	destination, err := os.Create(dst)
	if err != nil {
		return err
	}
	defer destination.Close()

	// Copy file content
	_, err = io.Copy(destination, source)
	return err
}

// MoveFile moves a file from source to destination
func MoveFile(src, dst string) error {
	// Try to rename first (fastest)
	if err := os.Rename(src, dst); err == nil {
		return nil
	}

	// If rename fails, copy and then remove
	if err := CopyFile(src, dst); err != nil {
		return err
	}

	return os.Remove(src)
}

// SafeDeleteFile safely deletes a file with backup
func SafeDeleteFile(filePath string, backupDir string) error {
	if !FileExists(filePath) {
		return nil
	}

	// Create backup directory if it doesn't exist
	if err := os.MkdirAll(backupDir, 0755); err != nil {
		return err
	}

	// Generate backup filename
	backupName := filepath.Base(filePath) + "." + time.Now().Format("20060102_150405") + ".bak"
	backupPath := filepath.Join(backupDir, backupName)

	// Copy file to backup
	if err := CopyFile(filePath, backupPath); err != nil {
		return err
	}

	// Delete original file
	return os.Remove(filePath)
}

// ReadFileLines reads a file and returns its lines
func ReadFileLines(filePath string) ([]string, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	var lines []string
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		lines = append(lines, scanner.Text())
	}

	if err := scanner.Err(); err != nil {
		return nil, err
	}

	return lines, nil
}

// WriteFileLines writes lines to a file
func WriteFileLines(filePath string, lines []string) error {
	file, err := os.Create(filePath)
	if err != nil {
		return err
	}
	defer file.Close()

	writer := bufio.NewWriter(file)
	for _, line := range lines {
		if _, err := writer.WriteString(line + "\n"); err != nil {
			return err
		}
	}

	return writer.Flush()
}

// AppendToFile appends content to a file
func AppendToFile(filePath string, content string) error {
	file, err := os.OpenFile(filePath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return err
	}
	defer file.Close()

	_, err = file.WriteString(content)
	return err
}

// FindFilesByPattern finds files matching a pattern recursively
func FindFilesByPattern(rootDir, pattern string) ([]string, error) {
	var matches []string

	err := filepath.Walk(rootDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		if !info.IsDir() {
			matched, err := filepath.Match(pattern, filepath.Base(path))
			if err != nil {
				return err
			}
			if matched {
				matches = append(matches, path)
			}
		}

		return nil
	})

	return matches, err
}

// FindFilesByExtension finds files with specific extension recursively
func FindFilesByExtension(rootDir, extension string) ([]string, error) {
	if !strings.HasPrefix(extension, ".") {
		extension = "." + extension
	}

	var files []string

	err := filepath.Walk(rootDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		if !info.IsDir() && strings.EqualFold(filepath.Ext(path), extension) {
			files = append(files, path)
		}

		return nil
	})

	return files, err
}

// GetFileSize returns the size of a file in bytes
func GetFileSize(filePath string) (int64, error) {
	info, err := os.Stat(filePath)
	if err != nil {
		return 0, err
	}
	return info.Size(), nil
}

// GetDirectorySize returns the total size of a directory
func GetDirectorySize(dirPath string) (int64, error) {
	var totalSize int64

	err := filepath.Walk(dirPath, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		if !info.IsDir() {
			totalSize += info.Size()
		}

		return nil
	})

	return totalSize, err
}

// CreateTempFile creates a temporary file with content
func CreateTempFile(content string, pattern string) (string, error) {
	tempFile, err := os.CreateTemp("", pattern)
	if err != nil {
		return "", err
	}
	defer tempFile.Close()

	if _, err := tempFile.WriteString(content); err != nil {
		return "", err
	}

	return tempFile.Name(), nil
}

// IsTextFile checks if a file is a text file
func IsTextFile(filePath string) bool {
	file, err := os.Open(filePath)
	if err != nil {
		return false
	}
	defer file.Close()

	buffer := make([]byte, 512)
	n, err := file.Read(buffer)
	if err != nil && err != io.EOF {
		return false
	}

	// Check if the content contains null bytes
	for i := 0; i < n; i++ {
		if buffer[i] == 0 {
			return false
		}
	}

	return true
}

// GetFileEncoding detects file encoding
func GetFileEncoding(filePath string) (string, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return "", err
	}
	defer file.Close()

	buffer := make([]byte, 1024)
	n, err := file.Read(buffer)
	if err != nil && err != io.EOF {
		return "", err
	}

	// Simple encoding detection
	if n >= 3 && buffer[0] == 0xEF && buffer[1] == 0xBB && buffer[2] == 0xBF {
		return "UTF-8", nil
	}
	if n >= 2 && buffer[0] == 0xFF && buffer[1] == 0xFE {
		return "UTF-16LE", nil
	}
	if n >= 2 && buffer[0] == 0xFE && buffer[1] == 0xFF {
		return "UTF-16BE", nil
	}

	// Try to detect UTF-8 by checking if it's valid UTF-8
	if IsValidUTF8(buffer[:n]) {
		return "UTF-8", nil
	}

	return "unknown", nil
}

// IsValidUTF8 checks if bytes are valid UTF-8
func IsValidUTF8(b []byte) bool {
	for i := 0; i < len(b); {
		if b[i] < 0x80 {
			i++
		} else if b[i] < 0xC0 {
			return false
		} else if b[i] < 0xE0 {
			if i+1 >= len(b) || (b[i+1]&0xC0) != 0x80 {
				return false
			}
			i += 2
		} else if b[i] < 0xF0 {
			if i+2 >= len(b) || (b[i+1]&0xC0) != 0x80 || (b[i+2]&0xC0) != 0x80 {
				return false
			}
			i += 3
		} else {
			return false
		}
	}
	return true
}

// FileExists checks if a file exists
func FileExists(filePath string) bool {
	_, err := os.Stat(filePath)
	return !os.IsNotExist(err)
}

// IsDirectory checks if a path is a directory
func IsDirectory(path string) bool {
	info, err := os.Stat(path)
	if err != nil {
		return false
	}
	return info.IsDir()
}

// CreateDirectory creates a directory recursively
func CreateDirectory(dirPath string) error {
	return os.MkdirAll(dirPath, 0755)
}

// DeleteDirectory deletes a directory recursively
func DeleteDirectory(dirPath string) error {
	return os.RemoveAll(dirPath)
}

// CleanDirectory deletes all files in a directory but keeps the directory
func CleanDirectory(dirPath string) error {
	entries, err := os.ReadDir(dirPath)
	if err != nil {
		return err
	}

	for _, entry := range entries {
		entryPath := filepath.Join(dirPath, entry.Name())
		if err := os.RemoveAll(entryPath); err != nil {
			return err
		}
	}

	return nil
}

// ListFiles lists all files in a directory
func ListFiles(dirPath string) ([]string, error) {
	entries, err := os.ReadDir(dirPath)
	if err != nil {
		return nil, err
	}

	var files []string
	for _, entry := range entries {
		if !entry.IsDir() {
			files = append(files, entry.Name())
		}
	}

	return files, nil
}

// ListDirectories lists all subdirectories in a directory
func ListDirectories(dirPath string) ([]string, error) {
	entries, err := os.ReadDir(dirPath)
	if err != nil {
		return nil, err
	}

	var dirs []string
	for _, entry := range entries {
		if entry.IsDir() {
			dirs = append(dirs, entry.Name())
		}
	}

	return dirs, nil
}

// GetFileNameWithoutExtension gets filename without extension
func GetFileNameWithoutExtension(filePath string) string {
	filename := filepath.Base(filePath)
	return strings.TrimSuffix(filename, filepath.Ext(filename))
}

// GetFileExtension gets file extension
func GetFileExtension(filePath string) string {
	return strings.ToLower(filepath.Ext(filePath))
}

// ChangeFileExtension changes file extension
func ChangeFileExtension(filePath, newExtension string) string {
	if !strings.HasPrefix(newExtension, ".") {
		newExtension = "." + newExtension
	}
	return filepath.Dir(filePath) + string(filepath.Separator) + GetFileNameWithoutExtension(filePath) + newExtension
}
