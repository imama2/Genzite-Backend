package dto

// DraftResponse is the HTTP response body for the GenerateDraft endpoint.
type DraftResponse struct {
	SiteID uint       `json:"site_id"`
	Slug   string     `json:"slug"`
	Status string     `json:"status"`
	Config SiteConfig `json:"config"`
}
