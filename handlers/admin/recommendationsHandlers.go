package admin

import (
	"net/http"
	"strconv"
	"ozinshe_production/logger"
	"ozinshe_production/models"
	"ozinshe_production/repositories"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

type RecommendationsHandler struct {
	recommendationsRepo *repositories.RecommendationsRepository
}

func NewRecommendationsHandler(repo *repositories.RecommendationsRepository) *RecommendationsHandler {
	return &RecommendationsHandler{recommendationsRepo: repo}
}

type createRecommendationRequest struct {
	MovieId  int `json:"movie_id"`
	Position int `json:"position"`
}

// FindAll godoc
// @Summary Get all recommendations
// @Description Retrieve all recommended movies
// @Tags recommendations
// @Accept json
// @Produce json
// @Success 200 {array} models.Recommendation
// @Failure 500 {object} models.ApiError
// @Router /admin/recommendations [get]
func (h *RecommendationsHandler) FindAll(c *gin.Context) {
	logger := logger.GetLogger()

	recommendations, err := h.recommendationsRepo.FindAll(c)
	if err != nil {
		logger.Error("Failed to load recommendations", zap.Error(err))
		c.JSON(http.StatusInternalServerError, models.NewApiError("Failed to load recommendations"))
		return
	}

	logger.Info("Recommendations loaded successfully", zap.Int("count", len(recommendations)))
	c.JSON(http.StatusOK, recommendations)
}

// FindById godoc
// @Summary Get recommendation by Id
// @Description Retrieve a recommendation by its Id
// @Tags recommendations
// @Accept json
// @Produce json
// @Param id path int true "Recommendation Id"
// @Success 200 {object} models.Recommendation
// @Failure 400 {object} models.ApiError
// @Failure 404 {object} models.ApiError
// @Router /admin/recommendations/{id} [get]
func (h *RecommendationsHandler) FindById(c *gin.Context) {
	logger := logger.GetLogger()

	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		logger.Error("Invalid recommendation Id", zap.String("id", c.Param("id")))
		c.JSON(http.StatusBadRequest, models.NewApiError("Invalid recommendation Id"))
		return
	}

	recommendation, err := h.recommendationsRepo.FindById(c, id)
	if err != nil {
		logger.Error("Failed to find recommendation", zap.Int("id", id), zap.Error(err))
		c.JSON(http.StatusNotFound, models.NewApiError("Failed to find recommendation"))
		return
	}

	logger.Info("Recommendation found", zap.Int("id", id))
	c.JSON(http.StatusOK, recommendation)
}

// Create godoc
// @Summary Create a new recommendation
// @Description Add a new movie to recommendations
// @Tags recommendations
// @Accept json
// @Produce json
// @Param recommendation body createRecommendationRequest true "Recommendation Information"
// @Success 201 {object} map[string]int "Id of the created recommendation"
// @Failure 400 {object} models.ApiError
// @Failure 500 {object} models.ApiError
// @Router /admin/recommendations [post]
func (h *RecommendationsHandler) Create(c *gin.Context) {
	logger := logger.GetLogger()

	var req createRecommendationRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		logger.Error("Invalid request format", zap.Error(err))
		c.JSON(http.StatusBadRequest, models.NewApiError("Invalid request format"))
		return
	}

	id, err := h.recommendationsRepo.Create(c, req.MovieId, req.Position)
	if err != nil {
		logger.Error("Failed to create recommendation", zap.Error(err))
		c.JSON(http.StatusInternalServerError, models.NewApiError("Failed to create recommendation"))
		return
	}

	logger.Info("Recommendation created successfully", zap.Int("id", id))
	c.JSON(http.StatusCreated, gin.H{"id": id})
}

// Delete godoc
// @Summary Delete a recommendation
// @Description Remove a recommendation by Id
// @Tags recommendations
// @Accept json
// @Produce json
// @Param id path int true "Recommendation Id"
// @Success 200 {object} map[string]string "Recommendation deleted successfully"
// @Failure 400 {object} models.ApiError
// @Failure 500 {object} models.ApiError
// @Router /admin/recommendations/{id} [delete]
func (h *RecommendationsHandler) Delete(c *gin.Context) {
	logger := logger.GetLogger()

	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		logger.Error("Invalid recommendation Id", zap.String("id", c.Param("id")))
		c.JSON(http.StatusBadRequest, models.NewApiError("Invalid recommendation Id"))
		return
	}

	if err := h.recommendationsRepo.Delete(c, id); err != nil {
		logger.Error("Failed to delete recommendation", zap.Int("id", id), zap.Error(err))
		c.JSON(http.StatusInternalServerError, models.NewApiError("Failed to delete recommendation"))
		return
	}

	logger.Info("Recommendation deleted successfully", zap.Int("id", id))
	c.JSON(http.StatusOK, gin.H{"message": "Recommendation deleted successfully"})
}
