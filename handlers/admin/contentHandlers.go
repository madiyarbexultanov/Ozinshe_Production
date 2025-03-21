package admin

import (
	"net/http"
	"ozinshe_production/logger"
	"ozinshe_production/models"
	"ozinshe_production/repositories"
	"strconv"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

type ContentHandler struct {
	seasonsRepo     *repositories.SeasonsRepository
	episodesRepo    *repositories.EpisodesRepository
}

type CreateSeasonRequest struct {
	MovieID 		int 					`json:"id"`
	Number          int         			`json:"number"`
	Episodes        []CreateEpisodeRequest  `json:"episodes"`
}

type UpdateSeasonRequest struct {
	Number   int                    `json:"number"`
	Episodes []UpdateEpisodeRequest `json:"episodes"`
}

type CreateEpisodeRequest struct {
	SeasonID int    `json:"seasonId"`
	Number   int    `json:"number"`
	VideoURL string `json:"videoURL"`
}

type UpdateEpisodeRequest struct {
	Id       int    `json:"id"`
	Number   int    `json:"number"`
	VideoURL string `json:"videoURL"`
	SeasonID int    `json:"seasonId"`
}

func NewContentsHandler(seasonsRepo *repositories.SeasonsRepository,
						episodesRepo *repositories.EpisodesRepository) *ContentHandler {
	return &ContentHandler{	seasonsRepo:    seasonsRepo,
							episodesRepo:   episodesRepo,}
}

// AddSeasonsAndEpisodes godoc
// @Summary      Add seasons and episodes to a movie
// @Description  Adds a new season with episodes to a movie
// @Tags         content
// @Accept       json
// @Produce      json
// @Param        id  path      int                              true  "Movie ID"
// @Param        request  body      admin.CreateSeasonRequest     true  "Season and episodes data"
// @Success      200      {object}  object{message=string}           "Seasons and episodes added successfully"
// @Failure      400      {object}  models.ApiError                 "Invalid movie id or payload"
// @Failure      500      {object}  models.ApiError                 "Internal server error"
// @Router       /admin/movies/{Id}/seasons [post]
func (h *ContentHandler) AddSeasonsAndEpisodes(c *gin.Context) {
	logger := logger.GetLogger()

	// Получаем movie Id
	movieID, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		logger.Error("Invalid movie id", zap.Error(err))
		c.JSON(http.StatusBadRequest, models.NewApiError("Invalid movie id"))
		return
	}

	var request CreateSeasonRequest
	request.MovieID = movieID

	// Привязываем данные
	if err := c.Bind(&request); err != nil {
		logger.Error("Failed to bind request payload", zap.Error(err))
		c.JSON(http.StatusBadRequest, models.NewApiError("Couldn't bind payload"))
		return
	}

	// Создаем сезон
	seasonID, err := h.seasonsRepo.Create(c, models.Season{MovieID: request.MovieID, Number: request.Number})
	if err != nil {
		logger.Error("Failed to create season", zap.Error(err))
		c.JSON(http.StatusBadRequest, models.NewApiError(err.Error()))
		return
	}

	// Создаем эпизоды
	for _, episodeReq := range request.Episodes {
		_, err := h.episodesRepo.Create(c, models.Episode{SeasonID: seasonID, Number: episodeReq.Number, VideoURL: episodeReq.VideoURL})
		if err != nil {
			logger.Warn("Skipping duplicate episode", zap.Error(err))
		}
	}

	logger.Info("Seasons and episodes added successfully")
	c.JSON(http.StatusOK, gin.H{"message": "Seasons and episodes added successfully"})
}

