package public

import (
	"net/http"
	"ozinshe_production/logger"
	"ozinshe_production/models"
	"ozinshe_production/repositories"
	"strconv"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

type WatchlistHandler struct {
	watchlistRepo *repositories.WatchlistRepository
}

func NewWatchlistHandler(
	watchlistRepo *repositories.WatchlistRepository) *WatchlistHandler {
	return &WatchlistHandler{
		watchlistRepo: watchlistRepo,
	}
}

type watchlistResponse struct {
	Message string `json:"message,omitempty"`
	Error   string `json:"error,omitempty"`
	Success bool   `json:"success"`
}

// AddToWatchlist godoc
// @Summary      Add movie to watchlist
// @Description  Adds a movie to the user's watchlist
// @Tags         watchlist
// @Accept       json
// @Produce      json
// @Param        movie_id path int true "Movie ID to add"
// @Success      200 {object} watchlistResponse "Movie successfully added to the watchlist"
// @Failure      400 {object} models.ApiError "Invalid movie ID"
// @Failure      500 {object} models.ApiError "Failed to add to watchlist"
// @Router       /watchlist/{movie_id} [post]
func (h *WatchlistHandler) AddToWatchlist(c *gin.Context) {
	logger := logger.GetLogger()

	userID := c.GetInt("userId")
	movieID, err := strconv.Atoi(c.Param("movie_id"))
	if err != nil {
		logger.Error("Invalid movie ID", zap.String("movie_id", c.Param("movie_id")))
		c.JSON(http.StatusBadRequest, models.NewApiError("Invalid movie ID"))
		return
	}

	err = h.watchlistRepo.AddToWatchlist(c, userID, movieID)
	if err != nil {
		logger.Error("Failed to add movie to watchlist", zap.Int("user_id", userID), zap.Int("movie_id", movieID), zap.Error(err))
		c.JSON(http.StatusInternalServerError, models.NewApiError("Failed to add to watchlist"))
		return
	}

	logger.Info("Movie added to watchlist", zap.Int("user_id", userID), zap.Int("movie_id", movieID))
	c.JSON(http.StatusOK, watchlistResponse{Message: "Movie added to watchlist", Success: true})
}

// GetWatchlist godoc
// @Summary      Get user's watchlist
// @Description  Retrieves the list of movies in the user's watchlist
// @Tags         watchlist
// @Accept       json
// @Produce      json
// @Success      200 {array} models.Movie "List of movies in the watchlist"
// @Failure      500 {object} models.ApiError "Failed to get watchlist"
// @Router       /watchlist [get]
func (h *WatchlistHandler) GetWatchlist(c *gin.Context) {
	logger := logger.GetLogger()

	userID := c.GetInt("userId")
	movies, err := h.watchlistRepo.GetWatchlist(c, userID)
	if err != nil {
		logger.Error("Failed to get watchlist", zap.Int("user_id", userID), zap.Error(err))
		c.JSON(http.StatusInternalServerError, models.NewApiError("Failed to get watchlist"))
		return
	}

	logger.Info("Watchlist retrieved", zap.Int("user_id", userID), zap.Int("movies_count", len(movies)))
	c.JSON(http.StatusOK, movies)
}

// RemoveFromWatchlist godoc
// @Summary      Remove movie from watchlist
// @Description  Removes a movie from the user's watchlist
// @Tags         watchlist
// @Accept       json
// @Produce      json
// @Param        movie_id path int true "Movie ID to remove"
// @Success      200 {object} watchlistResponse "Movie successfully removed from the watchlist"
// @Failure      400 {object} models.ApiError "Invalid movie ID"
// @Failure      500 {object} models.ApiError "Failed to remove from watchlist"
// @Router       /watchlist/{movie_id} [delete]
func (h *WatchlistHandler) RemoveFromWatchlist(c *gin.Context) {
	logger := logger.GetLogger()

	userID := c.GetInt("userId")
	movieID, err := strconv.Atoi(c.Param("movie_id"))
	if err != nil {
		logger.Error("Invalid movie ID", zap.String("movie_id", c.Param("movie_id")))
		c.JSON(http.StatusBadRequest, models.NewApiError("Invalid movie ID"))
		return
	}

	err = h.watchlistRepo.RemoveFromWatchlist(c, userID, movieID)
	if err != nil {
		logger.Error("Failed to remove movie from watchlist", zap.Int("user_id", userID), zap.Int("movie_id", movieID), zap.Error(err))
		c.JSON(http.StatusInternalServerError, models.NewApiError("Failed to remove from watchlist"))
		return
	}

	logger.Info("Movie removed from watchlist", zap.Int("user_id", userID), zap.Int("movie_id", movieID))
	c.JSON(http.StatusOK, watchlistResponse{Message: "Movie removed from watchlist", Success: true})
}

// IsInWatchlist godoc
// @Summary      Check if movie is in watchlist
// @Description  Checks if a movie is in the user's watchlist
// @Tags         watchlist
// @Accept       json
// @Produce      json
// @Param        movie_id path int true "Movie ID to check"
// @Success      200 {object} watchlistResponse "Success status (true if movie is in watchlist)"
// @Failure      400 {object} models.ApiError "Invalid movie ID"
// @Failure      500 {object} models.ApiError "Failed to check watchlist"
// @Router       /watchlist/{movie_id} [get]
func (h *WatchlistHandler) IsInWatchlist(c *gin.Context) {
	logger := logger.GetLogger()

	userID := c.GetInt("userId")
	movieID, err := strconv.Atoi(c.Param("movie_id"))
	if err != nil {
		logger.Error("Invalid movie ID", zap.String("movie_id", c.Param("movie_id")))
		c.JSON(http.StatusBadRequest, models.NewApiError("Invalid movie ID"))
		return
	}

	exists, err := h.watchlistRepo.IsInWatchlist(c, userID, movieID)
	if err != nil {
		logger.Error("Failed to check watchlist", zap.Int("user_id", userID), zap.Int("movie_id", movieID), zap.Error(err))
		c.JSON(http.StatusInternalServerError, models.NewApiError("Failed to check watchlist"))
		return
	}

	logger.Info("Checked if movie is in watchlist", zap.Int("user_id", userID), zap.Int("movie_id", movieID), zap.Bool("exists", exists))
	c.JSON(http.StatusOK, watchlistResponse{Success: exists})
}
