package api

import (
	"github.com/gin-gonic/gin"

	"github.com/minshurui/dockersphere/internal/docker"
	"github.com/minshurui/dockersphere/internal/model"
)

type SystemHandler struct {
	imageSvc  docker.ImageService
	systemSvc docker.SystemService
}

func NewSystemHandler(imageSvc docker.ImageService, systemSvc docker.SystemService) *SystemHandler {
	return &SystemHandler{imageSvc: imageSvc, systemSvc: systemSvc}
}

func (h *SystemHandler) Images(c *gin.Context) {
	images, err := h.imageSvc.List(c.Request.Context())
	if err != nil {
		model.InternalError(c, err.Error())
		return
	}
	model.OK(c, images)
}

func (h *SystemHandler) ImageRemove(c *gin.Context) {
	id := c.Param("id")
	if err := h.imageSvc.Remove(c.Request.Context(), id); err != nil {
		model.InternalError(c, err.Error())
		return
	}
	model.OK(c, "removed")
}

func (h *SystemHandler) Info(c *gin.Context) {
	info, err := h.systemSvc.Info(c.Request.Context())
	if err != nil {
		model.InternalError(c, err.Error())
		return
	}
	model.OK(c, info)
}

func (h *SystemHandler) DiskUsage(c *gin.Context) {
	du, err := h.systemSvc.DiskUsage(c.Request.Context())
	if err != nil {
		model.InternalError(c, err.Error())
		return
	}
	model.OK(c, du)
}
