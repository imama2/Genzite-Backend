package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/imama2/Genzite-Backend/internal/services/template/service"
)

type Handler struct {
	service *service.Service
}

func New(service *service.Service) *Handler {
	return &Handler{service: service}
}

func (h *Handler) Ping(c *gin.Context) {
	if err := h.service.Ping(c.Request.Context()); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "template service unavailable"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"status": "ok"})
}

