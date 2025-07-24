package cache

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"time"
)

type Manager struct {
	cacheDir         string
	notificationsFile string
	commitsFile      string
	data             *CacheData
	mutex            sync.RWMutex
}

type CacheData struct {
	LastUpdated   time.Time                      `json:"last_updated"`
	Notifications map[string]NotificationCache   `json:"notifications"`
	Commits       map[string]map[string]string   `json:"commits"` // repo -> branch -> sha
}

type NotificationCache struct {
	SeenIDs   map[string]bool `json:"seen_ids"`
	LastCheck time.Time       `json:"last_check"`
}

// New キャッシュマネージャーを作成します
func New(cacheDir string) (*Manager, error) {
	// パス展開
	if cacheDir[0] == '~' {
		homeDir, err := os.UserHomeDir()
		if err != nil {
			return nil, fmt.Errorf("ホームディレクトリの取得に失敗: %w", err)
		}
		cacheDir = filepath.Join(homeDir, cacheDir[1:])
	}

	// ディレクトリ作成
	if err := os.MkdirAll(cacheDir, 0755); err != nil {
		return nil, fmt.Errorf("キャッシュディレクトリの作成に失敗: %w", err)
	}

	manager := &Manager{
		cacheDir:          cacheDir,
		notificationsFile: filepath.Join(cacheDir, "notifications.json"),
		commitsFile:       filepath.Join(cacheDir, "commits.json"),
		data: &CacheData{
			Notifications: make(map[string]NotificationCache),
			Commits:       make(map[string]map[string]string),
		},
	}

	// 既存キャッシュの読み込み
	if err := manager.load(); err != nil {
		return nil, fmt.Errorf("キャッシュの読み込みに失敗: %w", err)
	}

	return manager, nil
}

// load キャッシュファイルを読み込みます
func (m *Manager) load() error {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	// 通知キャッシュの読み込み
	if _, err := os.Stat(m.notificationsFile); err == nil {
		data, err := os.ReadFile(m.notificationsFile)
		if err != nil {
			return fmt.Errorf("通知キャッシュファイルの読み込みに失敗: %w", err)
		}

		var notifCache struct {
			Notifications map[string]NotificationCache `json:"notifications"`
		}
		if err := json.Unmarshal(data, &notifCache); err == nil {
			m.data.Notifications = notifCache.Notifications
		}
	}

	// コミットキャッシュの読み込み
	if _, err := os.Stat(m.commitsFile); err == nil {
		data, err := os.ReadFile(m.commitsFile)
		if err != nil {
			return fmt.Errorf("コミットキャッシュファイルの読み込みに失敗: %w", err)
		}

		var commitCache struct {
			Commits map[string]map[string]string `json:"commits"`
		}
		if err := json.Unmarshal(data, &commitCache); err == nil {
			m.data.Commits = commitCache.Commits
		}
	}

	m.data.LastUpdated = time.Now()
	return nil
}

// Save キャッシュをファイルに保存します
func (m *Manager) Save() error {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	m.data.LastUpdated = time.Now()

	// 通知キャッシュの保存
	notifData := struct {
		LastUpdated   time.Time                    `json:"last_updated"`
		Notifications map[string]NotificationCache `json:"notifications"`
	}{
		LastUpdated:   m.data.LastUpdated,
		Notifications: m.data.Notifications,
	}

	data, err := json.MarshalIndent(notifData, "", "  ")
	if err != nil {
		return fmt.Errorf("通知キャッシュのJSONエンコードに失敗: %w", err)
	}

	if err := os.WriteFile(m.notificationsFile, data, 0644); err != nil {
		return fmt.Errorf("通知キャッシュファイルの保存に失敗: %w", err)
	}

	// コミットキャッシュの保存
	commitData := struct {
		LastUpdated time.Time                      `json:"last_updated"`
		Commits     map[string]map[string]string   `json:"commits"`
	}{
		LastUpdated: m.data.LastUpdated,
		Commits:     m.data.Commits,
	}

	data, err = json.MarshalIndent(commitData, "", "  ")
	if err != nil {
		return fmt.Errorf("コミットキャッシュのJSONエンコードに失敗: %w", err)
	}

	if err := os.WriteFile(m.commitsFile, data, 0644); err != nil {
		return fmt.Errorf("コミットキャッシュファイルの保存に失敗: %w", err)
	}

	return nil
}

// IsNotificationSeen 通知が既に表示済みかチェックします
func (m *Manager) IsNotificationSeen(notificationID string) bool {
	m.mutex.RLock()
	defer m.mutex.RUnlock()

	for _, notifCache := range m.data.Notifications {
		if notifCache.SeenIDs[notificationID] {
			return true
		}
	}
	return false
}

// MarkNotificationSeen 通知を表示済みとしてマークします
func (m *Manager) MarkNotificationSeen(notificationID string) {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	// "default" キーで通知を管理
	key := "default"
	if _, exists := m.data.Notifications[key]; !exists {
		m.data.Notifications[key] = NotificationCache{
			SeenIDs:   make(map[string]bool),
			LastCheck: time.Now(),
		}
	}

	notifCache := m.data.Notifications[key]
	notifCache.SeenIDs[notificationID] = true
	notifCache.LastCheck = time.Now()
	m.data.Notifications[key] = notifCache
}

// GetLastCommitSHA リポジトリ・ブランチの最後のコミットSHAを取得します
func (m *Manager) GetLastCommitSHA(repo, branch string) string {
	m.mutex.RLock()
	defer m.mutex.RUnlock()

	if repoCache, exists := m.data.Commits[repo]; exists {
		return repoCache[branch]
	}
	return ""
}

// SetLastCommitSHA リポジトリ・ブランチの最後のコミットSHAを設定します
func (m *Manager) SetLastCommitSHA(repo, branch, sha string) {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	if _, exists := m.data.Commits[repo]; !exists {
		m.data.Commits[repo] = make(map[string]string)
	}
	m.data.Commits[repo][branch] = sha
}

// CleanupOldNotifications 古い通知キャッシュをクリーンアップします
func (m *Manager) CleanupOldNotifications(maxAge time.Duration) {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	cutoff := time.Now().Add(-maxAge)

	for key, notifCache := range m.data.Notifications {
		if notifCache.LastCheck.Before(cutoff) {
			delete(m.data.Notifications, key)
		}
	}
}

// GetCacheStats キャッシュの統計情報を取得します
func (m *Manager) GetCacheStats() map[string]interface{} {
	m.mutex.RLock()
	defer m.mutex.RUnlock()

	notificationCount := 0
	for _, notifCache := range m.data.Notifications {
		notificationCount += len(notifCache.SeenIDs)
	}

	repoCount := len(m.data.Commits)
	branchCount := 0
	for _, branches := range m.data.Commits {
		branchCount += len(branches)
	}

	return map[string]interface{}{
		"last_updated":       m.data.LastUpdated,
		"notification_count": notificationCount,
		"repo_count":         repoCount,
		"branch_count":       branchCount,
		"cache_dir":          m.cacheDir,
	}
}