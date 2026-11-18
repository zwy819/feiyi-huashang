//go:build portable

package portable

import (
	"crypto/rand"
	"database/sql"
	"embed"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"io/fs"
	"log"
	"net"
	"net/http"
	"os"
	"os/exec"
	"path"
	"path/filepath"
	"strings"

	"github.com/gin-gonic/gin"
	"golang.org/x/crypto/bcrypt"
	"nonlegacy-garment-3d/server/internal/config"
	"nonlegacy-garment-3d/server/internal/model"
)

const Enabled = true

//go:embed all:assets
var assets embed.FS

// Prepare keeps mutable data beside the executable, independent of the caller's directory.
func Prepare() error {
	executable, err := os.Executable()
	if err != nil {
		return err
	}
	if err := os.Chdir(filepath.Dir(executable)); err != nil {
		return err
	}
	if err := fs.WalkDir(assets, "assets", func(name string, entry fs.DirEntry, walkErr error) error {
		if walkErr != nil {
			return walkErr
		}
		if entry.IsDir() || strings.HasPrefix(name, "assets/web/") || name == "assets/environment.example" {
			return nil
		}
		target := strings.TrimPrefix(name, "assets/")
		if strings.HasPrefix(target, "models/") {
			target = filepath.Join("data", "bundled-models", strings.TrimPrefix(target, "models/"))
		}
		if _, err := os.Stat(target); err == nil {
			return nil
		} else if !os.IsNotExist(err) {
			return err
		}
		if err := os.MkdirAll(filepath.Dir(target), 0755); err != nil {
			return err
		}
		source, err := assets.Open(name)
		if err != nil {
			return err
		}
		defer source.Close()
		file, err := os.OpenFile(target, os.O_CREATE|os.O_EXCL|os.O_WRONLY, 0644)
		if err != nil {
			return err
		}
		_, copyErr := io.Copy(file, source)
		closeErr := file.Close()
		if copyErr == nil {
			copyErr = closeErr
		}
		if copyErr != nil {
			os.Remove(target)
			return copyErr
		}
		return nil
	}); err != nil {
		return fmt.Errorf("释放运行资源: %w", err)
	}
	if _, err := os.Stat(".env"); err == nil {
		return nil
	} else if !os.IsNotExist(err) {
		return err
	}
	template, err := assets.ReadFile("assets/environment.example")
	if err != nil {
		return err
	}
	secret := make([]byte, 32)
	if _, err := rand.Read(secret); err != nil {
		return err
	}
	text := strings.ReplaceAll(string(template), "JWT_SECRET=nonlegacy_dev_secret_key", "JWT_SECRET="+hex.EncodeToString(secret))
	return os.WriteFile(".env", []byte(text), 0600)
}

