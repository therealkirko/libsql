package api

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/yourname/turbo-db/internal/manager"
	"github.com/yourname/turbo-db/internal/models"
)

// Handler handles HTTP requests
type Handler struct {
	manager *manager.Manager
}

// NewHandler creates a new API handler
func NewHandler(mgr *manager.Manager) *Handler {
	return &Handler{
		manager: mgr,
	}
}

// CreateDatabase handles database creation requests
// POST /api/v1/databases
func (h *Handler) CreateDatabase(c *gin.Context) {
	var req models.CreateDatabaseRequest

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{
			Error:   "invalid_request",
			Message: err.Error(),
		})
		return
	}

	db, err := h.manager.CreateDatabase(req.Name)
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.ErrorResponse{
			Error:   "creation_failed",
			Message: err.Error(),
		})
		return
	}

	c.JSON(http.StatusCreated, models.CreateDatabaseResponse{
		Database: *db,
		Message:  "Database created successfully",
	})
}

// ListDatabases handles listing all databases
// GET /api/v1/databases
func (h *Handler) ListDatabases(c *gin.Context) {
	databases, err := h.manager.ListDatabases()
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.ErrorResponse{
			Error:   "list_failed",
			Message: err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, models.ListDatabasesResponse{
		Databases: databases,
		Count:     len(databases),
	})
}

// GetDatabase handles retrieving a specific database
// GET /api/v1/databases/:id
func (h *Handler) GetDatabase(c *gin.Context) {
	id := c.Param("id")

	db, err := h.manager.GetDatabase(id)
	if err != nil {
		c.JSON(http.StatusNotFound, models.ErrorResponse{
			Error:   "not_found",
			Message: err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, db)
}

// DeleteDatabase handles database deletion
// DELETE /api/v1/databases/:id
func (h *Handler) DeleteDatabase(c *gin.Context) {
	id := c.Param("id")

	if err := h.manager.DeleteDatabase(id); err != nil {
		c.JSON(http.StatusInternalServerError, models.ErrorResponse{
			Error:   "deletion_failed",
			Message: err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Database deleted successfully",
	})
}

// StopDatabase handles stopping a database
// POST /api/v1/databases/:id/stop
func (h *Handler) StopDatabase(c *gin.Context) {
	id := c.Param("id")

	if err := h.manager.StopDatabase(id); err != nil {
		c.JSON(http.StatusInternalServerError, models.ErrorResponse{
			Error:   "stop_failed",
			Message: err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Database stopped successfully",
	})
}

// GetDatabaseConnection handles retrieving connection information
// GET /api/v1/databases/:id/connection
func (h *Handler) GetDatabaseConnection(c *gin.Context) {
	id := c.Param("id")

	db, err := h.manager.GetDatabase(id)
	if err != nil {
		c.JSON(http.StatusNotFound, models.ErrorResponse{
			Error:   "not_found",
			Message: err.Error(),
		})
		return
	}

	connInfo := models.DatabaseConnectionInfo{
		DatabaseID: db.ID,
		Name:       db.Name,
		HTTPUrl:    db.HTTPUrl,
		WSUrl:      "ws://localhost:" + string(rune(db.Port)),
		Status:     db.Status,
	}

	c.JSON(http.StatusOK, connInfo)
}

// HealthCheck handles health check requests
// GET /health
func (h *Handler) HealthCheck(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"status": "healthy",
		"service": "turbo-db",
	})
}
