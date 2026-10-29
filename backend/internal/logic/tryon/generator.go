package tryon

import (
	"context"
	"fmt"
	"strings"

	"github.com/gogf/gf/v2/frame/g"
)

// ImageGenerator 图像生成模型适配器接口
type ImageGenerator interface {
	// Generate 生成换装图。入参为两张图的 base64（人像 + 服装），返回结果图 base64
	Generate(ctx context.Context, personB64, garmentB64, prompt, size string) (string, error)
	// Name 模型别名（对外暴露的名称）
	Name() string
}

// ModelConfig 单个模型的配置项
type ModelConfig struct {
	Name     string // 模型别名（前端请求用），如 gpt-image / doubao / gemini
	Provider string // 协议类型：openai（OpenAI Images 兼容）/ gemini
	APIKey   string // API Key（可被环境变量覆盖）
	BaseURL  string // 接口地址（可配置，支持第三方/国内服务商）
	Model    string // 服务商侧真实模型名
	EnvKey   string // 覆盖用的环境变量名
}

// loadModelConfig 从配置文件加载指定模型的配置，并允许环境变量覆盖
func loadModelConfig(ctx context.Context, alias string) (*ModelConfig, error) {
	// 按列表读取：llm.models.[alias]
	prefix := "llm.models." + alias
	cfg := &ModelConfig{
		Name:     alias,
		Provider: g.Cfg().MustGet(ctx, prefix+".provider", "openai").String(),
		APIKey:   g.Cfg().MustGet(ctx, prefix+".apiKey", "").String(),
		BaseURL:  g.Cfg().MustGet(ctx, prefix+".baseURL", "").String(),
		Model:    g.Cfg().MustGet(ctx, prefix+".model", "").String(),
		EnvKey:   g.Cfg().MustGet(ctx, prefix+".envKey", "").String(),
	}

	// 环境变量覆盖（优先级最高）
	if cfg.EnvKey != "" {
		if v := lookupEnv(cfg.EnvKey); v != "" {
			cfg.APIKey = v
		}
	}

	// 未在 models 列表找到时，回退到旧的固定 key（兼容）
	if cfg.Model == "" && cfg.BaseURL == "" && cfg.APIKey == "" {
		cfg = loadLegacyConfig(ctx, alias)
	}

	if cfg.APIKey == "" {
		return nil, fmt.Errorf("模型 %s 未配置 API Key（请在 config.yaml 的 llm.models.%s.apiKey 或环境变量中配置）", alias, alias)
	}
	if cfg.Model == "" {
		return nil, fmt.Errorf("模型 %s 未配置 model 名（llm.models.%s.model）", alias, alias)
	}
	return cfg, nil
}

// loadLegacyConfig 兼容旧的 llm.openai / llm.gemini 配置写法
func loadLegacyConfig(ctx context.Context, alias string) *ModelConfig {
	switch alias {
	case "gpt-image":
		return &ModelConfig{
			Name:     alias,
			Provider: "openai",
			APIKey:   getKey("OPENAI_API_KEY", "llm.openai.apiKey"),
			BaseURL:  g.Cfg().MustGet(ctx, "llm.openai.baseURL", "https://api.openai.com/v1").String(),
			Model:    g.Cfg().MustGet(ctx, "llm.openai.model", "gpt-image-2").String(),
		}
	case "gemini", "gemini-3.1-flash-image", "banana", "nano-banana", "gemini-2.5-flash-image":
		return &ModelConfig{
			Name:     "gemini",
			Provider: "gemini",
			APIKey:   getKey("GEMINI_API_KEY", "llm.gemini.apiKey"),
			BaseURL:  g.Cfg().MustGet(ctx, "llm.gemini.baseURL", "https://generativelanguage.googleapis.com/v1beta").String(),
			Model:    g.Cfg().MustGet(ctx, "llm.gemini.model", "gemini-3.1-flash-image").String(),
		}
	}
	return nil
}

// NewGenerator 根据模型别名创建适配器
func NewGenerator(ctx context.Context, alias string) (ImageGenerator, error) {
	cfg, err := loadModelConfig(ctx, alias)
	if err != nil {
		return nil, err
	}

	switch strings.ToLower(cfg.Provider) {
	case "doubao", "ark", "volc", "volcengine", "huoshan":
		// 火山方舟（豆包 Seedream）：复用 /images/generations 路径但请求体为火山自定义格式
		return NewDoubao(cfg), nil
	case "openai", "qwen", "wanx", "zhipu", "siliconflow", "openai-compatible":
		// 所有 OpenAI Images 兼容协议（GPT-Image、通义万相、智谱等）走同一适配器
		return NewOpenAIImages(cfg), nil
	case "gemini", "google":
		return NewGemini(cfg), nil
	default:
		return nil, fmt.Errorf("unsupported provider: %s (supported: openai, doubao, gemini)", cfg.Provider)
	}
}

// getKey 优先取环境变量，其次取配置文件
func getKey(envName, cfgPath string) string {
	if v := lookupEnv(envName); v != "" {
		return v
	}
	return g.Cfg().MustGet(context.Background(), cfgPath).String()
}

// lookupEnv 兼容多个环境变量名（大小写不敏感地查找首个非空值）
func lookupEnv(names ...string) string {
	for _, n := range names {
		if n == "" {
			continue
		}
		if v := osGetenv(n); v != "" {
			return v
		}
	}
	return ""
}
