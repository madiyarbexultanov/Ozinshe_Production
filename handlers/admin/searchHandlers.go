package admin

import (
	"net/http"
	"ozinshe_production/logger"
	"ozinshe_production/models"
	"ozinshe_production/repositories"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

type SearchHandler struct {
	searchRepo *repositories.SearchRepository
}


func NewSearchHandler(searchRepo *repositories.SearchRepository) *SearchHandler {
	return &SearchHandler{searchRepo: searchRepo}
}


// SearchAll godoc
// @Summary Search for movies
// @Description Retrieve movies based on the search query
// @Tags search
// @Accept json
// @Produce json
// @Param q query string true "Search query"
// @Success 200 {array}  SearchResult
// @Failure 400 {object} models.ApiError "Invalid search query"
// @Failure 500 {object} models.ApiError "Error during search"
// @Router /search [get]
func (h *SearchHandler) SearchAll(c *gin.Context) {
    logger := logger.GetLogger()

    query := c.Query("q")
    if query == "" {
        logger.Warn("Empty search query received")
        c.JSON(http.StatusBadRequest, models.NewApiError("search request cannot be empty"))
        return
    }

    logger.Info("Search query received", zap.String("query", query))

    results, err := h.searchRepo.SearchAll(c, query)
    if err != nil {
        logger.Error("Error during search", zap.Error(err))
        c.JSON(http.StatusInternalServerError, models.NewApiError("error during search"))
        return
    }

    logger.Info("Search completed successfully", zap.Int("results_count", len(results)))
    c.JSON(http.StatusOK, results)
}