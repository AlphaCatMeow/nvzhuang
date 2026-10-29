package tryon

import (
	"context"

	"github.com/gogf/gf/v2/errors/gerror"
	"github.com/gogf/gf/v2/frame/g"
)

// DoTryon 执行换装：选择模型 → 构建提示词 → 调用生成 → 返回结果
func DoTryon(ctx context.Context, personB64, garmentB64, model, size, lang string) (resultB64, usedModel, usedPrompt string, err error) {
	// 参数校验
	if personB64 == "" {
		return "", "", "", gerror.New("人像照片不能为空")
	}
	if garmentB64 == "" {
		return "", "", "", gerror.New("服装图不能为空")
	}
	if err := decodeB64(personB64); err != nil {
		return "", "", "", gerror.New("人像照片 base64 格式错误")
	}
	if err := decodeB64(garmentB64); err != nil {
		return "", "", "", gerror.New("服装图 base64 格式错误")
	}

	// 默认模型：Gemini（Nano Banana 2），因为 GPT Image 系列对真人换装场景
	// 会被官方安全审核拦截（moderation_blocked / sexual），Gemini 无此限制。
	if model == "" {
		model = "gemini"
	}

	gen, err := NewGenerator(ctx, model)
	if err != nil {
		return "", "", "", err
	}

	// 构建提示词：Gemini 协议用英文模板，其余（GPT-Image / 豆包等）按语言选择
	var promptStr string
	if gen.Name() == "gemini" {
		promptStr = BuildGemini(lang)
	} else {
		promptStr = BuildGPT(lang)
	}

	// 调用生成
	result, err := gen.Generate(ctx, personB64, garmentB64, promptStr, size)
	if err != nil {
		g.Log().Errorf(ctx, "换装生成失败 model=%s: %v", model, err)
		return "", "", "", gerror.Wrap(err, "换装生成失败，请稍后重试")
	}

	return result, gen.Name(), promptStr, nil
}