// UpdateSeason godoc
// @Summary      Update season details
// @Description  Updates an existing season's details and episodes
// @Tags         content
// @Accept       json
// @Produce      json
// @Param        Id   path      int                          true  "Movie ID"
// @Param        seasonId  path      int                          true  "Season ID"
// @Param        request   body      admin.UpdateSeasonRequest true  "Season and episode update data"
// @Success      200       {object}  object{message=string}       "Season and episodes updated successfully"
// @Failure      400       {object}  models.ApiError             "Invalid movie Id or seasonId or payload"
// @Failure      404       {object}  models.ApiError             "Season not found"
// @Failure      500       {object}  models.ApiError             "Internal server error"
// @Router       /admin/movies/{Id}/seasons/{seasonId} [put]
func (h *ContentHandler) UpdateSeason(c *gin.Context) {
	logger := logger.GetLogger()

	movieID, err := strconv.Atoi(c.Param("Id"))
	if err != nil {
		logger.Error("Invalid movie id", zap.Error(err))
		c.JSON(http.StatusBadRequest, models.NewApiError("Invalid movie Id"))
		return
	}

	seasonID, err := strconv.Atoi(c.Param("seasonId"))
	if err != nil {
		logger.Error("Invalid seasonId", zap.Error(err))
		c.JSON(http.StatusBadRequest, models.NewApiError("Invalid seasonId"))
		return
	}

	// Проверяем, что сезон принадлежит фильму
	season, err := h.seasonsRepo.FindById(c, seasonID)
	if err != nil {
		logger.Error("Season not found", zap.Error(err))
		c.JSON(http.StatusNotFound, models.NewApiError("Season not found"))
		return
	}

	if season.MovieID != movieID {
		logger.Error("Season does not belong to the specified movie")
		c.JSON(http.StatusBadRequest, models.NewApiError("Season does not belong to the specified movie"))
		return
	}

	var request UpdateSeasonRequest
	if err := c.Bind(&request); err != nil {
		logger.Error("Failed to bind request payload", zap.Error(err))
		c.JSON(http.StatusBadRequest, models.NewApiError("Couldn't bind payload"))
		return
	}

	// Обновляем сезон
	err = h.seasonsRepo.Update(c, models.Season{
		Id:      seasonID,
		MovieID: movieID,
		Number:  request.Number,
	})
	if err != nil {
		logger.Error("Failed to update season", zap.Error(err))
		c.JSON(http.StatusInternalServerError, models.NewApiError("Failed to update season"))
		return
	}

	// Обновляем эпизоды
	for _, episodeReq := range request.Episodes {
		episode, err := h.episodesRepo.FindById(c, episodeReq.Id)
		if err != nil {
			logger.Warn("Episode not found", zap.Int("episodeId", episodeReq.Id))
			continue
		}

		if episode.SeasonID != seasonID {
			logger.Warn("Episode does not belong to the specified season", zap.Int("episodeId", episodeReq.Id))
			continue
		}

		err = h.episodesRepo.Update(c, models.Episode{
			Id:       episodeReq.Id,
			SeasonID: seasonID,
			Number:   episodeReq.Number,
			VideoURL: episodeReq.VideoURL,
		})
		if err != nil {
			logger.Warn("Failed to update episode", zap.Int("episodeId", episodeReq.Id), zap.Error(err))
		}
	}

	logger.Info("Season and episodes updated successfully")
	c.JSON(http.StatusOK, gin.H{"message": "Season and episodes updated successfully"})
}

// DeleteSeason godoc
// @Summary      Delete a season from a movie
// @Description  Removes a season and its episodes from a movie
// @Tags         content
// @Accept       json
// @Produce      json
// @Param        Id   path      int   true  "Movie ID"
// @Param        seasonId  path      int   true  "Season ID"
// @Success      200       {object}  object{message=string} "Season deleted successfully"
// @Failure      400       {object}  models.ApiError        "Invalid movie id or seasonId"
// @Failure      500       {object}  models.ApiError        "Internal server error"
// @Router       /admin/movies/{Id}/seasons/{seasonId} [delete]
func (h *ContentHandler) DeleteSeason(c *gin.Context) {
	logger := logger.GetLogger()

	// Получаем movie Id и seasonId из параметров URL
	movieID, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		logger.Error("Invalid movie id", zap.Error(err))
		c.JSON(http.StatusBadRequest, models.NewApiError("Invalid movie id"))
		return
	}

	seasonID, err := strconv.Atoi(c.Param("seasonId"))
	if err != nil {
		logger.Error("Invalid seasonId", zap.Error(err))
		c.JSON(http.StatusBadRequest, models.NewApiError("Invalid seasonId"))
		return
	}

	season, err := h.seasonsRepo.FindById(c, seasonID)
	if err != nil || season.MovieID != movieID {
		c.JSON(http.StatusBadRequest, models.NewApiError("Invalid season or movie"))
		return
	}

	// Удаляем сезон
	err = h.seasonsRepo.Delete(c, seasonID)
	if err != nil {
		logger.Error("Failed to delete season", zap.Error(err))
		c.JSON(http.StatusBadRequest, models.NewApiError(err.Error()))
		return
	}

	logger.Info("Season deleted successfully")
	c.JSON(http.StatusOK, gin.H{"message": "Season deleted successfully"})
}

