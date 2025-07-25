# GitHub Notify

GitHubのIssue、プルリクエスト、コミットの変更を監視してデスクトップ通知を送信するツールです。

## 特徴

- ✅ **GitHub通知監視**: Issue、PRの新着通知
- ✅ **コミット監視**: リポジトリ・ブランチ別の新しいコミット
- ✅ **全ブランチ対応**: `*` 指定で全ブランチ監視可能
- ✅ **効率的なキャッシュ**: 重複通知を防ぐSHAベースキャッシュ
- ✅ **クロスプラットフォーム**: Linux, macOS, Windows対応
- ✅ **単一バイナリ**: 依存関係なしで配布可能
- ✅ **YAML設定**: 柔軟な設定管理

## インストール

### 要件

- Go 1.21+
- GitHub Personal Access Token

### ビルドとインストール

```bash
# リポジトリをクローン
git clone <repository>
cd github-notify

# 依存関係をダウンロード
make deps

# ビルド
make build

# インストール (~/.local/bin)
make install

# または、システム全体にインストール
make install-system
```

### バイナリリリース

```bash
# 複数プラットフォーム用バイナリを作成
make build-release
```

## 設定

### GitHub Token の取得

1. [GitHub Settings > Developer settings > Personal access tokens](https://github.com/settings/tokens) にアクセス
2. "Generate new token (classic)" をクリック
3. 以下のスコープを選択:
   - `repo` (プライベートリポジトリの場合)
   - `public_repo` (パブリックリポジトリのみの場合)
   - `notifications`
4. トークンをコピーして保存

### 設定ファイル

初回実行時に `~/.config/github-notify/config.yaml` が自動作成されます。

```bash
# 設定ファイルの場所を確認
make config-info

# 初回実行（設定ファイルを作成）
github-notify
```

設定例:

```yaml
# GitHub トークン (環境変数 GITHUB_TOKEN も使用可能)
github_token: "ghp_xxxxxxxxxxxxxxxxxxxx"

# 監視対象リポジトリ
repositories:
  # デフォルトブランチのみ
  - repo: "microsoft/vscode"
    branches: ["main"]
  
  # 特定ブランチ
  - repo: "torvalds/linux"
    branches: ["master", "stable"]
  
  # 全ブランチ監視
  - repo: "golang/go"
    branches: ["*"]

# 通知設定
notifications:
  enable_issues: true   # Issue 通知
  enable_prs: true      # プルリクエスト通知
  enable_commits: true  # コミット通知
```

## 使用方法

### 基本的な使用

```bash
# 通常の監視実行
github-notify

# 詳細ログ付き
github-notify --verbose

# 全ブランチを強制監視
github-notify --all-branches

# 設定ファイルを指定
github-notify --config /path/to/config.yaml
```

### 定期実行の設定

#### systemd timer (Linux)

```bash
# systemdタイマーを設定
make setup-systemd

# 有効化
systemctl --user enable --now github-notify.timer

# 状態確認
systemctl --user status github-notify.timer

# ログ確認
journalctl --user -u github-notify.service -f
```

#### cron

```bash
# 5分毎に実行
echo "*/5 * * * * $HOME/.local/bin/github-notify" | crontab -
```

#### launchd (macOS)

`~/Library/LaunchAgents/com.github.notify.plist`:

```xml
<?xml version="1.0" encoding="UTF-8"?>
<!DOCTYPE plist PUBLIC "-//Apple//DTD PLIST 1.0//EN" "http://www.apple.com/DTDs/PropertyList-1.0.dtd">
<plist version="1.0">
<dict>
    <key>Label</key>
    <string>com.github.notify</string>
    <key>ProgramArguments</key>
    <array>
        <string>/Users/username/.local/bin/github-notify</string>
    </array>
    <key>StartInterval</key>
    <integer>300</integer>
</dict>
</plist>
```

```bash
launchctl load ~/Library/LaunchAgents/com.github.notify.plist
```

## トラブルシューティング

### GitHub API レート制限

```bash
# レート制限を確認（実装予定）
github-notify --check-rate-limit
```

### 通知が表示されない

1. **Linux**: `libnotify` がインストールされていることを確認
   ```bash
   sudo apt install libnotify-bin  # Ubuntu/Debian
   sudo dnf install libnotify      # Fedora
   ```

2. **macOS**: 通知権限が有効になっていることを確認

3. **Windows**: Action Center が有効になっていることを確認

### デバッグ

```bash
# 詳細ログ付きで実行
github-notify --verbose

# 設定ファイルの確認
cat ~/.config/github-notify/config.yaml

# キャッシュファイルの確認
ls -la ~/.cache/github-notify/
```

## 開発

### テスト実行

```bash
make test
```

### 開発用ビルド

```bash
make build-dev
```

### デバッグ情報

```bash
# プラットフォーム情報を表示（実装予定）
github-notify --debug-info
```

## 貢献

1. フォークする
2. フィーチャーブランチを作成 (`git checkout -b feature/amazing-feature`)
3. 変更をコミット (`git commit -m 'Add amazing feature'`)
4. ブランチにプッシュ (`git push origin feature/amazing-feature`)
5. プルリクエストを作成

## 関連プロジェクト

- [gh CLI](https://cli.github.com/) - GitHub 公式 CLI
- [beeep](https://github.com/gen2brain/beeep) - クロスプラットフォーム通知ライブラリ

## ライセンス

```
   Copyright 2025 Satsuki Akiba <anosatsuk124@gmail.com>

   Licensed under the Apache License, Version 2.0 (the "License");
   you may not use this file except in compliance with the License.
   You may obtain a copy of the License at

       http://www.apache.org/licenses/LICENSE-2.0

   Unless required by applicable law or agreed to in writing, software
   distributed under the License is distributed on an "AS IS" BASIS,
   WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
   See the License for the specific language governing permissions and
   limitations under the License.
```

