# DownloadFromPlaylist

Goで書かれたシンプルなCLIツール。YouTube Data API v3でプレイリスト内の動画IDを取得し、[yt-dlp](https://github.com/yt-dlp/yt-dlp)（[goutubedl](https://github.com/wader/goutubedl)経由）を使って動画をダウンロードします。

## 特徴

- 指定したYouTubeプレイリストから全動画のIDを取得
- 最良の映像＋音声フォーマットでダウンロード
- ダウンロードした動画を自動的に作成される `Video/` ディレクトリに保存
- 環境変数（`.env`）で簡単に設定可能
- Go Modules対応、依存関係はローカルで完結（`yt-dlp` 以外のグローバルインストール不要）

## 必要要件

- **Go** 1.24 以上
- **YouTube Data API v3** の API キー
- **yt-dlp** がインストールされ、`PATH` に登録されていること

```bash
  # 例: pip でインストール
  pip install yt-dlp
```

## はじめに

1. **リポジトリをクローン**

   ```bash
   git clone https://github.com/yourusername/DownloadFromPlaylist.git
   cd DownloadFromPlaylist
   ```

2. **依存関係を取得**

   ```bash
   go mod tidy
   ```

3. **`.env` ファイルを作成**（プロジェクトルート）

   ```dotenv
   YOUTUBE_API_KEY=あなたのAPI_KEY
   YOUTUBE_PLAYLIST_ID=対象のプレイリストID
   ```

4. **バイナリをビルド**（任意）

   ```bash
   go build -o download-playlist main.go
   ```

5. **実行**

   ```bash
   # ビルド済みバイナリの場合
   ./download-playlist

   # または
   go run main.go
   ```

   実行すると以下を行います：

   1. `.env` を読み込み、`Video/` ディレクトリを自動生成
   2. プレイリストから動画IDを取得
   3. 各動画を `Video/<VIDEO_ID>.webm` としてダウンロード（最良フォーマット）
   4. ダウンロード間に1秒のインターバル

## ディレクトリ構成

```tree
.
├── go.mod           # Go Modules定義
├── go.sum           # 依存チェックサム
├── main.go          # メイン処理: プレイリスト取得＆ダウンロード
├── README.md        # （このファイル）
├── test/
│   └── README/
│       └── main.go  # 単一URLダウンロードの最小サンプル
├── tmp/
│   └── tmp.go       # 実験用: 並列ダウンロード＋YouTube API直叩き
└── Video/           # ダウンロード出力先（自動生成）
```

## 設定

* **環境変数**

  * `YOUTUBE_API_KEY` — YouTube Data API v3キー
  * `YOUTUBE_PLAYLIST_ID` — ダウンロード対象のプレイリストID

* **yt-dlpのパス変更**
  必要に応じて `main.go` 内の `goutubedl.Path = "yt-dlp"` を編集するか、環境変数で上書きしてください。

## カスタマイズ & 上級者向け

* ダウンロードオプション（フォーマット指定など）は `sess.Download(ctx, "bestvideo+bestaudio")` 部分を変更
* 大量プレイリストやリトライ処理は `tmp/tmp.go` を参照（並列ダウンロード＋タイムアウト例）
* `test/README/main.go` では単一動画のフォーマット調査・ダウンロード手順を確認可能

README.md made by [ChatGPT](https://chat.openai.com/) with hitto.
