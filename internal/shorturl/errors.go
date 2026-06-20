package shorturl

import "errors"

var (
	ErrDuplicateCode    = errors.New("duplicate short code")
	ErrInvalidURL       = errors.New("invalid url")
	ErrInvalidURLScheme = errors.New("url must use http or https")
	ErrShortURLNotFound = errors.New("short url not found")
	ErrShortURLInactive = errors.New("short url is inactive")
	ErrShortURLExpired  = errors.New("short url is expired")
	ErrCacheMiss        = errors.New("cache miss")
)
