package database

import (
	"database/sql"
	"log"
)

func RunMigrations(db *sql.DB) error {
	migrations := []string{
		// 用户表
		`CREATE TABLE IF NOT EXISTS users (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			username TEXT UNIQUE NOT NULL,
			email TEXT UNIQUE NOT NULL,
			password_hash TEXT NOT NULL,
			avatar_url TEXT DEFAULT '',
			role TEXT NOT NULL DEFAULT 'user',
			created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
			updated_at DATETIME DEFAULT CURRENT_TIMESTAMP
		)`,

		// 服饰表
		`CREATE TABLE IF NOT EXISTS garments (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			name TEXT NOT NULL,
			dynasty TEXT DEFAULT '',
			ethnic TEXT DEFAULT '',
			craft TEXT DEFAULT '',
			material TEXT DEFAULT '',
			pattern TEXT DEFAULT '',
			cultural_meaning TEXT DEFAULT '',
			province TEXT DEFAULT '',
			description TEXT DEFAULT '',
			image_url TEXT DEFAULT '',
			model_url TEXT DEFAULT '',
			model_id INTEGER,
			created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
			updated_at DATETIME DEFAULT CURRENT_TIMESTAMP,
			FOREIGN KEY (model_id) REFERENCES models(id) ON DELETE SET NULL
		)`,

		// 3D模型文件表
		`CREATE TABLE IF NOT EXISTS models (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			filename TEXT NOT NULL,
			original_name TEXT NOT NULL,
			file_size INTEGER NOT NULL DEFAULT 0,
			mime_type TEXT NOT NULL DEFAULT 'model/gltf-binary',
			file_path TEXT NOT NULL DEFAULT '',
			garment_id INTEGER,
			created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
			updated_at DATETIME DEFAULT CURRENT_TIMESTAMP,
			FOREIGN KEY (garment_id) REFERENCES garments(id) ON DELETE SET NULL
		)`,

		// 收藏表
		`CREATE TABLE IF NOT EXISTS favorites (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			user_id INTEGER NOT NULL,
			garment_id INTEGER NOT NULL,
			created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
			UNIQUE(user_id, garment_id),
			FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE,
			FOREIGN KEY (garment_id) REFERENCES garments(id) ON DELETE CASCADE
		)`,

		// 每日推荐表
		`CREATE TABLE IF NOT EXISTS daily_picks (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			garment_id INTEGER NOT NULL,
			pick_date DATE UNIQUE NOT NULL,
			created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
			FOREIGN KEY (garment_id) REFERENCES garments(id) ON DELETE CASCADE
		)`,

		// 索引
		`CREATE INDEX IF NOT EXISTS idx_garments_dynasty ON garments(dynasty)`,
		`CREATE INDEX IF NOT EXISTS idx_garments_ethnic ON garments(ethnic)`,
		`CREATE INDEX IF NOT EXISTS idx_garments_craft ON garments(craft)`,
		`CREATE INDEX IF NOT EXISTS idx_garments_province ON garments(province)`,
		`CREATE INDEX IF NOT EXISTS idx_garments_model_id ON garments(model_id)`,
		`CREATE INDEX IF NOT EXISTS idx_models_garment_id ON models(garment_id)`,
		`CREATE INDEX IF NOT EXISTS idx_favorites_user_id ON favorites(user_id)`,
		`CREATE INDEX IF NOT EXISTS idx_favorites_garment_id ON favorites(garment_id)`,
		`CREATE INDEX IF NOT EXISTS idx_daily_picks_date ON daily_picks(pick_date)`,

		// 用户图片投稿表
		`CREATE TABLE IF NOT EXISTS submissions (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			user_id INTEGER NOT NULL,
			image_filename TEXT NOT NULL,
			image_original_name TEXT NOT NULL,
			image_file_size INTEGER NOT NULL DEFAULT 0,
			image_mime_type TEXT NOT NULL DEFAULT 'image/png',
			image_file_path TEXT NOT NULL DEFAULT '',
			name TEXT NOT NULL DEFAULT '',
			dynasty TEXT DEFAULT '',
			ethnic TEXT DEFAULT '',
			craft TEXT DEFAULT '',
			material TEXT DEFAULT '',
			pattern TEXT DEFAULT '',
			cultural_meaning TEXT DEFAULT '',
			province TEXT DEFAULT '',
			description TEXT DEFAULT '',
			status TEXT NOT NULL DEFAULT 'pending' CHECK(status IN ('pending','approved','rejected','completed','retracted')),
			reject_reason TEXT DEFAULT '',
			admin_note TEXT DEFAULT '',
			created_garment_id INTEGER,
			created_model_id INTEGER,
			created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
			updated_at DATETIME DEFAULT CURRENT_TIMESTAMP,
			FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE,
			FOREIGN KEY (created_garment_id) REFERENCES garments(id) ON DELETE SET NULL,
			FOREIGN KEY (created_model_id) REFERENCES models(id) ON DELETE SET NULL
		)`,
		`CREATE INDEX IF NOT EXISTS idx_submissions_user_id ON submissions(user_id)`,
		`CREATE INDEX IF NOT EXISTS idx_submissions_status ON submissions(status)`,

		// 为已有数据库补加 retracted 状态支持（重建 submissions 表以更新 CHECK 约束）
		`CREATE TABLE IF NOT EXISTS submissions_v2 (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			user_id INTEGER NOT NULL,
			image_filename TEXT NOT NULL,
			image_original_name TEXT NOT NULL,
			image_file_size INTEGER NOT NULL DEFAULT 0,
			image_mime_type TEXT NOT NULL DEFAULT 'image/png',
			image_file_path TEXT NOT NULL DEFAULT '',
			name TEXT NOT NULL DEFAULT '',
			dynasty TEXT DEFAULT '',
			ethnic TEXT DEFAULT '',
			craft TEXT DEFAULT '',
			material TEXT DEFAULT '',
			pattern TEXT DEFAULT '',
			cultural_meaning TEXT DEFAULT '',
			province TEXT DEFAULT '',
			description TEXT DEFAULT '',
			status TEXT NOT NULL DEFAULT 'pending' CHECK(status IN ('pending','approved','rejected','completed','retracted')),
			reject_reason TEXT DEFAULT '',
			admin_note TEXT DEFAULT '',
			created_garment_id INTEGER,
			created_model_id INTEGER,
			created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
			updated_at DATETIME DEFAULT CURRENT_TIMESTAMP,
			FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE,
			FOREIGN KEY (created_garment_id) REFERENCES garments(id) ON DELETE SET NULL,
			FOREIGN KEY (created_model_id) REFERENCES models(id) ON DELETE SET NULL
		)`,
		`INSERT OR IGNORE INTO submissions_v2 SELECT * FROM submissions`,
		`DROP TABLE IF EXISTS submissions`,
		`ALTER TABLE submissions_v2 RENAME TO submissions`,
		`CREATE INDEX IF NOT EXISTS idx_submissions_user_id ON submissions(user_id)`,
		`CREATE INDEX IF NOT EXISTS idx_submissions_status ON submissions(status)`,
	}

	for i, m := range migrations {
		if _, err := db.Exec(m); err != nil {
			return err
		}
		log.Printf("[Migrate] 迁移 %d/%d 完成", i+1, len(migrations))
	}

	log.Println("[Migrate] 所有迁移执行完毕")
	return nil
}
