package handlers

import (
	"net/http"
	"strconv"
	"support-app-backend/internal/models"
	"support-app-backend/internal/services"

	"github.com/gin-gonic/gin"
)

// SupportRequestHandler handles HTTP requests for support requests
type SupportRequestHandler struct {
	service services.SupportRequestService
}

// NewSupportRequestHandler creates a new support request handler
func NewSupportRequestHandler(service services.SupportRequestService) *SupportRequestHandler {
	return &SupportRequestHandler{
		service: service,
	}
}

// CreateSupportRequest handles POST /api/v1/support-request
func (h *SupportRequestHandler) CreateSupportRequest(c *gin.Context) {
	var req models.CreateSupportRequestRequest

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	response, err := h.service.CreateSupportRequest(&req)
	if err != nil {
		if err == services.ErrInvalidRequest {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create support request"})
		return
	}

	c.JSON(http.StatusCreated, gin.H{"data": response})
}

// GetSupportRequest handles GET /api/v1/support-requests/:id
func (h *SupportRequestHandler) GetSupportRequest(c *gin.Context) {
	idParam := c.Param("id")
	id, err := strconv.ParseUint(idParam, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid ID format"})
		return
	}

	response, err := h.service.GetSupportRequest(uint(id))
	if err != nil {
		if err == services.ErrSupportRequestNotFound {
			c.JSON(http.StatusNotFound, gin.H{"error": "Support request not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get support request"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"data": response})
}

// GetAllSupportRequests handles GET /api/v1/support-requests
func (h *SupportRequestHandler) GetAllSupportRequests(c *gin.Context) {
	// Parse pagination parameters
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "20"))

	responses, total, err := h.service.GetAllSupportRequests(page, pageSize)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get support requests"})
		return
	}

	// Calculate pagination metadata
	totalPages := (int(total) + pageSize - 1) / pageSize

	c.JSON(http.StatusOK, gin.H{
		"data": responses,
		"pagination": gin.H{
			"page":        page,
			"page_size":   pageSize,
			"total":       total,
			"total_pages": totalPages,
		},
	})
}

// UpdateSupportRequest handles PATCH /api/v1/support-requests/:id
func (h *SupportRequestHandler) UpdateSupportRequest(c *gin.Context) {
	idParam := c.Param("id")
	id, err := strconv.ParseUint(idParam, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid ID format"})
		return
	}

	var req models.UpdateSupportRequestRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	response, err := h.service.UpdateSupportRequest(uint(id), &req)
	if err != nil {
		if err == services.ErrSupportRequestNotFound {
			c.JSON(http.StatusNotFound, gin.H{"error": "Support request not found"})
			return
		}
		if err == services.ErrInvalidRequest {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update support request"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"data": response})
}

// DeleteSupportRequest handles DELETE /api/v1/support-requests/:id
func (h *SupportRequestHandler) DeleteSupportRequest(c *gin.Context) {
	idParam := c.Param("id")
	id, err := strconv.ParseUint(idParam, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid ID format"})
		return
	}

	err = h.service.DeleteSupportRequest(uint(id))
	if err != nil {
		if err == services.ErrSupportRequestNotFound {
			c.JSON(http.StatusNotFound, gin.H{"error": "Support request not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete support request"})
		return
	}

	c.JSON(http.StatusNoContent, nil)
}

// HealthCheck handles GET /health
func (h *SupportRequestHandler) HealthCheck(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"status":  "healthy",
		"service": "support-app-backend",
	})
}
