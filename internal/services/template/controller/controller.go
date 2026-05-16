package controller

import (
	"github.com/gin-gonic/gin"
	"github.com/imama2/Genzite-Backend/internal/services/template/service"
	"github.com/imama2/Genzite-Backend/internal/utils/responses"
)

type Controller struct {
	service *service.Service
}

func New(service *service.Service) *Controller {
	return &Controller{service: service}
}

func (h *Controller) Ping(c *gin.Context) {
	if err := h.service.Ping(c.Request.Context()); err != nil {
		responses.ServerError(c, "template service unavailable")
		return
	}
	responses.Ok(c, "ok")
}
