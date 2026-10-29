package tryon

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"sync/atomic"
	"testing"
	"time"
)

func TestComfyUIGenerateFixedWorkflow(t *testing.T) {
	workflow := map[string]interface{}{
		"1": map[string]interface{}{"class_type": "LoadImage", "inputs": map[string]interface{}{"image": "person"}},
		"2": map[string]interface{}{"class_type": "LoadImage", "inputs": map[string]interface{}{"image": "garment"}},
		"3": map[string]interface{}{"class_type": "PromptNode", "inputs": map[string]interface{}{"prompt": "placeholder"}},
		"4": map[string]interface{}{"class_type": "RandomNoise", "inputs": map[string]interface{}{"noise_seed": float64(0)}},
		"5": map[string]interface{}{"class_type": "SaveImage", "inputs": map[string]interface{}{"filename_prefix": "tryon"}},
	}
	workflowData, err := json.Marshal(workflow)
	if err != nil {
		t.Fatal(err)
	}
	workflowPath := filepath.Join(t.TempDir(), "workflow.json")
	if err := os.WriteFile(workflowPath, workflowData, 0o600); err != nil {
		t.Fatal(err)
	}

	var uploads atomic.Int32
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch {
		case r.URL.Path == "/system_stats":
			_, _ = io.WriteString(w, `{}`)
		case r.URL.Path == "/object_info":
			_, _ = io.WriteString(w, `{"LoadImage":{},"PromptNode":{},"RandomNoise":{},"SaveImage":{}}`)
		case r.URL.Path == "/upload/image":
			if err := r.ParseMultipartForm(1 << 20); err != nil {
				t.Errorf("parse multipart: %v", err)
			}
			count := uploads.Add(1)
			w.Header().Set("Content-Type", "application/json")
			_ = json.NewEncoder(w).Encode(map[string]interface{}{"name": "uploaded-" + string(rune('0'+count)) + ".png", "type": "input"})
		case r.URL.Path == "/prompt":
			var request comfyPromptRequest
			if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
				t.Errorf("decode prompt: %v", err)
			}
			person := request.Prompt["1"].(map[string]interface{})["inputs"].(map[string]interface{})["image"]
			garment := request.Prompt["2"].(map[string]interface{})["inputs"].(map[string]interface{})["image"]
			prompt := request.Prompt["3"].(map[string]interface{})["inputs"].(map[string]interface{})["prompt"].(string)
			seed := request.Prompt["4"].(map[string]interface{})["inputs"].(map[string]interface{})["noise_seed"]
			if person == "person" || garment == "garment" {
				t.Error("uploaded image names were not bound into workflow")
			}
			if !strings.Contains(prompt, "subject_definitions:") || !strings.Contains(prompt, "retention_analysis:") || !strings.Contains(prompt, "Picture 1") || !strings.Contains(prompt, "Picture 2") {
				t.Error("try-on prompt wrapper was not applied")
			}
			if seed != float64(246813579) && seed != int64(246813579) {
				t.Errorf("unexpected fixed seed: %v", seed)
			}
			_, _ = io.WriteString(w, `{"prompt_id":"prompt-1","node_errors":{}}`)
		case r.URL.Path == "/history/prompt-1":
			_, _ = io.WriteString(w, `{"prompt-1":{"outputs":{"5":{"images":[{"filename":"result.png","subfolder":"","type":"output"}]}},"status":{"completed":true,"status_str":"success"}}}`)
		case r.URL.Path == "/view":
			w.Header().Set("Content-Type", "image/png")
			_, _ = w.Write([]byte("generated-image"))
		default:
			http.NotFound(w, r)
		}
	}))
	defer server.Close()

	adapter := NewComfyUI(&ModelConfig{
		Name:           "comfyui",
		Provider:       "comfyui",
		BaseURL:        server.URL,
		Model:          "test-workflow",
		WorkflowPath:   workflowPath,
		PersonNode:     "1",
		GarmentNode:    "2",
		PromptNode:     "3",
		SeedNode:       "4",
		Seed:           246813579,
		OutputNode:     "5",
		PollInterval:   time.Millisecond,
		RequestTimeout: time.Second,
	})

	input := base64.StdEncoding.EncodeToString([]byte("input-image"))
	result, err := adapter.Generate(context.Background(), input, input, "fashion try-on", "1024x1536")
	if err != nil {
		t.Fatalf("Generate returned error: %v", err)
	}
	if uploads.Load() != 2 {
		t.Fatalf("expected 2 uploads, got %d", uploads.Load())
	}
	decoded, err := base64.StdEncoding.DecodeString(result)
	if err != nil {
		t.Fatal(err)
	}
	if string(decoded) != "generated-image" {
		t.Fatalf("unexpected image result: %q", decoded)
	}
}

