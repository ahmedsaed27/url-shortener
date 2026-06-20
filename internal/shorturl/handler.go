package shorturl

import (
	"context"
	"errors"
	"net/http"
	"time"

	"github.com/as9840935/url-shortener/internal/analytics"
	"github.com/as9840935/url-shortener/internal/request"
	"github.com/as9840935/url-shortener/internal/response"
	"github.com/go-chi/chi/v5"
	"github.com/go-playground/validator/v10"
)

type Handler struct {
	service       *Service
	validate      *validator.Validate
	clickProducer ClickProducerContract
}

type ClickProducerContract interface {
	TrackClick(ctx context.Context, event analytics.ClickEvent) error
}

func NewHandler(service *Service, validate *validator.Validate, clickProducer ClickProducerContract) *Handler {
	return &Handler{
		service:       service,
		validate:      validate,
		clickProducer: clickProducer,
	}
}

func (h *Handler) Create(w http.ResponseWriter, r *http.Request) {
	var req CreateURLRequest

	if err := request.DecodeJSON(r, &req); err != nil {
		switch {
		case errors.Is(err, request.ErrEmptyBody):
			response.Error(w, http.StatusBadRequest, request.ErrEmptyBody.Error())
		case errors.Is(err, request.ErrInvalidJSON):
			response.Error(w, http.StatusBadRequest, request.ErrInvalidJSON.Error())
		case errors.Is(err, request.ErrMultipleJSON):
			response.Error(w, http.StatusBadRequest, request.ErrMultipleJSON.Error())
		default:
			response.Error(w, http.StatusBadRequest, "invalid request")
		}
		return
	}

	if err := request.Validate(h.validate, req); err != nil {
		if errors.Is(err, request.ErrInvalidFields) {
			response.Error(w, http.StatusBadRequest, request.ErrInvalidFields.Error())
			return
		}

		response.Error(w, http.StatusBadRequest, "invalid request")
		return
	}

	createdURL, err := h.service.Create(r.Context(), req.URL)
	if err != nil {
		switch {
		case errors.Is(err, ErrInvalidURL):
			response.Error(w, http.StatusBadRequest, ErrInvalidURL.Error())
		case errors.Is(err, ErrInvalidURLScheme):
			response.Error(w, http.StatusBadRequest, ErrInvalidURLScheme.Error())
		default:
			response.Error(w, http.StatusInternalServerError, "internal server error")
		}
		return
	}

	response.JSON(w, http.StatusCreated, createdURL)
}

func (h *Handler) Redirect(w http.ResponseWriter, r *http.Request) {
	code := chi.URLParam(r, "code")

	originalURL, err := h.service.Resolve(r.Context(), code)
	if err != nil {
		switch {
		case errors.Is(err, ErrShortURLNotFound):
			response.Error(w, http.StatusNotFound, ErrShortURLNotFound.Error())
		case errors.Is(err, ErrShortURLInactive), errors.Is(err, ErrShortURLExpired):
			response.Error(w, http.StatusGone, err.Error())
		default:
			response.Error(w, http.StatusInternalServerError, "internal server error")
		}
		return
	}

	if h.clickProducer != nil {
		_ = h.clickProducer.TrackClick(r.Context(), analytics.ClickEvent{
			Code:      code,
			IP:        r.RemoteAddr,
			UserAgent: r.UserAgent(),
			Referer:   r.Referer(),
			ClickedAt: time.Now().UTC(),
		})
	}

	http.Redirect(w, r, originalURL, http.StatusFound)
}
