package cli

import (
	"context"
	"fmt"

	"github.com/erikkubica/gmcp/internal/services/youtube"
	"github.com/spf13/cobra"
)

var ytCmd = &cobra.Command{
	Use:   "yt",
	Short: "Search and inspect YouTube videos",
}

var ytMax int64

var ytSearchCmd = &cobra.Command{
	Use:   "search [query]",
	Short: "Search YouTube videos",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		svc, err := youtube.NewService(context.Background())
		if err != nil {
			return err
		}
		videos, err := svc.SearchVideos(args[0], ytMax)
		if err != nil {
			return err
		}
		for _, v := range videos {
			fmt.Printf("[%s] %s\n  Channel: %s | Published: %s\n\n", v.ID, v.Title, v.ChannelName, v.PublishedAt)
		}
		return nil
	},
}

var ytInfoCmd = &cobra.Command{
	Use:   "info [video_id]",
	Short: "Get statistics and description for a video",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		svc, err := youtube.NewService(context.Background())
		if err != nil {
			return err
		}
		info, err := svc.GetVideoDetails(args[0])
		if err != nil {
			return err
		}
		fmt.Printf("Title: %s\nChannel: %s\nViews: %d | Likes: %d\n\nDescription:\n%s\n",
			info.Title, info.ChannelName, info.ViewCount, info.LikeCount, info.Description)
		return nil
	},
}

func init() {
	ytSearchCmd.Flags().Int64VarP(&ytMax, "max", "m", 10, "Max results")
	ytCmd.AddCommand(ytSearchCmd)
	ytCmd.AddCommand(ytInfoCmd)
}
