package common

import (
	"embed"
	"github.com/gin-contrib/static"
	"github.com/quantumclaw/quantumclaw/common/logger"
	"io/fs"
	"net/http"
)

// Credit: https://github.com/gin-contrib/static/issues/19

type embedFileSystem struct {
	http.FileSystem
}

func (e embedFileSystem) Exists(prefix string, path string) bool {
	_, err := e.Open(path)
	return err == nil
}

func EmbedFolder(fsEmbed embed.FS, targetPath string) static.ServeFileSystem {
	efs, err := fs.Sub(fsEmbed, targetPath)
	if err != nil {
		logger.SysWarn("embedded frontend not found at " + targetPath + ", serving without UI. Run 'npm run build' in web/ directory.")
		// 返回一个空的文件系统，避免 panic
		return embedFileSystem{
			FileSystem: http.FS(failFS{}),
		}
	}
	return embedFileSystem{
		FileSystem: http.FS(efs),
	}
}

// failFS 是一个空文件系统，所有操作返回错误
type failFS struct{}

func (f failFS) Open(name string) (fs.File, error) {
	return nil, fs.ErrNotExist
}
