# CLAUDE.md

このファイルは、Claude Code (claude.ai/code) がこのリポジトリで作業する際のガイダンスを提供します。

## 開発コマンド

### ビルドと実行
```bash
# アプリケーションをビルド
go build -o bin/survey-bot ./cmd

# アプリケーションを実行
./bin/survey-bot

# ビルド成果物をクリーンアップ
rm -rf bin/
```

### 依存関係管理
```bash
# 依存関係をインストール/更新
go mod tidy

# 新しい依存関係を追加
go get <package-name>
```

### 開発環境
```bash
# 開発時にライブリロードで実行
PORT=8080 ./bin/survey-bot

# ビルドせずにコンパイルエラーをチェック
go build ./...
```

## アーキテクチャ概要

これは **Domain Driven Design (DDD)** で設計されたSlack botで、Slackチャンネルに投稿された学術論文URLを自動的にOpenAI APIを使って要約します。

### コアフロー
1. **Slack Event API** がURLを含むメッセージイベントを受信
2. **Paper Repository** が学術サイトから論文メタデータとコンテンツを抽出
3. **Summary Service** がOpenAI API経由で日本語要約を生成
4. **Slack Client** がスレッド返信として要約を投稿

### DDDレイヤー構造

**Domain層** (`domain/`):
- `Paper` エンティティ（ID、メタデータ、コンテンツ）
- `Summary` 値オブジェクト（生成された要約）
- `PaperRepository` と `SummaryService` インターフェースが契約を定義

**UseCase層** (`usecase/`):
- `SurveyUseCase` が論文取得と要約生成を統合
- `SlackEventHandler` がSlackイベント処理とURL抽出を担当

**Infrastructure層** (`infrastructure/`):
- `IEEEPaperRepository` がIEEE Xplore専用の論文スクレイピングを実装
- `OpenAISummaryService` がOpenAI APIとの統合を実装
- `SlackClient` がslack-go/slackライブラリをラップ
- `prompts/system_prompt.txt` がAIシステムプロンプトを格納（go:embedで埋め込み）

**Presentation層** (`presentation/`):
- `SlackEventController` がHTTPエンドポイントとSlack署名検証を処理

### 主要な設計パターン

**依存性注入**: 全ての依存関係は `cmd/main.go` でワイヤリングされ、インターフェースが層間の契約を定義

**Repositoryパターン**: `PaperRepository` が論文ソース実装を抽象化。現在はIEEE Xploreのみだが、arXiv、ACM等への拡張を想定した設計

**イベント駆動**: URLを含むSlackメッセージイベントを処理し、対象チャンネルとメッセージタイプでフィルタリング

## 環境変数

必須:
- `SLACK_BOT_TOKEN` - Slack bot OAuthトークン
- `SLACK_SIGNING_SECRET` - Slackリクエスト検証用
- `OPENAI_API_KEY` - OpenAI API認証
- `TARGET_CHANNEL_ID` - 監視対象のSlackチャンネル

オプション:
- `OPENAI_MODEL` - OpenAIモデル（デフォルト: "gpt-4.1"）
- `PORT` - HTTPサーバーポート（デフォルト: "8080"）

## 主要な統合

**Slack Event API**: slack-go/slackライブラリを使用してEvents API、署名検証、スレッド対応メッセージ投稿を実装

**OpenAI API**: 日本語学術論文要約用の埋め込みシステムプロンプトを使用したチャット補完のための直接HTTP統合

**Webスクレイピング**: PuerkitoBio/goqueryを使用してIEEE Xplore論文を解析 - タイトル、著者、発行日、アブストラクト、キーワードを抽出

## 拡張ポイント

**新しい論文ソース**: 新しい学術サイト用に `infrastructure/` で `domain.PaperRepository` インターフェースを実装

**要約カスタマイズ**: `infrastructure/prompts/system_prompt.txt` を変更するか、異なる要約形式用に `SummaryService` インターフェースを拡張

**イベント処理**: 追加のSlackイベントタイプや処理パターンをサポートするために `SlackEventHandler` を拡張

## 機能仕様

### 論文URL検出・解析
- **トリガー**: Slack Event APIによるメッセージ監視
- **対象**: 特定チャンネルに投稿されたメッセージ内のURL
- **初期対応サイト**: IEEE Xplore
- **拡張性**: Infrastructure層での抽象化により他サイト対応可能

### 論文情報取得
- **メタデータ**: 決定的な方法でタイトル・著者・発行年等を取得
- **本文**: 論文サイトから本文テキストを抽出
- **実装**: 論文サイトごとにInfrastructure層で個別実装

### 要約生成
- **API**: OpenAI API使用
- **言語**: 日本語での要約生成
- **内容**: 論文全体の要約（概要・手法・結果・意義）

### Slack応答
- **方式**: 元メッセージのスレッドに返信
- **フォーマット**: 
  ```
  📄 [論文タイトル]
  👥 著者: [著者名]
  📅 発行年: [年]
  
  ## 要約
  [OpenAI APIによる日本語要約]
  ```

### エラーハンドリング
- **URL解析失敗**: Ephemeral messageでエラー通知
- **API制限**: Ephemeral messageで制限通知
- **論文取得不可**: Ephemeral messageで取得不可通知

## 拡張予定
- arXiv対応
- ACM Digital Library対応
- Google Scholar対応
- 要約形式のカスタマイズ
- 処理履歴の保存

## 制限事項
- 初期版はIEEE Xploreのみ対応
- データ永続化なし
- 同時処理制限なし（OpenAI API制限に依存）

## デプロイメント
- **設定管理**: Secret Manager使用
- **認証情報**: Secret Manager経由で環境変数取得