package handler

import (
	"errors"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/imama2/Genzite-Backend/internal/migrations/service"
)

type MigrationHandler struct {
	service *service.MigrationService
}

func New(service *service.MigrationService) *MigrationHandler {
	return &MigrationHandler{service: service}
}

func (h *MigrationHandler) Up(c *gin.Context) {
	steps, err := parseSteps(c, false)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	status, err := h.service.Up(c.Request.Context(), steps)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to run migrations"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"status": "ok", "version": status.Version, "dirty": status.Dirty})
}

func (h *MigrationHandler) Down(c *gin.Context) {
	steps, err := parseSteps(c, true)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	status, err := h.service.Down(c.Request.Context(), steps)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to run migrations"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"status": "ok", "version": status.Version, "dirty": status.Dirty})
}

func (h *MigrationHandler) Seed(c *gin.Context) {
	if err := h.service.Seed(c.Request.Context()); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"status": "ok"})
}

func (h *MigrationHandler) Version(c *gin.Context) {
	status, err := h.service.Status(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to read migration version"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"status": "ok", "version": status.Version, "dirty": status.Dirty})
}

func parseSteps(c *gin.Context, required bool) (int, error) {
	stepsRaw := c.Query("steps")
	if stepsRaw == "" {
		if required {
			return 0, errors.New("steps is required")
		}
		return 0, nil
	}

	steps, err := strconv.Atoi(stepsRaw)
	if err != nil || steps <= 0 {
		return 0, errors.New("steps must be a positive integer")
	}

	return steps, nil
}
