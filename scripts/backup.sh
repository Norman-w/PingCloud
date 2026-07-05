#!/bin/bash
# PingCloud 一键备份/恢复/迁移脚本
# 用法: bash scripts/backup.sh

set -e
SCRIPT_DIR="$(cd "$(dirname "$0")" && pwd)"
PROJECT_DIR="$(dirname "$SCRIPT_DIR")"
BACKUP_DIR="${PROJECT_DIR}/backups"
TIMESTAMP=$(date +%Y%m%d_%H%M%S)
DB_HOST="${DB_HOST:-127.0.0.1}"
DB_PORT="${DB_PORT:-5432}"
DB_USER="${DB_USER:-postgres}"
DB_NAME="${DB_NAME:-pingpong}"
ENV_FILE="${PROJECT_DIR}/backend/.env"

RED='\033[0;31m'; GREEN='\033[0;32m'; YELLOW='\033[1;33m'; BLUE='\033[0;34m'; NC='\033[0m'
info()  { echo -e "${BLUE}[*]${NC} $1"; }
ok()    { echo -e "${GREEN}[✓]${NC} $1"; }
warn()  { echo -e "${YELLOW}[!]${NC} $1"; }
err()   { echo -e "${RED}[✗]${NC} $1"; }

choose() {
  echo ""
  echo -e "${BLUE}========================================${NC}"
  echo -e "${BLUE}   🏓 PingCloud 数据管理中心${NC}"
  echo -e "${BLUE}========================================${NC}"
  echo ""
  echo "  1) 💾 备份全部数据"
  echo "  2) 📥 恢复数据（从本地备份）"
  echo "  3) 📤 一键迁移（打包传输用）"
  echo "  4) 📋 查看备份列表"
  echo "  5) 🗑️  清理旧备份"
  echo "  0) 退出"
  echo ""
  read -p "  请选择 [0-5]: " choice
  case $choice in
    1) do_backup ;;
    2) do_restore ;;
    3) do_migrate ;;
    4) list_backups ;;
    5) clean_backups ;;
    0) exit 0 ;;
    *) err "无效选择"; choose ;;
  esac
}

