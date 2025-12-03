package handler

import (
	"goboot/internal/model"
	"goboot/internal/service"
	"goboot/pkg/response"

	"github.com/gofiber/fiber/v3"
)

type UploadHandler struct {
	uploadService *service.UploadService
	auditService  *service.AuditService
}

func NewUploadHandler() *UploadHandler {
	return &UploadHandler{
		uploadService: service.NewUploadService(),
		auditService:  service.NewAuditService(),
	}
}

// UploadFile 上传单个文件
// @Summary 上传文件
// @Description 上传单个文件，支持多种格式
// @Tags 文件上传
// @Accept multipart/form-data
// @Produce json
// @Param file formance file true "上传的文件"
// @Param category formance string false "文件分类目录"
// @Success 200 {object} response.Response{data=service.FileInfo}
// @Router /api/upload/file [post]
func (h *UploadHandler) UploadFile(c fiber.Ctx) error {
	// 获取上传的文件
	file, err := c.FormFile("file")
	if err != nil {
		return response.Fail(c, "获取上传文件失败: "+err.Error())
	}

	// 获取分类目录(可选)
	category := c.FormValue("category", "files")

	// 上传文件
	fileInfo, err := h.uploadService.UploadFile(file, category)
	if err != nil {
		h.auditService.LogFail(c, model.ActionUpload, model.ModuleFile, file.Filename, err.Error())
		return response.Fail(c, err.Error())
	}

	// 记录审计日志
	h.auditService.LogSuccess(c, model.ActionUpload, model.ModuleFile, fileInfo.Path, "上传文件成功")

	return response.Success(c, fileInfo)
}

// UploadImage 上传图片
// @Summary 上传图片
// @Description 上传单个图片，仅支持图片格式
// @Tags 文件上传
// @Accept multipart/form-data
// @Produce json
// @Param file formData file true "上传的图片"
// @Param category formData string false "图片分类目录"
// @Success 200 {object} response.Response{data=service.FileInfo}
// @Router /api/upload/image [post]
func (h *UploadHandler) UploadImage(c fiber.Ctx) error {
	// 获取上传的文件
	file, err := c.FormFile("file")
	if err != nil {
		return response.Fail(c, "获取上传文件失败: "+err.Error())
	}

	// 获取分类目录(可选)
	category := c.FormValue("category", "images")

	// 上传图片
	fileInfo, err := h.uploadService.UploadImage(file, category)
	if err != nil {
		h.auditService.LogFail(c, model.ActionUpload, model.ModuleFile, file.Filename, err.Error())
		return response.Fail(c, err.Error())
	}

	// 记录审计日志
	h.auditService.LogSuccess(c, model.ActionUpload, model.ModuleFile, fileInfo.Path, "上传图片成功")

	return response.Success(c, fileInfo)
}

// UploadFiles 批量上传文件
// @Summary 批量上传文件
// @Description 同时上传多个文件
// @Tags 文件上传
// @Accept multipart/form-data
// @Produce json
// @Param files formData file true "上传的文件列表"
// @Param category formData string false "文件分类目录"
// @Success 200 {object} response.Response{data=UploadFilesResponse}
// @Router /api/upload/files [post]
func (h *UploadHandler) UploadFiles(c fiber.Ctx) error {
	// 获取表单
	form, err := c.MultipartForm()
	if err != nil {
		return response.Fail(c, "解析表单失败: "+err.Error())
	}

	// 获取文件列表
	files := form.File["files"]
	if len(files) == 0 {
		return response.Fail(c, "请选择要上传的文件")
	}

	// 获取分类目录(可选)
	category := c.FormValue("category", "files")

	// 批量上传
	results, errs := h.uploadService.UploadFiles(files, category)

	// 构建错误信息
	var errMsgs []string
	for _, e := range errs {
		errMsgs = append(errMsgs, e.Error())
	}

	// 记录审计日志
	if len(results) > 0 {
		h.auditService.LogSuccess(c, model.ActionUpload, model.ModuleFile, "",
			"批量上传成功"+string(rune(len(results)))+"个文件")
	}

	return response.Success(c, fiber.Map{
		"success": results,
		"errors":  errMsgs,
		"total":   len(files),
		"failed":  len(errs),
	})
}

// DeleteFile 删除文件
// @Summary 删除文件
// @Description 根据路径删除文件
// @Tags 文件上传
// @Accept json
// @Produce json
// @Param body body DeleteFileRequest true "删除文件请求"
// @Success 200 {object} response.Response
// @Router /api/upload/delete [post]
func (h *UploadHandler) DeleteFile(c fiber.Ctx) error {
	var req DeleteFileRequest
	if err := c.Bind().Body(&req); err != nil {
		return response.Fail(c, "参数错误: "+err.Error())
	}

	if req.Path == "" {
		return response.Fail(c, "文件路径不能为空")
	}

	// 删除文件
	if err := h.uploadService.DeleteFile(req.Path); err != nil {
		h.auditService.LogFail(c, model.ActionDelete, model.ModuleFile, req.Path, err.Error())
		return response.Fail(c, "删除文件失败: "+err.Error())
	}

	// 记录审计日志
	h.auditService.LogSuccess(c, model.ActionDelete, model.ModuleFile, req.Path, "删除文件成功")

	return response.SuccessWithMessage(c, "删除成功", nil)
}

// GetFileInfo 获取文件信息
// @Summary 获取文件信息
// @Description 根据路径获取文件信息
// @Tags 文件上传
// @Accept json
// @Produce json
// @Param path query string true "文件路径"
// @Success 200 {object} response.Response{data=service.FileInfo}
// @Router /api/upload/info [get]
func (h *UploadHandler) GetFileInfo(c fiber.Ctx) error {
	path := c.Query("path")
	if path == "" {
		return response.Fail(c, "文件路径不能为空")
	}

	// 获取文件信息
	info, err := h.uploadService.GetFileInfo(path)
	if err != nil {
		return response.Fail(c, err.Error())
	}

	return response.Success(c, info)
}

// DeleteFileRequest 删除文件请求
type DeleteFileRequest struct {
	Path string `json:"path" validate:"required"`
}
