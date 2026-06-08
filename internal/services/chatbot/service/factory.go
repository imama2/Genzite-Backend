package service

import (
	webbuilder "github.com/imama2/Genzite-Backend/internal/services/webbuilder/service"
)

type ChatService struct {
	ai         AIClient
	webBuilder webbuilder.SiteManager
}

func New(aiClient AIClient, webBuilder webbuilder.SiteManager) *ChatService {
	return &ChatService{
		ai:         aiClient,
		webBuilder: webBuilder,
	}
}
