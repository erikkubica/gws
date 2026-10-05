package drive

import (
	"fmt"
	"io"
	"strings"
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
