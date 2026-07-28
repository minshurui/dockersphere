package api

import (
	"bytes"
	"os"
	"os/exec"
	"path/filepath"

	"github.com/gin-gonic/gin"

	"github.com/minshurui/dockersphere/internal/model"
)

type ComposeHandler struct {
	deployDir string
}

func NewComposeHandler() *ComposeHandler {
	return &ComposeHandler{deployDir: "/tmp/dockersphere-deploy"}
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

func (h *ComposeHandler) Update(c *gin.Context) {
	var req struct {
		Path    string `json:"path" binding:"required"`
		Content string `json:"content" binding:"required"`
	}
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

type deployRequest struct {
	Project string `json:"project" binding:"required"`
	Content string `json:"content" binding:"required"`
}

func (h *ComposeHandler) Deploy(c *gin.Context) {
	var req deployRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		model.BadRequest(c, "project and content required")
		return
	}

	dir := filepath.Join(h.deployDir, req.Project)
	if err := os.MkdirAll(dir, 0755); err != nil {
		model.InternalError(c, "create dir: "+err.Error())
		return
	}
	composePath := filepath.Join(dir, "docker-compose.yml")
	if err := os.WriteFile(composePath, []byte(req.Content), 0644); err != nil {
		model.InternalError(c, "write file: "+err.Error())
		return
	}

	cmd := exec.Command("docker", "compose", "-p", req.Project, "-f", composePath, "up", "-d")
	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	if err := cmd.Run(); err != nil {
		model.OK(c, gin.H{"success": false, "output": stdout.String() + stderr.String(), "error": err.Error()})
		return
	}
	model.OK(c, gin.H{"success": true, "output": stdout.String() + stderr.String()})
}
