package service

import (
	"fmt"
	"io"
	"mime/multipart"
	"os"
	"path/filepath"
	"strings"
	"time"

	"nonlegacy-garment-3d/server/internal/model"
)

type ModelService struct {
	modelStore   *model.ModelStore
	garmentStore *model.GarmentStore
	uploadDir    string
	maxSize      int64
}

func NewModelService(modelStore *model.ModelStore, garmentStore *model.GarmentStore, uploadDir string, maxSize int64) *ModelService {
	return &ModelService{
		modelStore:   modelStore,
		garmentStore: garmentStore,
		uploadDir:    uploadDir,
		maxSize:      maxSize,
	}
}

func (s *ModelService) Upload(file *multipart.FileHeader, garmentID *int64) (*model.Model3D, error) {
	// 验证文件扩展名
	ext := strings.ToLower(filepath.Ext(file.Filename))
	if ext != ".glb" && ext != ".gltf" {
		return nil, fmt.Errorf("不支持的文件格式，仅支持 .glb 和 .gltf")
	}

	// 验证文件大小
	if file.Size > s.maxSize {
		return nil, fmt.Errorf("文件大小超过限制 (50MB)")
	}

	// 验证 garment_id（如果提供）
	if garmentID != nil {
		if _, err := s.garmentStore.GetByID(*garmentID); err != nil {
			return nil, fmt.Errorf("关联的服饰不存在")
		}
	}

	// 创建模型记录（先获取 ID）
	m := &model.Model3D{
		OriginalName: file.Filename,
		FileSize:     file.Size,
		MimeType:     "model/gltf-binary",
		GarmentID:    garmentID,
	}

	id, err := s.modelStore.Create(m)
	if err != nil {
		return nil, fmt.Errorf("创建模型记录失败: %w", err)
	}

	// 生成存储文件名
	timestamp := time.Now().Unix()
	safeName := fmt.Sprintf("%d_%s", timestamp, sanitizeFilename(file.Filename))
	modelDir := filepath.Join(s.uploadDir, fmt.Sprintf("%d", id))
	if err := os.MkdirAll(modelDir, 0755); err != nil {
		s.modelStore.Delete(id)
		return nil, fmt.Errorf("创建存储目录失败: %w", err)
	}

	filePath := filepath.Join(modelDir, safeName)
	relPath := filepath.Join(s.uploadDir, fmt.Sprintf("%d", id), safeName)

	// 保存文件
	src, err := file.Open()
	if err != nil {
		s.modelStore.Delete(id)
		return nil, fmt.Errorf("读取上传文件失败: %w", err)
	}
	defer src.Close()

	dst, err := os.Create(filePath)
	if err != nil {
		s.modelStore.Delete(id)
		return nil, fmt.Errorf("创建文件失败: %w", err)
	}
	defer dst.Close()

	if _, err := io.Copy(dst, src); err != nil {
		s.modelStore.Delete(id)
		os.Remove(filePath)
		return nil, fmt.Errorf("保存文件失败: %w", err)
	}

	// 更新模型记录
	m.ID = id
	m.Filename = safeName
	m.FilePath = relPath
	if err := s.modelStore.Update(id, m); err != nil {
		return nil, fmt.Errorf("更新模型记录失败: %w", err)
	}

	// 如果关联了 garment，更新 garment 的 model_id
	if garmentID != nil {
		s.garmentStore.UpdateModelID(*garmentID, &id)
	}

	return s.modelStore.GetByID(id)
}

func (s *ModelService) Delete(id int64) error {
	m, err := s.modelStore.GetByID(id)
	if err != nil {
		return fmt.Errorf("模型不存在")
	}

	// 如果关联了 garment，解除关联
	if m.GarmentID != nil {
		s.garmentStore.UpdateModelID(*m.GarmentID, nil)
	}

	// 删除数据库记录
	if err := s.modelStore.Delete(id); err != nil {
		return fmt.Errorf("删除模型记录失败: %w", err)
	}

	// 删除文件
	modelDir := filepath.Join(s.uploadDir, fmt.Sprintf("%d", id))
	os.RemoveAll(modelDir)

	return nil
}

func sanitizeFilename(name string) string {
	name = strings.ReplaceAll(name, " ", "_")
	name = strings.Map(func(r rune) rune {
		if (r >= 'a' && r <= 'z') || (r >= 'A' && r <= 'Z') || (r >= '0' && r <= '9') || r == '_' || r == '.' || r == '-' {
			return r
		}
		return '_'
	}, name)
	return name
}
