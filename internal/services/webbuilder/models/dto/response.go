package dto

type SiteResponse struct {
	ID          uint   `json:"id"`
	Slug        string `json:"slug"`
	Status      string `json:"status"`
	PublishedAt string `json:"published_at,omitempty"`
}
