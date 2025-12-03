package service

import (
	"errors"
	"fmt"
	"mime/multipart"
	"path/filepath"
	"strings"
	"time"

	"goboot/config"
)

// UploadService 文件上传服务
type UploadService struct {
	storage Storage
	config  *config.UploadConfig
}

// NewUploadService 创建上传服务实例
func NewUploadService() *UploadService {
	cfg := &config.AppConfig.Upload

	// 根据配置选择存储后端
	var storage Storage
	switch cfg.StorageType {
	case "local":
		storage = NewLocalStorage()
	// case "oss":
	//     storage = NewOSSStorage()
	// case "s3":
	//     storage = NewS3Storage()
	default:
		storage = NewLocalStorage()
	}

	return &UploadService{
		storage: storage,
		config:  cfg,
	}
}

// NewUploadServiceWithStorage 使用自定义存储后端创建上传服务
func NewUploadServiceWithStorage(storage Storage) *UploadService {
	return &UploadService{
		storage: storage,
		config:  &config.AppConfig.Upload,
	}
}

// SetStorage 设置存储后端
func (s *UploadService) SetStorage(storage Storage) {
	s.storage = storage
}

// UploadFile 上传单个文件
func (s *UploadService) UploadFile(file *multipart.FileHeader, category string) (*FileInfo, error) {
	// 检查是否启用
	if !s.config.Enabled {
		return nil, errors.New("文件上传服务未启用")
	}

	// 验证文件大小
	if err := s.validateFileSize(file.Size); err != nil {
		return nil, err
	}

	// 验证文件类型
	ext := strings.ToLower(filepath.Ext(file.Filename))
	if err := s.validateFileType(ext); err != nil {
		return nil, err
	}

	// 生成存储路径
	path := s.generatePath(category)

	// 上传文件
	return s.storage.Upload(file, path, "")
}

// UploadImage 上传图片(仅允许图片格式)
func (s *UploadService) UploadImage(file *multipart.FileHeader, category string) (*FileInfo, error) {
	// 检查是否启用
	if !s.config.Enabled {
		return nil, errors.New("文件上传服务未启用")
	}

	// 验证文件大小
	if err := s.validateImageSize(file.Size); err != nil {
		return nil, err
	}

	// 验证是否为图片
	ext := strings.ToLower(filepath.Ext(file.Filename))
	if !s.isImageExt(ext) {
		return nil, fmt.Errorf("不支持的图片格式: %s，允许的格式: %v", ext, s.config.ImageExts)
	}

	// 生成存储路径
	path := s.generatePath(category)

	// 上传文件
	return s.storage.Upload(file, path, "")
}

// UploadFiles 批量上传文件
func (s *UploadService) UploadFiles(files []*multipart.FileHeader, category string) ([]*FileInfo, []error) {
	results := make([]*FileInfo, 0, len(files))
	errs := make([]error, 0)

	for _, file := range files {
		info, err := s.UploadFile(file, category)
		if err != nil {
			errs = append(errs, fmt.Errorf("%s: %v", file.Filename, err))
			continue
		}
		results = append(results, info)
	}

	return results, errs
}

// DeleteFile 删除文件
func (s *UploadService) DeleteFile(path string) error {
	return s.storage.Delete(path)
}

// GetFileInfo 获取文件信息
func (s *UploadService) GetFileInfo(path string) (*FileInfo, error) {
	return s.storage.GetInfo(path)
}

// FileExists 检查文件是否存在
func (s *UploadService) FileExists(path string) (bool, error) {
	return s.storage.Exists(path)
}

// GetFileURL 获取文件访问URL
func (s *UploadService) GetFileURL(path string) string {
	return s.storage.GetURL(path)
}

// validateFileSize 验证文件大小
func (s *UploadService) validateFileSize(size int64) error {
	maxSize := int64(s.config.MaxSize) * 1024 * 1024 // MB转字节
	if size > maxSize {
		return fmt.Errorf("文件大小超出限制，最大允许 %dMB", s.config.MaxSize)
	}
	return nil
}

// validateImageSize 验证图片大小
func (s *UploadService) validateImageSize(size int64) error {
	maxSize := int64(s.config.MaxImageSize) * 1024 * 1024 // MB转字节
	if size > maxSize {
		return fmt.Errorf("图片大小超出限制，最大允许 %dMB", s.config.MaxImageSize)
	}
	return nil
}

// validateFileType 验证文件类型
func (s *UploadService) validateFileType(ext string) error {
	// 检查是否在允许列表中
	for _, allowed := range s.config.AllowedExts {
		if ext == allowed {
			return nil
		}
	}
	return fmt.Errorf("不支持的文件格式: %s，允许的格式: %v", ext, s.config.AllowedExts)
}

// isImageExt 检查是否为图片扩展名
func (s *UploadService) isImageExt(ext string) bool {
	for _, imgExt := range s.config.ImageExts {
		if ext == imgExt {
			return true
		}
	}
	return false
}

// generatePath 生成存储路径
func (s *UploadService) generatePath(category string) string {
	now := time.Now()
	// 按日期分目录: category/2024/01/15
	return filepath.Join(category, now.Format("2006"), now.Format("01"), now.Format("02"))
}
