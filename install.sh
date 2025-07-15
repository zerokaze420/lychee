#!/bin/bash

# install.sh - 為 lychee 專案在 Ubuntu 上進行安裝和服務設置的腳本
#
# 功能:
# 1. 安裝依賴 (git, curl, Nix)
# 2. 創建專用系統用戶
# 3. 克隆源碼並使用 Nix Flakes 構建
# 4. 安裝二進制文件
# 5. 設置、啟用並啟動 systemd 服務

# --- 配置 ---
set -e  # 如果任何命令失敗，立即退出
set -u  # 如果使用未定義的變量，立即退出

APP_NAME="lychee"
GIT_REPO="https://github.com/hashcowuwu/lychee.git"
SERVICE_USER="lychee"
INSTALL_SRC_DIR="/opt/${APP_NAME}-src"
INSTALL_BIN_DIR="/usr/local/bin"
SYSTEMD_FILE_PATH="/etc/systemd/system/${APP_NAME}.service"

# --- 腳本開始 ---

# 1. 權限檢查
if [ "$(id -u)" -ne 0 ]; then
  echo "錯誤：此腳本需要以 root 權限運行。請使用 'sudo ./install.sh'"
  exit 1
fi

echo "--- lychee 監控工具安裝程序 ---"

# 2. 安裝系統依賴
echo "正在更新包列表並安裝依賴 (git, curl)..."
apt-get update >/dev/null
apt-get install -y git curl >/dev/null

# 3. 安裝 Nix 包管理器 (如果需要)
if ! command -v nix &> /dev/null; then
  echo "Nix 未安裝。正在使用官方腳本進行多用戶安裝..."
  # 以非交互模式運行 Nix 安裝程序
  sh <(curl -L https://nixos.org/nix/install) --daemon
  
  # 使 Nix 命令在當前 shell 會話中可用
  # 注意：如果安裝後立即在同一個腳本中使用 nix，需要手動 source profile
  if [ -e '/nix/var/nix/profiles/default/etc/profile.d/nix-daemon.sh' ]; then
    . '/nix/var/nix/profiles/default/etc/profile.d/nix-daemon.sh'
  fi
  echo "Nix 安裝完成。"
else
  echo "Nix 已安裝，跳過安裝步驟。"
  # 確保 Nix profile 被加載
  if [ -e '/nix/var/nix/profiles/default/etc/profile.d/nix-daemon.sh' ]; then
    . '/nix/var/nix/profiles/default/etc/profile.d/nix-daemon.sh'
  fi
fi

# 啟用 Nix Flakes 功能
# 檢查 nix.conf 是否存在，並添加配置
NIX_CONF_DIR="/etc/nix"
NIX_CONF_FILE="${NIX_CONF_DIR}/nix.conf"
mkdir -p "$NIX_CONF_DIR"
touch "$NIX_CONF_FILE"
if ! grep -q "experimental-features" "$NIX_CONF_FILE"; then
    echo "正在啟用 Nix Flakes 功能..."
    echo "experimental-features = nix-command flakes" >> "$NIX_CONF_FILE"
fi


# 4. 創建服務用戶
if ! id -u "$SERVICE_USER" >/dev/null 2>&1; then
  echo "正在創建系統用戶 '$SERVICE_USER'..."
  useradd --system --no-create-home --shell /bin/false "$SERVICE_USER"
else
  echo "系統用戶 '$SERVICE_USER' 已存在。"
fi

# 5. 克隆並構建項目
echo "正在從 GitHub 克隆源碼..."
rm -rf "$INSTALL_SRC_DIR"
git clone "$GIT_REPO" "$INSTALL_SRC_DIR"
cd "$INSTALL_SRC_DIR"

echo "正在使用 Nix Flakes 進行構建（這可能需要一些時間）..."
# 使用 `nix build`，它會讀取 flake.nix 和 flake.lock
# 這裡的 .#lychee 是指定要構建 flake 中的 'lychee' 輸出
nix build .#lychee

# 6. 安裝構建好的二進制文件
echo "正在安裝 '${APP_NAME}' 到 ${INSTALL_BIN_DIR}..."
cp -f ./result/bin/${APP_NAME} "${INSTALL_BIN_DIR}/${APP_NAME}"
chown root:root "${INSTALL_BIN_DIR}/${APP_NAME}"
chmod 755 "${INSTALL_BIN_DIR}/${APP_NAME}"

# 7. 創建 systemd 服務文件
echo "正在創建 systemd 服務文件..."
cat <<EOF > "$SYSTEMD_FILE_PATH"
[Unit]
Description=Lychee Monitoring Tool
Documentation=https://github.com/hashcowuwu/lychee
After=network.target

[Service]
# 使用我們創建的專用用戶運行
User=${SERVICE_USER}
Group=${SERVICE_USER}

# 服務的類型和啟動命令
Type=simple
ExecStart=${INSTALL_BIN_DIR}/${APP_NAME}

# 日誌將會被重定向到 systemd-journald
StandardOutput=journal
StandardError=journal

# 自動重啟策略
Restart=on-failure
RestartSec=5

[Install]
WantedBy=multi-user.target
EOF

# 8. 啟用並啟動服務
echo "正在重載 systemd，並啟用/啟動 '${APP_NAME}' 服務..."
systemctl daemon-reload
systemctl enable "${APP_NAME}.service"
systemctl start "${APP_NAME}.service"

# 9. 清理
echo "正在清理臨時的源碼文件..."
rm -rf "$INSTALL_SRC_DIR"

echo ""
echo "--- 安裝完成 ---"
echo "Lychee 服務已成功安裝並正在運行。"
echo ""
echo "常用命令:"
echo "  查看服務狀態: sudo systemctl status ${APP_NAME}"
echo "  查看服務日誌: sudo journalctl -u ${APP_NAME} -f"
echo "  停止服務:     sudo systemctl stop ${APP_NAME}"
echo "  啟動服務:     sudo systemctl start ${APP_NAME}"
echo ""