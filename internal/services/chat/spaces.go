package chat

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"sync"

	"google.golang.org/api/chat/v1"
)

// SpaceInfo represents a Google Chat space or direct message room.
type SpaceInfo struct {
	Name           string   `json:"name"`
	DisplayName    string   `json:"display_name"`
	Type           string   `json:"type"`
	SpaceType      string   `json:"space_type"`
	LastActiveTime string   `json:"last_active_time,omitempty"`
	Members        []string `json:"members,omitempty"`
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

// ListSpaces returns all joined spaces and direct message rooms with resolved participant names.
func (s *Service) ListSpaces(pageSize int64) ([]*SpaceInfo, error) {
	if pageSize <= 0 {
		pageSize = 20
	}
	res, err := s.client.Spaces.List().PageSize(pageSize).Do()
	if err != nil {
		return nil, fmt.Errorf("list spaces: %w", err)
	}

	memberMap := s.fetchMembersConcurrently(res.Spaces)
	currentUser := s.getCurrentUserName()
	if currentUser == "" {
		currentUser = detectCurrentUserName(res.Spaces, memberMap)
	}

	spaces := make([]*SpaceInfo, len(res.Spaces))
	for i, sp := range res.Spaces {
		spaces[i] = s.resolveSpaceInfo(sp, memberMap[sp.Name], currentUser)
	}
	return spaces, nil
}

func (s *Service) fetchMembersConcurrently(spaces []*chat.Space) map[string][]string {
	var mu sync.Mutex
	memberMap := make(map[string][]string)
	var wg sync.WaitGroup
	for _, sp := range spaces {
		if sp.DisplayName != "" {
			continue
		}
		wg.Add(1)
		go func(spaceName string) {
			defer wg.Done()
			names := s.fetchSpaceMemberNames(spaceName)
			mu.Lock()
			memberMap[spaceName] = names
			mu.Unlock()
		}(sp.Name)
	}
	wg.Wait()
	return memberMap
}

func (s *Service) fetchSpaceMemberNames(spaceName string) []string {
	memList, err := s.client.Spaces.Members.List(spaceName).PageSize(10).Do()
	if err != nil {
		return nil
	}
	var names []string
	for _, m := range memList.Memberships {
		if m.Member != nil && m.Member.DisplayName != "" {
			names = append(names, m.Member.DisplayName)
		}
	}
	return names
}

func detectCurrentUserName(spaces []*chat.Space, memberMap map[string][]string) string {
	counts := make(map[string]int)
	dmCount := 0
	for _, sp := range spaces {
		if sp.SpaceType == "DIRECT_MESSAGE" {
			dmCount++
			for _, m := range memberMap[sp.Name] {
				counts[m]++
			}
		}
	}
	if dmCount < 2 {
		return ""
	}
	bestName := ""
	maxCount := 0
	for name, count := range counts {
		if count > maxCount && count >= dmCount/2 {
			maxCount = count
			bestName = name
		}
	}
	return bestName
}

func (s *Service) resolveSpaceInfo(sp *chat.Space, members []string, currentUser string) *SpaceInfo {
	info := &SpaceInfo{
		Name:           sp.Name,
		DisplayName:    sp.DisplayName,
		Type:           sp.Type,
		SpaceType:      sp.SpaceType,
		LastActiveTime: sp.LastActiveTime,
		Members:        members,
	}
	if info.DisplayName != "" {
		return info
	}
	if len(members) == 0 {
		info.DisplayName = defaultDisplayName(sp.SpaceType)
		return info
	}
	info.DisplayName = formatMembersDisplayName(sp.SpaceType, members, currentUser)
	return info
}

func defaultDisplayName(spaceType string) string {
	if spaceType == "DIRECT_MESSAGE" {
		return "(Direct Message)"
	}
	return "(Unnamed Space)"
}

func formatMembersDisplayName(spaceType string, members []string, currentUser string) string {
	var otherNames []string
	for _, name := range members {
		if currentUser == "" || !strings.EqualFold(name, currentUser) {
			otherNames = append(otherNames, name)
		}
	}
	if spaceType == "DIRECT_MESSAGE" {
		return formatDMName(members, otherNames)
	}
	if len(otherNames) > 0 {
		return "Group: " + strings.Join(otherNames, ", ")
	}
	return "Group: " + strings.Join(members, ", ")
}

func formatDMName(allMembers, otherNames []string) string {
	if len(otherNames) == 1 {
		return "DM: " + otherNames[0]
	}
	if len(otherNames) > 1 {
		return "DM: " + strings.Join(otherNames, ", ")
	}
	if len(allMembers) > 0 {
		return "DM: " + allMembers[0] + " (You)"
	}
	return "(Direct Message)"
}

func (s *Service) getCurrentUserName() string {
	req, err := http.NewRequest(http.MethodGet, "https://www.googleapis.com/oauth2/v2/userinfo", nil)
	if err != nil {
		return ""
	}
	resp, err := s.httpClient.Do(req)
	if err != nil || resp.StatusCode != http.StatusOK {
		return ""
	}
	defer resp.Body.Close()
	var info struct {
		Name string `json:"name"`
	}
	_ = json.NewDecoder(resp.Body).Decode(&info)
	return strings.TrimSpace(info.Name)
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
	members := s.fetchSpaceMemberNames(sp.Name)
	return s.resolveSpaceInfo(sp, members, s.getCurrentUserName()), nil
}
