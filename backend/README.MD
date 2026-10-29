# 拍片哥女装 · AI 换装后端服务

基于 GoFrame v2 的 AI 换装（虚拟试衣）后端，为「拍片哥女装」网站提供 `POST /tryon` 接口。灵感源自 X 推文作者 [@giaaibusi](https://x.com/giaaibusi)。

## 效果演示

用户上传人像照片（图 2）与选中的服装图（图 1），调用后端 AI 换装接口合成生成换装效果图（图 3）：

| ① 选中的服装（图 1） | ② 上传的模特照片（图 2） | ③ AI 换装合成效果（图 3） |
|:---:|:---:|:---:|
| <img src="../docs/tryon-1.png" width="260" alt="服装图" /> | <img src="../docs/tryon-2.png" width="260" alt="模特照片" /> | <img src="../docs/tryon-3.png" width="260" alt="换装效果" /> |
| **选定女装参考图** | **人物模特自拍原图** | **合成试穿效果（图 1 + 图 2 ➔ 图 3）** |

## 功能

- 接收人像照片 + 服装图（base64），调用大模型生成换装结果图
- 支持多模型配置化插拔：
  - **GPT-Image 2**（OpenAI Images 协议）
  - **豆包 Seedream**（火山方舟协议，国内免翻墙首选）
  - **Nano Banana 2 / Gemini**（Google generateContent 协议）
  - 其他 OpenAI 兼容模型（通义万相、智谱 CogView、硅基流动等）

## 快速开始（优先 Docker）

### 推荐方式：Docker Compose

```bash
cd backend

# 1. 复制环境变量并填入你的 Key（如 ARK_API_KEY 或 OPENAI_API_KEY）
cp .env.example .env

# 2. 一键启动
docker compose up -d
```

### 方式二：Docker 单容器运行

```bash
cd backend
docker build -f manifest/docker/Dockerfile -t costume-tryon .

# 注入对应的 Key 环境变量运行
docker run -d --name costume-tryon -p 8000:8000 \
  -e ARK_API_KEY="你的豆包ArkKey" \
  -e OPENAI_API_KEY="sk-你的OpenAIKey" \
  costume-tryon
```

### 方式三：本地直接运行

- **Windows 用户**：直接双击运行 `restart.bat`。
- **源码运行**：
  ```bash
  cd backend
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
  "model": "gemini",
  "size": "1024x1536",
  "keep_face": true,
  "lang": "zh",
  "client_config": {
    "provider": "gemini",
    "base_url": "https://generativelanguage.googleapis.com/v1beta",
    "api_key": "仅本次请求使用的 Key",
    "model": "gemini-3.1-flash-image"
  }
}
```

| 字段 | 说明 | 默认 |
|---|---|---|
| person_image | 人像照片（全身/半身站立照），base64 | 必填 |
| garment_image | 服装图，base64 | 必填 |
| model | `gemini` / `gpt-image` / `doubao` | `gemini` |
| size | 输出尺寸 `1024x1024`/`1024x1536`/`1536x1024` | `1024x1536` |
| lang | 提示词语言 `zh`/`en` | `zh` |
| client_config | 可选的浏览器个人模型配置；仅在当前请求内存中使用，不持久化、不记录 Key/Base URL | 空（使用服务器预设） |

`client_config.provider` 支持 `openai`、`doubao`、`gemini`。个人 Base URL 必须是 HTTP/HTTPS 公网地址，且不能包含账号、查询参数或片段；localhost、内网和保留地址会被拒绝，以降低 SSRF 风险。

响应：

```json
{
  "image": "换装结果图 base64",
  "model": "gemini",
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
