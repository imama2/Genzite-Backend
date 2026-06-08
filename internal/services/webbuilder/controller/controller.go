package controller

import (
	"errors"
	"os"
	"path/filepath"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/imama2/Genzite-Backend/internal/core/middleware"
	"github.com/imama2/Genzite-Backend/internal/services/webbuilder/models/dto"
	errorUtils "github.com/imama2/Genzite-Backend/internal/utils/errors"
	"github.com/imama2/Genzite-Backend/internal/utils/responses"
)

func (h *SiteController) CreateSite(c *gin.Context) {
	userID, ok := getUserID(c)
	if !ok {
		responses.Unauthorized(c, "unauthorized")
		return
	}

	var req dto.CreateSiteRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		responses.BadRequest(c, "invalid request")
		return
	}

	site, err := h.service.CreateSite(c.Request.Context(), userID, dto.CreateSiteInput{
		Slug:   req.Slug,
		Config: req.Config,
	})
	if err != nil {
		switch {
		case errors.Is(err, errorUtils.ErrSlugTaken):
			responses.ConflictResponse(c, "slug already taken")
		case errors.Is(err, errorUtils.ErrInvalidInput):
			responses.BadRequest(c, "invalid input")
		default:
			responses.ServerError(c, "failed to create site")
		}
		return
	}

	responses.CreatedResponse(c, "site created", dto.SiteResponse{
		ID:     site.ID,
		Slug:   site.Slug,
		Status: site.Status,
	})
}

func (h *SiteController) PublishSite(c *gin.Context) {
	userID, ok := getUserID(c)
	if !ok {
		responses.Unauthorized(c, "unauthorized")
		return
	}

	siteID, err := parseUintParam(c, "id")
	if err != nil {
		responses.BadRequest(c, "invalid site id")
		return
	}

	site, err := h.service.PublishSite(c.Request.Context(), userID, siteID)
	if err != nil {
		switch {
		case errors.Is(err, errorUtils.ErrUnauthorized):
			responses.ForbiddenResponse(c, "forbidden")
		case errors.Is(err, errorUtils.ErrSiteNotFound):
			responses.NotFound(c, "site not found")
		default:
			responses.ServerError(c, "failed to publish site")
		}
		return
	}

	publishedAt := ""
	if site.PublishedAt != nil {
		publishedAt = site.PublishedAt.UTC().Format(time.RFC3339)
	}

	responses.Ok(c, "site published", dto.SiteResponse{
		ID:          site.ID,
		Slug:        site.Slug,
		Status:      site.Status,
		PublishedAt: publishedAt,
	})
}

func (h *SiteController) ServeSite(c *gin.Context) {
	slug := c.Param("slug")
	if slug == "" {
		responses.NotFound(c, "not found")
		return
	}

	site, err := h.service.GetPublishedSiteBySlug(c.Request.Context(), slug)
	if err != nil {
		responses.NotFound(c, "not found")
		return
	}

	path := filepath.Clean(site.OutputPath)
	if _, err := os.Stat(path); err != nil {
		responses.NotFound(c, "not found")
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
