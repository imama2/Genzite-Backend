package iam

import "github.com/imama2/Genzite-Backend/internal/core/middleware"

func (m *Module) AuthProvider() middleware.AuthProvider {
	return m.service
}
