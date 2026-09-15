# FeishuShow

把飞书里的今天摊开：日程、发呆文档、待办，再对一场会一键出纪要。

不和豆包工作抢「会干活的 Agent」。FeishuShow 做的是官方闭源产品不强调的那一层——**先看清自己的飞书**，模型可选，数据默认只读本人权限。

- 后端：Go + Gin + SQLite
- 前端：Vue 3 + Vite + Pinia
- 没配飞书应用也能先看演示数据

## 今天能做什么

- **今日雷达**：今天的日程、未完成任务、最近文档、最近会话
- **发呆文档**：超过 30 天没改过的知识标出来
- **会后纪要**：选一场会，生成决议 / 待办 / 风险卡片
- **写回飞书**（需授权）：纪要落到云文档，待办落到任务中心
- **可选豆包**：配置 `DOUBAO_API_KEY` 后用模型润色纪要；不配则用本地模板

## 启动

需要本机已安装 Go 1.21+ 和 Node.js 18+。开两个终端：

**1. 启动后端**

```bash
cd backend
go mod tidy
go run .
```

API：http://localhost:8080

**2. 启动前端**

```bash
cd frontend
npm install
npm run dev
```

浏览器打开 http://localhost:5173 ，点「先看演示数据」即可走通一屏。

## 接上真实飞书

1. 打开 [飞书开发者后台](https://open.feishu.cn/app) 创建企业自建应用
2. 复制仓库根目录 `.env.example` 为 `.env`，填入 `FEISHU_APP_ID` / `FEISHU_APP_SECRET`
3. 安全设置里把重定向 URL 配成 `http://localhost:5173/callback`
4. 在「权限管理」申请下列权限，并发布版本、给自己开通可用范围

| 权限 | 用途 |
| --- | --- |
| `offline_access` | 刷新登录态 |
| `auth:user.id:read` | 读取当前用户 |
| `calendar:calendar:readonly` / `calendar:calendar` | 今日日程、会后写纪要 |
| `drive:drive:readonly` / `drive:drive` | 最近文档 |
| `docx:document:readonly` / `docx:document` | 写回纪要文档 |
| `task:task:read` / `task:task:write` | 待办列表与写回 |
| `im:chat:readonly` | 最近会话 |

权限继承你本人在飞书里已经能看到的范围，应用不会读到你看不到的群或文档。

## 用豆包润色纪要

在 `.env` 里填火山方舟兼容接口：

```
DOUBAO_API_KEY=你的key
DOUBAO_BASE_URL=https://ark.cn-beijing.volces.com/api/v3
DOUBAO_MODEL=doubao-seed-1-6-250615
```

不填也能生成纪要，只是用本地模板。

## 目录

```
backend/     Go API（SQLite 在 backend/data/feishushow.db）
frontend/    Vue 今日雷达
.env.example 飞书 / 豆包配置样例
```

## 下一步

周报草稿、群订阅摘要、把技能挂到豆包工作伙伴当三方 Agent。v0.1 先把「今天这一屏」做稳。
