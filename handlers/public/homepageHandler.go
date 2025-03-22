package public

import (
	"fmt"
	"net/http"
	"ozinshe_production/logger"
	"ozinshe_production/models"
	"ozinshe_production/repositories"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

type HomepageHandler struct {
	hompageRepo 		*repositories.HomepageRepository
	moviesRepo 			*repositories.MoviesRepository
	genresRepo 			*repositories.GenresRepository
	categoriesRepo  	*repositories.CategoriesRepository
	agesRepo 			*repositories.AgesRepository
}

func NewHomepageHandler(hompageRepo 	*repositories.HomepageRepository,
						moviesRepo 		*repositories.MoviesRepository, 
					  	genresRepo 		*repositories.GenresRepository,
						categoriesRepo 	*repositories.CategoriesRepository,
					  	agesRepo 		*repositories.AgesRepository,
					  	) *HomepageHandler {
	return &HomepageHandler{hompageRepo: 	hompageRepo,
							moviesRepo:		moviesRepo,
							genresRepo: 	genresRepo,
							categoriesRepo: categoriesRepo,
							agesRepo: 		agesRepo,
						}
}

func (h *HomepageHandler) GetMainScreen(c *gin.Context) {
	logger := logger.GetLogger()

	// Получаем рекомендованные фильмы
	recommended, err := h.hompageRepo.GetRecommendedMovies(c)
	if err != nil {
		logger.Error("Failed to load recommended movies", zap.Error(err))
		c.Status(http.StatusInternalServerError)
		return
	}

	// Получаем категории
	categories, err := h.categoriesRepo.FindAll(c)
	if err != nil {
		logger.Error("Failed to load categories", zap.Error(err))
		c.Status(http.StatusInternalServerError)
		return
	}

	// Получаем все доступные жанры
	genres, err := h.genresRepo.FindAll(c)
	if err != nil {
		logger.Error("Failed to load genres", zap.Error(err))
		c.Status(http.StatusInternalServerError)
		return
	}

	// Получаем возрастные рейтинги
	ages, err := h.agesRepo.FindAll(c)
	if err != nil {
		logger.Error("Failed to load ages", zap.Error(err))
		c.Status(http.StatusInternalServerError)
		return
	}

	// Получаем фильмы по категориям, жанрам и возрастам
	moviesByCategory := make(map[string][]models.Movie)
	for _, category := range categories {
		filters := models.Moviesfilters{
			CategoryIds: fmt.Sprint(category.Id),
		}

		movies, err := h.moviesRepo.FindAll(c, filters)
		if err != nil {
			logger.Error("Failed to load movies for category", zap.String("category", category.Title), zap.Error(err))
			return
		}
		moviesByCategory[category.Title] = movies
	}

	// Формируем JSON-ответ
	response := gin.H{
		"recommended":       	recommended,
		"movies_by_category":	moviesByCategory,
		"genres":            	genres,
		"ages":              	ages,
	}

	logger.Info("Main screen data loaded successfully", zap.Int("recommended_count", len(recommended)))
	c.JSON(http.StatusOK, response)
}
