package config

import (
	"fmt"
	"os"
	"path/filepath"

	"gopkg.in/yaml.v3"
)

type Config struct {
	GitHubToken   string         `yaml:"github_token"`
	CacheDir      string         `yaml:"cache_dir"`
	Repositories  []Repository   `yaml:"repositories"`
	Notifications Notifications  `yaml:"notifications"`
}

type Repository struct {
	Repo     string   `yaml:"repo"`
	Branches []string `yaml:"branches"`
}

type Notifications struct {
	EnableIssues  bool `yaml:"enable_issues"`
	EnablePRs     bool `yaml:"enable_prs"`
	EnableCommits bool `yaml:"enable_commits"`
}

// Load 設定ファイルを読み込みます
func Load(configPath string) (*Config, error) {
	if configPath == "" {
		homeDir, err := os.UserHomeDir()
		if err != nil {
			return nil, fmt.Errorf("ホームディレクトリの取得に失敗: %w", err)
		}
		configPath = filepath.Join(homeDir, ".config", "github-notify", "config.yaml")
	}

	// 設定ファイルが存在しない場合はデフォルト設定を作成
	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		if err := createDefaultConfig(configPath); err != nil {
			return nil, fmt.Errorf("デフォルト設定の作成に失敗: %w", err)
		}
		fmt.Printf("デフォルト設定ファイルを作成しました: %s\n", configPath)
		fmt.Println("設定を編集してから再実行してください。")
		os.Exit(0)
	}

	data, err := os.ReadFile(configPath)
	if err != nil {
		return nil, fmt.Errorf("設定ファイルの読み込みに失敗: %w", err)
	}

	var config Config
	if err := yaml.Unmarshal(data, &config); err != nil {
		return nil, fmt.Errorf("設定ファイルのパースに失敗: %w", err)
	}

	// デフォルト値の設定
	if config.CacheDir == "" {
		homeDir, _ := os.UserHomeDir()
		config.CacheDir = filepath.Join(homeDir, ".cache", "github-notify")
	}

	// GitHub Token の環境変数チェック
	if config.GitHubToken == "" {
		config.GitHubToken = os.Getenv("GITHUB_TOKEN")
	}

	if config.GitHubToken == "" {
		return nil, fmt.Errorf("GitHub トークンが設定されていません。GITHUB_TOKEN 環境変数または設定ファイルで指定してください")
	}

	// 設定の検証
	if err := validateConfig(&config); err != nil {
		return nil, fmt.Errorf("設定の検証に失敗: %w", err)
	}

	return &config, nil
}

// createDefaultConfig デフォルト設定ファイルを作成します
func createDefaultConfig(configPath string) error {
	// ディレクトリ作成
	dir := filepath.Dir(configPath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return err
	}

	defaultConfig := Config{
		GitHubToken: "", // 環境変数 GITHUB_TOKEN を使用
		CacheDir:    "", // デフォルト: ~/.cache/github-notify
		Repositories: []Repository{
			{
				Repo:     "microsoft/vscode",
				Branches: []string{"main"},
			},
			{
				Repo:     "torvalds/linux",
				Branches: []string{"master"},
			},
			{
				Repo:     "golang/go",
				Branches: []string{"*"}, // 全ブランチ
			},
		},
		Notifications: Notifications{
			EnableIssues:  true,
			EnablePRs:     true,
			EnableCommits: true,
		},
	}

	_, err := yaml.Marshal(&defaultConfig)
	if err != nil {
		return err
	}

	// コメント付きの設定ファイルを作成
	configContent := `# GitHub 通知監視ツール設定ファイル
# 
# GitHub トークンの設定:
#   - GITHUB_TOKEN 環境変数を使用するか
#   - 以下の github_token フィールドで指定
# github_token: "ghp_xxxxxxxxxxxxxxxxxxxx"

# キャッシュディレクトリ (空の場合は ~/.cache/github-notify を使用)
# cache_dir: "/path/to/cache"

# 監視対象リポジトリ
repositories:
  # デフォルトブランチのみ監視
  - repo: "microsoft/vscode"
    branches: ["main"]
  
  # 特定ブランチを監視
  - repo: "torvalds/linux"
    branches: ["master"]
  
  # 全ブランチを監視
  - repo: "golang/go"
    branches: ["*"]
  
  # 複数ブランチを指定
  # - repo: "myorg/myrepo"
  #   branches: ["main", "develop", "staging"]

# 通知設定
notifications:
  enable_issues: true   # Issue 通知
  enable_prs: true      # プルリクエスト通知
  enable_commits: true  # コミット通知
`

	return os.WriteFile(configPath, []byte(configContent), 0644)
}

// validateConfig 設定の妥当性を検証します
func validateConfig(config *Config) error {
	if len(config.Repositories) == 0 {
		return fmt.Errorf("監視対象リポジトリが設定されていません")
	}

	for i, repo := range config.Repositories {
		if repo.Repo == "" {
			return fmt.Errorf("リポジトリ %d: リポジトリ名が空です", i+1)
		}
		if len(repo.Branches) == 0 {
			return fmt.Errorf("リポジトリ %s: ブランチが設定されていません", repo.Repo)
		}
	}

	return nil
}

// ExpandPath パス内の ~ を展開します
func ExpandPath(path string) string {
	if path == "" {
		return path
	}
	if path[0] == '~' {
		homeDir, _ := os.UserHomeDir()
		return filepath.Join(homeDir, path[1:])
	}
	return path
}