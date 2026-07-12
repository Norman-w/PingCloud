# Workflow

## 开发方法论

轻量级增量开发，无需 TDD，但每个功能完成后需手动验证。

## Git 约定

- **分支**: `main` 直接提交（个人项目）
- **提交格式**: `emoji 简短描述`，如 `🏓 车轮战并列排名`
- **提交用户**: Norman Wang

## 质量门禁

| 门禁 | 要求 |
|------|------|
| 编译 | Go 编译通过 (`go build .`) |
| 类型检查 | `vue-tsc -b` 无错误 |
| 手动测试 | 浏览器端验证关键流程 |
| 风格一致性 | 遵循现有代码风格和设计系统 |

## 后端开发模式

### 新增功能步骤
1. `models/` — 定义数据结构（Go struct）
2. `handlers/` — 实现 HTTP handler（CRUD）
3. `main.go` — 注册路由
4. 数据库 — 执行建表 SQL

### Handler 规范
- 使用 `db.DB.Query` / `db.DB.QueryRow` 直接查询
- 使用 `writeJSON(w, v)` 统一 JSON 响应
- CORS 通过 `cors(w)` 函数处理
- 错误返回 `http.Error(w, msg, code)`

## 前端开发模式

### 新增页面步骤
1. `views/NewPage.vue` — 创建视图组件
2. `main.ts` — 注册路由
3. `App.vue` — 如需底部 Tab 入口，添加 tab 配置
4. `api.ts` — 如需新 API 接口，添加类型和请求方法

### Vue 组件规范
- `<script setup lang="ts">` 语法
- 复用 `style.css` 中的设计系统（CSS 变量、card、stats-row 等）
- 使用 Vant 组件库（van-button、van-cell、van-popup 等）
- 图标使用 `@tabler/icons-vue`

## 部署

```bash
# 后端
cd backend && go build -o pingpong-server .
sudo systemctl restart pingpong

# 前端
cd frontend && npm run build
cp -r dist/* /var/www/pingpong/
```
