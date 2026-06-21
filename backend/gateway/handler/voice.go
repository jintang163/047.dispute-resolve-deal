package handler

import (
	"context"
	"encoding/base64"
	"fmt"
	"io"
	"net/http"
	"path/filepath"
	"strings"

	"github.com/dispute-resolve/common/aliyun"
	"github.com/dispute-resolve/common/response"

	"github.com/cloudwego/hertz/pkg/app"
)

type VoiceRecognizeRequest struct {
	FileName   string `json:"fileName"`
	FileBase64 string `json:"fileBase64"`
	Format     string `json:"format"`
}

func VoiceRecognize(ctx context.Context, c *app.RequestContext) {
	contentType := string(c.ContentType())
	var audioData []byte
	var format string
	var fileName string

	if strings.Contains(contentType, "multipart/form-data") {
		file, err := c.FormFile("file")
		if err != nil {
			c.JSON(http.StatusBadRequest, response.BadRequest("请选择要识别的音频文件"))
			return
		}

		const maxFileSize = 10 * 1024 * 1024
		if file.Size > maxFileSize {
			c.JSON(http.StatusBadRequest, response.BadRequest("音频文件大小不能超过10MB"))
			return
		}

		ext := strings.ToLower(strings.TrimPrefix(filepath.Ext(file.Filename), "."))
		allowedFormats := map[string]bool{
			"mp3": true, "wav": true, "m4a": true, "amr": true, "aac": true, "webm": true,
		}
		if !allowedFormats[ext] {
			c.JSON(http.StatusBadRequest, response.BadRequest(
				fmt.Sprintf("不支持的音频格式 %s，支持格式：mp3, wav, m4a, amr, aac, webm", ext),
			))
			return
		}

		src, err := file.Open()
		if err != nil {
			c.JSON(http.StatusInternalServerError, response.ServerError("文件读取失败"))
			return
		}
		defer src.Close()

		data, err := io.ReadAll(src)
		if err != nil {
			c.JSON(http.StatusInternalServerError, response.ServerError("文件读取失败"))
			return
		}

		audioData = data
		format = ext
		fileName = file.Filename
	} else {
		var req VoiceRecognizeRequest
		if err := c.BindAndValidate(&req); err != nil {
			c.JSON(http.StatusBadRequest, response.BadRequest("参数错误: "+err.Error()))
			return
		}

		if req.FileBase64 == "" {
			c.JSON(http.StatusBadRequest, response.BadRequest("音频数据不能为空"))
			return
		}

		if req.Format == "" && req.FileName != "" {
			req.Format = strings.ToLower(strings.TrimPrefix(filepath.Ext(req.FileName), "."))
		}

		allowedFormats := map[string]bool{
			"mp3": true, "wav": true, "m4a": true, "amr": true, "aac": true, "webm": true,
		}
		if !allowedFormats[req.Format] {
			c.JSON(http.StatusBadRequest, response.BadRequest(
				fmt.Sprintf("不支持的音频格式 %s，支持格式：mp3, wav, m4a, amr, aac, webm", req.Format),
			))
			return
		}

		base64Str := req.FileBase64
		if strings.Contains(base64Str, ",") {
			base64Str = strings.SplitN(base64Str, ",", 2)[1]
		}

		data, err := base64.StdEncoding.DecodeString(base64Str)
		if err != nil {
			c.JSON(http.StatusBadRequest, response.BadRequest("音频base64解码失败"))
			return
		}

		const maxFileSize = 10 * 1024 * 1024
		if len(data) > maxFileSize {
			c.JSON(http.StatusBadRequest, response.BadRequest("音频文件大小不能超过10MB"))
			return
		}

		audioData = data
		format = req.Format
		fileName = req.FileName
		if fileName == "" {
			fileName = "audio." + format
		}
	}

	client := aliyun.GetVoiceClient()
	result, err := client.RecognizeSpeech(audioData, format)
	if err != nil {
		c.JSON(http.StatusInternalServerError, response.ServerError("语音识别失败："+err.Error()))
		return
	}

	c.JSON(http.StatusOK, response.Success(map[string]interface{}{
		"text":     result.Text,
		"duration": result.Duration,
		"taskId":   result.TaskID,
		"format":   format,
		"fileSize": len(audioData),
		"fileName": fileName,
	}))
}
