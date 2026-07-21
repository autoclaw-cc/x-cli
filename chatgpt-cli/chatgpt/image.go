package chatgpt

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"chatgpt-cli/browser"
)

const imagesURL = "https://chatgpt.com/images/"

type ImageOptions struct {
	Prompt  string
	OutDir  string
	Timeout time.Duration
}

type ImageResult struct {
	Prompt    string `json:"prompt"`
	Path      string `json:"path"`
	Bytes     int    `json:"bytes"`
	MediaType string `json:"media_type"`
	Caption   string `json:"caption,omitempty"`
	ElapsedMS int64  `json:"elapsed_ms"`
}

type imageCandidate struct {
	Src      string `json:"src"`
	Alt      string `json:"alt"`
	FileID   string `json:"fileId"`
	Width    int    `json:"width"`
	Height   int    `json:"height"`
	Complete bool   `json:"complete"`
}

type downloadPayload struct {
	OK          bool   `json:"ok"`
	Error       string `json:"error"`
	Status      int    `json:"status"`
	ContentType string `json:"contentType"`
	Base64      string `json:"base64"`
}

const ImageReadyScript = `(()=>{
  function chatgptImageReady(){
    const prompt=document.querySelector('#prompt-textarea');
    const form=prompt?.closest('form');
    return {ready:location.pathname.startsWith('/images')&&!!prompt&&!!form&&form.querySelectorAll('button').length>=3};
  }
  return JSON.stringify(chatgptImageReady());
})()`

const ImageBaselineScript = `(()=>{
  function chatgptImageBaseline(){
    const fileIds=[];
    for(const img of document.querySelectorAll('main img[src*="/backend-api/estuary/content"]')){
      const match=(img.src||'').match(/[?&]id=(file_[A-Za-z0-9]+)/);
      if(match) fileIds.push(match[1]);
    }
    return {fileIds:[...new Set(fileIds)]};
  }
  return JSON.stringify(chatgptImageBaseline());
})()`

const ImageSnapshotScript = `(()=>{
  function chatgptImageSnapshot(){
    const images=[];
    for(const img of document.querySelectorAll('main img[src*="/backend-api/estuary/content"]')){
      const src=img.src||'';
      const match=src.match(/[?&]id=(file_[A-Za-z0-9]+)/);
      images.push({src,alt:img.alt||'',fileId:match?match[1]:'',width:img.naturalWidth||0,height:img.naturalHeight||0,complete:!!img.complete&&img.naturalWidth>0});
    }
    return {images};
  }
  return JSON.stringify(chatgptImageSnapshot());
})()`

func DownloadImageScript(fileID string) string {
	encoded, _ := json.Marshal(fileID)
	return fmt.Sprintf(`(async()=>{
  async function chatgptDownloadImage(){
    try{
      const fileId=%s;
      const img=document.querySelector('main img[src*="'+fileId+'"]');
      if(!img||!img.src) return {ok:false,error:'image_missing'};
      const response=await fetch(img.src,{credentials:'include'});
      if(!response.ok) return {ok:false,error:'fetch_failed',status:response.status};
      const bytes=new Uint8Array(await response.arrayBuffer());
      let binary='';
      for(let i=0;i<bytes.length;i+=32768){binary+=String.fromCharCode.apply(null,bytes.subarray(i,i+32768));}
      return {ok:true,contentType:response.headers.get('content-type')||'',base64:btoa(binary)};
    }catch(error){return {ok:false,error:String(error).slice(0,300)};}
  }
  return JSON.stringify(await chatgptDownloadImage());
})()`, string(encoded))
}

