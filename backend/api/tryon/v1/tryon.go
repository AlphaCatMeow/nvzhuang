package v1

import "github.com/gogf/gf/v2/frame/g"

// TryonReq 换装请求
type TryonReq struct {
	g.Meta `path:"/tryon" method:"post" tags:"换装" summary:"AI换装图生成"`
	// 人像照片（全身/半身站立照，base64 或 上传文件）
	PersonImage string `json:"person_image" v:"required" dc:"人像照片，base64编码"`
	// 服装图（base64）
	GarmentImage string `json:"garment_image" v:"required" dc:"服装图，base64编码"`
	// 模型选择：gemini / gpt-image / doubao（可在后端 config.yaml 的 llm.models 中扩展），默认 gemini
	Model string `json:"model" d:"gemini" dc:"模型别名：gemini / gpt-image / doubao（可在 config.yaml 扩展）"`
	// 生成尺寸：1024x1024 / 1024x1536 / 1536x1024
	Size string `json:"size" d:"1024x1536" dc:"输出尺寸"`
	// 是否保留人物面部（保持同一个人）
	KeepFace bool `json:"keep_face" d:"true" dc:"是否保持人物面部一致"`
	// 语言：zh / en
	Lang string `json:"lang" d:"zh" dc:"提示词语言"`
}

// TryonRes 换装响应
type TryonRes struct {
	// 生成的换装结果图（base64）
	Image string `json:"image" dc:"换装结果图，base64"`
	// 使用的模型
	Model string `json:"model" dc:"实际使用的模型"`
	// 提示词（调试用，可空）
	Prompt string `json:"prompt,omitempty" dc:"实际使用的提示词"`
}
