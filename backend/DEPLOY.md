# AI 换装后端部署指南

本文档说明如何把 `backend/`（GoFrame v2 后端）部署到**你自己的服务器**。

## 一、前置条件

- 一台能访问对应服务商的服务器（调国外模型需能访问 OpenAI / Google，用豆包等国内模型则只需能访问国内网络即可）
- 服务器已安装 Go 1.18+ 或 Docker
- 已申请 API Key（按需，用哪个模型就配哪个；若使用本地 ComfyUI 则完全不需要申请 Key）：
  - OpenAI GPT-Image：`sk-...`（https://platform.openai.com）
  - 豆包（火山方舟）：`ARK...`（https://console.volcengine.com/ark）
  - Google Gemini：`AIza...`（https://aistudio.google.com）
  - 其他：通义万相、智谱 CogView、硅基流动等 OpenAI 兼容服务均可接入
  - ComfyUI 本地离线部署：若主机拥有独立显卡，支持直接接入本地 ComfyUI（优先 Qwen-Image-Edit-2511 4步极速生图，兼容 MiniMax H3），无需任何云端 API 费用。详见 `README.MD`。

## 二、推荐部署方式（优先 Docker）

### 方式 A：Docker Compose（首选推荐，最省事）

```bash
cd backend

# 1. 复制环境变量示例文件
cp .env.example .env

# 2. 编辑 .env 文件，填入你使用的模型 API Key（按需填写，国内优先推荐豆包 ARK_API_KEY）
# 例如：ARK_API_KEY=xxxx

# 3. 一键构建并启动后台容器
docker compose up -d

# 查看日志
docker compose logs -f
```

### 方式 B：Docker Run 单命令运行

```bash
# 1. 在 backend 目录构建镜像
cd backend
docker build -f manifest/docker/Dockerfile -t costume-tryon .

# 2. 运行容器，通过环境变量注入 Key（按需填写，豆包用 ARK_API_KEY）
docker run -d --name costume-tryon \
  -p 8000:8000 \
  --restart unless-stopped \
  -e OPENAI_API_KEY="sk-你的key" \
  -e ARK_API_KEY="你的豆包ArkKey" \
  -e GEMINI_API_KEY="AIza你的key" \
  costume-tryon
```

### 方式 C：直接编译运行（无 Docker）

```bash
cd backend

# 编译 Linux 可执行文件（如在本机 Windows 交叉编译）
GOOS=linux GOARCH=amd64 go build -o main .

# 上传 main 到服务器，设置环境变量后运行
export OPENAI_API_KEY="sk-你的key"
export ARK_API_KEY="你的豆包ArkKey"
export GEMINI_API_KEY="AIza你的key"
./main
```

> 服务默认监听 `:8000`。可用环境变量 `GF_GCFG_FILE` 指定配置文件路径，或用 `--gf.gcfg.file` 参数。

## 二点五、模型与 Key 配置说明

所有模型的 **API Key、Base URL、模型名** 都在 `manifest/config/config.yaml` 的 `llm.models` 下配置，均可自由修改：

```yaml
llm:
  models:
    gpt-image:            # 前端请求用的别名
      provider: openai    # 协议：openai（OpenAI Images 兼容）/ gemini
      apiKey: ""          # 直接填 Key，或留空用环境变量
      envKey: "OPENAI_API_KEY"   # 环境变量名（优先于 apiKey）
      baseURL: "https://api.openai.com/v1"
      model: "gpt-image-2"
    doubao:               # 豆包（火山方舟），国内可用
      provider: doubao    # 火山方舟协议（复用 images/generations，但请求体为火山格式）
      envKey: "ARK_API_KEY"
      baseURL: "https://ark.cn-beijing.volces.com/api/v3"
      model: "doubao-seedream-3-0-t2i-250415"
    gemini:
      provider: gemini
      envKey: "GEMINI_API_KEY"
      baseURL: "https://generativelanguage.googleapis.com/v1beta"
      model: "gemini-3.1-flash-image"
```

**接入任意 OpenAI 兼容的国内生图模型**（通义万相、智谱 CogView、硅基流动、各类中转站），只需新增一个条目：

