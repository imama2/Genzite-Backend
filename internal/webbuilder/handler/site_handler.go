package handler

import (
	"errors"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/imama2/Genzite-Backend/internal/core/middleware"
	"github.com/imama2/Genzite-Backend/internal/webbuilder/service"
)

type SiteHandler struct {
	service *service.SiteService
}

func New(service *service.SiteService) *SiteHandler {
	return &SiteHandler{service: service}
}

type createSiteRequest struct {
	Slug   string             `json:"slug" binding:"required"`
	Config service.SiteConfig `json:"config" binding:"required"`
}

type siteResponse struct {
	ID          uint   `json:"id"`
	Slug        string `json:"slug"`
	Status      string `json:"status"`
	PublishedAt string `json:"published_at,omitempty"`
}

func (h *SiteHandler) CreateSite(c *gin.Context) {
	userID, ok := getUserID(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

	var req createSiteRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request"})
		return
	}

	site, err := h.service.CreateSite(c.Request.Context(), userID, service.CreateSiteInput{
		Slug:   req.Slug,
		Config: req.Config,
	})
	if err != nil {
		switch {
		case errors.Is(err, service.ErrSlugTaken):
			c.JSON(http.StatusConflict, gin.H{"error": "slug already taken"})
		case errors.Is(err, service.ErrInvalidInput):
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid input"})
		default:
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to create site"})
		}
		return
	}

	c.JSON(http.StatusCreated, siteResponse{
		ID:     site.ID,
		Slug:   site.Slug,
		Status: site.Status,
	})
}

func (h *SiteHandler) PublishSite(c *gin.Context) {
	userID, ok := getUserID(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

	siteID, err := parseUintParam(c, "id")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid site id"})
		return
	}

	site, err := h.service.PublishSite(c.Request.Context(), userID, siteID)
	if err != nil {
		switch {
		case errors.Is(err, service.ErrUnauthorized):
			c.JSON(http.StatusForbidden, gin.H{"error": "forbidden"})
		case errors.Is(err, service.ErrSiteNotFound):
			c.JSON(http.StatusNotFound, gin.H{"error": "site not found"})
		default:
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to publish site"})
		}
		return
	}

	publishedAt := ""
	if site.PublishedAt != nil {
		publishedAt = site.PublishedAt.UTC().Format(time.RFC3339)
	}

	c.JSON(http.StatusOK, siteResponse{
		ID:          site.ID,
		Slug:        site.Slug,
		Status:      site.Status,
		PublishedAt: publishedAt,
	})
}

func (h *SiteHandler) ServeSite(c *gin.Context) {
	slug := c.Param("slug")
	if slug == "" {
		c.JSON(http.StatusNotFound, gin.H{"error": "not found"})
		return
	}

	site, err := h.service.GetPublishedSiteBySlug(c.Request.Context(), slug)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "not found"})
		return
	}

	path := filepath.Clean(site.OutputPath)
	if _, err := os.Stat(path); err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "not found"})
		return
	}

	c.File(path)
}

func getUserID(c *gin.Context) (uint, bool) {
	value, ok := c.Get(middleware.ContextUserIDKey)
	if !ok {
		return 0, false
	}
	userID, ok := value.(uint)
	return userID, ok
}

func parseUintParam(c *gin.Context, name string) (uint, error) {
	raw := c.Param(name)
	if raw == "" {
		return 0, errors.New("missing param")
	}
	parsed, err := strconv.ParseUint(raw, 10, 64)
	if err != nil {
		return 0, err
	}
	return uint(parsed), nil
}
