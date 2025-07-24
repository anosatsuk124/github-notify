package notifier

import (
	"fmt"
	"runtime"
	"strings"

	"github.com/gen2brain/beeep"
)

type Client struct {
	appName string
}

// New 通知クライアントを作成します
func New() *Client {
	return &Client{
		appName: "GitHub Notify",
	}
}

// SendNotification GitHub通知のデスクトップ通知を送信します
func (c *Client) SendNotification(notificationType, repository, title string) error {
	subject := fmt.Sprintf("%s @ %s", notificationType, repository)
	
	// アイコンの選択
	icon := c.getIcon()
	
	return beeep.Notify(subject, title, icon)
}

// SendCommitNotification コミット通知のデスクトップ通知を送信します
func (c *Client) SendCommitNotification(author, message, repository, branch, url string) error {
	// メッセージを適切な長さに制限
	if len(message) > 100 {
		message = message[:97] + "..."
	}

	subject := fmt.Sprintf("New commit @ %s:%s", repository, branch)
	body := fmt.Sprintf("Author: %s\nMessage: %s", author, message)
	
	// アイコンの選択
	icon := c.getIcon()
	
	return beeep.Notify(subject, body, icon)
}

// getIcon プラットフォームに適したアイコンを取得します
func (c *Client) getIcon() string {
	switch runtime.GOOS {
	case "linux":
		// Linux の場合、システムアイコンを使用
		icons := []string{
			"github",
			"git",
			"vcs-git",
			"application-x-git",
			"github-desktop",
		}
		
		// 最初に見つかったアイコンを使用（実際は存在チェックが必要）
		for _, icon := range icons {
			return icon
		}
		return "info"
		
	case "darwin":
		// macOS の場合
		return ""
		
	case "windows":
		// Windows の場合
		return ""
		
	default:
		return ""
	}
}

// SendBulkNotifications 複数の通知をまとめて送信します
func (c *Client) SendBulkNotifications(notifications []BulkNotification) error {
	for _, notif := range notifications {
		var err error
		
		switch notif.Type {
		case "commit":
			err = c.SendCommitNotification(
				notif.Author,
				notif.Message,
				notif.Repository,
				notif.Branch,
				notif.URL,
			)
		case "issue", "pull_request":
			err = c.SendNotification(
				notif.Type,
				notif.Repository,
				notif.Title,
			)
		default:
			err = fmt.Errorf("不明な通知タイプ: %s", notif.Type)
		}
		
		if err != nil {
			return fmt.Errorf("通知送信エラー (%s): %w", notif.Type, err)
		}
	}
	
	return nil
}

// BulkNotification まとめて送信する通知の構造体
type BulkNotification struct {
	Type       string
	Repository string
	Title      string
	Message    string
	Author     string
	Branch     string
	URL        string
}

// TestNotification テスト通知を送信します
func (c *Client) TestNotification() error {
	return beeep.Notify(
		"GitHub Notify テスト",
		"通知システムが正常に動作しています",
		c.getIcon(),
	)
}

// SendSummaryNotification 複数の変更をまとめたサマリー通知を送信します
func (c *Client) SendSummaryNotification(commitCount, issueCount, prCount int, repositories []string) error {
	if commitCount == 0 && issueCount == 0 && prCount == 0 {
		return nil
	}

	var parts []string
	if commitCount > 0 {
		parts = append(parts, fmt.Sprintf("%d個の新しいコミット", commitCount))
	}
	if issueCount > 0 {
		parts = append(parts, fmt.Sprintf("%d個の新しいIssue", issueCount))
	}
	if prCount > 0 {
		parts = append(parts, fmt.Sprintf("%d個の新しいPR", prCount))
	}

	subject := "GitHub Activity Summary"
	body := strings.Join(parts, "、") + "\n"
	
	if len(repositories) > 0 {
		if len(repositories) <= 3 {
			body += "Repository: " + strings.Join(repositories, ", ")
		} else {
			body += fmt.Sprintf("Repository: %s and %d others", 
				strings.Join(repositories[:2], ", "), 
				len(repositories)-2)
		}
	}

	return beeep.Notify(subject, body, c.getIcon())
}

// IsSupported 現在のプラットフォームで通知がサポートされているかチェックします
func (c *Client) IsSupported() bool {
	switch runtime.GOOS {
	case "linux", "darwin", "windows":
		return true
	default:
		return false
	}
}

// GetPlatformInfo プラットフォーム情報を取得します
func (c *Client) GetPlatformInfo() map[string]string {
	return map[string]string{
		"os":        runtime.GOOS,
		"arch":      runtime.GOARCH,
		"supported": fmt.Sprintf("%t", c.IsSupported()),
		"icon":      c.getIcon(),
	}
}