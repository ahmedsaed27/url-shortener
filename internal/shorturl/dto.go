package shorturl

type CreateURLRequest struct {
	URL string `json:"url" validate:"required,url"`
}

type CreateURLResponse struct {
	Code        string `json:"code"`
	ShortURL    string `json:"short_url"`
	OriginalURL string `json:"original_url"`
}
