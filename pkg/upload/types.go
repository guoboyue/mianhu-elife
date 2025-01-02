package upload

import (
	"fmt"
	"image"
	_ "image/gif"
	_ "image/jpeg"
	_ "image/png"
	"io"
	"mime/multipart"
	"path"
	"strings"
)

// CheckUploadType 检查上传类型是否合法
func CheckUploadType(uploadType string) bool {
	_, exists := TypeConfigs[uploadType]
	return exists
}

// ValidateUpload 验证上传文件
func ValidateUpload(file multipart.File, header *multipart.FileHeader, uploadType string) error {
	config, exists := TypeConfigs[uploadType]
	if !exists {
		return fmt.Errorf("不支持的上传类型: %s", uploadType)
	}

	// 1. 验证文件大小
	if header.Size > config.MaxSize {
		return fmt.Errorf("文件大小超过限制，最大允许 %d bytes", config.MaxSize)
	}

	// 2. 验证文件扩展名
	ext := strings.ToLower(path.Ext(header.Filename))
	validExt := false
	for _, allowExt := range config.AllowExts {
		if ext == allowExt {
			validExt = true
			break
		}
	}
	if !validExt {
		return fmt.Errorf("不支持的文件类型: %s", ext)
	}

	// 3. 验证图片尺寸（如果配置了尺寸限制）
	if config.MaxWidth > 0 || config.MaxHeight > 0 {
		// 获取图片信息
		imgConfig, _, err := image.DecodeConfig(file)
		if err != nil {
			return fmt.Errorf("无法解析图片: %v", err)
		}

		// 重置文件指针
		file.Seek(0, io.SeekStart)

		// 检查尺寸
		if config.MaxWidth > 0 && imgConfig.Width > config.MaxWidth {
			return fmt.Errorf("图片宽度超过限制，最大允许 %d 像素", config.MaxWidth)
		}
		if config.MaxHeight > 0 && imgConfig.Height > config.MaxHeight {
			return fmt.Errorf("图片高度超过限制，最大允许 %d 像素", config.MaxHeight)
		}
		if config.MinWidth > 0 && imgConfig.Width < config.MinWidth {
			return fmt.Errorf("图片宽度不足，最小要求 %d 像素", config.MinWidth)
		}
		if config.MinHeight > 0 && imgConfig.Height < config.MinHeight {
			return fmt.Errorf("图片高度不足，最小要求 %d 像素", config.MinHeight)
		}
	}

	return nil
}
