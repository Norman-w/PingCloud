# 🏓 乒云 PingCloud

> 📱 乒乓球积分排名系统 —— 开球网风格的轻量级开源替代。

[![License](https://img.shields.io/badge/license-MIT-blue.svg)](LICENSE)
[![Go](https://img.shields.io/badge/Go-1.22+-00ADD8?logo=go)](https://go.dev/)
[![Vue](https://img.shields.io/badge/Vue-3.x-4FC08D?logo=vue.js)](https://vuejs.org/)
[![PostgreSQL](https://img.shields.io/badge/PostgreSQL-14+-4169E1?logo=postgresql)](https://www.postgresql.org/)

---

## ✨ 功能

| 功能 | 说明 |
|------|------|
| 👤 球员管理 | 快速录入，自定义初始积分（默认 1500），排名查看 |
| ⚔️ 组局比赛 | 创建活动 → 勾选参赛球员 → 自动生成循环赛对阵表 |
| 📊 实时排名 | 逐场录入比分后即时更新积分和排名 |
| 🎯 三档赛制 | 三局两胜 / 五局三胜 / 七局四胜，快捷比分按钮自动适配 |
| 🔄 中场拉人 | 活动进行中可添加新球员，已打场次不动，未打场次重新编排 |
| ✏️ 比分纠错 | 活动结束前可修改已录入比分，自动回滚旧积分变化 |
| 🌱 种子排序 | 按积分高低自动排阵，高分选手在最后几轮相遇 |
| 📱 移动端优先 | 手机浏览器直接访问，底部导航栏，触控友好 |
| 📋 积分明细 | 活动结束后显示每人初始分→变化→最终分，场次编号+选手编号 |

---

## 🧮 积分算法（USATT Elo 变体）

### 核心公式

```
预期胜率ₐ = 1 / ( 1 + 10^((R_b − R_a) / 400) )

积分变化ₐ = K × ( 实际结果ₐ − 预期胜率ₐ )

实际结果：赢 = 1.0，输 = 0.0
变化值四舍五入取整
```

### 预期胜率速查表

| 分差 | 高分预期胜率 | 低分预期胜率 |
|------|:----------:|:----------:|
| 0 分 | 50% | 50% |
| 50 分 | 57% | 43% |
| 100 分 | 64% | 36% |
| 150 分 | 70% | 30% |
| 200 分 | 76% | 24% |
| 250 分 | 81% | 19% |
| 300 分 | 85% | 15% |
| 400 分 | 91% | 9% |
| 500 分 | 95% | 5% |

> 📐 公式解释：分差 400 时，10^(400/400) = 10，所以高分胜率 = 1/(1+10) ≈ 9%。差越大，预期越高但永远达不到 100%。

### K 因子（动态调整）

| 分差范围 | K 值 | 说明 |
|----------|:----:|------|
| 0 – 200 | **32** | 势均力敌，波动充分 |
| 201 – 400 | **24** | 有差距，适当缩小波动 |
| 400+ | **16** | 差距悬殊，防止刷分 |
| 🎯 冷门加权 | **×1.2** | 低分赢高分（差 >100 分）时 K 上浮 20% |

### 实战举例

#### 同分段（1500 vs 1500）

| 结果 | 胜者变化 | 败者变化 |
|------|:------:|:------:|
| 高分赢 | **+16** | −16 |
| 低分赢 | **+16** | −16 |

#### 差 100 分（1600 vs 1500）

| 结果 | 胜者变化 | 败者变化 | 说明 |
|------|:------:|:------:|------|
| 1600 赢 | **+12** | −12 | 高分赢预期内，少加 |
| 1500 赢 | **+25** | −25 | 冷门！×1.2 = +25 |

#### 差 300 分（1800 vs 1500）

| 结果 | 胜者变化 | 败者变化 | 说明 |
|------|:------:|:------:|------|
| 1800 赢 | **+4** | −4 | 碾压局，几乎不加 |
| 1500 赢 | **+24** | −24 | 大冷门！×1.2 = +24 |

#### 差 500 分（2000 vs 1500）

| 结果 | 胜者变化 | 败者变化 | 说明 |
|------|:------:|:------:|------|
| 2000 赢 | **+1** | −1 | 毫无波澜 |
| 1500 赢 | **+18** | −18 | 惊天大冷！×1.2 = +18 |

### 赛前积分冻结机制

活动创建时，所有参赛选手的当前积分被**快照冻结**为「赛前积分」。活动内的所有比赛均使用冻结积分计算，**活动结束时一次性结算**到真实积分。

- 🔒 **公平一致**：无论先打后打，每场都用同一基准分计算
- 📊 **可预测**：开赛前就知道胜负对应的积分变化，不因顺序不同而变化
- 🎯 **适合业余球局**：今天的积分按今天来的时候算，打完统一更新

> 💡 单场快速记分（非活动模式）仍然采用逐场即时结算。

---

## 📖 使用说明

### 手机访问

浏览器打开 `http://你的服务器IP:8080`，底部三个 Tab：

| Tab | 功能 |
|-----|------|
| 🏆 **排名** | 总积分排行榜，点击球员看详情 |
| ⚔️ **活动** | 开新局、录比分、看排名 |
| 👤 **球员** | 添加新球员 |

### 完整流程

```
第一步：点「球员」→ 添加所有参赛球员
第二步：点「活动」→ 创建新活动 → 勾选球员 → 生成对阵表
第三步：按对阵表顺序 → 逐场点开录入比分 → 提交
第四步：全部打完 → 点「结束活动」→ 看最终排名和积分明细
```

### 中场拉人

活动进行中 → 点「+ 拉人加入本场」→ 选球员 → 确认

- ✅ 已打完的场次：保留不动
- 🔄 未打的场次：删除后重新编排（所有选手循环赛）

### 比分纠错

活动未结束时 → 点已录入的比赛 → 修改比分 → 提交

- 🔙 旧的积分变化自动回滚
- ✅ 新的积分变化重新计算

---

## 🛠️ 开发说明

### 环境要求

- Go 1.22+
- Node.js 20+
- PostgreSQL 14+

### 本地开发

```bash
# 1. 安装依赖
cd frontend && npm install

# 2. 创建数据库（见下方 SQL）

# 3. 启动后端（终端 A）
cd backend
export DB_PASSWORD=yourpass
go run .
# → http://localhost:8090

# 4. 启动前端（终端 B）
cd frontend
npm run dev
# → http://localhost:5173（自动代理 /api 到 :8090）
```

### 数据库建表

```sql
CREATE DATABASE pingpong;

CREATE TABLE players (
    id SERIAL PRIMARY KEY,
    name VARCHAR(50) NOT NULL,
    initial_rating INT NOT NULL DEFAULT 1500,
    current_rating INT NOT NULL DEFAULT 1500,
    created_at TIMESTAMP NOT NULL DEFAULT NOW()
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

CREATE INDEX idx_players_rating ON players(current_rating DESC);
CREATE INDEX idx_matches_played ON matches(played_at DESC);
```

### API 接口

| 方法 | 路径 | 说明 |
|------|------|------|
| `GET` | `/api/rankings` | 全排名（含胜率、场次） |
| `GET` | `/api/players` | 球员列表 |
| `POST` | `/api/players` | 添加球员 |
| `GET` | `/api/players/:id` | 球员详情+对战史 |
| `POST` | `/api/matches` | 记录单场比赛 |
| `GET` | `/api/matches` | 比赛列表 |
| `GET` | `/api/sessions` | 活动列表（含人数、场数） |
| `POST` | `/api/sessions` | 创建活动+自动生成循环赛 |
| `GET` | `/api/sessions/:id` | 活动详情（对阵表+排名） |
| `PUT` | `/api/sessions/:id` | 修改活动名称 |
| `POST` | `/api/sessions/:id/players` | 中场拉人 |
| `POST` | `/api/sessions/:id/matches/:mid` | 录入/纠正比分 |
| `POST` | `/api/sessions/:id/complete` | 结束活动 |

### 生产部署

```bash
# 后端
cd backend && go build -o pingpong-server .
sudo systemctl enable --now pingpong

# 前端
cd frontend && npm run build
cp -r dist/* /var/www/pingpong/

# systemd 服务文件
# [Service]
# ExecStart=/path/to/pingpong-server
# Environment="DB_PASSWORD=yourpass"
```

### 循环赛编排算法

采用 **Circle Method（圆桌轮转法）**：

- N 个选手固定一个位置，其余顺时针轮转
- 每轮 N/2 或 (N−1)/2 场（奇人数补空位 bye）
- 生成前按积分**从高到低排序**，确保高分选手在最后几轮相遇（种子制）

### 项目结构

```
├── backend/
│   ├── main.go              # HTTP 路由注册
│   ├── db/db.go             # PostgreSQL 连接
│   ├── handlers/
│   │   ├── players.go       # 球员 CRUD
│   │   ├── matches.go       # 比赛记录
│   │   ├── sessions.go      # 活动管理（创建/拉人/结束/纠错）
│   │   ├── rankings.go      # 积分排名
│   │   └── roundrobin.go    # 循环赛编排+种子排序
│   ├── models/              # 数据模型
│   └── rating/elo.go        # USATT 积分计算
├── frontend/
│   ├── src/
│   │   ├── views/
│   │   │   ├── Home.vue         # 🏆 排名首页
│   │   │   ├── SessionView.vue  # ⚔️ 活动/组局（四步流程）
│   │   │   ├── AddPlayer.vue    # 👤 添加球员
│   │   │   ├── RecordMatch.vue  # ⚡ 快速单场记分
│   │   │   └── PlayerDetail.vue # 📋 球员详情
│   │   ├── components/
│   │   │   ├── ScoreDialog.vue      # 比分录入/纠错弹窗
│   │   │   └── AddPlayerDialog.vue  # 拉人弹窗
│   │   ├── api.ts             # API 封装
│   │   ├── session-utils.ts   # 活动工具函数
│   │   └── style.css          # 全局设计系统
│   └── index.html
└── README.md
```

---

## 📄 许可

MIT License — 自由使用、修改、分发。
