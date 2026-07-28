package api

import (
	"os"
	"path/filepath"

	"github.com/gin-gonic/gin"

	"github.com/minshurui/dockersphere/internal/model"
)

type ComposeHandler struct{}

func NewComposeHandler() *ComposeHandler {
	return &ComposeHandler{}
}

func (h *ComposeHandler) List(c *gin.Context) {
	// Collect unique compose file paths from cached container data
	// This is a lightweight endpoint - just returns projects with their file paths
	project := c.Query("project")
	if project == "" {
		model.BadRequest(c, "project required")
		return
	}
	// We don't have the container data here, so we scan common locations
	// User's known compose directories
	dirs := []string{
		"/root/media-hub",
		"/root",
		"/mnt/c/Users/Administrator/media-hub",
	}
	var files []string
	for _, dir := range dirs {
		matches, _ := filepath.Glob(filepath.Join(dir, "docker-compose*.yml"))
		files = append(files, matches...)
		matches2, _ := filepath.Glob(filepath.Join(dir, "compose*.yml"))
		files = append(files, matches2...)
	}
	model.OK(c, files)
}

func (h *ComposeHandler) Read(c *gin.Context) {
	path := c.Query("path")
	if path == "" {
		model.BadRequest(c, "path required")
		return
	}
	data, err := os.ReadFile(path)
	if err != nil {
		model.NotFound(c, "file not found: "+err.Error())
		return
	}
	model.OK(c, string(data))
}

type composeUpdateRequest struct {
	Path    string `json:"path" binding:"required"`
	Content string `json:"content" binding:"required"`
}

func (h *ComposeHandler) Update(c *gin.Context) {
	var req composeUpdateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		model.BadRequest(c, "invalid request")
		return
	}
	if err := os.WriteFile(req.Path, []byte(req.Content), 0644); err != nil {
		model.InternalError(c, "write failed: "+err.Error())
		return
	}
	model.OK(c, "saved")
}
