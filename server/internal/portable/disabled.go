//go:build !portable

package portable

import (
	"database/sql"

	"github.com/gin-gonic/gin"
	"nonlegacy-garment-3d/server/internal/config"
)

const Enabled = false

func Prepare() error                                    { return nil }
func Seed(_ *sql.DB) error                              { return nil }
func Mount(_ *gin.Engine)                               {}
func Run(_ *gin.Engine, _ *config.Config, _ bool) error { return nil }
