package service

import (
	"io"
	"mime/multipart"
	"time"
)

// FileInfo 文件信息
type FileInfo struct {
	Name      string    `json:"name"`      // 原始文件名
	Path      string    `json:"path"`      // 存储路径
	URL       string    `json:"url"`       // 访问URL
	Size      int64     `json:"size"`      // 文件大小(字节)
	MimeType  string    `json:"mimeType"`  // MIME类型
	Extension string    `json:"extension"` // 文件扩展名
	CreatedAt time.Time `json:"createdAt"` // 创建时间
}

// Storage 存储接口
// 实现此接口可以支持不同的存储后端(本地、OSS、S3等)
type Storage interface {
	// Upload 上传文件
	// file: 上传的文件
	// path: 存储路径(不含文件名)
	// filename: 文件名(为空则自动生成)
	Upload(file *multipart.FileHeader, path string, filename string) (*FileInfo, error)

	// UploadFromReader 从Reader上传文件
	// reader: 文件内容读取器
	// size: 文件大小
	// path: 存储路径(不含文件名)
	// filename: 文件名
	// mimeType: MIME类型
	UploadFromReader(reader io.Reader, size int64, path string, filename string, mimeType string) (*FileInfo, error)

	// Delete 删除文件
	// path: 文件完整路径
	Delete(path string) error

	// Exists 检查文件是否存在
	// path: 文件完整路径
	Exists(path string) (bool, error)

	// GetURL 获取文件访问URL
	// path: 文件完整路径
	GetURL(path string) string

	// GetInfo 获取文件信息
	// path: 文件完整路径
	GetInfo(path string) (*FileInfo, error)
}