do_backup() {
  mkdir -p "$BACKUP_DIR"
  local NAME="pingcloud_${TIMESTAMP}"
  local DIR="${BACKUP_DIR}/${NAME}"
  mkdir -p "$DIR"

  info "开始备份..."

  # 1. PostgreSQL 全库导出
  info "导出数据库 ${DB_NAME}..."
  PGPASSWORD="${DB_PASSWORD}" pg_dump -h "$DB_HOST" -p "$DB_PORT" -U "$DB_USER" -d "$DB_NAME" \
    --no-owner --no-acl --clean --if-exists \
    > "${DIR}/database.sql" 2>/dev/null
  ok "数据库已导出 ($(du -h "${DIR}/database.sql" | cut -f1))"

  # 2. 环境配置
  if [ -f "$ENV_FILE" ]; then
    cp "$ENV_FILE" "${DIR}/.env"
    ok ".env 已备份"
  fi

  # 3. Nginx 配置
  if [ -f /etc/nginx/sites-enabled/pingpong.conf ]; then
    cp /etc/nginx/sites-enabled/pingpong.conf "${DIR}/nginx.conf"
    ok "Nginx 配置已备份"
  fi

  # 4. Systemd 服务
  if [ -f /etc/systemd/system/pingpong.service ]; then
    cp /etc/systemd/system/pingpong.service "${DIR}/pingpong.service"
    ok "Systemd 服务已备份"
  fi

  # 5. 系统环境变量 (SMS等)
  env | grep -E 'ALIBABA_ACCESS_KEY|ALIBABA_ACCESS_SECRET|ALIBABA_SMS_SIGN|ALIBABA_SMS_TEMPLATE|DB_' > "${DIR}/sys_env.txt" 2>/dev/null
  if [ -s "${DIR}/sys_env.txt" ]; then ok "系统环境变量已备份"; fi

  # 6. Crontab 定时任务
  crontab -l > "${DIR}/crontab.txt" 2>/dev/null
  if [ -s "${DIR}/crontab.txt" ]; then ok "Crontab 已备份"; fi

  # 7. 数据库连接信息
  cat > "${DIR}/restore.conf" << EOF
# 恢复配置 - 修改目标机器信息
DB_HOST=${DB_HOST}
DB_PORT=${DB_PORT}
DB_USER=${DB_USER}
DB_NAME=${DB_NAME}
BACKUP_DATE=${TIMESTAMP}
EOF
  ok "恢复配置已生成"

  # 6. 打包
  cd "$BACKUP_DIR"
  tar -czf "${NAME}.tar.gz" "$NAME" 2>/dev/null
  rm -rf "$NAME"
  ok "备份完成: ${BACKUP_DIR}/${NAME}.tar.gz ($(du -h "${BACKUP_DIR}/${NAME}.tar.gz" | cut -f1))"

  # 7. 保留最近 10 个备份
  ls -t "${BACKUP_DIR}"/*.tar.gz 2>/dev/null | tail -n +11 | xargs rm -f 2>/dev/null || true

  choose
}

do_restore() {
  local files=($(ls -t "${BACKUP_DIR}"/*.tar.gz 2>/dev/null))
  if [ ${#files[@]} -eq 0 ]; then
    err "没有找到备份文件"; choose; return
  fi

  echo ""
  echo "  可用备份:"
  for i in "${!files[@]}"; do
    echo "  $((i+1))) $(basename "${files[$i]}") ($(du -h "${files[$i]}" | cut -f1))"
  done
  echo ""
  read -p "  选择要恢复的备份 [1-${#files[@]}]: " idx
  if [ -z "$idx" ] || [ "$idx" -lt 1 ] || [ "$idx" -gt "${#files[@]}" ]; then
    err "无效选择"; choose; return
  fi

  local file="${files[$((idx-1))]}"
  echo ""
  warn "⚠  即将恢复数据库，当前数据将被覆盖！"
  read -p "  确认恢复？(输入 yes 继续): " confirm
  if [ "$confirm" != "yes" ]; then
    info "已取消"; choose; return
  fi

  info "解压备份..."
  local tmpdir=$(mktemp -d)
  tar -xzf "$file" -C "$tmpdir"

  # 尝试加载恢复配置覆盖默认值
  if [ -f "${tmpdir}/"*/restore.conf ]; then
    source "${tmpdir}/"*/restore.conf 2>/dev/null || true
  fi

  # 恢复数据库
  info "恢复数据库..."
  local sqlfile=$(find "$tmpdir" -name "database.sql" | head -1)
  if [ -f "$sqlfile" ]; then
    PGPASSWORD="${DB_PASSWORD}" psql -h "$DB_HOST" -p "$DB_PORT" -U "$DB_USER" -d "$DB_NAME" -f "$sqlfile" > /dev/null 2>&1
    ok "数据库已恢复"
  fi

  # 恢复 .env
  local envfile=$(find "$tmpdir" -name ".env" | head -1)
  if [ -f "$envfile" ]; then
    cp "$envfile" "$ENV_FILE"
    ok ".env 已恢复"
  fi

  rm -rf "$tmpdir"
  ok "恢复完成！"
  choose
}

