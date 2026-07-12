# Tech Stack

## 语言 & 框架

| 技术 | 版本 | 用途 |
|------|------|------|
| Go | 1.25.0 | 后端 API 服务器 |
| Vue | 3.5.x | 前端 UI 框架 |
| TypeScript | 6.0.x | 前端类型系统 |
| Vite | 8.1.x | 前端构建工具 |
| Vue Router | 4.6.x | 前端路由 (hash mode) |

## 关键依赖

| 包 | 版本 | 理由 |
|---|------|------|
| lib/pq | 1.12.3 | PostgreSQL 驱动，纯 Go、稳定 |
| Vant | 4.9.x | 移动端 UI 组件库（轻量） |
| @tabler/icons-vue | 3.44.x | 图标库，替代 Feather Icons |

## 基础设施

| 组件 | 选择 | 备注 |
|------|------|------|
| 数据库 | PostgreSQL 14+ | 主数据存储 |
| 后端端口 | :8090 | HTTP 直接服务 |
| 前端端口 | :5173 (dev) | Vite dev server，代理 /api → :8090 |
| 部署 | systemd + 静态文件 | 后端编译二进制，前端 build 到 /var/www |

## 架构决策

### 无 ORM
- **决策**: 使用 `database/sql` + 原生 SQL 查询
- **理由**: 项目规模小、查询简单，无需 ORM 额外复杂度；`lib/pq` 直接操作 PostgreSQL

### 无登录使用
- **决策**: 核心功能（球员管理、比赛、排名）无需登录
- **理由**: 降低使用门槛，适合球友圈快速上手
- **例外**: 管理后台需登录；短信验证登录用于"我的"功能

### 前端 Hash 路由
- **决策**: `createWebHashHistory`
- **理由**: 单页应用部署为静态文件，避免服务端路由配置

### 活动积分冻结
- **决策**: 活动创建时快照冻结参赛球员积分
- **理由**: 确保活动内所有比赛使用同一基准分，公平一致

## 开发工具

| 工具 | 用途 | 配置 |
|------|------|------|
| go run . | 后端开发运行 | backend/ |
| npm run dev | 前端开发运行 | frontend/ |
| npm run build | 前端构建 | 输出到 frontend/dist/ |
| vue-tsc | 前端类型检查 | npm run build:check |