// Seed imports public catalog fields only; user data and provider credentials are not bundled.
func Seed(db *sql.DB) error {
	type garment struct {
		model.Garment
		ModelFile string `json:"model_file"`
	}
	var garments []garment
	raw, err := assets.ReadFile("assets/seed/catalog.json")
	if err != nil {
		return err
	}
	if err := json.Unmarshal(raw, &garments); err != nil {
		return err
	}
	var annotations []struct {
		GarmentName string  `json:"garment_name"`
		Title       string  `json:"title"`
		Content     string  `json:"content"`
		X           float64 `json:"position_x"`
		Y           float64 `json:"position_y"`
		Z           float64 `json:"position_z"`
		SortOrder   int     `json:"sort_order"`
	}
	raw, err = assets.ReadFile("assets/seed/annotations.json")
	if err != nil {
		return err
	}
	if err := json.Unmarshal(raw, &annotations); err != nil {
		return err
	}
	tx, err := db.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()
	ids := map[string]int64{}
	for _, g := range garments {
		result, err := tx.Exec(`INSERT INTO garments (name, dynasty, ethnic, craft, material, pattern, cultural_meaning, province, description, image_url, model_url) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`, g.Name, g.Dynasty, g.Ethnic, g.Craft, g.Material, g.Pattern, g.CulturalMeaning, g.Province, g.Description, g.ImageURL, g.ModelURL)
		if err != nil {
			return err
		}
		id, err := result.LastInsertId()
		if err != nil {
			return err
		}
		ids[g.Name] = id
		modelPath := filepath.Join("data", "bundled-models", g.ModelFile)
		info, err := os.Stat(modelPath)
		if err != nil {
			return err
		}
		result, err = tx.Exec(`INSERT INTO models (filename, original_name, file_size, mime_type, file_path, garment_id) VALUES (?, ?, ?, ?, ?, ?)`, g.ModelFile, g.ModelFile, info.Size(), "model/gltf-binary", modelPath, id)
		if err != nil {
			return err
		}
		modelID, err := result.LastInsertId()
		if err != nil {
			return err
		}
		if _, err := tx.Exec("UPDATE garments SET model_id = ? WHERE id = ?", modelID, id); err != nil {
			return err
		}
	}
	for _, a := range annotations {
		id, ok := ids[a.GarmentName]
		if !ok {
			return fmt.Errorf("热点对应馆藏不存在: %s", a.GarmentName)
		}
		if _, err := tx.Exec(`INSERT INTO annotations (garment_id, title, content, position_x, position_y, position_z, sort_order) VALUES (?, ?, ?, ?, ?, ?, ?)`, id, a.Title, a.Content, a.X, a.Y, a.Z, a.SortOrder); err != nil {
			return err
		}
	}
	for _, user := range []struct{ name, password, role string }{{"admin", "admin123", "admin"}, {"user", "user123", "user"}} {
		hash, err := bcrypt.GenerateFromPassword([]byte(user.password), bcrypt.DefaultCost)
		if err != nil {
			return err
		}
		if _, err := tx.Exec(`INSERT OR IGNORE INTO users (username, email, password_hash, role) VALUES (?, ?, ?, ?)`, user.name, user.name+"@nonlegacy.cn", string(hash), user.role); err != nil {
			return err
		}
	}
	if err := tx.Commit(); err != nil {
		return err
	}
	log.Printf("[Portable] 已初始化 %d 件馆藏、%d 个热点", len(garments), len(annotations))
	return nil
}

func Mount(r *gin.Engine) {
	web, _ := fs.Sub(assets, "assets/web")
	files := http.FileServer(http.FS(web))
	r.NoRoute(func(c *gin.Context) {
		if (c.Request.Method != http.MethodGet && c.Request.Method != http.MethodHead) || c.Request.URL.Path == "/api" || strings.HasPrefix(c.Request.URL.Path, "/api/") {
			c.Status(http.StatusNotFound)
			return
		}
		name := strings.TrimPrefix(path.Clean(c.Request.URL.Path), "/")
		if info, err := fs.Stat(web, name); err == nil && !info.IsDir() {
			files.ServeHTTP(c.Writer, c.Request)
			return
		}
		if path.Ext(name) != "" {
			c.Status(http.StatusNotFound)
			return
		}
		index, _ := assets.ReadFile("assets/web/index.html")
		c.Data(http.StatusOK, "text/html; charset=utf-8", index)
	})
}

func Run(r *gin.Engine, cfg *config.Config, openBrowser bool) error {
	listener, err := net.Listen("tcp", net.JoinHostPort("127.0.0.1", cfg.Port))
	if err != nil {
		return err
	}
	url := "http://localhost:" + cfg.Port + "/"
	log.Printf("[Portable] 网站已启动: %s；关闭本窗口或按 Ctrl+C 停止", url)
	if openBrowser {
		if err := exec.Command("rundll32", "url.dll,FileProtocolHandler", url).Start(); err != nil {
			log.Printf("请在浏览器打开 %s: %v", url, err)
		}
	}
	return http.Serve(listener, r)
}
