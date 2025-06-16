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
	"time"
	// "context"
	// "strings"

	"github.com/kkdai/youtube/v2"

	"github.com/joho/godotenv"
)

func init() {
	// カレントディレクトリの .env を自動で読み込む
	if err := godotenv.Load(); err != nil {
			log.Println(".env ファイルが見つかりませんでした（環境変数が既にセットされている可能性があります）")
	}
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


// headerTransport は全リクエストにブラウザ風 User-Agent を付与
type headerTransport struct {
	rt http.RoundTripper
}

func (t *headerTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	req.Header.Set("User-Agent",
			"Mozilla/5.0 (Windows NT 10.0; Win64; x64) " +
					"AppleWebKit/537.36 (KHTML, like Gecko) " +
					"Chrome/114.0.0.0 Safari/537.36")
	return t.rt.RoundTrip(req)
}

// ExampleDownload : Example code for how to use this package for download video.
func ExampleClient(videoID string) {
	// 1. カスタム HTTP クライアントを作成
	httpClient := &http.Client{
		Timeout:       60 * time.Second, // 全リクエストを 60 秒で打ち切る
		// Unlimited redirects: 常に nil を返してリダイレクトを追従
    CheckRedirect: func(_ *http.Request, _ []*http.Request) error { return nil },
		// 実際のブラウザ風 UA
    Transport:     &headerTransport{rt: http.DefaultTransport},
	}
	client := youtube.Client{
			HTTPClient: httpClient,
	}

	video, err := client.GetVideo(videoID)
	if err != nil {
		panic(err)
	}

	formats := video.Formats.WithAudioChannels() // only get videos with audio
	stream, _, err := client.GetStream(video, &formats[0])
	if err != nil {
		panic(err)
	}
	defer stream.Close()

	file, err := os.Create(videoID + ".mp4")
	if err != nil {
		panic(err)
	}
	defer file.Close()

	_, err = io.Copy(file, stream)
	if err != nil {
		panic(err)
	}
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
	for _, videoID := range videoIDs {
		fmt.Printf("Downloading video: %s\n", videoID)
		ExampleClient(videoID)
		fmt.Printf("Finished downloading video: %s\n", videoID)
	}
	fmt.Println("All videos downloaded successfully.")
}