func GenerateImage(client *browser.Client, options ImageOptions) (*ImageResult, error) {
	started := time.Now()
	options.Prompt = strings.TrimSpace(options.Prompt)
	if options.Prompt == "" {
		return nil, fmt.Errorf("prompt is required")
	}
	if options.OutDir == "" {
		options.OutDir = "."
	}
	if options.Timeout <= 0 {
		options.Timeout = 3 * time.Minute
	}
	if err := client.Navigate(imagesURL, true); err != nil {
		return nil, err
	}
	if err := client.BringToFront(); err != nil {
		return nil, err
	}
	if err := waitForImagePage(client, options.Timeout); err != nil {
		return nil, err
	}
	var baselinePayload struct {
		FileIDs []string `json:"fileIds"`
	}
	if err := client.EvaluateValue(ImageBaselineScript, &baselinePayload); err != nil {
		return nil, err
	}
	baseline := make(map[string]bool, len(baselinePayload.FileIDs))
	for _, fileID := range baselinePayload.FileIDs {
		baseline[fileID] = true
	}
	if err := client.Fill("#prompt-textarea", options.Prompt); err != nil {
		return nil, err
	}
	if err := waitForImageSubmit(client, options.Timeout); err != nil {
		return nil, err
	}
	candidate, err := waitForNewImage(client, baseline, options.Timeout)
	if err != nil {
		return nil, err
	}
	var download downloadPayload
	if err := client.EvaluateValue(DownloadImageScript(candidate.FileID), &download); err != nil {
		return nil, err
	}
	imageBytes, err := decodeImagePayload(download)
	if err != nil {
		return nil, err
	}
	absDir, err := filepath.Abs(options.OutDir)
	if err != nil {
		return nil, err
	}
	if err := os.MkdirAll(absDir, 0o755); err != nil {
		return nil, err
	}
	extension := imageExtension(download.ContentType)
	path := filepath.Join(absDir, "chatgpt-"+time.Now().Format("20060102-150405")+extension)
	if err := os.WriteFile(path, imageBytes, 0o644); err != nil {
		return nil, err
	}
	return &ImageResult{
		Prompt:    options.Prompt,
		Path:      path,
		Bytes:     len(imageBytes),
		MediaType: download.ContentType,
		Caption:   candidate.Alt,
		ElapsedMS: time.Since(started).Milliseconds(),
	}, nil
}

func waitForImagePage(client *browser.Client, timeout time.Duration) error {
	deadline := time.Now().Add(timeout)
	for time.Now().Before(deadline) {
		var ready struct {
			Ready bool `json:"ready"`
		}
		if err := client.EvaluateValue(ImageReadyScript, &ready); err != nil {
			return err
		}
		if ready.Ready {
			return nil
		}
		time.Sleep(250 * time.Millisecond)
	}
	return fmt.Errorf("timed out waiting for ChatGPT image composer")
}

func waitForImageSubmit(client *browser.Client, timeout time.Duration) error {
	deadline := time.Now().Add(timeout)
	for time.Now().Before(deadline) {
		var submit SubmitResult
		if err := client.EvaluateValue(SubmitPromptScript, &submit); err != nil {
			return err
		}
		if submit.OK {
			return nil
		}
		if submit.Error != "send_button_not_ready" {
			return fmt.Errorf("submit image prompt: %s", submit.Error)
		}
		time.Sleep(250 * time.Millisecond)
	}
	return fmt.Errorf("timed out waiting for ChatGPT image send button")
}

func waitForNewImage(client *browser.Client, baseline map[string]bool, timeout time.Duration) (imageCandidate, error) {
	deadline := time.Now().Add(timeout)
	for time.Now().Before(deadline) {
		var snapshot struct {
			Images []imageCandidate `json:"images"`
		}
		if err := client.EvaluateValue(ImageSnapshotScript, &snapshot); err != nil {
			return imageCandidate{}, err
		}
		if candidate, ok := selectNewGeneratedImage(snapshot.Images, baseline); ok {
			return candidate, nil
		}
		time.Sleep(time.Second)
	}
	return imageCandidate{}, fmt.Errorf("timed out waiting for generated image")
}

func selectNewGeneratedImage(candidates []imageCandidate, baseline map[string]bool) (imageCandidate, bool) {
	for _, candidate := range candidates {
		if candidate.FileID == "" || baseline[candidate.FileID] || !candidate.Complete {
			continue
		}
		if candidate.Width < 400 || candidate.Height < 400 {
			continue
		}
		return candidate, true
	}
	return imageCandidate{}, false
}

func decodeImagePayload(payload downloadPayload) ([]byte, error) {
	if !payload.OK {
		return nil, fmt.Errorf("download image: %s (status=%d)", payload.Error, payload.Status)
	}
	if !strings.HasPrefix(strings.ToLower(payload.ContentType), "image/") {
		return nil, fmt.Errorf("unexpected image content-type %q", payload.ContentType)
	}
	decoded, err := base64.StdEncoding.DecodeString(payload.Base64)
	if err != nil {
		return nil, fmt.Errorf("decode image: %w", err)
	}
	return decoded, nil
}

func imageExtension(contentType string) string {
	switch strings.ToLower(strings.Split(contentType, ";")[0]) {
	case "image/jpeg":
		return ".jpg"
	case "image/webp":
		return ".webp"
	case "image/gif":
		return ".gif"
	default:
		return ".png"
	}
}
