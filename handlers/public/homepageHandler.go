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

// GetMainScreen godoc
// @Summary      Get Main Screen Data
// @Description  Retrieves the main screen data, including recommended movies, categories, genres, and age ratings
// @Tags         homepage
// @Accept       json
// @Produce      json
// @Success      200 {object} gin.H "Main screen data including recommended movies, categories, genres, and ages"
// @Failure      500 {object} models.ApiError "Server error: failed to load data"
// @Router       /homepage [get]
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

// SearchMovies godoc
// @Summary      Search Movies
// @Description  Searches for movies based on a query string
// @Tags         homepage
// @Accept       json
// @Produce      json
// @Param        query query string true "Search query"
// @Success      200 {object} gin.H "Search results with matching movies"
// @Failure      400 {object} models.ApiError "Bad request: empty search query"
// @Failure      500 {object} models.ApiError "Server error: failed to search movies"
// @Router       /search [get]
func (h *HomepageHandler) SearchMovies(c *gin.Context) {
	logger := logger.GetLogger()
	query := c.Query("query")

	if query == "" {
		c.JSON(http.StatusBadRequest, models.NewApiError("Empty search query"))
		return
	}

	movies, err := h.moviesRepo.SearchMovies(c, query)
	if err != nil {
		logger.Error("Failed to search movies", zap.Error(err))
		c.Status(http.StatusInternalServerError)
		return
	}

	c.JSON(http.StatusOK, gin.H{"search_results": movies})
}

// GetMoviesByCategory godoc
// @Summary      Get Movies by Category
// @Description  Retrieves movies based on the provided category ID
// @Tags         homepage
// @Accept       json
// @Produce      json
// @Param        category_id path string true "Category ID"
// @Success      200 {object} gin.H "Movies belonging to the specified category"
// @Failure      400 {object} models.ApiError "Bad request: category ID is required"
// @Failure      500 {object} models.ApiError "Server error: failed to load movies for category"
// @Router       /search/{category_id} [get]
func (h *HomepageHandler) GetMoviesByCategory(c *gin.Context) {
	logger := logger.GetLogger()
	categoryID := c.Param("category_id")

	if categoryID == "" {
		c.JSON(http.StatusBadRequest, models.NewApiError("Category id required"))
		return
	}

	filters := models.Moviesfilters{
		CategoryIds: categoryID,
	}

	movies, err := h.moviesRepo.FindAll(c, filters)
	if err != nil {
		logger.Error("Failed to load movies by category", zap.Error(err))
		c.Status(http.StatusInternalServerError)
		return
	}

	c.JSON(http.StatusOK, gin.H{"movies": movies})
}
