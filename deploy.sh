#!/bin/bash
# ==============================================
# 🏓 PingCloud 一键部署脚本
# 用法: bash deploy.sh
# 功能: 备份数据库 → 同步代码 → 编译 → 部署前端 → 重启服务
# ==============================================
set -e

RED='\033[0;31m'; GREEN='\033[0;32m'; YELLOW='\033[1;33m'; BLUE='\033[0;34m'; NC='\033[0m'
info()  { echo -e "${BLUE}[*]${NC} $1"; }
ok()    { echo -e "${GREEN}[✓]${NC} $1"; }
warn()  { echo -e "${YELLOW}[!]${NC} $1"; }
err()   { echo -e "${RED}[✗]${NC} $1"; }

# ── Config ──
REMOTE="root@pingpang.norman.wang"
PROJECT_DIR="/usr/local/ClaudeProjects/PingCloud"
LOCAL_DIR="$(cd "$(dirname "$0")" && pwd)"
TIMESTAMP=$(date +%Y%m%d_%H%M%S)

echo ""
echo -e "${BLUE}╔══════════════════════════════════════╗${NC}"
echo -e "${BLUE}║   🏓 PingCloud 一键部署            ║${NC}"
echo -e "${BLUE}╚══════════════════════════════════════╝${NC}"
echo ""

# ── Step 0: Detect build location ──
info "检测构建环境..."
LOCAL_GO=$(go version 2>/dev/null | grep -oP 'go\K[0-9.]+' | head -1 || echo "0")
REMOTE_GO=$(ssh $REMOTE "go version 2>/dev/null" | grep -oP 'go\K[0-9.]+' | head -1 || echo "0")
LOCAL_NODE=$(node --version 2>/dev/null | grep -oP 'v\K[0-9]+' || echo "0")
REMOTE_NODE=$(ssh $REMOTE "node --version 2>/dev/null" | grep -oP 'v\K[0-9]+' || echo "0")

BUILD_LOCAL=false
if [ "$LOCAL_GO" != "0" ] && [ "$LOCAL_NODE" != "0" ]; then
  BUILD_LOCAL=true
  ok "本地构建 (Go $LOCAL_GO, Node $LOCAL_NODE)"
else
  ok "远程构建 (Go $REMOTE_GO, Node $REMOTE_NODE)"
fi

# ── Step 1: Backup database ──
info "备份生产数据库..."
ssh $REMOTE "source $PROJECT_DIR/backend/.env 2>/dev/null
  BACKUP_FILE=/tmp/pingcloud_backup_${TIMESTAMP}.sql
  sudo -u postgres pg_dump -d pingpong --no-owner --no-acl --clean --if-exists > \$BACKUP_FILE 2>&1
  echo \$BACKUP_FILE"
BACKUP_FILE=$(ssh $REMOTE "ls -t /tmp/pingcloud_backup_*.sql 2>/dev/null | head -1")
if [ -n "$BACKUP_FILE" ]; then
  BACKUP_SIZE=$(ssh $REMOTE "du -h $BACKUP_FILE | cut -f1")
  ok "数据库已备份: $BACKUP_SIZE ($BACKUP_FILE)"
else
  err "备份失败！"
  exit 1
fi

# ── Step 2: Sync source code ──
info "同步源代码..."
rsync -az --delete \
  --exclude='.git' \
  --exclude='node_modules' \
  --exclude='dist' \
  --exclude='backups' \
  --exclude='pingpong-server' \
  --exclude='.env' \
  "$LOCAL_DIR/" "$REMOTE:$PROJECT_DIR/" 2>&1 | tail -1
ok "代码已同步"

# ── Step 3: Run database migrations ──
info "执行数据库迁移..."
MIGRATIONS=$(ssh $REMOTE "ls $PROJECT_DIR/backend/db/migrations/*.sql 2>/dev/null" || echo "")
if [ -n "$MIGRATIONS" ]; then
  for f in $MIGRATIONS; do
    MIG_NAME=$(basename "$f")
    info "  执行 $MIG_NAME..."
    ssh $REMOTE "sudo -u postgres psql -d pingpong -f $f 2>&1" | while read line; do
      echo "    $line"
    done
  done
  ok "迁移完成"
else
  info "无新迁移"
fi

# ── Step 4: Build ──
if $BUILD_LOCAL; then
  # ── Local build ──
  info "本地编译后端..."
  cd "$LOCAL_DIR/backend"
  GOPROXY=https://goproxy.cn,direct go build -o pingpong-server . 2>&1
  ok "后端编译完成"

  info "本地构建前端..."
  cd "$LOCAL_DIR/frontend"
  npm run build 2>&1 | tail -3
  ok "前端构建完成"

  # Upload binaries (stop service first to allow overwrite)
  info "上传构建产物..."
  ssh $REMOTE "systemctl stop pingpong" 2>&1
  scp "$LOCAL_DIR/backend/pingpong-server" "$REMOTE:$PROJECT_DIR/backend/pingpong-server" 2>&1
  rsync -az --delete "$LOCAL_DIR/frontend/dist/" "$REMOTE:$PROJECT_DIR/frontend/dist/" 2>&1
  ok "构建产物已上传"
else
  # ── Remote build ──
  info "远程编译后端..."
  ssh $REMOTE "cd $PROJECT_DIR/backend && GOPROXY=https://goproxy.cn,direct go build -o pingpong-server . 2>&1"
  ok "后端编译完成"

  info "远程构建前端..."
  ssh $REMOTE "cd $PROJECT_DIR/frontend && npm run build 2>&1" | tail -5
  ok "前端构建完成"
fi

# ── Step 5: Deploy frontend static files ──
info "部署前端静态文件..."
ssh $REMOTE "rsync -aq --delete $PROJECT_DIR/frontend/dist/ /var/www/pingpong/ 2>&1"
ok "前端已部署"

# ── Step 6: Restart service ──
info "重启服务..."
ssh $REMOTE "systemctl restart pingpong && sleep 1 && systemctl is-active pingpong" 2>&1
ok "服务已重启"

# ── Step 7: Health check ──
info "健康检查..."
HEALTH=$(ssh $REMOTE "curl -s http://127.0.0.1:8090/api/health 2>&1")
if echo "$HEALTH" | grep -q '"ok"'; then
  ok "健康检查通过: $HEALTH"
else
  warn "健康检查异常: $HEALTH"
fi

# ── Step 8: Cleanup old backups (keep last 10) ──
ssh $REMOTE "ls -t /tmp/pingcloud_backup_*.sql 2>/dev/null | tail -n +11 | xargs rm -f 2>/dev/null || true"

echo ""
echo -e "${GREEN}╔══════════════════════════════════════╗${NC}"
echo -e "${GREEN}║   ✅ 部署完成！                    ║${NC}"
echo -e "${GREEN}║   🌐 https://pingpang.norman.wang   ║${NC}"
echo -e "${GREEN}╚══════════════════════════════════════╝${NC}"
echo ""
