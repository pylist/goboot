package service

import (
	"fmt"
	"io"
	"mime/multipart"
	"os"
	"path/filepath"
	"strings"
	"time"

	"goboot/config"

	"github.com/google/uuid"
)

// LocalStorage 本地文件存储实现
type LocalStorage struct {
	basePath string // 文件存储根目录
	baseURL  string // 文件访问URL前缀
}

// NewLocalStorage 创建本地存储实例
func NewLocalStorage() *LocalStorage {
	cfg := &config.AppConfig.Upload
	return &LocalStorage{
		basePath: cfg.LocalPath,
		baseURL:  cfg.BaseURL,
	}
}

// Upload 上传文件
func (s *LocalStorage) Upload(file *multipart.FileHeader, path string, filename string) (*FileInfo, error) {
	// 打开上传的文件
	src, err := file.Open()
	if err != nil {
		return nil, fmt.Errorf("打开上传文件失败: %v", err)
	}
	defer src.Close()

	// 获取文件扩展名
	ext := strings.ToLower(filepath.Ext(file.Filename))

	// 生成文件名
	if filename == "" {
		filename = s.generateFilename(ext)
	} else if !strings.HasSuffix(strings.ToLower(filename), ext) {
		filename = filename + ext
	}

	// 完整存储路径
	fullPath := filepath.Join(s.basePath, path)

	// 确保目录存在
	if err := os.MkdirAll(fullPath, 0755); err != nil {
		return nil, fmt.Errorf("创建目录失败: %v", err)
	}

	// 完整文件路径
	filePath := filepath.Join(fullPath, filename)

	// 创建目标文件
	dst, err := os.Create(filePath)
	if err != nil {
		return nil, fmt.Errorf("创建目标文件失败: %v", err)
	}
	defer dst.Close()

	// 复制文件内容
	if _, err := io.Copy(dst, src); err != nil {
		os.Remove(filePath) // 清理失败的文件
		return nil, fmt.Errorf("写入文件失败: %v", err)
	}

	// 获取文件信息
	stat, err := os.Stat(filePath)
	if err != nil {
		return nil, fmt.Errorf("获取文件信息失败: %v", err)
	}

	// 相对路径(用于存储和URL)
	relativePath := filepath.Join(path, filename)

	return &FileInfo{
		Name:      file.Filename,
		Path:      relativePath,
		URL:       s.GetURL(relativePath),
		Size:      stat.Size(),
		MimeType:  file.Header.Get("Content-Type"),
		Extension: ext,
		CreatedAt: time.Now(),
	}, nil
}

// UploadFromReader 从Reader上传文件
func (s *LocalStorage) UploadFromReader(reader io.Reader, size int64, path string, filename string, mimeType string) (*FileInfo, error) {
	// 获取扩展名
	ext := strings.ToLower(filepath.Ext(filename))

	// 完整存储路径
	fullPath := filepath.Join(s.basePath, path)

	// 确保目录存在
	if err := os.MkdirAll(fullPath, 0755); err != nil {
		return nil, fmt.Errorf("创建目录失败: %v", err)
	}

	// 完整文件路径
	filePath := filepath.Join(fullPath, filename)

	// 创建目标文件
	dst, err := os.Create(filePath)
	if err != nil {
		return nil, fmt.Errorf("创建目标文件失败: %v", err)
	}
	defer dst.Close()

	// 复制文件内容
	written, err := io.Copy(dst, reader)
	if err != nil {
		os.Remove(filePath) // 清理失败的文件
		return nil, fmt.Errorf("写入文件失败: %v", err)
	}

	// 相对路径(用于存储和URL)
	relativePath := filepath.Join(path, filename)

	return &FileInfo{
		Name:      filename,
		Path:      relativePath,
		URL:       s.GetURL(relativePath),
		Size:      written,
		MimeType:  mimeType,
		Extension: ext,
		CreatedAt: time.Now(),
	}, nil
}

// Delete 删除文件
func (s *LocalStorage) Delete(path string) error {
	fullPath := filepath.Join(s.basePath, path)

	// 检查文件是否存在
	if _, err := os.Stat(fullPath); os.IsNotExist(err) {
		return nil // 文件不存在，视为删除成功
	}

	if err := os.Remove(fullPath); err != nil {
		return fmt.Errorf("删除文件失败: %v", err)
	}

	return nil
}

// Exists 检查文件是否存在
func (s *LocalStorage) Exists(path string) (bool, error) {
	fullPath := filepath.Join(s.basePath, path)
	_, err := os.Stat(fullPath)
	if err == nil {
		return true, nil
	}
	if os.IsNotExist(err) {
		return false, nil
	}
	return false, err
}

// GetURL 获取文件访问URL
func (s *LocalStorage) GetURL(path string) string {
	// 将路径分隔符统一为URL格式
	urlPath := strings.ReplaceAll(path, string(os.PathSeparator), "/")
	return s.baseURL + "/" + urlPath
}

// GetInfo 获取文件信息
func (s *LocalStorage) GetInfo(path string) (*FileInfo, error) {
	fullPath := filepath.Join(s.basePath, path)

	stat, err := os.Stat(fullPath)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, fmt.Errorf("文件不存在")
		}
		return nil, fmt.Errorf("获取文件信息失败: %v", err)
	}

	ext := strings.ToLower(filepath.Ext(path))
	mimeType := getMimeType(ext)

	return &FileInfo{
		Name:      stat.Name(),
		Path:      path,
		URL:       s.GetURL(path),
		Size:      stat.Size(),
		MimeType:  mimeType,
		Extension: ext,
		CreatedAt: stat.ModTime(),
	}, nil
}

// generateFilename 生成唯一文件名
func (s *LocalStorage) generateFilename(ext string) string {
	return uuid.New().String() + ext
}

// getMimeType 根据扩展名获取MIME类型
func getMimeType(ext string) string {
	mimeTypes := map[string]string{
		".jpg":  "image/jpeg",
		".jpeg": "image/jpeg",
		".png":  "image/png",
		".gif":  "image/gif",
		".webp": "image/webp",
		".svg":  "image/svg+xml",
		".ico":  "image/x-icon",
		".bmp":  "image/bmp",
		".pdf":  "application/pdf",
		".doc":  "application/msword",
		".docx": "application/vnd.openxmlformats-officedocument.wordprocessingml.document",
		".xls":  "application/vnd.ms-excel",
		".xlsx": "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet",
		".ppt":  "application/vnd.ms-powerpoint",
		".pptx": "application/vnd.openxmlformats-officedocument.presentationml.presentation",
		".zip":  "application/zip",
		".rar":  "application/x-rar-compressed",
		".7z":   "application/x-7z-compressed",
		".tar":  "application/x-tar",
		".gz":   "application/gzip",
		".mp3":  "audio/mpeg",
		".wav":  "audio/wav",
		".mp4":  "video/mp4",
		".avi":  "video/x-msvideo",
		".mov":  "video/quicktime",
		".txt":  "text/plain",
		".html": "text/html",
		".css":  "text/css",
		".js":   "application/javascript",
		".json": "application/json",
		".xml":  "application/xml",
	}

	if mime, ok := mimeTypes[ext]; ok {
		return mime
	}
	return "application/octet-stream"
}
