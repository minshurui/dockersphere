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

type projectActionRequest struct {
	Project string `form:"project"`
	Action  string `json:"action"`
}

func (h *ComposeHandler) ProjectAction(c *gin.Context) {
	project := c.Param("project")
	action := c.Param("action")

	// Find compose file from the project's deployed dir or common locations
	dirs := []string{
		filepath.Join(h.deployDir, project),
		"/root/" + project,
		"/mnt/c/Users/Administrator/" + project,
	}
	var composePath string
	for _, dir := range dirs {
		// Prefer the main docker-compose.yml over variant files
		for _, name := range []string{"docker-compose.yml", "compose.yml", "docker-compose.yaml"} {
			candidate := filepath.Join(dir, name)
			if _, err := os.Stat(candidate); err == nil {
				composePath = candidate
				break
			}
		}
		if composePath != "" {
			break
		}
	}
	if composePath == "" {
		model.NotFound(c, "compose file not found for project: "+project)
		return
	}

	args := []string{"compose", "-p", project, "-f", composePath}
	switch action {
	case "up":
		args = append(args, "up", "-d")
	case "down":
		args = append(args, "down")
	case "restart":
		args = append(args, "restart")
	case "stop":
		args = append(args, "stop")
	case "start":
		args = append(args, "start")
	default:
		model.BadRequest(c, "invalid action: "+action)
		return
	}

	cmd := exec.Command("docker", args...)
	// Run from the compose file's directory to pick up .env files
	cmd.Dir = filepath.Dir(composePath)
	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	if err := cmd.Run(); err != nil {
		model.OK(c, gin.H{"success": false, "output": stdout.String() + stderr.String(), "error": err.Error()})
		return
	}
	model.OK(c, gin.H{"success": true, "output": stdout.String() + stderr.String()})
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
	cmd.Dir = filepath.Dir(composePath)
	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	if err := cmd.Run(); err != nil {
		model.OK(c, gin.H{"success": false, "output": stdout.String() + stderr.String(), "error": err.Error()})
		return
	}
	model.OK(c, gin.H{"success": true, "output": stdout.String() + stderr.String()})
}
