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
		log.Fatal("YOUTUBE_PLAYLIST_ID environment variable is not set")
	}

	return &Config{
		YouTubeAPIKey: apikey,
		PlaylistId:   id,
	}, nil
}


type playlistItemsResponse struct {
	NextPageToken string
	Items         []struct {
		ContentDetails struct {
			VideoID     string
		}
	}
}

func getPlaylistVideos(cfg *Config) ([]string, error) {
	const baseURL = "https://www.googleapis.com/youtube/v3/playlistItems"
	// var (
	// 	videoIDs []string
	// 	pageToken string
	// )
	// for {
	// }
	params := url.Values{}
	params.Set("part", "contentDetails")
	// params.Set("part", "snippet")
	params.Set("playlistId", cfg.PlaylistId)
	params.Set("key", cfg.YouTubeAPIKey)
	// ここでリクエスト
	resp, err := http.Get(baseURL + "?" + params.Encode())
	if err != nil {
		log.Fatalf("HTTP GET Err: %v", err)
	}
	defer resp.Body.Close()

	// レスポンスボディをそのまま標準出力に
	if _, err := io.Copy(os.Stdout, resp.Body); err != nil {
		log.Fatalf("レスポンス読み込みエラー: %v", err)
	}

	var testdata []string

	return testdata, nil
	// return videoIDs, nil



	// res, err := http.Get(cfg.PlaylistId)
	// if err != nil {
	// 	log.Fatal(err)
	// }
	// defer res.Body.Close()
	// return []string{"video1", "video2", "video3"}, nil
}

func main() {
	cfg, err := LoadConfig()
	if err != nil {
		fmt.Fprintf(os.Stderr, "failed to load config: %v\n", err)
		os.Exit(1)
	}

	_, _ = getPlaylistVideos(cfg)
}
