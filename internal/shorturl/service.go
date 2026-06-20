package shorturl

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	neturl "net/url"
	"strings"
	"time"
)

type RepositoryContract interface {
	Create(ctx context.Context, code string, originalURL string, urlHash string) (ShortURL, error)
	FindByCode(ctx context.Context, code string) (ShortURL, error)
}

type Service struct {
	repo             RepositoryContract
	cache            CacheContract
	appBaseURL       string
	codeLength       int
	cacheTTL         time.Duration
	negativeCacheTTL time.Duration
}

func NewService(
	repo RepositoryContract,
	cache CacheContract,
	appBaseURL string,
	codeLength int,
	cacheTTL time.Duration,
	negativeCacheTTL time.Duration,
) *Service {
	return &Service{
		repo:             repo,
		cache:            cache,
		appBaseURL:       strings.TrimRight(appBaseURL, "/"),
		codeLength:       codeLength,
		cacheTTL:         cacheTTL,
		negativeCacheTTL: negativeCacheTTL,
	}
}

func (s *Service) Create(ctx context.Context, originalURL string) (CreateURLResponse, error) {
	normalizedURL, err := normalizeURL(originalURL)
	if err != nil {
		return CreateURLResponse{}, err
	}

	urlHash := hashURL(normalizedURL)

	for attempt := 0; attempt < 5; attempt++ {
		code, err := GenerateCode(s.codeLength)
		if err != nil {
			return CreateURLResponse{}, fmt.Errorf("generate short code: %w", err)
		}

		createdURL, err := s.repo.Create(ctx, code, normalizedURL, urlHash)
		if err != nil {
			if errors.Is(err, ErrDuplicateCode) {
				continue
			}

			return CreateURLResponse{}, err
		}

		_ = s.cache.SetOriginalURL(ctx, createdURL.Code, createdURL.OriginalURL, s.cacheTTL)

		return CreateURLResponse{
			Code:        createdURL.Code,
			ShortURL:    s.appBaseURL + "/" + createdURL.Code,
			OriginalURL: createdURL.OriginalURL,
		}, nil
	}

	return CreateURLResponse{}, fmt.Errorf("could not generate a unique short code after 5 attempts")
}

func (s *Service) Resolve(ctx context.Context, code string) (string, error) {
	code = strings.TrimSpace(code)
	if code == "" {
		return "", ErrShortURLNotFound
	}

	isNotFoundCached, err := s.cache.IsNotFoundCached(ctx, code)
	if err == nil && isNotFoundCached {
		return "", ErrShortURLNotFound
	}

	originalURL, err := s.cache.GetOriginalURL(ctx, code)
	if err == nil {
		return originalURL, nil
	}

	shortURL, err := s.repo.FindByCode(ctx, code)
	if err != nil {
		if errors.Is(err, ErrShortURLNotFound) {
			_ = s.cache.SetNotFound(ctx, code, s.negativeCacheTTL)
		}

		return "", err
	}

	if !shortURL.IsActive {
		return "", ErrShortURLInactive
	}

	if shortURL.ExpiresAt != nil && time.Now().After(*shortURL.ExpiresAt) {
		return "", ErrShortURLExpired
	}

	_ = s.cache.SetOriginalURL(ctx, code, shortURL.OriginalURL, s.cacheTTL)

	return shortURL.OriginalURL, nil
}

func normalizeURL(rawURL string) (string, error) {
	parsedURL, err := neturl.ParseRequestURI(strings.TrimSpace(rawURL))
	if err != nil {
		return "", ErrInvalidURL
	}

	if parsedURL.Scheme != "http" && parsedURL.Scheme != "https" {
		return "", ErrInvalidURLScheme
	}

	if parsedURL.Host == "" {
		return "", ErrInvalidURL
	}

	return parsedURL.String(), nil
}

func hashURL(originalURL string) string {
	sum := sha256.Sum256([]byte(originalURL))
	return hex.EncodeToString(sum[:])
}
