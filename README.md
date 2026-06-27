# 乒云 (PingCloud)

乒乓球积分排名系统 —— 开球网风格的轻量级开源替代。

## 功能

- **球员管理**：快速录入球员，支持自定义初始积分
- **组局比赛**：创建活动 → 选人 → 自动生成循环赛对阵表
- **积分排名**：USATT Elo 变体积分系统，实时排名
- **赛制灵活**：三局两胜 / 五局三胜 / 七局四胜
- **中场拉人**：比赛中途可加人，已打场次保留，未打场次重新编排
- **比分纠错**：活动结束前可修改已录入比分，自动回滚旧积分
- **移动端优先**：手机浏览器访问，无需安装
- **种子排序**：按积分自动排阵，高分选手最后相遇

## 技术栈

| 层 | 技术 |
|----|------|
| 后端 | Go + net/http |
| 前端 | Vue 3 + TypeScript + Vite |
| UI | Tabler Icons（MIT 协议 SVG） |
| 数据库 | PostgreSQL |
| 部署 | systemd + nginx |

## 快速开始

### 环境要求

- Go 1.22+
- Node.js 20+
- PostgreSQL 14+

### 1. 创建数据库

```sql
CREATE DATABASE pingpong;

CREATE TABLE players (
    id SERIAL PRIMARY KEY,
    name VARCHAR(50) NOT NULL,
    initial_rating INT NOT NULL DEFAULT 1500,
    current_rating INT NOT NULL DEFAULT 1500,
    created_at TIMESTAMP NOT NULL DEFAULT NOW()
);

CREATE TABLE matches (
    id SERIAL PRIMARY KEY,
    player_a_id INT NOT NULL REFERENCES players(id),
    player_b_id INT NOT NULL REFERENCES players(id),
    score_a INT,
    score_b INT,
    rating_change_a INT NOT NULL DEFAULT 0,
    rating_change_b INT NOT NULL DEFAULT 0,
    winner_id INT REFERENCES players(id),
    session_id INT REFERENCES sessions(id) ON DELETE SET NULL,
    round INT DEFAULT 0,
    played_at TIMESTAMP NOT NULL DEFAULT NOW()
);

CREATE TABLE sessions (
    id SERIAL PRIMARY KEY,
    name VARCHAR(100) NOT NULL DEFAULT '乒乒活动',
    status VARCHAR(20) DEFAULT 'active',
    created_at TIMESTAMP DEFAULT NOW()
);

CREATE TABLE session_players (
    session_id INT REFERENCES sessions(id) ON DELETE CASCADE,
    player_id INT REFERENCES players(id),
    PRIMARY KEY (session_id, player_id)
);
```

### 2. 启动后端

```bash
cd backend
export DB_HOST=127.0.0.1 DB_PORT=5432 DB_USER=postgres DB_PASSWORD=yourpass DB_NAME=pingpong
go run .
# API started on :8090
```

### 3. 启动前端

```bash
cd frontend
npm install
npm run dev
# Dev server on :5173, proxies /api to :8090
```

### 4. 生产部署

```bash
# 前端构建
cd frontend && npm run build
cp -r dist/* /var/www/pingpong/

# 后端编译
cd backend && go build -o pingpong-server .

# nginx 配置示例
server {
    listen 80;
    root /var/www/pingpong;
    index index.html;
    location /api/ { proxy_pass http://127.0.0.1:8090; }
    location / { try_files $uri /index.html; }
}

# systemd 服务
[Service]
ExecStart=/path/to/pingpong-server
Environment="DB_PASSWORD=yourpass"
```

## 积分算法

基于 USATT Elo 变体：

```
预期胜率 = 1 / (1 + 10^((对手分 - 自己分) / 400))
积分变化 = K × (实际结果 - 预期胜率)
```

K 因子动态调整：

| 分差 | K 值 | 说明 |
|------|------|------|
| 0-200 | 32 | 势均力敌 |
| 201-400 | 24 | 有差距 |
| 400+ | 16 | 差距大 |
| 冷门 | ×1.2 | 低分赢高分加权 |

## 项目结构

```
├── backend/
│   ├── main.go              # 服务入口
│   ├── db/db.go             # 数据库连接
│   ├── handlers/
│   │   ├── players.go       # 球员 CRUD
│   │   ├── matches.go       # 比赛记录
│   │   ├── sessions.go      # 活动管理
│   │   ├── rankings.go      # 积分排名
│   │   └── roundrobin.go    # 循环赛编排
│   ├── models/              # 数据模型
│   └── rating/elo.go        # USATT 积分算法
├── frontend/
│   ├── src/
│   │   ├── views/
│   │   │   ├── Home.vue         # 排名首页
│   │   │   ├── SessionView.vue  # 活动/组局
│   │   │   ├── AddPlayer.vue    # 添加球员
│   │   │   ├── RecordMatch.vue  # 快速记分
│   │   │   └── PlayerDetail.vue # 球员详情
│   │   ├── components/
│   │   │   ├── ScoreDialog.vue      # 比分录入弹窗
│   │   │   └── AddPlayerDialog.vue  # 拉人弹窗
│   │   ├── api.ts             # API 调用
│   │   ├── session-utils.ts   # 活动工具函数
│   │   └── style.css          # 全局样式
│   └── index.html
└── README.md
```

## 许可

MIT License
