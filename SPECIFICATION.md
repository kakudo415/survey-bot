# 論文Survey支援Slack Bot仕様書

## 概要
学術論文のURLがSlackに投稿された際に、自動的に論文を解析して日本語要約を返すSlack botをGoで実装する。

## アーキテクチャ
- **設計手法**: Domain Driven Design (DDD)
- **言語**: Go
- **外部API**: Slack Event API、OpenAI API

## 機能仕様

### 1. 論文URL検出・解析
- **トリガー**: Slack Event APIによるメッセージ監視
- **対象**: 特定チャンネルに投稿されたメッセージ内のURL
- **初期対応サイト**: IEEE Xplore
- **拡張性**: Infrastructure層での抽象化により他サイト対応可能

### 2. 論文情報取得
- **メタデータ**: 決定的な方法でタイトル・著者・発行年等を取得
- **本文**: 論文サイトから本文テキストを抽出
- **実装**: 論文サイトごとにInfrastructure層で個別実装

### 3. 要約生成
- **API**: OpenAI API使用
- **言語**: 日本語での要約生成
- **内容**: 論文全体の要約（概要・手法・結果・意義）

### 4. Slack応答
- **方式**: 元メッセージのスレッドに返信
- **フォーマット**: 
  ```
  📄 [論文タイトル]
  👥 著者: [著者名]
  📅 発行年: [年]
  
  ## 要約
  [OpenAI APIによる日本語要約]
  ```

### 5. エラーハンドリング
- **URL解析失敗**: Ephemeral messageでエラー通知
- **API制限**: Ephemeral messageで制限通知
- **論文取得不可**: Ephemeral messageで取得不可通知

## 技術仕様

### アーキテクチャレイヤー

#### Domain層
- `Paper` エンティティ（ID、タイトル、著者、URL、本文）
- `Summary` 値オブジェクト（日本語要約テキスト）
- `PaperRepository` インターフェース
- `SummaryService` インターフェース

#### Application層
- `SurveyUseCase` - 論文解析・要約の統合処理
- `SlackEventHandler` - Slack Event API処理

#### Infrastructure層
- `IEEEPaperRepository` - IEEE Xplore専用実装
- `OpenAISummaryService` - OpenAI API実装
- `SlackClient` - Slack API実装

#### Presentation層
- `SlackEventController` - Slack Event API endpoint

### 環境設定
- **認証情報**: Secret Manager経由で環境変数取得
  - `SLACK_BOT_TOKEN`
  - `SLACK_SIGNING_SECRET`
  - `OPENAI_API_KEY`
- **監視対象**: `TARGET_CHANNEL_ID`

### デプロイメント
- **方式**: （要相談）
- **設定管理**: Secret Manager使用

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