do_migrate() {
  do_backup

  # 生成迁移脚本
  local latest=$(ls -t "${BACKUP_DIR}"/*.tar.gz 2>/dev/null | head -1)
  if [ -z "$latest" ]; then
    err "备份失败，无法生成迁移包"; choose; return
  fi

  local migrant="${BACKUP_DIR}/migrate_${TIMESTAMP}.sh"
  cat > "$migrant" << 'MIGRATE_SCRIPT'
#!/bin/bash
# PingCloud 一键迁移恢复脚本
# 在目标机器上执行: bash migrate_XXXXX.sh

set -e
RED='\033[0;31m'; GREEN='\033[0;32m'; BLUE='\033[0;34m'; NC='\033[0m'
info() { echo -e "${BLUE}[*]${NC} $1"; }
ok()   { echo -e "${GREEN}[✓]${NC} $1"; }

# ==== 配置区（按需修改） ====
DB_HOST="${DB_HOST:-127.0.0.1}"
DB_PORT="${DB_PORT:-5432}"
DB_USER="${DB_USER:-postgres}"
DB_NAME="${DB_NAME:-pingpong}"
PROJECT_DIR="${PROJECT_DIR:-/usr/local/ClaudeProjects/PingCloud}"
# =======================

# 自解压: 找到本脚本末尾的 tar.gz 数据
ARCHIVE=$(awk '/^__ARCHIVE__/ {print NR+1}' "$0")
tail -n +$ARCHIVE "$0" | tar -xz -C /tmp/

TMPDIR=$(ls -dt /tmp/pingcloud_* 2>/dev/null | head -1)
if [ -z "$TMPDIR" ]; then
  echo "解压失败"; exit 1
fi

info "开始迁移恢复..."

# 1. 安装依赖
info "检查环境..."
command -v psql >/dev/null 2>&1 || { echo "请先安装 PostgreSQL 客户端"; exit 1; }
command -v go >/dev/null 2>&1 || { echo "请先安装 Go (用于编译后端)"; }
command -v node >/dev/null 2>&1 || { echo "请先安装 Node.js (用于编译前端)"; }

# 2. 创建数据库
info "创建数据库 ${DB_NAME}..."
sudo -u postgres createdb "$DB_NAME" 2>/dev/null || info "数据库已存在，跳过创建"

# 3. 恢复数据
if [ -f "${TMPDIR}/database.sql" ]; then
  info "恢复数据库..."
  PGPASSWORD="${DB_PASSWORD}" psql -h "$DB_HOST" -p "$DB_PORT" -U "$DB_USER" -d "$DB_NAME" -f "${TMPDIR}/database.sql"
  ok "数据库已恢复"
fi

# 4. 恢复配置
if [ -f "${TMPDIR}/.env" ]; then
  mkdir -p "${PROJECT_DIR}/backend"
  cp "${TMPDIR}/.env" "${PROJECT_DIR}/backend/.env"
  ok ".env 已恢复"
fi

# 5. Nginx
if [ -f "${TMPDIR}/nginx.conf" ]; then
  sudo cp "${TMPDIR}/nginx.conf" /etc/nginx/sites-enabled/pingpong.conf
  sudo nginx -t && sudo systemctl reload nginx
  ok "Nginx 已配置"
fi

# 6. Systemd
if [ -f "${TMPDIR}/pingpong.service" ]; then
  sudo cp "${TMPDIR}/pingpong.service" /etc/systemd/system/pingpong.service
  sudo systemctl daemon-reload
  sudo systemctl enable pingpong
  ok "Systemd 服务已注册"
fi

# 7. 编译启动
info "编译后端..."
cd "${PROJECT_DIR}/backend" && go build -o pingpong-server .
info "编译前端..."
cd "${PROJECT_DIR}/frontend" && npm install --silent && npm run build

# 8. 部署前端
sudo rsync -aq --delete "${PROJECT_DIR}/frontend/dist/" /var/www/pingpong/

# 9. 启动
sudo systemctl restart pingpong
ok "迁移完成！访问 https://你的域名 查看"

rm -rf "$TMPDIR"
exit 0
__ARCHIVE__
MIGRATE_SCRIPT

  # 追加备份数据到脚本末尾
  cat "$latest" >> "$migrant"
  chmod +x "$migrant"

  ok "迁移脚本已生成: ${migrant}"
  ok "将此文件传输到目标机器，执行 bash $(basename "$migrant") 即可一键迁移"
  choose
}

list_backups() {
  echo ""
  if [ ! -d "$BACKUP_DIR" ] || [ -z "$(ls -A "$BACKUP_DIR" 2>/dev/null)" ]; then
    info "暂无备份"; choose; return
  fi
  echo "  备份列表 (${BACKUP_DIR}/):"
  echo "  ──────────────────────────────────────"
  local total=0
  for f in "${BACKUP_DIR}"/*.tar.gz; do
    [ -f "$f" ] || continue
    local size=$(du -h "$f" | cut -f1)
    local name=$(basename "$f" .tar.gz)
    echo "  📦 ${name}  (${size})"
    total=$((total+1))
  done
  echo "  ──────────────────────────────────────"
  echo "  共 ${total} 个备份"
  choose
}

clean_backups() {
  local count=$(ls "${BACKUP_DIR}"/*.tar.gz 2>/dev/null | wc -l)
  if [ "$count" -eq 0 ]; then
    info "没有可清理的备份"; choose; return
  fi
  echo ""
  read -p "  保留最近几个备份？[默认 5]: " keep
  keep=${keep:-5}
  ls -t "${BACKUP_DIR}"/*.tar.gz 2>/dev/null | tail -n +$((keep+1)) | xargs rm -f 2>/dev/null || true
  ok "已清理，保留最近 ${keep} 个备份"
  choose
}

# 加载 .env
if [ -f "$ENV_FILE" ]; then
  set -a; source "$ENV_FILE" 2>/dev/null; set +a
fi

choose
