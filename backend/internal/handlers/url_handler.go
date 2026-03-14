
package handlers

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/yourusername/url-shortener/internal/models"
	"github.com/yourusername/url-shortener/internal/service"
)

type URLHandler struct {
	urlService service.URLService
}

func NewURLHandler(urlService service.URLService) *URLHandler {
	return &URLHandler{
		urlService: urlService,
	}
}

// CreateShortURL handles POST /api/v1/shorten
func (h *URLHandler) CreateShortURL(c *gin.Context) {
	var req models.CreateURLRequest

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, models.APIResponse{
			Success: false,
			Error: &models.APIError{
				Code:    "INVALID_REQUEST",
				Message: "Invalid request body: " + err.Error(),
			},
		})
		return
	}

	response, err := h.urlService.CreateShortURL(c.Request.Context(), &req)
	if err != nil {
		h.handleServiceError(c, err)
		return
	}

	c.JSON(http.StatusCreated, models.APIResponse{
		Success: true,
		Data:    response,
	})
}

// RedirectToOriginal handles GET /:shortCode
func (h *URLHandler) RedirectToOriginal(c *gin.Context) {
	shortCode := c.Param("shortCode")

	if shortCode == "" {
		c.JSON(http.StatusBadRequest, models.APIResponse{
			Success: false,
			Error: &models.APIError{
				Code:    "INVALID_REQUEST",
				Message: "Short code is required",
			},
		})
		return
	}

	originalURL, err := h.urlService.GetOriginalURL(c.Request.Context(), shortCode)
	if err != nil {
		h.handleServiceError(c, err)
		return
	}

	c.Redirect(http.StatusFound, originalURL)  // 302 = Temporary (browser won't cache)
}

// GetAnalytics handles GET /api/v1/analytics/:shortCode
func (h *URLHandler) GetAnalytics(c *gin.Context) {
	shortCode := c.Param("shortCode")

	if shortCode == "" {
		c.JSON(http.StatusBadRequest, models.APIResponse{
			Success: false,
			Error: &models.APIError{
				Code:    "INVALID_REQUEST",
				Message: "Short code is required",
			},
		})
		return
	}

	analytics, err := h.urlService.GetAnalytics(c.Request.Context(), shortCode)
	if err != nil {
		h.handleServiceError(c, err)
		return
	}

	c.JSON(http.StatusOK, models.APIResponse{
		Success: true,
		Data:    analytics,
	})
}

// GetURLDetails handles GET /api/v1/url/:shortCode
func (h *URLHandler) GetURLDetails(c *gin.Context) {
	shortCode := c.Param("shortCode")

	if shortCode == "" {
		c.JSON(http.StatusBadRequest, models.APIResponse{
			Success: false,
			Error: &models.APIError{
				Code:    "INVALID_REQUEST",
				Message: "Short code is required",
			},
		})
		return
	}

	details, err := h.urlService.GetURLDetails(c.Request.Context(), shortCode)
	if err != nil {
		h.handleServiceError(c, err)
		return
	}

	c.JSON(http.StatusOK, models.APIResponse{
		Success: true,
		Data:    details,
	})
}

// DeleteURL handles DELETE /api/v1/url/:shortCode
func (h *URLHandler) DeleteURL(c *gin.Context) {
	shortCode := c.Param("shortCode")

	if shortCode == "" {
		c.JSON(http.StatusBadRequest, models.APIResponse{
			Success: false,
			Error: &models.APIError{
				Code:    "INVALID_REQUEST",
				Message: "Short code is required",
			},
		})
		return
	}

	err := h.urlService.DeleteURL(c.Request.Context(), shortCode)
	if err != nil {
		h.handleServiceError(c, err)
		return
	}

	c.JSON(http.StatusOK, models.APIResponse{
		Success: true,
		Data: map[string]string{
			"message": "Short URL deleted successfully",
		},
	})
}

// handleServiceError converts service errors to HTTP responses
func (h *URLHandler) handleServiceError(c *gin.Context, err error) {
	errMsg := err.Error()

	// Parse error to determine HTTP status and error code
	var status int
	var code string
	var message string

	if strings.HasPrefix(errMsg, "INVALID_URL:") {
		status = http.StatusBadRequest
		code = "INVALID_URL"
		message = strings.TrimPrefix(errMsg, "INVALID_URL: ")
	} else if strings.HasPrefix(errMsg, "INVALID_ALIAS:") {
		status = http.StatusBadRequest
		code = "INVALID_ALIAS"
		message = strings.TrimPrefix(errMsg, "INVALID_ALIAS: ")
	} else if strings.HasPrefix(errMsg, "ALIAS_RESERVED:") {
		status = http.StatusConflict
		code = "ALIAS_RESERVED"
		message = strings.TrimPrefix(errMsg, "ALIAS_RESERVED: ")
	} else if strings.HasPrefix(errMsg, "ALIAS_EXISTS:") {
		status = http.StatusConflict
		code = "ALIAS_EXISTS"
		message = strings.TrimPrefix(errMsg, "ALIAS_EXISTS: ")
	} else if strings.HasPrefix(errMsg, "URL_NOT_FOUND:") {
		status = http.StatusNotFound
		code = "URL_NOT_FOUND"
		message = strings.TrimPrefix(errMsg, "URL_NOT_FOUND: ")
	} else if strings.HasPrefix(errMsg, "URL_EXPIRED:") {
		status = http.StatusGone
		code = "URL_EXPIRED"
		message = strings.TrimPrefix(errMsg, "URL_EXPIRED: ")
	} else if strings.HasPrefix(errMsg, "URL_TOO_LONG:") {
		status = http.StatusBadRequest
		code = "URL_TOO_LONG"
		message = strings.TrimPrefix(errMsg, "URL_TOO_LONG: ")
	} else {
		status = http.StatusInternalServerError
		code = "INTERNAL_ERROR"
		message = "An internal error occurred"
	}

	c.JSON(status, models.APIResponse{
		Success: false,
		Error: &models.APIError{
			Code:    code,
			Message: message,
		},
	})
}
