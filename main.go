package main

// download from youtube-playlist
import (
	"fmt"
	"log"
	"net/http"
	"net/url"
	"os"
	"errors"
	"io"
	"encoding/json"
	"context"
	"time"

	"github.com/wader/goutubedl"

	"github.com/joho/godotenv"
)

func init() {
	// カレントディレクトリの .env を自動で読み込む
	if err := godotenv.Load(); err != nil {
			log.Println(".env ファイルが見つかりませんでした（環境変数が既にセットされている可能性があります）")
	}
	goutubedl.Path = "yt-dlp"
}

type Config struct {
	YouTubeAPIKey    string
	PlaylistId       string
}

func LoadConfig()(*Config, error) {
	apikey := os.Getenv("YOUTUBE_API_KEY")
	if apikey == "" {
		return nil, errors.New("YOUTUBE_API_KEY is nil")
	}

	id := os.Getenv("YOUTUBE_PLAYLIST_ID")
	if id == "" {
		return nil, errors.New("YOUTUBE_PLAYLIST_ID is nil")
	}

	return &Config{
		YouTubeAPIKey: apikey,
		PlaylistId:   id,
	}, nil
}

func getPlaylistVideos(cfg *Config) ([]string, error) {
	// apidocs: https://developers.google.com/youtube/v3/docs/playlistItems/list?hl=ja

	const baseURL = "https://www.googleapis.com/youtube/v3/playlistItems"

	type playlistItemsResponse struct {
		NextPageToken string `json:"nextPageToken"`
		Items         []struct {
			ContentDetails struct {
				VideoID string `json:"videoId"`
			} `json:"contentDetails"`
		} `json:"items"`
		PageInfo       struct {
			TotalResults int `json:"totalResults"`
			ResultsPerPage int `json:"resultsPerPage"`
		} `json:"pageInfo"`
	}

	var (
		videoIDs []string
		pageToken string
	)

	for {
		// 組み立てる
		params := url.Values{}
		params.Set("part", "contentDetails")
		// params.Set("part", "snippet")
		params.Set("playlistId", cfg.PlaylistId)
		params.Set("key", cfg.YouTubeAPIKey)
		params.Set("pageToken", pageToken)
		// params.Set("maxResults", "50")
		// ここでリクエスト
		fmt.Println("GET request initiated")
		resp, err := http.Get(baseURL + "?" + params.Encode())
		if err != nil {
			return nil, fmt.Errorf("HTTP GET error: %w", err)
		}
		// defer resp.Body.Close()

		// fmt.Println("RESP: -------")
		// fmt.Println(resp.StatusCode)
		// fmt.Println("RESP: -------")
		// body, _ := io.ReadAll(resp.Body)
		// fmt.Println(string(body))
		// fmt.Println("-------------")

		if resp.StatusCode != http.StatusOK {
			// エラー
			body, _ := io.ReadAll(resp.Body)
			resp.Body.Close()
			return nil, fmt.Errorf("API error: status %d: %s", resp.StatusCode, string(body))
		}

		// jsonをデコード
		var result playlistItemsResponse
		if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
			resp.Body.Close()
			return nil, fmt.Errorf("JSON decode error: %w", err)
		}

		for _, item := range result.Items {
			videoIDs = append(videoIDs, item.ContentDetails.VideoID)
		}

		fmt.Println("Video IDs:", videoIDs)
		fmt.Println("Page Info:", result.PageInfo)
		if result.NextPageToken == "" {
			resp.Body.Close()
			break
		}
		pageToken = result.NextPageToken
		fmt.Println("Next Page Token:", result.NextPageToken)
	}
	return videoIDs, nil
}

func downloadVideo(ctx context.Context, videoID string) error {
	videoURL := "https://www.youtube.com/watch?v=" + videoID
	out := videoID + ".mp4"

	// タイムアウト付きコンテキスト（例：動画あたり 60 秒）
	ctx, cancel := context.WithTimeout(ctx, 60*time.Second)
	defer cancel()

	opts := goutubedl.Options{}
	sess, err := goutubedl.New(ctx, videoURL, opts)
	if err != nil {
		return fmt.Errorf("goutubedl.New: %w", err)
	}
	dl, err := sess.Download(ctx, "best")
	if err != nil {
		return fmt.Errorf("Download failed: %w", err)
	}
	defer dl.Close()

	f, err := os.Create(out)
	if err != nil {
		return fmt.Errorf("file create failed: %w", err)
	}
	defer f.Close()

	if _, err := io.Copy(f, dl); err != nil {
		return fmt.Errorf("file write failed: %w", err)
	}
	log.Printf("Saved %s", out)
	return nil
}

func main() {
	cfg, err := LoadConfig()
	if err != nil {
		fmt.Fprintf(os.Stderr, "failed to load config: %v\n", err)
		os.Exit(1)
	}
	fmt.Println("---func getPlaylistVideos(cfg *Config) Started---")
	videoIDs, err := getPlaylistVideos(cfg)
	fmt.Println("---func getPlaylistVideos(cfg *Config) Finished---")
	if err != nil {
		fmt.Fprintf(os.Stderr, "failed to get playlist videos: %v\n", err)
		os.Exit(1)
	}
	for i, videoID := range videoIDs {
		fmt.Printf("Video ID:%d: %s\n", i, videoID)
	}
	fmt.Println("Total videos:", len(videoIDs))

	ctx := context.Background()
	for _, videoID := range videoIDs {
		fmt.Printf("Downloading video: %s\n", videoID)
		if err := downloadVideo(ctx, videoID); err != nil {
			fmt.Fprintf(os.Stderr, "failed to download video %s: %v\n", videoID, err)
			continue // エラーが発生しても次の動画をダウンロード
		}
		fmt.Printf("Successfully downloaded video: %s\n", videoID)
		time.Sleep(1 * time.Second) // ダウンロード間隔を空ける
	}

	fmt.Println("All videos downloaded successfully.")
}
