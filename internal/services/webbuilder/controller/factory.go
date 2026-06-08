package controller

import "github.com/imama2/Genzite-Backend/internal/services/webbuilder/service"

type SiteController struct {
	service service.SiteManager
}

func New(service service.SiteManager) *SiteController {
	return &SiteController{service: service}
}
