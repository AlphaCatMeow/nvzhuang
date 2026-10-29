# AI 换装后端服务

基于 GoFrame v2 的 AI 换装（虚拟试衣）后端，为「衣装集」网站提供 `POST /tryon` 接口。

## 功能

- 接收人像照片 + 服装图（base64），调用大模型生成换装结果图
- 支持两种模型，可配置：
  - **GPT-Image**（OpenAI，默认）
  - **Gemini 2.5 Flash Image**（Google，即 "Banana Pro"）

## 快速开始

```bash
cd backend
gf run main.go
# 或
go run main.go
```

服务默认监听 `:8000`。

## 配置 API Key

**切勿把真实 Key 提交到仓库。** Key 通过环境变量注入，优先级高于配置文件：

```bash
# Windows PowerShell
$env:OPENAI_API_KEY = "sk-xxxx"
$env:GEMINI_API_KEY  = "AIza-xxxx"

go run main.go
```

或复制模板改配置（`manifest/config/config.yaml` 已被 .gitignore 忽略）：

```bash
cp manifest/config/config.example.yaml manifest/config/config.yaml
# 编辑 config.yaml，填入 llm.openai.apiKey / llm.gemini.apiKey
```

## 接口说明

`POST /tryon`，请求体（JSON）：

```json
{
  "person_image": "人像照片 base64",
  "garment_image": "服装图 base64",
  "model": "gpt-image",
  "size": "1024x1536",
  "keep_face": true,
  "lang": "zh"
}
```

| 字段 | 说明 | 默认 |
|---|---|---|
| person_image | 人像照片（全身/半身站立照），base64 | 必填 |
| garment_image | 服装图，base64 | 必填 |
| model | `gpt-image` 或 `gemini` | `gpt-image` |
| size | 输出尺寸 `1024x1024`/`1024x1536`/`1536x1024` | `1024x1536` |
| lang | 提示词语言 `zh`/`en` | `zh` |

响应：

```json
{
  "image": "换装结果图 base64",
  "model": "gpt-image",
  "prompt": "实际使用的提示词"
}
```

## 目录结构

```
backend/
├── api/tryon/            # API 定义（请求/响应结构）
├── internal/
│   ├── cmd/              # 服务启动 + 路由 + CORS
│   ├── controller/tryon/ # 控制器
│   └── logic/tryon/      # 业务逻辑
│       ├── tryon.go      # 换装主流程
│       ├── prompt.go     # 提示词模板（针对模型优化）
│       ├── generator.go  # 模型适配器接口
│       ├── gpt_image.go  # GPT-Image 适配器
│       └── gemini.go     # Gemini 适配器
└── manifest/config/      # 配置文件
```
