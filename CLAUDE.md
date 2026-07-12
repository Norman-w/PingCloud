# PingCloud 项目开发指南

> 🏓 乒云 PingCloud — 乒乓球积分排名系统（Go 后端 + Vue 3 前端 + PostgreSQL）

## 项目结构

```
backend/          — Go HTTP 服务器 (:8090)
  main.go         — 路由注册
  db/db.go        — PostgreSQL 连接
  models/         — 数据模型 struct
  handlers/       — HTTP handler（CRUD）
  rating/elo.go   — USATT Elo 积分计算
frontend/         — Vue 3 + Vite + TypeScript
  src/
    views/        — 页面级组件
    components/   — 可复用组件
    api.ts        — API 客户端封装
    style.css     — 全局设计系统（CSS 变量 + 全局样式）
    main.ts       — 路由定义 + App 挂载
conductor/        — 项目上下文文档
```

## 开发模式

### 后端 (Go)
- 无 ORM，使用 `database/sql` + 原生 SQL
- Handler 函数签名: `func HandlerName(w http.ResponseWriter, r *http.Request)`
- 统一 JSON 响应: `writeJSON(w, v)` (定义在 `handlers/players.go`)
- CORS: `cors(w)` 函数
- 路由注册在 `main.go`，模式为 `/api/<resource>/`

### 前端 (Vue 3)
- `<script setup lang="ts">` 语法
- 使用 Vant 4.x 移动端 UI 组件库
- 图标: `@tabler/icons-vue`
- 复用 `style.css` 的 CSS 变量和全局样式类
- 路由: Hash 模式 (`createWebHashHistory`)

### 数据库 (PostgreSQL)
- 表名: 小写下划线 (`training_logs`)
- 主键: `id SERIAL PRIMARY KEY`
- 时间: `TIMESTAMP DEFAULT NOW()`

## 当前任务

参见 [conductor/tracks.md](conductor/tracks.md) — P0: 练功记录页

## 代码风格

- Go: 标准库风格、英文注释、error 返回 `http.Error`
- Vue: Composition API、Tab 缩进、中文 UI 文案
- CSS: 使用 CSS 变量（`var(--c-primary)` 等）