// UpdateEpisode godoc
// @Summary      Update episode details
// @Description  Updates an existing episode's details
// @Tags         content
// @Accept       json
// @Produce      json
// @Param        episodeId path      int                        true  "Episode ID"
// @Param        request   body     admin.UpdateEpisodeRequest true  "Episode update data"
// @Success      200       {object} object{message=string}   "Episode updated successfully"
// @Failure      400       {object} models.ApiError         "Invalid episodeId or payload"
// @Failure      404       {object} models.ApiError         "Episode not found"
// @Failure      500       {object} models.ApiError         "Internal server error"
// @Router       /admin/seasons/{seasonId}/episodes/{episodeId} [put]
func (h *ContentHandler) UpdateEpisode(c *gin.Context) {
	logger := logger.GetLogger()

	episodeID, err := strconv.Atoi(c.Param("episodeId"))
	if err != nil {
		logger.Error("Invalid episodeId", zap.Error(err))
		c.JSON(http.StatusBadRequest, models.NewApiError("Invalid episodeId"))
		return
	}

	var request UpdateEpisodeRequest
	if err := c.Bind(&request); err != nil {
		logger.Error("Failed to bind request payload", zap.Error(err))
		c.JSON(http.StatusBadRequest, models.NewApiError("Couldn't bind payload"))
		return
	}

	// Проверяем, что эпизод принадлежит сезону
	episode, err := h.episodesRepo.FindById(c, episodeID)
	if err != nil {
		logger.Error("Episode not found", zap.Error(err))
		c.JSON(http.StatusNotFound, models.NewApiError("Episode not found"))
		return
	}

	if episode.SeasonID != request.SeasonID {
		logger.Error("Episode does not belong to the specified season")
		c.JSON(http.StatusBadRequest, models.NewApiError("Episode does not belong to the specified season"))
		return
	}

	// Обновляем эпизод
	err = h.episodesRepo.Update(c, models.Episode{
		Id:       episodeID,
		SeasonID: episode.SeasonID,
		Number:   request.Number,
		VideoURL: request.VideoURL,
	})
	if err != nil {
		logger.Error("Failed to update episode", zap.Error(err))
		c.JSON(http.StatusInternalServerError, models.NewApiError("Failed to update episode"))
		return
	}

	logger.Info("Episode updated successfully")
	c.JSON(http.StatusOK, gin.H{"message": "Episode updated successfully"})
}

// DeleteEpisode godoc
// @Summary      Delete an episode
// @Description  Removes an episode from a season
// @Tags         content
// @Accept       json
// @Produce      json
// @Param        episodeId path      int   true  "Episode ID"
// @Success      200       {object} object{message=string} "Episode deleted successfully"
// @Failure      400       {object} models.ApiError        "Invalid episodeId"
// @Failure      500       {object} models.ApiError        "Internal server error"
// @Router       /admin/seasons/{seasonId}/episodes/{episodeId} [delete]
func (h *ContentHandler) DeleteEpisode(c *gin.Context) {
	logger := logger.GetLogger()

	// Получаем episodeId из параметров URL
	episodeID, err := strconv.Atoi(c.Param("episodeId"))
	if err != nil {
		logger.Error("Invalid episodeId", zap.Error(err))
		c.JSON(http.StatusBadRequest, models.NewApiError("Invalid episodeId"))
		return
	}

	// Получаем эпизод, чтобы узнать его seasonId
	episode, err := h.episodesRepo.FindById(c, episodeID)
	if err != nil {
		logger.Error("Failed to retrieve episode", zap.Error(err))
		c.JSON(http.StatusBadRequest, models.NewApiError("Failed to retrieve episode"))
		return
	}

	// Удаляем эпизод
	err = h.episodesRepo.Delete(c, episode.Id)
	if err != nil {
		logger.Error("Failed to delete episode", zap.Error(err))
		c.JSON(http.StatusBadRequest, models.NewApiError(err.Error()))
		return
	}

	logger.Info("Episode deleted successfully")
	c.JSON(http.StatusOK, gin.H{"message": "Episode deleted successfully"})
}
