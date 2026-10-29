package tryon

import (
	"context"
	"strings"
	"testing"
)

func TestBuildPrompt(t *testing.T) {
	gptZh := BuildGPT("zh")
	if !strings.Contains(gptZh, "虚拟试衣") {
		t.Errorf("BuildGPT(zh) should contain 虚拟试衣, got: %s", gptZh)
	}

	gptEn := BuildGPT("en")
	if !strings.Contains(gptEn, "virtual try-on") {
		t.Errorf("BuildGPT(en) should contain virtual try-on, got: %s", gptEn)
	}

	geminiPrompt := BuildGemini("zh")
	if !strings.Contains(geminiPrompt, "virtual try-on") {
		t.Errorf("BuildGemini should return stable prompt, got: %s", geminiPrompt)
	}
}

func TestDecodeB64(t *testing.T) {
	validB64 := "iVBORw0KGgoAAAANSUhEUgAAAAEAAAABCAQAAAC1HAwCAAAAC0lEQVR42mNkYAAAAAYAAjCB0C8AAAAASUVORK5CYII="
	if err := decodeB64(validB64); err != nil {
		t.Errorf("decodeB64 should succeed for valid base64: %v", err)
	}

	dataUri := "data:image/png;base64," + validB64
	if err := decodeB64(dataUri); err != nil {
		t.Errorf("decodeB64 should succeed for data uri: %v", err)
	}

	invalidB64 := "!!!not-valid-base64!!!"
	if err := decodeB64(invalidB64); err == nil {
		t.Errorf("decodeB64 should fail for invalid base64")
	}
}

func TestDoTryonValidation(t *testing.T) {
	ctx := context.Background()

	// Empty person
	_, _, _, err := DoTryon(ctx, "", "somegarment", "gemini", "1024x1536", "zh")
	if err == nil || !strings.Contains(err.Error(), "人像照片不能为空") {
		t.Errorf("expected empty person error, got: %v", err)
	}

	// Empty garment
	_, _, _, err = DoTryon(ctx, "someperson", "", "gemini", "1024x1536", "zh")
	if err == nil || !strings.Contains(err.Error(), "服装图不能为空") {
		t.Errorf("expected empty garment error, got: %v", err)
	}
}
