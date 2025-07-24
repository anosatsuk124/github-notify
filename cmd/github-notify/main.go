package main

import (
	"context"
	"fmt"
	"log"
	"os"

	"github-notify/internal/config"
	"github-notify/internal/github"
	"github-notify/internal/cache"
	"github-notify/internal/notifier"

	"github.com/spf13/cobra"
)

var (
	configPath string
	verbose    bool
	allBranches bool
)

func main() {
	rootCmd := &cobra.Command{
		Use:   "github-notify",
		Short: "GitHub通知監視ツール",
		Long: `GitHubのIssue、PR、コミットの変更を監視してデスクトップ通知を送信します。
設定ファイル: ~/.config/github-notify/config.yaml`,
		Run: runNotify,
	}

	rootCmd.Flags().StringVarP(&configPath, "config", "c", "", "設定ファイルのパス")
	rootCmd.Flags().BoolVarP(&verbose, "verbose", "v", false, "詳細ログを表示")
	rootCmd.Flags().BoolVar(&allBranches, "all-branches", false, "全ブランチを監視")

	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func runNotify(cmd *cobra.Command, args []string) {
	ctx := context.Background()

	// 設定読み込み
	cfg, err := config.Load(configPath)
	if err != nil {
		log.Fatalf("設定の読み込みに失敗: %v", err)
	}

	// キャッシュマネージャー初期化
	cacheManager, err := cache.New(cfg.CacheDir)
	if err != nil {
		log.Fatalf("キャッシュの初期化に失敗: %v", err)
	}

	// GitHub クライアント初期化
	ghClient, err := github.NewClient(cfg.GitHubToken)
	if err != nil {
		log.Fatalf("GitHub クライアントの初期化に失敗: %v", err)
	}

	// 通知クライアント初期化
	notifyClient := notifier.New()

	// 通知チェック実行
	if err := checkAndNotify(ctx, cfg, cacheManager, ghClient, notifyClient); err != nil {
		log.Fatalf("通知チェックに失敗: %v", err)
	}

	if verbose {
		fmt.Println("通知チェック完了")
	}
}

func checkAndNotify(ctx context.Context, cfg *config.Config, cache *cache.Manager, gh *github.Client, notify *notifier.Client) error {
	// 通知チェック
	if cfg.Notifications.EnableIssues || cfg.Notifications.EnablePRs {
		if err := checkNotifications(ctx, cfg, cache, gh, notify); err != nil {
			return fmt.Errorf("通知チェックエラー: %w", err)
		}
	}

	// コミットチェック
	if cfg.Notifications.EnableCommits {
		if err := checkCommits(ctx, cfg, cache, gh, notify); err != nil {
			return fmt.Errorf("コミットチェックエラー: %w", err)
		}
	}

	return nil
}

func checkNotifications(ctx context.Context, cfg *config.Config, cache *cache.Manager, gh *github.Client, notify *notifier.Client) error {
	// GitHub 通知取得と処理
	notifications, err := gh.GetNotifications(ctx)
	if err != nil {
		return err
	}

	for _, n := range notifications {
		if !cache.IsNotificationSeen(n.ID) {
			if err := notify.SendNotification(n.Type, n.Repository, n.Title); err != nil {
				log.Printf("通知送信エラー: %v", err)
				continue
			}
			cache.MarkNotificationSeen(n.ID)
		}
	}

	return cache.Save()
}

func checkCommits(ctx context.Context, cfg *config.Config, cache *cache.Manager, gh *github.Client, notify *notifier.Client) error {
	for _, repo := range cfg.Repositories {
		branches := repo.Branches
		if allBranches || contains(branches, "*") {
			// 全ブランチ取得
			allBranches, err := gh.GetBranches(ctx, repo.Repo)
			if err != nil {
				log.Printf("ブランチ取得エラー (%s): %v", repo.Repo, err)
				continue
			}
			branches = allBranches
		}

		for _, branch := range branches {
			if branch == "*" {
				continue
			}

			lastSHA := cache.GetLastCommitSHA(repo.Repo, branch)
			commits, err := gh.GetCommits(ctx, repo.Repo, branch, lastSHA)
			if err != nil {
				log.Printf("コミット取得エラー (%s:%s): %v", repo.Repo, branch, err)
				continue
			}

			for _, commit := range commits {
				if err := notify.SendCommitNotification(commit.Author, commit.Message, repo.Repo, branch, commit.URL); err != nil {
					log.Printf("コミット通知送信エラー: %v", err)
					continue
				}
			}

			if len(commits) > 0 {
				cache.SetLastCommitSHA(repo.Repo, branch, commits[0].SHA)
			}
		}
	}

	return cache.Save()
}

func contains(slice []string, item string) bool {
	for _, s := range slice {
		if s == item {
			return true
		}
	}
	return false
}