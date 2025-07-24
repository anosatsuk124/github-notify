package github

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/google/go-github/v62/github"
	"golang.org/x/oauth2"
)

type Client struct {
	client *github.Client
}

type Notification struct {
	ID         string
	Type       string
	Repository string
	Title      string
	URL        string
}

type Commit struct {
	SHA     string
	Author  string
	Message string
	Date    time.Time
	URL     string
}

// NewClient GitHub API クライアントを作成します
func NewClient(token string) (*Client, error) {
	if token == "" {
		return nil, fmt.Errorf("GitHub トークンが必要です")
	}

	ctx := context.Background()
	ts := oauth2.StaticTokenSource(
		&oauth2.Token{AccessToken: token},
	)
	tc := oauth2.NewClient(ctx, ts)

	client := github.NewClient(tc)

	return &Client{client: client}, nil
}

// GetNotifications GitHub 通知を取得します
func (c *Client) GetNotifications(ctx context.Context) ([]*Notification, error) {
	opts := &github.NotificationListOptions{
		All:           true,
		Participating: false,
		ListOptions: github.ListOptions{
			PerPage: 50,
		},
	}

	var allNotifications []*Notification

	for {
		notifications, resp, err := c.client.Activity.ListNotifications(ctx, opts)
		if err != nil {
			return nil, fmt.Errorf("通知の取得に失敗: %w", err)
		}

		for _, n := range notifications {
			notification := &Notification{
				ID:         n.GetID(),
				Type:       n.GetSubject().GetType(),
				Repository: n.GetRepository().GetFullName(),
				Title:      n.GetSubject().GetTitle(),
				URL:        n.GetSubject().GetURL(),
			}
			allNotifications = append(allNotifications, notification)
		}

		if resp.NextPage == 0 {
			break
		}
		opts.Page = resp.NextPage
	}

	return allNotifications, nil
}

// GetBranches リポジトリのブランチ一覧を取得します
func (c *Client) GetBranches(ctx context.Context, repoFullName string) ([]string, error) {
	parts := strings.Split(repoFullName, "/")
	if len(parts) != 2 {
		return nil, fmt.Errorf("不正なリポジトリ名: %s", repoFullName)
	}

	owner, repo := parts[0], parts[1]

	opts := &github.BranchListOptions{
		ListOptions: github.ListOptions{
			PerPage: 100,
		},
	}

	var allBranches []string

	for {
		branches, resp, err := c.client.Repositories.ListBranches(ctx, owner, repo, opts)
		if err != nil {
			return nil, fmt.Errorf("ブランチ一覧の取得に失敗 (%s): %w", repoFullName, err)
		}

		for _, branch := range branches {
			allBranches = append(allBranches, branch.GetName())
		}

		if resp.NextPage == 0 {
			break
		}
		opts.Page = resp.NextPage
	}

	return allBranches, nil
}

// GetCommits 指定されたブランチの新しいコミットを取得します
func (c *Client) GetCommits(ctx context.Context, repoFullName, branch, sinceCommit string) ([]*Commit, error) {
	parts := strings.Split(repoFullName, "/")
	if len(parts) != 2 {
		return nil, fmt.Errorf("不正なリポジトリ名: %s", repoFullName)
	}

	owner, repo := parts[0], parts[1]

	opts := &github.CommitsListOptions{
		SHA: branch,
		ListOptions: github.ListOptions{
			PerPage: 50,
		},
	}

	// 時間ベースフィルタリング (24時間以内)
	since := time.Now().Add(-24 * time.Hour)
	opts.Since = since

	var newCommits []*Commit

	for {
		commits, resp, err := c.client.Repositories.ListCommits(ctx, owner, repo, opts)
		if err != nil {
			return nil, fmt.Errorf("コミット一覧の取得に失敗 (%s:%s): %w", repoFullName, branch, err)
		}

		for _, commit := range commits {
			sha := commit.GetSHA()
			
			// 既知のコミットに到達したら停止
			if sha == sinceCommit {
				return newCommits, nil
			}

			author := "unknown"
			if commit.GetAuthor() != nil {
				author = commit.GetAuthor().GetLogin()
			} else if commit.GetCommit().GetAuthor() != nil {
				author = commit.GetCommit().GetAuthor().GetName()
			}

			message := commit.GetCommit().GetMessage()
			// メッセージの最初の行のみ取得
			if idx := strings.Index(message, "\n"); idx != -1 {
				message = message[:idx]
			}

			newCommit := &Commit{
				SHA:     sha,
				Author:  author,
				Message: message,
				Date:    commit.GetCommit().GetAuthor().GetDate().Time,
				URL:     commit.GetHTMLURL(),
			}

			newCommits = append(newCommits, newCommit)
		}

		if resp.NextPage == 0 {
			break
		}
		opts.Page = resp.NextPage
	}

	return newCommits, nil
}

// GetDefaultBranch リポジトリのデフォルトブランチを取得します
func (c *Client) GetDefaultBranch(ctx context.Context, repoFullName string) (string, error) {
	parts := strings.Split(repoFullName, "/")
	if len(parts) != 2 {
		return "", fmt.Errorf("不正なリポジトリ名: %s", repoFullName)
	}

	owner, repo := parts[0], parts[1]

	repository, _, err := c.client.Repositories.Get(ctx, owner, repo)
	if err != nil {
		return "", fmt.Errorf("リポジトリ情報の取得に失敗 (%s): %w", repoFullName, err)
	}

	return repository.GetDefaultBranch(), nil
}

// CheckRateLimit API レート制限をチェックします
func (c *Client) CheckRateLimit(ctx context.Context) (*github.RateLimits, error) {
	rateLimit, _, err := c.client.RateLimit.Get(ctx)
	if err != nil {
		return nil, fmt.Errorf("レート制限の確認に失敗: %w", err)
	}

	return rateLimit, nil
}

// IsRateLimited API レート制限に達しているかチェックします
func (c *Client) IsRateLimited(ctx context.Context) (bool, time.Duration, error) {
	rateLimit, err := c.CheckRateLimit(ctx)
	if err != nil {
		return false, 0, err
	}

	core := rateLimit.GetCore()
	remaining := core.Remaining
	
	if remaining <= 10 { // 10回以下になったら制限とみなす
		resetTime := core.Reset.Time
		waitDuration := time.Until(resetTime)
		return true, waitDuration, nil
	}

	return false, 0, nil
}