package admin

import (
	"mime/multipart"
	"net/http"
	"ozinshe_production/logger"
	"ozinshe_production/models"
	"ozinshe_production/repositories"
	"strconv"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

type MoviesHandler struct {
	moviesRepo 		*repositories.MoviesRepository
	genresRepo 		*repositories.GenresRepository
	categoriesRepo  *repositories.CategoriesRepository
	agesRepo 		*repositories.AgesRepository
	seasonsRepo     *repositories.SeasonsRepository
	episodesRepo    *repositories.EpisodesRepository
}

type createMovieRequest struct {
	Title 			string					`json:"title"`
	CategoryIds 	[]int					`json:"categories"`
	GenreIds 		[]int					`json:"genres"`
	AgeIds 			[]int					`json:"ages"`
	ReleaseYear 	int						`json:"releaseYear"`
	Runtime 		int						`json:"runtime"`
	KeyWords 		[]string				`json:"keywords"`
	Description 	string					`json:"description"`
	Director 		string					`json:"director"`
	Producer 		string					`json:"producer"`
}


type createSeasonRequest struct {
	MovieID 		int 					`json:"movieId"`
	Number          int         			`json:"number"`
	Episodes        []createEpisodeRequest  `json:"episodes"`
}

type createEpisodeRequest struct {
	Number 	int 		`json:"number"`
	VideoURL	string 	`json:"videoURL"`
}

type addMediaRequest struct {
	Cover  		*multipart.FileHeader       `form:"cover"`
	Screenshots []*multipart.FileHeader 	`form:"screenshots"`
}


func NewMoviesHandler(moviesRepo *repositories.MoviesRepository, 
					  genresRepo *repositories.GenresRepository,
					  agesRepo 	 *repositories.AgesRepository,
					  categoriesRepo *repositories.CategoriesRepository,
					  seasonsRepo *repositories.SeasonsRepository,
					  episodesRepo *repositories.EpisodesRepository) *MoviesHandler {
	return &MoviesHandler{
		moviesRepo: 	moviesRepo,
		genresRepo: 	genresRepo,
		categoriesRepo: categoriesRepo,
		agesRepo: 		agesRepo,
		seasonsRepo:    seasonsRepo,
		episodesRepo:   episodesRepo,
	}
}

func (h *MoviesHandler) FindById(c *gin.Context) {
	logger := logger.GetLogger()
	idStr := c.Param("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		logger.Error("Invalid movie ID", zap.String("id", idStr), zap.Error(err))
		c.JSON(http.StatusBadRequest, models.NewApiError("Invalid movie id"))
		return
	}

	movie, err := h.moviesRepo.FindById(c, id)
	if err != nil {
		logger.Error("Movie not found", zap.Int("id", id), zap.Error(err))
		c.JSON(http.StatusNotFound, models.NewApiError(err.Error()))
		return
	}

	logger.Info("Movie found", zap.Int("id", id))
	c.JSON(http.StatusOK, movie)
}

func (h *MoviesHandler) FindAll(c *gin.Context) {
	logger := logger.GetLogger()
	movies, err := h.moviesRepo.FindAll(c)
	if err != nil {
		logger.Error("Failed to load movies", zap.Error(err))
		c.JSON(http.StatusNotFound, models.NewApiError(err.Error()))
		return
	}

	logger.Info("Movies loaded successfully", zap.Int("count", len(movies)))
	c.JSON(http.StatusOK, movies)
}

func (h *MoviesHandler) Create(c *gin.Context) {
	logger := logger.GetLogger()
	var request createMovieRequest

	if err := c.Bind(&request); err != nil {
		logger.Error("Failed to bind request payload", zap.Error(err))
		c.JSON(http.StatusBadRequest, models.NewApiError("Couldn't bind payload"))
		return
	}

	logger.Info("Received movie creation request", zap.String("title", request.Title))

	genres, err := h.genresRepo.FindAllByIds(c, request.GenreIds)
	if err != nil {
		logger.Error("Genres not found", zap.Any("genreIds", request.GenreIds), zap.Error(err))
		c.JSON(http.StatusNotFound, models.NewApiError(err.Error()))
		return
	}

	categories, err := h.categoriesRepo.FindAllByIds(c, request.CategoryIds)
	if err != nil {
		logger.Error("Categories not found", zap.Any("categoryIds", request.CategoryIds), zap.Error(err))
		c.JSON(http.StatusNotFound, models.NewApiError(err.Error()))
		return
	}

	ages, err := h.agesRepo.FindAllByIds(c, request.AgeIds)
	if err != nil {
		logger.Error("Ages not found", zap.Any("ageIds", request.AgeIds), zap.Error(err))
		c.JSON(http.StatusNotFound, models.NewApiError(err.Error()))
		return
	}

	movie := models.Movie {
		Title:        request.Title,
		Categories:   categories,
		Genres:       genres,
		Ages:         ages,
		ReleaseYear:  request.ReleaseYear,
		Runtime:      request.Runtime,
		KeyWords:     request.KeyWords,
		Description:  request.Description,
		Director:     request.Director,
		Producer:     request.Producer,
	}

	movieID, err := h.moviesRepo.Create(c, movie)
	if err != nil {
		logger.Error("Failed to create movie", zap.String("title", movie.Title), zap.Error(err))
		c.JSON(http.StatusInternalServerError, models.NewApiError("Failed to create movie"))
		return
	}

	logger.Info("Movie created successfully", zap.Int("movieID", movieID))
	c.JSON(http.StatusOK, gin.H{"movieID": movieID})
}

func (h *MoviesHandler) AddSeasonsAndEpisodes(c *gin.Context) {
	logger := logger.GetLogger()

	// Получаем movieId из параметров URL
	movieID := c.Param("movieId")
	if movieID == "" {
		logger.Error("movieId is required")
		c.JSON(http.StatusBadRequest, models.NewApiError("movieId is required"))
		return
	}

	var request createSeasonRequest
	// Приводим movieId из строки в int
	id, err := strconv.Atoi(movieID)
	if err != nil {
		logger.Error("Invalid movieId", zap.Error(err))
		c.JSON(http.StatusBadRequest, models.NewApiError("Invalid movieId"))
		return
	}

	// Встраиваем movieId в тело запроса
	request.MovieID = id

	// Привязываем остальные данные из тела запроса
	if err := c.Bind(&request); err != nil {
		logger.Error("Failed to bind request payload", zap.Error(err))
		c.JSON(http.StatusBadRequest, models.NewApiError("Couldn't bind payload"))
		return
	}

	// Создаем сезоны
	seasonID, err := h.seasonsRepo.Create(c, models.Season{
		MovieID: request.MovieID,
		Number:  request.Number,
	})
	if err != nil {
		logger.Error("Failed to create season", zap.Error(err))
		c.JSON(http.StatusInternalServerError, models.NewApiError("Failed to create season"))
		return
	}

	// Создаем эпизоды для сезона
	for _, episodeReq := range request.Episodes {
		episode := models.Episode{
			SeasonID: seasonID,
			Number:   episodeReq.Number,
			VideoURL: episodeReq.VideoURL,
		}
		_, err := h.episodesRepo.Create(c, episode)
		if err != nil {
			logger.Error("Failed to create episode", zap.Error(err))
			c.JSON(http.StatusInternalServerError, models.NewApiError("Failed to create episode"))
			return
		}
	}

	logger.Info("Seasons and episodes added successfully", zap.Int("movieID", request.MovieID))
	c.JSON(http.StatusOK, gin.H{"message": "Seasons and episodes added successfully"})
}

func (h *MoviesHandler) AddMedia(c *gin.Context) {
	logger := logger.GetLogger()

	// Получаем ID фильма из параметров URL
	idStr := c.Param("movieId")
	logger.Info("Received movie id", zap.String("movieId", idStr))
	movieId, err := strconv.Atoi(idStr)
	if err != nil {
		logger.Error("Invalid movie id", zap.String("movieId", idStr), zap.Error(err))
		c.JSON(http.StatusBadRequest, models.NewApiError("Invalid movie id"))
		return
	}

	// Проверяем, существует ли фильм
	movie, err := h.moviesRepo.FindById(c, movieId)
	if err != nil {
		logger.Error("Movie not found", zap.Int("movieId", movieId), zap.Error(err))
		c.JSON(http.StatusNotFound, models.NewApiError("Movie not found"))
		return
	}

	// Используем структуру addMediaRequest для парсинга данных
	var request addMediaRequest
	if err := c.ShouldBind(&request); err != nil {
		logger.Error("Failed to bind data", zap.Error(err))
		c.JSON(http.StatusBadRequest, models.NewApiError("Failed to bind data"))
		return
	}

	// Обрабатываем загрузку обложки
	if request.Cover != nil {
		filename, err := savePoster(c, request.Cover)
		if err != nil {
			logger.Error("Couldn't save cover", zap.Error(err))
			c.JSON(http.StatusInternalServerError, models.NewApiError("Couldn't save cover"))
			return
		}
		movie.Cover = filename
		logger.Info("Cover saved successfully", zap.String("filename", filename))
	}

	// Обрабатываем загрузку скриншотов
	var screenshotFilenames []string
	for _, screenshot := range request.Screenshots {
		filename, err := savePoster(c, screenshot)
		if err != nil {
			logger.Error("Couldn't save screenshot", zap.Error(err))
			c.JSON(http.StatusInternalServerError, models.NewApiError("Couldn't save screenshot"))
			return
		}
		screenshotFilenames = append(screenshotFilenames, filename)
		logger.Info("Screenshot saved successfully", zap.String("filename", filename))
	}

	// Обновляем фильм в БД
	err = h.moviesRepo.UpdateCoverAndScreenshots(c, movieId, movie.Cover, screenshotFilenames)
	if err != nil {
		logger.Error("Couldn't update movie", zap.Error(err))
		c.JSON(http.StatusInternalServerError, models.NewApiError("Couldn't update movie"))
		return
	}

	logger.Info("Media has been successfully added", zap.Int("movieId", movieId))
	c.JSON(http.StatusOK, gin.H{"message": "Media has been successfully added"})
}