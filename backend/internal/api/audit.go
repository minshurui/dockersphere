package api

import (
	"strconv"

	"github.com/gin-gonic/gin"

	"github.com/yourname/dockersphere/internal/audit"
	"github.com/yourname/dockersphere/internal/model"
)

// AuditHandler handles audit log requests.
type AuditHandler struct {
	store *audit.Store
}

// NewAuditHandler creates a new AuditHandler.
func NewAuditHandler(store *audit.Store) *AuditHandler {
	return &AuditHandler{store: store}
}

// List returns recent audit records.
func (h *AuditHandler) List(c *gin.Context) {
	limit := 50
	if l := c.Query("limit"); l != "" {
		if n, err := strconv.Atoi(l); err == nil && n > 0 {
			limit = n
		}
	}
	records, err := h.store.List(c.Request.Context(), limit)
	if err != nil {
		model.InternalError(c, err.Error())
		return
	}
	model.OK(c, records)
}
