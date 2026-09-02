package service

import (
	"fmt"
	"io"
	"log"
	"mime/multipart"
	"os"
	"path/filepath"
	"strings"
	"time"

	"nonlegacy-garment-3d/server/internal/model"
)

// SubmissionMetadata 投稿元数据（从表单中提取）
type SubmissionMetadata struct {
	Name            string
	Dynasty         string
	Ethnic          string
	Craft           string
	Material        string
	Pattern         string
	CulturalMeaning string
	Province        string
	Description     string
}

// SubmissionService 投稿业务逻辑
type SubmissionService struct {
	submissionStore *model.SubmissionStore
	uploadDir       string
	maxImageSize    int64
}

// NewSubmissionService 创建投稿服务
func NewSubmissionService(submissionStore *model.SubmissionStore, uploadDir string, maxImageSize int64) *SubmissionService {
	return &SubmissionService{
		submissionStore: submissionStore,
		uploadDir:       uploadDir,
		maxImageSize:    maxImageSize,
	}
}

// AllowedImageExts 允许的图片扩展名
var AllowedImageExts = map[string]string{
	".jpg":  "image/jpeg",
	".jpeg": "image/jpeg",
	".png":  "image/png",
	".webp": "image/webp",
}

// Upload 保存用户上传的图片并创建投稿记录
func (s *SubmissionService) Upload(file *multipart.FileHeader, userID int64, metadata SubmissionMetadata) (*model.Submission, error) {
	// 验证文件扩展名
	ext := strings.ToLower(filepath.Ext(file.Filename))
	mimeType, ok := AllowedImageExts[ext]
	if !ok {
		return nil, fmt.Errorf("不支持的图片格式，仅支持 JPG、PNG、WebP")
	}

	// 验证文件大小
	if file.Size > s.maxImageSize {
		return nil, fmt.Errorf("图片大小超过限制 (%dMB)", s.maxImageSize/(1024*1024))
	}

	// 创建投稿记录（先获取 ID）
	sub := &model.Submission{
		UserID:            userID,
		ImageOriginalName: file.Filename,
		ImageFileSize:     file.Size,
		ImageMimeType:     mimeType,
		Name:              metadata.Name,
		Dynasty:           metadata.Dynasty,
		Ethnic:            metadata.Ethnic,
		Craft:             metadata.Craft,
		Material:          metadata.Material,
		Pattern:           metadata.Pattern,
		CulturalMeaning:   metadata.CulturalMeaning,
		Province:          metadata.Province,
		Description:       metadata.Description,
		Status:            "pending",
	}

	id, err := s.submissionStore.Create(sub)
	if err != nil {
		return nil, fmt.Errorf("创建投稿记录失败: %w", err)
	}

	// 生成存储文件名并保存图片
	timestamp := time.Now().Unix()
	safeName := fmt.Sprintf("%d_%s", timestamp, sanitizeImageFilename(file.Filename))
	imageDir := filepath.Join(s.uploadDir, "submissions", fmt.Sprintf("%d", id))
	if err := os.MkdirAll(imageDir, 0755); err != nil {
		s.submissionStore.Delete(id)
		return nil, fmt.Errorf("创建存储目录失败: %w", err)
	}

	filePath := filepath.Join(imageDir, safeName)
	relPath := filepath.Join(s.uploadDir, "submissions", fmt.Sprintf("%d", id), safeName)

	// 写入文件
	src, err := file.Open()
	if err != nil {
		s.submissionStore.Delete(id)
		return nil, fmt.Errorf("读取上传图片失败: %w", err)
	}
	defer src.Close()

	dst, err := os.Create(filePath)
	if err != nil {
		s.submissionStore.Delete(id)
		return nil, fmt.Errorf("创建文件失败: %w", err)
	}
	defer dst.Close()

	if _, err := io.Copy(dst, src); err != nil {
		s.submissionStore.Delete(id)
		os.Remove(filePath)
		return nil, fmt.Errorf("保存图片失败: %w", err)
	}

	// 更新记录（写入文件路径信息）
	sub.ID = id
	sub.ImageFilename = safeName
	sub.ImageFilePath = relPath
	_ = s.submissionStore.UpdateFilePath(id, safeName, relPath)

	return s.submissionStore.GetByID(id)
}

// DeleteSubmission 删除投稿（含文件）
func (s *SubmissionService) DeleteSubmission(id int64) error {
	sub, err := s.submissionStore.GetByID(id)
	if err != nil {
		return fmt.Errorf("投稿不存在")
	}

	// 先删除数据库记录，再清理文件（避免文件锁导致 DB 记录残留）
	if err := s.submissionStore.Delete(id); err != nil {
		return fmt.Errorf("删除投稿记录失败: %w", err)
	}

	// 删除图片文件（best-effort，失败不影响整体结果）
	if sub.ImageFilePath != "" {
		imageDir := filepath.Dir(sub.ImageFilePath)
		if err := os.RemoveAll(imageDir); err != nil {
			log.Printf("[Submission] 删除投稿 #%d 的图片目录失败: %v (路径: %s)", id, err, imageDir)
		}
	}

	return nil
}

// sanitizeImageFilename 净化文件名
func sanitizeImageFilename(name string) string {
	name = strings.ReplaceAll(name, " ", "_")
	name = strings.Map(func(r rune) rune {
		if (r >= 'a' && r <= 'z') || (r >= 'A' && r <= 'Z') || (r >= '0' && r <= '9') || r == '_' || r == '.' || r == '-' {
			return r
		}
		return '_'
	}, name)
	return name
}
