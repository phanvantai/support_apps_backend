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
// @Summary Create support request
// @Description Create a new support request (public endpoint with rate limiting)
// @Tags Support Requests
// @Accept json
// @Produce json
// @Param request body models.CreateSupportRequestRequest true "Support request data"
// @Success 201 {object} map[string]interface{} "Support request created successfully"
// @Failure 400 {object} map[string]interface{} "Invalid request"
// @Failure 429 {object} map[string]interface{} "Rate limit exceeded"
// @Router /support-request [post]
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
// @Summary Get support request by ID (Admin only)
// @Description Get support request details by ID (requires admin authentication)
// @Tags Support Requests
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path int true "Support Request ID"
// @Success 200 {object} map[string]interface{} "Support request details"
// @Failure 400 {object} map[string]interface{} "Invalid ID format"
// @Failure 401 {object} map[string]interface{} "Unauthorized"
// @Failure 403 {object} map[string]interface{} "Forbidden - Admin access required"
// @Failure 404 {object} map[string]interface{} "Support request not found"
// @Router /support-requests/{id} [get]
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
// @Summary Get all support requests (Admin only)
// @Description Get paginated list of all support requests (requires admin authentication)
// @Tags Support Requests
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param page query int false "Page number" default(1)
// @Param page_size query int false "Page size" default(20)
// @Success 200 {object} map[string]interface{} "Support requests list"
// @Failure 401 {object} map[string]interface{} "Unauthorized"
// @Failure 403 {object} map[string]interface{} "Forbidden - Admin access required"
// @Router /support-requests [get]
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
// @Summary Update support request (Admin only)
// @Description Update support request details (requires admin authentication)
// @Tags Support Requests
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path int true "Support Request ID"
// @Param request body models.UpdateSupportRequestRequest true "Support request update data"
// @Success 200 {object} map[string]interface{} "Support request updated successfully"
// @Failure 400 {object} map[string]interface{} "Invalid request"
// @Failure 401 {object} map[string]interface{} "Unauthorized"
// @Failure 403 {object} map[string]interface{} "Forbidden - Admin access required"
// @Failure 404 {object} map[string]interface{} "Support request not found"
// @Router /support-requests/{id} [patch]
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
// @Summary Delete support request (Admin only)
// @Description Delete support request (requires admin authentication)
// @Tags Support Requests
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path int true "Support Request ID"
// @Success 204 "Support request deleted successfully"
// @Failure 400 {object} map[string]interface{} "Invalid ID format"
// @Failure 401 {object} map[string]interface{} "Unauthorized"
// @Failure 403 {object} map[string]interface{} "Forbidden - Admin access required"
// @Failure 404 {object} map[string]interface{} "Support request not found"
// @Router /support-requests/{id} [delete]
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
// @Summary Health check
// @Description Check if the service is running and healthy
// @Tags Health
// @Accept json
// @Produce json
// @Success 200 {object} map[string]interface{} "Service is healthy"
// @Router /health [get]
func (h *SupportRequestHandler) HealthCheck(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"status":  "healthy",
		"service": "support-app-backend",
	})
}
