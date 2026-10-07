package chat

import (
	"fmt"
	"strings"
)

// SpaceInfo represents a Google Chat space or direct message room.
type SpaceInfo struct {
	Name        string `json:"name"`
	DisplayName string `json:"display_name"`
	Type        string `json:"type"`
	SpaceType   string `json:"space_type"`
}

// NormalizeSpaceName ensures the resource name has the required 'spaces/' prefix.
func NormalizeSpaceName(name string) string {
	clean := strings.TrimSpace(name)
	if clean == "" {
		return ""
	}
	if !strings.HasPrefix(clean, "spaces/") {
		return "spaces/" + clean
	}
	return clean
}

// ListSpaces returns all joined spaces and direct message rooms.
func (s *Service) ListSpaces(pageSize int64) ([]*SpaceInfo, error) {
	if pageSize <= 0 {
		pageSize = 20
	}
	call := s.client.Spaces.List().PageSize(pageSize)
	res, err := call.Do()
	if err != nil {
		return nil, fmt.Errorf("list spaces: %w", err)
	}

	var spaces []*SpaceInfo
	for _, sp := range res.Spaces {
		spaces = append(spaces, &SpaceInfo{
			Name:        sp.Name,
			DisplayName: sp.DisplayName,
			Type:        sp.Type,
			SpaceType:   sp.SpaceType,
		})
	}
	return spaces, nil
}

// GetSpace fetches metadata for a specific space.
func (s *Service) GetSpace(name string) (*SpaceInfo, error) {
	resName := NormalizeSpaceName(name)
	if resName == "" {
		return nil, fmt.Errorf("space name cannot be empty")
	}
	sp, err := s.client.Spaces.Get(resName).Do()
	if err != nil {
		return nil, fmt.Errorf("get space %s: %w", resName, err)
	}
	return &SpaceInfo{
		Name:        sp.Name,
		DisplayName: sp.DisplayName,
		Type:        sp.Type,
		SpaceType:   sp.SpaceType,
	}, nil
}