```yaml
    qwen:                 # 通义万相示例
      provider: openai
      envKey: "DASHSCOPE_API_KEY"
      baseURL: "https://dashscope.aliyuncs.com/compatible-mode/v1"
      model: "wanx2.1-t2i-turbo"

    comfyui:              # 本地 ComfyUI 离线方案（默认 Qwen 优先，自动回退 H3）
      provider: comfyui
      baseURL: "http://127.0.0.1:8188"
      model: "qwen-image-edit-2511"
      workflowPath: "workflows/comfyui_qwen_image_edit_2511_api.json"
```

> 关键点：
> - `provider: openai` 表示走标准的 `POST {baseURL}/images/generations` + `input` 数组协议（OpenAI 官方 GPT-Image、通义万相、智谱 CogView、多数中转站）。把 baseURL 换成服务商地址、model 换成其模型 ID 即可。
> - `provider: doubao` 走火山方舟协议（同样是 `/images/generations`，但请求体用 `prompt` + `image[]` 多图生图，返回 `url`），适配豆包 Seedream 换装。
> - `provider: gemini` 走 Google 原生 `generateContent` 协议。
> 前端模型下拉框在 `index.html` 的 `<select id="aiModel">` 中增减 `<option>` 即可与后端别名对应。

## 三、配置前端指向后端

前端 `index.html` 按以下优先级获取后端地址：

1. **设置中心**：用户在当前浏览器保存的后端 API 地址
2. **全局变量**：在页面加载前设置 `window.AI_API_URL = "https://api.你的域名"`
3. **URL 参数**：`https://你的前端/?tryon=https://api.你的域名`
4. **本地自动回退**：访问宿主为 `localhost` 或 `127.0.0.1` 时，自动尝试回退连接 `http://127.0.0.1:8000`

> 建议：若部署公网，把后端反向代理到 `https://api.你的域名`（Nginx 配置见下），前端部署时在 `index.html` 顶部加一句：
> ```html
> <script>window.AI_API_URL="https://api.你的域名";</script>
> ```

## 四、Nginx 反向代理（可选但推荐）

用 HTTPS 域名 + 反代，避免前端跨域和明文 Key 问题：

```nginx
server {
    listen 443 ssl;
    server_name api.你的域名.com;

    ssl_certificate     /path/to/cert.pem;
    ssl_certificate_key /path/to/key.pem;

    location / {
        proxy_pass http://127.0.0.1:8000;
        proxy_set_header Host $host;
        proxy_set_header X-Real-IP $remote_addr;
        # 大图片 base64 请求体，调大超时和体积上限
        client_max_body_size 30m;
        proxy_read_timeout 180s;
        proxy_send_timeout 180s;
    }
}
```

## 五、安全与成本控制提醒

1. **默认仅建议本地或内网自用**：容器或后端默认仅供开发者本机或私有网络自用。**不建议在无鉴权、无限流的情况下直接将 `:8000` 端口裸露到公网**，以防被他人批量刷取产生高额大模型账单。
2. **公网开放防护建议**：若确实需要部署公网供朋友或用户体验，必须通过 Nginx 等反向代理层配置：
   - 严格的 IP 限流与防刷机制（如 `limit_req_zone`）
   - 单次请求 Body 上限控制（限制在 20MB 以内，防止大文件 OOM）
   - 服务端简易鉴权（Token 或验证码）
3. **API Key 隔离与保护**：
   - Key 仅保存在服务器本地环境变量或 `.env`，禁止将包含真实 Key 的 `config.yaml` 或 `.env` 提交到公开仓库。
   - 默认的「服务器预设」模式下，前端不接触服务器 API Key，所有大模型通信均由后端代理中转。
   - 可选的「个人模型配置」由用户主动保存在自己的浏览器，并只随单次 `/tryon` 请求发送。后端不写数据库、配置文件或业务日志；Gemini Key 使用请求头传递，不进入上游 URL。
   - 个人 Base URL 默认拒绝 localhost、内网与保留地址，且不允许查询参数，避免公开后端被滥用为 SSRF 代理或把 Key 写入 URL。
4. **模型内容安全遵从**：各模型提供商（如 OpenAI、Google、火山引擎）对真人图像生成均有内容安全审核（Moderation），请遵守相关法律法规与平台服务条款，勿用于违背公序良俗的场景。

## 六、验证部署

部署后测试：

```bash
# 健康检查（应返回校验错误，说明服务在线）
curl -X POST https://api.你的域名/tryon -H "Content-Type: application/json" -d '{}'

# 预期返回：{"code":51,"message":"The PersonImage field is required","data":null}
```

看到这个响应，说明后端已正常运行，即可在前端使用「AI 换装」。
