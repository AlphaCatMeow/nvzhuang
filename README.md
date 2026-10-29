# 衣装集 COSTUME-DB-01 · AI 服装参考图集

一个深色主题的服装参考图集网站，收录 **1407 张** 人物不入镜的 AI 生成服装图像（2:3 比例），支持中英文双语浏览与检索。

> 🔗 **在线预览**：https://paipiange.github.io/nvzhuang/

## 📸 预览

| 中文模式 | 英文模式 | 关键词搜索 |
|---|---|---|
| ![中文模式](docs/screenshot-zh.png) | ![英文模式](docs/screenshot-en.png) | ![关键词搜索](docs/screenshot-search.png) |

## ✨ 功能特性

- **8 大分类**：春（215）/ 夏（254）/ 秋（244）/ 冬（253）/ 内衣（199）/ 派对（48）/ 泳装（54）/ 性感（140）
- **中英文双语**：默认中文，右上角一键切换 English
- **双向检索**：中文（白、蕾丝、连衣裙…）和英文（white、lace、dress…）关键词都能搜索
- **关键词本地化**：中文模式自动将服装属性翻译为中文展示
- **交互体验**：分类筛选、实时搜索、无限滚动加载、放大查看、键盘 ←/→ 切换、Esc 关闭、原尺寸 PNG 下载
- **AI 试穿搭配台**：上传人像 + 选服装，在线预览、拖拽缩放、抠图合成，或一键 AI 换装（需自备后端）

## 🤖 AI 换装（可选功能）

「AI 换装」需要一个后端服务，把服装真实地「穿」到人像上。后端位于 `backend/`（GoFrame v2），支持**多模型可插拔**：

| 模型别名 | 模型 | 协议 | 说明 |
|---|---|---|---|
| `gpt-image` | **GPT Image 2**（`gpt-image-2`） | OpenAI Images | 默认模型，官方当前主推 |
| `doubao` | 豆包 Seedream（火山方舟） | 火山方舟 | 国内可用，无需科学上网 |
| `gemini` | **Nano Banana 2**（`gemini-3.1-flash-image`） | Google generateContent | Google 图像模型 |
| `qwen` / `zhipu` / `siliconflow` … | 通义万相 / 智谱 CogView / 硅基流动 | OpenAI Images 兼容 | 按需在配置中开启 |

> 三个关键点都可配置：**API Key、Base URL、模型名**。加一个国内模型只需在配置里新增一段，无需改代码。

### 配置模型

编辑 `backend/manifest/config/config.yaml`（模板见 `config.example.yaml`）：

```yaml
llm:
  models:
    gpt-image:                     # 前端请求用的别名
      provider: openai             # 协议：openai / doubao / gemini
      apiKey: ""                   # 直接填 Key，或留空用环境变量
      envKey: "OPENAI_API_KEY"     # 环境变量名（优先级高于 apiKey）
      baseURL: "https://api.openai.com/v1"   # Base URL 可改（支持中转/代理）
      model: "gpt-image-2"         # 真实模型名

    doubao:                        # 豆包（国内可用）
      provider: doubao
      envKey: "ARK_API_KEY"
      baseURL: "https://ark.cn-beijing.volces.com/api/v3"
      model: "doubao-seedream-3-0-t2i-250415"

    gemini:
      provider: gemini
      envKey: "GEMINI_API_KEY"
      baseURL: "https://generativelanguage.googleapis.com/v1beta"
      model: "gemini-3.1-flash-image"
```

**接入任意 OpenAI 兼容的国内生图模型**（通义万相、智谱、硅基流动等），加一段即可：

```yaml
    qwen:                          # 通义万相示例
      provider: openai
      envKey: "DASHSCOPE_API_KEY"
      baseURL: "https://dashscope.aliyuncs.com/compatible-mode/v1"
      model: "wanx2.1-t2i-turbo"
```

### 启动后端 & 对接前端

```bash
cd backend
export OPENAI_API_KEY="sk-..."      # 按需注入对应 Key
go run main.go                      # 服务监听 :8000
```

前端通过以下方式之一指向后端（优先级从高到低）：

1. URL 参数：`https://你的前端/?tryon=http://localhost:8000`
2. 全局变量：`window.AI_API_URL = "http://localhost:8000"`

> 完整的 Docker 部署、Nginx 反代、安全加固说明见 [`backend/DEPLOY.md`](backend/DEPLOY.md)。

## 🚀 快速开始

本项目是纯静态站点，无需构建。任选一种方式运行：

```bash
# 方式一：Python 内置服务器
python3 -m http.server 8000

# 方式二：Node 简单服务器
npx serve .

# 方式三：直接双击 index.html（部分浏览器可能有跨域限制，推荐前两种）
```

然后访问 `http://localhost:8000`。

## 📁 目录结构

```
nvzhuang/
├── index.html          # 页面结构 + 样式 + 交互逻辑（含双语字典、AI 试穿搭配台）
├── data.js             # 1407 条服装数据（id / 分类 / 关键词 / 图片路径）
├── assets/
│   ├── thumbs/         # 缩略图（webp，网格卡片用）
│   └── img/            # 展示图（webp，弹窗放大用）
├── backend/            # AI 换装后端（GoFrame v2，可选）
│   ├── manifest/config/config.yaml        # 模型配置（Key / BaseURL / 模型名）
│   ├── internal/logic/tryon/              # 模型适配器（openai / doubao / gemini）
│   └── DEPLOY.md                          # 部署指南
└── .gitignore
```

> 注：原尺寸 PNG（约 1GB+）未纳入仓库，「下载原尺寸」按钮指向原站资源。

## 🔗 URL 参数

支持通过 URL 参数直接分享特定状态：

```
?lang=en          # 初始语言（zh 为中文，en 为英文，默认 zh）
?q=lace           # 初始搜索词
?lang=en&q=white  # 组合使用
```

## 📄 使用许可

- 可免费用于作品创作（含商业使用、编辑、参考使用）
- 图像数据本身（原尺寸或加工后）禁止再分发、转载、再销售
- 所有图像均为 AI 生成（t2i），非真实商品照片

## 📝 数据说明

数据存于 `data.js`，每条记录包含：

| 字段 | 说明 |
|---|---|
| `id` | 编号（如 `SPR-0001`） |
| `カテゴリ` | 分类（日文分类名，如「春」） |
| `キーワード` | 关键词（英文，空格分隔） |
| `サムネ` | 缩略图路径 |
| `画像` | 展示图路径 |
| `原寸` | 原尺寸 PNG 路径 |

> 说明：`カテゴリ`、`キーワード`、`サムネ`、`画像`、`原寸` 是数据原始的日文字段名，为保持与原站数据兼容而保留。网站会在前端自动将它们本地化为中文或英文展示。

### Data Fields (English)

Each record in `data.js` contains the following fields:

| Field | Description |
|---|---|
| `id` | Identifier (e.g. `SPR-0001`) |
| `カテゴリ` | Category (Japanese category name, e.g. "春" for Spring) |
| `キーワード` | Keywords (English, space-separated) |
| `サムネ` | Thumbnail image path |
| `画像` | Display image path |
| `原寸` | Full-size PNG path |

> Note: `カテゴリ`, `キーワード`, `サムネ`, `画像`, and `原寸` are the original Japanese field names kept for compatibility with the source data. The site localizes them to Chinese or English on the frontend automatically.
