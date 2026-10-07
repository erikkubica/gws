package drive

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"google.golang.org/api/drive/v3"
)

// FileSummary represents minimal file metadata.
type FileSummary struct {
	ID       string `json:"id"`
	Name     string `json:"name"`
	MimeType string `json:"mimeType"`
	Size     int64  `json:"size,omitempty"`
}

// ListFiles searches and lists files from Google Drive.
func (s *Service) ListFiles(query string, max int64) ([]FileSummary, error) {
	if max <= 0 {
		max = 10
	}
	req := s.client.Files.List().PageSize(max).
		Fields("files(id, name, mimeType, size)").OrderBy("modifiedTime desc").
		Q(buildDriveQuery(query))

	res, err := req.Do()
	if err != nil {
		return nil, fmt.Errorf("list drive files: %w", err)
	}

	var files []FileSummary
	for _, f := range res.Files {
		files = append(files, FileSummary{
			ID:       f.Id,
			Name:     f.Name,
			MimeType: f.MimeType,
			Size:     f.Size,
		})
	}
	return files, nil
}

func buildDriveQuery(query string) string {
	clean := strings.TrimSpace(query)
	if clean == "" {
		return "trashed = false"
	}
	escaped := strings.ReplaceAll(clean, "'", "\\'")
	return fmt.Sprintf("name contains '%s' and trashed = false", escaped)
}

// MaxReadFileSize limits in-memory file reading to 10MB to avoid Out of Memory (OOM) crashes.
const MaxReadFileSize int64 = 10 * 1024 * 1024

func readLimitedStream(r io.Reader, maxBytes int64) (string, error) {
	limited := io.LimitReader(r, maxBytes+1)
	b, err := io.ReadAll(limited)
	if err != nil {
		return "", fmt.Errorf("read stream: %w", err)
	}
	if int64(len(b)) > maxBytes {
		return "", fmt.Errorf("file content exceeds maximum limit of %d bytes", maxBytes)
	}
	return string(b), nil
}

// ReadFile exports or downloads a file content as UTF-8 string.
func (s *Service) ReadFile(fileID string) (string, error) {
	f, err := s.client.Files.Get(fileID).Fields("id, name, mimeType, size").Do()
	if err != nil {
		return "", fmt.Errorf("inspect file metadata: %w", err)
	}

	if f.Size > MaxReadFileSize {
		return "", fmt.Errorf("file size (%d bytes) exceeds maximum limit of %d bytes", f.Size, MaxReadFileSize)
	}

	if strings.Contains(f.MimeType, "google-apps.document") {
		return s.exportFile(fileID, "text/plain")
	}
	if strings.Contains(f.MimeType, "google-apps.spreadsheet") {
		return s.exportFile(fileID, "text/csv")
	}

	res, err := s.client.Files.Get(fileID).Download()
	if err != nil {
		return "", fmt.Errorf("download file %s: %w", fileID, err)
	}
	defer res.Body.Close()

	return readLimitedStream(res.Body, MaxReadFileSize)
}

// exportFile exports a Google Workspace doc/sheet to plain text or csv.
func (s *Service) exportFile(fileID, exportMime string) (string, error) {
	res, err := s.client.Files.Export(fileID, exportMime).Download()
	if err != nil {
		return "", fmt.Errorf("export file %s: %w", fileID, err)
	}
	defer res.Body.Close()

	return readLimitedStream(res.Body, MaxReadFileSize)
}

// UploadFile uploads a local file to Google Drive.
func (s *Service) UploadFile(localPath, remoteName string) (*FileSummary, error) {
	file, err := os.Open(localPath)
	if err != nil {
		return nil, fmt.Errorf("open local file: %w", err)
	}
	defer file.Close()

	if remoteName == "" {
		remoteName = filepath.Base(localPath)
	}

	driveFile := &drive.File{Name: remoteName}
	res, err := s.client.Files.Create(driveFile).Media(file).Fields("id, name, mimeType, size").Do()
	if err != nil {
		return nil, fmt.Errorf("drive upload: %w", err)
	}
	return &FileSummary{ID: res.Id, Name: res.Name, MimeType: res.MimeType, Size: res.Size}, nil
}

// DeleteFile permanently deletes a file from Google Drive by its file ID.
func (s *Service) DeleteFile(fileID string) error {
	if err := s.client.Files.Delete(fileID).Do(); err != nil {
		return fmt.Errorf("delete drive file %s: %w", fileID, err)
	}
	return nil
}

// DownloadFile downloads a remote file to a local destination path.
func (s *Service) DownloadFile(fileID, destPath string) error {
	f, err := s.client.Files.Get(fileID).Fields("id, name, mimeType").Do()
	if err != nil {
		return fmt.Errorf("inspect file metadata: %w", err)
	}

	body, err := s.getDownloadStream(fileID, f.MimeType)
	if err != nil {
		return err
	}
	defer body.Close()

	return writeStreamToFile(destPath, body)
}

func writeStreamToFile(destPath string, body io.Reader) error {
	tmpPath := destPath + ".tmp"
	out, err := os.Create(tmpPath)
	if err != nil {
		return fmt.Errorf("create temp destination file: %w", err)
	}

	_, copyErr := io.Copy(out, body)
	closeErr := out.Close()
	if copyErr != nil {
		_ = os.Remove(tmpPath)
		return fmt.Errorf("write stream to disk: %w", copyErr)
	}
	if closeErr != nil {
		_ = os.Remove(tmpPath)
		return fmt.Errorf("close temp destination file: %w", closeErr)
	}

	if err := os.Rename(tmpPath, destPath); err != nil {
		_ = os.Remove(tmpPath)
		return fmt.Errorf("rename temporary file: %w", err)
	}
	return nil
}

func (s *Service) getDownloadStream(fileID, mimeType string) (io.ReadCloser, error) {
	if strings.Contains(mimeType, "google-apps.document") {
		res, err := s.client.Files.Export(fileID, "application/pdf").Download()
		if err != nil {
			return nil, fmt.Errorf("export doc as pdf: %w", err)
		}
		return res.Body, nil
	}
	if strings.Contains(mimeType, "google-apps.spreadsheet") {
		res, err := s.client.Files.Export(fileID, "text/csv").Download()
		if err != nil {
			return nil, fmt.Errorf("export sheet as csv: %w", err)
		}
		return res.Body, nil
	}
	res, err := s.client.Files.Get(fileID).Download()
	if err != nil {
		return nil, fmt.Errorf("download binary file: %w", err)
	}
	return res.Body, nil
}

// CreateEmptyFile creates a new empty file or doc in Google Drive.
func (s *Service) CreateEmptyFile(name, mimeType string) (*FileSummary, error) {
	if mimeType == "" {
		mimeType = "text/plain"
	}
	driveFile := &drive.File{Name: name, MimeType: mimeType}
	res, err := s.client.Files.Create(driveFile).Fields("id, name, mimeType").Do()
	if err != nil {
		return nil, fmt.Errorf("create file: %w", err)
	}
	return &FileSummary{ID: res.Id, Name: res.Name, MimeType: res.MimeType}, nil
}
