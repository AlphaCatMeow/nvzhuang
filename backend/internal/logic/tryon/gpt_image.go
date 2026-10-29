package tryon

import (
	"bytes"
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
	"time"
)

// OpenAIImages OpenAI Images API 兼容适配器。
// 覆盖 GPT-Image-1、豆包（火山方舟）、通义万相、智谱 CogView、硅基流动等
// 所有走 POST {baseURL}/images/generations 协议的服务商。
type OpenAIImages struct {
	apiKey  string
	baseURL string
	model   string
	name    string
	client  *http.Client
}

// NewOpenAIImages 从配置创建适配器
func NewOpenAIImages(cfg *ModelConfig) *OpenAIImages {
	return &OpenAIImages{
		apiKey:  cfg.APIKey,
		baseURL: strings.TrimRight(cfg.BaseURL, "/"),
		model:   cfg.Model,
		name:    cfg.Name,
		client:  &http.Client{Timeout: 180 * time.Second},
	}
}

func (o *OpenAIImages) Name() string { return o.name }

// gptImageRequest 图片生成请求体。
// 同时兼容两种协议：
//   - prompt + image[]（火山方舟 / 多数 OpenAI 兼容中转站的多图生图格式，换装必须用这个）
//   - input 数组（OpenAI 官方 gpt-image 格式，作为兜底）
type gptImageRequest struct {
	Model   string          `json:"model"`
	Prompt  string          `json:"prompt,omitempty"`
	Image   []string        `json:"image,omitempty"`
	Input   []gptImageInput `json:"input,omitempty"`
	Size    string          `json:"size,omitempty"`
	Quality string          `json:"quality,omitempty"`
	N       int             `json:"n,omitempty"`
}

type gptImageInput struct {
	Role    string            `json:"role"`
	Content []gptImageContent `json:"content"`
}

type gptImageContent struct {
	Type     string    `json:"type"`
	Text     string    `json:"text,omitempty"`
	ImageURL *imageURL `json:"image_url,omitempty"`
}

type imageURL struct {
	URL string `json:"url"`
}

// gptImageResponse OpenAI Images API 响应
type gptImageResponse struct {
	Data  []gptImageData `json:"data"`
	Error *gptAPIError   `json:"error,omitempty"`
}

type gptImageData struct {
	B64JSON string `json:"b64_json"`
	URL     string `json:"url"`
}

type gptAPIError struct {
	Message string `json:"message"`
	Type    string `json:"type"`
	Code    string `json:"code"`
}

func (o *OpenAIImages) Generate(ctx context.Context, personB64, garmentB64, prompt, size string) (string, error) {
	if o.apiKey == "" {
		return "", fmt.Errorf("模型 %s 未配置 API Key", o.name)
	}

	// 换装场景需要「人像 + 服装」两张参考图。
	// 绝大多数 OpenAI 兼容中转站（含用户的中转站 cpa.paipiange.com）对 gpt-image-2 走
	// 火山方舟式多图生图协议：prompt 必填 + image[] 传参考图，返回 b64_json。
	// 这里统一用 prompt + image[]，对官方 OpenAI 协议则由后端中转站自行兼容。
	reqBody := gptImageRequest{
		Model:   o.model,
		Prompt:  prompt,
		Image: []string{
			"data:image/png;base64," + ensureBase64(personB64),
			"data:image/png;base64," + ensureBase64(garmentB64),
		},
		Size:    normalizeOpenAISize(size),
		Quality: "high",
	}

	body, err := json.Marshal(reqBody)
	if err != nil {
		return "", err
	}

	req, err := http.NewRequestWithContext(ctx, "POST", o.baseURL+"/images/generations", bytes.NewReader(body))
	if err != nil {
		return "", err
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+o.apiKey)

	resp, err := o.client.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	respBody, _ := io.ReadAll(resp.Body)
	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("图像 API 错误 (status %d): %s", resp.StatusCode, truncate(string(respBody), 500))
	}

	var result gptImageResponse
	if err := json.Unmarshal(respBody, &result); err != nil {
		return "", err
	}
	if result.Error != nil {
		return "", fmt.Errorf("图像 API 错误: %s", result.Error.Message)
	}
	if len(result.Data) == 0 {
		return "", fmt.Errorf("图像 API 未返回图片")
	}
	// 优先 b64_json，其次 url（部分服务商返回 URL）
	if result.Data[0].B64JSON != "" {
		return result.Data[0].B64JSON, nil
	}
	if result.Data[0].URL != "" {
		return o.downloadImage(ctx, result.Data[0].URL)
	}
	return "", fmt.Errorf("图像 API 返回格式无法解析")
}

// downloadImage 下载 URL 图片并转 base64（部分服务商只返回 URL）
func (o *OpenAIImages) downloadImage(ctx context.Context, url string) (string, error) {
	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return "", err
	}
	resp, err := o.client.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()
	data, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}
	return base64.StdEncoding.EncodeToString(data), nil
}

// ensureBase64 兼容：如果传入的是 data URL，去掉前缀
func ensureBase64(s string) string {
	if len(s) > 22 && s[:5] == "data:" {
		if idx := bytes.IndexByte([]byte(s), ','); idx >= 0 {
			return s[idx+1:]
		}
	}
	return s
}

// decodeB64 辅助：校验 base64 合法性
func decodeB64(s string) error {
	s = ensureBase64(s)
	_, err := base64.StdEncoding.DecodeString(s)
	return err
}

// truncate 截断字符串，避免错误信息过长
func truncate(s string, n int) string {
	if len(s) <= n {
		return s
	}
	return s[:n] + "..."
}

// normalizeOpenAISize 归一化尺寸。换装默认竖版 1024x1536（2:3）。
// GPT-Image 支持 1024x1024 / 1536x1024 / 1024x1536 / auto。
func normalizeOpenAISize(size string) string {
	switch size {
	case "1024x1024", "1536x1024", "1024x1536", "auto":
		return size
	case "", "1024x1792", "1792x1024":
		return "1024x1536"
	default:
		return "1024x1536"
	}
}

// osGetenv 读取环境变量
func osGetenv(name string) string { return os.Getenv(name) }

// encodeB64 字节转标准 base64
func encodeB64(data []byte) string { return base64.StdEncoding.EncodeToString(data) }
