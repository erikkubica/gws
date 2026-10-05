package youtube

import (
	"fmt"
)

// VideoSummary holds summary metadata for YouTube videos.
type VideoSummary struct {
	ID          string `json:"id"`
	Title       string `json:"title"`
	ChannelName string `json:"channelName"`
	Description string `json:"description"`
	PublishedAt string `json:"publishedAt"`
}

// VideoDetails holds detailed statistics and description.
type VideoDetails struct {
	VideoSummary
	ViewCount uint64 `json:"viewCount,omitempty"`
	LikeCount uint64 `json:"likeCount,omitempty"`
}

// SearchVideos queries the YouTube search API for video items.
func (s *Service) SearchVideos(query string, max int64) ([]VideoSummary, error) {
	if max <= 0 {
		max = 10
	}
	call := s.client.Search.List([]string{"snippet"}).Q(query).Type("video").MaxResults(max)
	res, err := call.Do()
	if err != nil {
		return nil, fmt.Errorf("search youtube videos: %w", err)
	}

	var videos []VideoSummary
	for _, item := range res.Items {
		if item.Id.VideoId == "" {
			continue
		}
		videos = append(videos, VideoSummary{
			ID:          item.Id.VideoId,
			Title:       item.Snippet.Title,
			ChannelName: item.Snippet.ChannelTitle,
			Description: item.Snippet.Description,
			PublishedAt: item.Snippet.PublishedAt,
		})
	}
	return videos, nil
}

// GetVideoDetails retrieves metadata and statistics for a specific video ID.
func (s *Service) GetVideoDetails(videoID string) (*VideoDetails, error) {
	call := s.client.Videos.List([]string{"snippet", "statistics"}).Id(videoID)
	res, err := call.Do()
	if err != nil || len(res.Items) == 0 {
		return nil, fmt.Errorf("retrieve video details %s: %w", videoID, err)
	}

	item := res.Items[0]
	return &VideoDetails{
		VideoSummary: VideoSummary{
			ID:          item.Id,
			Title:       item.Snippet.Title,
			ChannelName: item.Snippet.ChannelTitle,
			Description: item.Snippet.Description,
			PublishedAt: item.Snippet.PublishedAt,
		},
		ViewCount: item.Statistics.ViewCount,
		LikeCount: item.Statistics.LikeCount,
	}, nil
}
