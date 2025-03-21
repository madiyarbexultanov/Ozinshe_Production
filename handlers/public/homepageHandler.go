package public

import (
	"net/http"
	"ozinshe_production/logger"
	"ozinshe_production/repositories"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

type HomepageHandler struct {
	hompageRepo *repositories.HomepageRepository
}

func NewHomepageHandler(repo *repositories.HomepageRepository) *HomepageHandler {
	return &HomepageHandler{hompageRepo: repo}
}

// GetMainScreen godoc
// @Summary      Get main screen data
// @Description  Retrieves recommended movies and all movies
// @Tags         main-screen
// @Produce      json
// @Success      200 {object} models.MainScreenResponse "Main screen movies"
// @Failure      500 {object} models.ApiError "Internal server error"
// @Router       /public/main-screen [get]
func (h *HomepageHandler) GetMainScreen(c *gin.Context) {
	logger := logger.GetLogger()

	// Получаем рекомендованные фильмы
	recommended, err := h.hompageRepo.GetRecommendedMovies(c)
	if err != nil {
		logger.Error("Failed to load recommended movies", zap.Error(err))
		c.Status(http.StatusInternalServerError)
		return
	}
	// Формируем JSON-ответ
	response := recommended

	logger.Info("Main screen data loaded successfully", zap.Int("recommended_count", len(recommended)))
	c.JSON(http.StatusOK, response)
}
