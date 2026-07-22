# 锦标赛功能部署指南

## 部署步骤

在 65.49.218.149 生产服务器上执行：

```bash
# 1. 拉取最新代码
cd /usr/local/ClaudeProjects/PingCloud
git pull origin main

# 2. 数据库迁移
PGPASSWORD=你的数据库密码 psql -h 127.0.0.1 -U postgres -d pingpong \
  -f backend/db/migrations/010_tournament.sql

# 3. 编译后端
cd backend
go build -o pingpong .

# 4. 构建前端
cd ../frontend
npm install   # 如有新依赖
npx vite build

# 5. 创建/更新 systemd 服务
sudo tee /etc/systemd/system/pingcloud.service <<'EOF'
[Unit]
Description=PingCloud PingPong HTTP Server
After=network.target postgresql.service

[Service]
Type=simple
WorkingDirectory=/usr/local/ClaudeProjects/PingCloud/backend
ExecStart=/usr/local/ClaudeProjects/PingCloud/backend/pingpong
Restart=always
RestartSec=5
User=root
Group=root
Environment=DB_HOST=127.0.0.1
Environment=DB_PORT=5432
Environment=DB_USER=postgres
Environment=DB_PASSWORD=你的数据库密码
Environment=DB_NAME=pingpong

[Install]
WantedBy=multi-user.target
EOF

# 6. 重启服务
sudo systemctl daemon-reload
sudo systemctl enable pingcloud
sudo systemctl restart pingcloud

# 7. 验证
curl -s -o /dev/null -w "Frontend: %{http_code}\n" http://localhost:8090/
curl -s -o /dev/null -w "API: %{http_code}\n" http://localhost:8090/api/tournaments
```

## 说明

- 后端现在内置了前端静态文件服务（从 `frontend/dist/` 目录），不再需要 nginx 单独 serve 前端
- 如果生产 nginx 配置有 proxy_pass 到后端，保持原样即可
- 数据库密码请替换为实际值（查看生产环境变量或 `.env` 文件）

## 新增文件清单

| 文件 | 说明 |
|---|---|
| `backend/handlers/tournament.go` | 锦标赛全部后端逻辑 |
| `backend/db/migrations/010_tournament.sql` | 6 张新表 |
| `backend/main.go` | 新增路由 + 静态文件服务 |
| `frontend/src/views/Tournament.vue` | 锦标赛主页（状态机） |
| `frontend/src/components/TournamentGroupView.vue` | 小组赛积分榜+赛程 |
| `frontend/src/components/TournamentKnockout.vue` | 淘汰赛对阵图 |
| `frontend/src/components/TournamentCardDraw.vue` | 趣味卡抽取 |
| `frontend/src/main.ts` | 新增 /tournament 路由 |
| `frontend/src/App.vue` | 新增底部"锦标赛"导航 |
