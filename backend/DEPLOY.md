# AI 换装后端部署指南

本文档说明如何把 `backend/`（GoFrame v2 后端）部署到**你自己的服务器**。

## 一、前置条件

- 一台能访问对应服务商的服务器（调国外模型需能访问 OpenAI / Google，用豆包等国内模型则只需能访问国内网络即可）
- 服务器已安装 Go 1.18+ 或 Docker
- 已申请 API Key（按需，用哪个模型就配哪个）：
  - OpenAI GPT-Image：`sk-...`（https://platform.openai.com）
  - 豆包（火山方舟）：`ARK...`（https://console.volcengine.com/ark）
  - Google Gemini：`AIza...`（https://aistudio.google.com）
  - 其他：通义万相、智谱 CogView、硅基流动等 OpenAI 兼容服务均可接入

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
```

> 关键点：
> - `provider: openai` 表示走标准的 `POST {baseURL}/images/generations` + `input` 数组协议（OpenAI 官方 GPT-Image、通义万相、智谱 CogView、多数中转站）。把 baseURL 换成服务商地址、model 换成其模型 ID 即可。
> - `provider: doubao` 走火山方舟协议（同样是 `/images/generations`，但请求体用 `prompt` + `image[]` 多图生图，返回 `url`），适配豆包 Seedream 换装。
> - `provider: gemini` 走 Google 原生 `generateContent` 协议。
> 前端模型下拉框在 `index.html` 的 `<select id="aiModel">` 中增减 `<option>` 即可与后端别名对应。

## 三、配置前端指向后端

前端 `index.html` 通过以下方式之一知道后端地址（优先级从高到低）：

1. **URL 参数**：`https://你的前端/?tryon=https://api.你的域名`
2. **全局变量**：在页面加载前设置 `window.AI_API_URL = "https://api.你的域名"`

> 建议：把后端反向代理到 `https://api.你的域名`（Nginx 配置见下），前端部署时在 `index.html` 顶部加一句：
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

## 五、安全提醒

1. **API Key 只放在服务器环境变量**，绝不要提交到 Git 仓库（`manifest/config/config.yaml` 已在 `.gitignore` 中忽略）。
2. 建议给接口加**访问限流**或简单鉴权（当前是开放接口，任何人知道地址都能调用你的付费 API）。如需加鉴权，可后续在 `cmd.go` 中间件里加一个 Token 校验。
3. 图片 base64 体积较大，Nginx 已调大 `client_max_body_size`。

## 六、验证部署

部署后测试：

```bash
# 健康检查（应返回校验错误，说明服务在线）
curl -X POST https://api.你的域名/tryon -H "Content-Type: application/json" -d '{}'

# 预期返回：{"code":51,"message":"The PersonImage field is required","data":null}
```

看到这个响应，说明后端已正常运行，即可在前端使用「AI 换装」。
