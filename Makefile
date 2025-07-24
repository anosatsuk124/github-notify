# GitHub Notify Makefile

.PHONY: build install clean test run deps help

# デフォルトターゲット
help:
	@echo "利用可能なコマンド:"
	@echo "  make build    - バイナリをビルド"
	@echo "  make install  - バイナリをインストール"
	@echo "  make test     - テストを実行"
	@echo "  make run      - プログラムを実行"
	@echo "  make deps     - 依存関係をダウンロード"
	@echo "  make clean    - ビルドファイルを削除"

# 依存関係のダウンロード
deps:
	go mod download
	go mod tidy

# ビルド
build: deps
	go build -o bin/github-notify ./cmd/github-notify

# インストール (ホームディレクトリの .local/bin にインストール)
install: build
	mkdir -p $(HOME)/.local/bin
	cp bin/github-notify $(HOME)/.local/bin/
	@echo "github-notify を $(HOME)/.local/bin/ にインストールしました"
	@echo "PATH に $(HOME)/.local/bin が含まれていることを確認してください"

# システム全体にインストール (root権限が必要)
install-system: build
	sudo cp bin/github-notify /usr/local/bin/
	@echo "github-notify を /usr/local/bin/ にインストールしました"

# テスト実行
test:
	go test -v ./...

# プログラム実行
run: build
	./bin/github-notify

# テスト通知送信
test-notify: build
	./bin/github-notify --test

# クリーンアップ
clean:
	rm -rf bin/
	go clean

# 開発用ビルド (デバッグ情報付き)
build-dev: deps
	go build -gcflags="all=-N -l" -o bin/github-notify-dev ./cmd/github-notify

# リリース用ビルド (複数プラットフォーム)
build-release: deps
	mkdir -p bin/release
	# Linux
	GOOS=linux GOARCH=amd64 go build -ldflags="-s -w" -o bin/release/github-notify-linux-amd64 ./cmd/github-notify
	# macOS
	GOOS=darwin GOARCH=amd64 go build -ldflags="-s -w" -o bin/release/github-notify-darwin-amd64 ./cmd/github-notify
	GOOS=darwin GOARCH=arm64 go build -ldflags="-s -w" -o bin/release/github-notify-darwin-arm64 ./cmd/github-notify
	# Windows
	GOOS=windows GOARCH=amd64 go build -ldflags="-s -w" -o bin/release/github-notify-windows-amd64.exe ./cmd/github-notify
	@echo "リリースバイナリを bin/release/ に作成しました"

# 設定ファイルの場所を表示
config-info:
	@echo "設定ファイルの場所:"
	@echo "  デフォルト: $(HOME)/.config/github-notify/config.yaml"
	@echo "  例: config.example.yaml"

# systemdタイマーのセットアップ（Linux用）
setup-systemd:
	@echo "systemdタイマーのセットアップ"
	mkdir -p $(HOME)/.config/systemd/user
	@echo "[Unit]" > $(HOME)/.config/systemd/user/github-notify.service
	@echo "Description=GitHub Notification Monitor" >> $(HOME)/.config/systemd/user/github-notify.service
	@echo "" >> $(HOME)/.config/systemd/user/github-notify.service
	@echo "[Service]" >> $(HOME)/.config/systemd/user/github-notify.service
	@echo "Type=oneshot" >> $(HOME)/.config/systemd/user/github-notify.service
	@echo "ExecStart=$(HOME)/.local/bin/github-notify" >> $(HOME)/.config/systemd/user/github-notify.service
	@echo "" >> $(HOME)/.config/systemd/user/github-notify.service
	@echo "[Unit]" > $(HOME)/.config/systemd/user/github-notify.timer
	@echo "Description=GitHub Notification Monitor Timer" >> $(HOME)/.config/systemd/user/github-notify.timer
	@echo "" >> $(HOME)/.config/systemd/user/github-notify.timer
	@echo "[Timer]" >> $(HOME)/.config/systemd/user/github-notify.timer
	@echo "OnCalendar=*:0/5" >> $(HOME)/.config/systemd/user/github-notify.timer
	@echo "Persistent=true" >> $(HOME)/.config/systemd/user/github-notify.timer
	@echo "" >> $(HOME)/.config/systemd/user/github-notify.timer
	@echo "[Install]" >> $(HOME)/.config/systemd/user/github-notify.timer
	@echo "WantedBy=timers.target" >> $(HOME)/.config/systemd/user/github-notify.timer
	systemctl --user daemon-reload
	@echo "systemdタイマーを作成しました"
	@echo "有効化: systemctl --user enable --now github-notify.timer"
	@echo "状態確認: systemctl --user status github-notify.timer"