func TestComfyUIHealthRejectsMissingNodeClass(t *testing.T) {
	workflowPath := filepath.Join(t.TempDir(), "workflow.json")
	if err := os.WriteFile(workflowPath, []byte(`{"1":{"class_type":"MissingNode","inputs":{}},"2":{"class_type":"LoadImage","inputs":{}},"3":{"class_type":"PromptNode","inputs":{}},"4":{"class_type":"SaveImage","inputs":{}}}`), 0o600); err != nil {
		t.Fatal(err)
	}
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/system_stats" {
			_, _ = io.WriteString(w, `{}`)
			return
		}
		_, _ = io.WriteString(w, `{"LoadImage":{},"PromptNode":{},"SaveImage":{}}`)
	}))
	defer server.Close()

	adapter := NewComfyUI(&ModelConfig{
		BaseURL:      server.URL,
		WorkflowPath: workflowPath,
		PersonNode:   "1",
		GarmentNode:  "2",
		PromptNode:   "3",
		OutputNode:   "4",
	})
	if err := adapter.Health(context.Background()); err == nil {
		t.Fatal("expected missing node class error")
	}
}

func TestComfyTryonPromptUsesGarmentCompositionAnchor(t *testing.T) {
	prompt := buildComfyTryonPrompt("ignored generic prompt")
	for _, required := range []string{
		"Do not retain Picture 1's original clothing, close-up crop",
		"full-body pose, body placement and catalog composition come from Picture 2",
		"entire head and all hair",
		"no zoom, no push-in, no reframing",
	} {
		if !strings.Contains(prompt, required) {
			t.Fatalf("ComfyUI prompt is missing framing constraint %q", required)
		}
	}
}

func TestComfyWorkflowReturnsEarlyCompositedFrame(t *testing.T) {
	workflowData, err := os.ReadFile(filepath.Join("..", "..", "..", "workflows", "comfyui_h3_ref2va_still_api.json"))
	if err != nil {
		t.Fatal(err)
	}
	var workflow map[string]interface{}
	if err := json.Unmarshal(workflowData, &workflow); err != nil {
		t.Fatal(err)
	}
	frameNode := workflow["18"].(map[string]interface{})
	inputs := frameNode["inputs"].(map[string]interface{})
	if inputs["batch_index"] != float64(31) {
		t.Fatalf("expected early composited frame 31, got %v", inputs["batch_index"])
	}
}

// TestComfyUIIntegration exercises the real local ComfyUI installation when
// explicitly enabled. It stays skipped in normal CI and never uses user photos.
func TestComfyUIIntegration(t *testing.T) {
	if os.Getenv("COMFYUI_INTEGRATION") != "1" {
		t.Skip("set COMFYUI_INTEGRATION=1 to run against local ComfyUI")
	}

	personData, err := os.ReadFile(filepath.Join("..", "..", "..", "..", "docs", "tryon-3.png"))
	if err != nil {
		t.Fatal(err)
	}
	garmentData, err := os.ReadFile(filepath.Join("..", "..", "..", "..", "assets", "img", "SUM-0156.webp"))
	if err != nil {
		t.Fatal(err)
	}

	gen := NewComfyUI(&ModelConfig{
		Name:           "comfyui",
		Provider:       "comfyui",
		BaseURL:        "http://127.0.0.1:8188",
		Model:          "h3-ref2va-still",
		WorkflowPath:   filepath.Join("..", "..", "..", "workflows", "comfyui_h3_ref2va_still_api.json"),
		PersonNode:     "16",
		GarmentNode:    "17",
		PromptNode:     "6",
		SeedNode:       "10",
		Seed:           246813579,
		OutputNode:     "19",
		PollInterval:   time.Second,
		RequestTimeout: 15 * time.Minute,
	})

	ctx, cancel := context.WithTimeout(context.Background(), 16*time.Minute)
	defer cancel()
	result, err := gen.Generate(
		ctx,
		base64.StdEncoding.EncodeToString(personData),
		base64.StdEncoding.EncodeToString(garmentData),
		"Keep the subject front-facing and preserve the black-and-white gingham dress exactly.",
		"352x640",
	)
	if err != nil {
		t.Fatal(err)
	}
	imageData, err := base64.StdEncoding.DecodeString(result)
	if err != nil {
		t.Fatal(err)
	}
	if contentType := http.DetectContentType(imageData); !strings.HasPrefix(contentType, "image/") {
		t.Fatalf("unexpected result content type: %s", contentType)
	}
}
