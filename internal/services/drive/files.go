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
		Fields("files(id, name, mimeType, size)").OrderBy("modifiedTime desc")
	if query != "" {
		req = req.Q(fmt.Sprintf("name contains '%s' and trashed = false", query))
	} else {
		req = req.Q("trashed = false")
	}

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

// ReadFile exports or downloads a file content as UTF-8 string.
func (s *Service) ReadFile(fileID string) (string, error) {
	f, err := s.client.Files.Get(fileID).Fields("id, name, mimeType").Do()
	if err != nil {
		return "", fmt.Errorf("inspect file metadata: %w", err)
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

	b, err := io.ReadAll(res.Body)
	if err != nil {
		return "", fmt.Errorf("read stream: %w", err)
	}
	return string(b), nil
}

// exportFile exports a Google Workspace doc/sheet to plain text or csv.
func (s *Service) exportFile(fileID, exportMime string) (string, error) {
	res, err := s.client.Files.Export(fileID, exportMime).Download()
	if err != nil {
		return "", fmt.Errorf("export file %s: %w", fileID, err)
	}
	defer res.Body.Close()

	b, err := io.ReadAll(res.Body)
	if err != nil {
		return "", fmt.Errorf("read export stream: %w", err)
	}
	return string(b), nil
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

	out, err := os.Create(destPath)
	if err != nil {
		return fmt.Errorf("create destination file: %w", err)
	}
	defer out.Close()

	if _, err := io.Copy(out, body); err != nil {
		return fmt.Errorf("write stream to disk: %w", err)
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
