package web

import (
	"embed"
	"io/fs"
)

//go:embed all:dist
var distFS embed.FS

// GetStaticFS 获取内嵌的前端静态资源文件系统
func GetStaticFS() (fs.FS, error) {
	return fs.Sub(distFS, "dist")
}
