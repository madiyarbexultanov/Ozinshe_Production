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
	moviesRepo 			*repositories.MoviesRepository
	movieTypeRepo 		*repositories.MovieTypesRepository
	genresRepo 			*repositories.GenresRepository
	categoriesRepo  	*repositories.CategoriesRepository
	agesRepo 			*repositories.AgesRepository
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
	MovieTypeId     int             		`json:"movieTypeId"`
}

type addMediaRequest struct {
	Cover  		*multipart.FileHeader       `form:"cover"`
	Screenshots []*multipart.FileHeader 	`form:"screenshots"`
}

func NewMoviesHandler(moviesRepo *repositories.MoviesRepository, 
					  movieTypeRepo *repositories.MovieTypesRepository,
					  genresRepo *repositories.GenresRepository,
					  agesRepo 	 *repositories.AgesRepository,
					  categoriesRepo *repositories.CategoriesRepository) *MoviesHandler {
	return &MoviesHandler{
		moviesRepo: 	moviesRepo,
		movieTypeRepo: 	movieTypeRepo,
		genresRepo: 	genresRepo,
		categoriesRepo: categoriesRepo,
		agesRepo: 		agesRepo,
	}
}

// @Summary Get movie by ID
// @Description Get a movie by its ID
// @Tags Movies
// @Accept json
// @Produce json
// @Param id path int true "Movie ID"
// @Success 200 {object} models.Movie
// @Failure 400 {object} models.ApiError
// @Failure 404 {object} models.ApiError
// @Router /admin/movies/{id} [get]
func (h *MoviesHandler) FindById(c *gin.Context) {
	logger := logger.GetLogger()
	idStr := c.Param("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		// Логируем ошибку и возвращаем ответ с ошибкой
		logger.Error("Invalid movie ID", zap.String("id", idStr), zap.Error(err))
		c.JSON(http.StatusBadRequest, models.NewApiError("Invalid movie id"))
		return
	}

	// Пытаемся найти фильм в репозитории
	movie, err := h.moviesRepo.FindById(c, id)
	if err != nil {
		// Если фильм не найден, логируем ошибку и возвращаем ответ с ошибкой
		logger.Error("Movie not found", zap.Int("id", id), zap.Error(err))
		c.JSON(http.StatusNotFound, models.NewApiError(err.Error()))
		return
	}

	// Если фильм найден, возвращаем его в ответе
	logger.Info("Movie found", zap.Int("id", id))
	c.JSON(http.StatusOK, movie)
}

// @Summary Get all movies
// @Description Get a list of all movies
// @Tags Movies
// @Accept json
// @Produce json
// @Success 200 {array} models.Movie
// @Failure 404 {object} models.ApiError
// @Router /admin/movies [get]
func (h *MoviesHandler) FindAll(c *gin.Context) {
	logger := logger.GetLogger()

	filters := models.Moviesfilters {
		GenreIds: 	c.Query("genreids"),
		CategoryIds: c.Query("categoryids"),
		TypeIds: 	c.Query("typeids"),
	}

	movies, err := h.moviesRepo.FindAll(c, filters)
	if err != nil {
		// Логируем ошибку и возвращаем ответ с ошибкой
		logger.Error("Failed to load movies", zap.Error(err))
		c.JSON(http.StatusNotFound, models.NewApiError(err.Error()))
		return
	}

	// Если фильмы найдены, возвращаем их в ответе
	logger.Info("Movies loaded successfully", zap.Int("count", len(movies)))
	c.JSON(http.StatusOK, movies)
}

// @Summary Create a new movie
// @Description Create a new movie with the specified information
// @Tags Movies
// @Accept json
// @Produce json
// @Param movie body createMovieRequest true "Movie information"
// @Success 200 {object} map[string]int "Movie ID"
// @Failure 400 {object} models.ApiError
// @Failure 404 {object} models.ApiError
// @Failure 500 {object} models.ApiError
// @Router /admin/movies [post]
func (h *MoviesHandler) Create(c *gin.Context) {
	logger := logger.GetLogger()
	var request createMovieRequest

	// Пробуем распарсить тело запроса
	if err := c.Bind(&request); err != nil {
		// Логируем ошибку и возвращаем ответ с ошибкой
		logger.Error("Failed to bind request payload", zap.Error(err))
		c.JSON(http.StatusBadRequest, models.NewApiError("Couldn't bind payload"))
		return
	}

	// Проверяем, существует ли тип фильма
	movieType, err := h.movieTypeRepo.FindById(c, request.MovieTypeId)
	if err != nil {
		logger.Error("Movie type not found", zap.Int("movieTypeId", request.MovieTypeId), zap.Error(err))
		c.JSON(http.StatusNotFound, models.NewApiError("Movie type not found"))
		return
	}

	// Проверяем, существуют ли жанры
	genres, err := h.genresRepo.FindAllByIds(c, request.GenreIds)
	if err != nil {
		logger.Error("Genres not found", zap.Any("genreIds", request.GenreIds), zap.Error(err))
		c.JSON(http.StatusNotFound, models.NewApiError(err.Error()))
		return
	}

	// Проверяем, существуют ли категории
	categories, err := h.categoriesRepo.FindAllByIds(c, request.CategoryIds)
	if err != nil {
		logger.Error("Categories not found", zap.Any("categoryIds", request.CategoryIds), zap.Error(err))
		c.JSON(http.StatusNotFound, models.NewApiError(err.Error()))
		return
	}

	// Проверяем, существуют ли возрастные категории
	ages, err := h.agesRepo.FindAllByIds(c, request.AgeIds)
	if err != nil {
		logger.Error("Ages not found", zap.Any("ageIds", request.AgeIds), zap.Error(err))
		c.JSON(http.StatusNotFound, models.NewApiError(err.Error()))
		return
	}

	// Создаем новый фильм
	movie := models.Movie{
		Title:       request.Title,
		Categories:  categories,
		Genres:      genres,
		Ages:        ages,
		ReleaseYear: request.ReleaseYear,
		Runtime:     request.Runtime,
		KeyWords:    request.KeyWords,
		Description: request.Description,
		Director:    request.Director,
		Producer:    request.Producer,
		MovieTypeId: movieType.Id,
	}

	// Сохраняем фильм в базе данных
	movieID, err := h.moviesRepo.Create(c, movie)
	if err != nil {
		// Логируем ошибку и возвращаем ответ с ошибкой
		logger.Error("Failed to create movie", zap.String("title", movie.Title), zap.Error(err))
		c.JSON(http.StatusInternalServerError, models.NewApiError("Failed to create movie"))
		return
	}

	// Возвращаем ID нового фильма
	logger.Info("Movie created successfully", zap.Int("movieID", movieID))
	c.JSON(http.StatusOK, gin.H{"movieID": movieID})
}

// @Summary Update movie
// @Description Update an existing movie by its ID
// @Tags Movies
// @Accept json
// @Produce json
// @Param id path int true "Movie ID"
// @Param movie body createMovieRequest true "Movie information"
// @Success 200 {object} map[string]string "Movie updated successfully"
// @Failure 400 {object} models.ApiError
// @Failure 404 {object} models.ApiError
// @Failure 500 {object} models.ApiError
// @Router /admin/movies/{id} [patch]
func (h *MoviesHandler) Update(c *gin.Context) {
    logger := logger.GetLogger()

    // Получаем ID фильма из параметров URL
    idStr := c.Param("id")
    movieId, err := strconv.Atoi(idStr)
    if err != nil {
        logger.Error("Invalid movie ID", zap.String("id", idStr), zap.Error(err))
        c.JSON(http.StatusBadRequest, models.NewApiError("Invalid movie ID"))
        return
    }

    // Проверяем, существует ли фильм
    movie, err := h.moviesRepo.FindById(c, movieId)
    if err != nil {
        logger.Error("Movie not found", zap.Int("id", movieId), zap.Error(err))
        c.JSON(http.StatusNotFound, models.NewApiError("Movie not found"))
        return
    }

    // Создаем структуру запроса для обновления
    var request createMovieRequest
    if err := c.Bind(&request); err != nil {
        logger.Error("Failed to bind request payload", zap.Error(err))
        c.JSON(http.StatusBadRequest, models.NewApiError("Couldn't bind payload"))
        return
    }

    // Обновляем фильм
    movie.Title = request.Title
    movie.ReleaseYear = request.ReleaseYear
    movie.Runtime = request.Runtime
    movie.KeyWords = request.KeyWords
    movie.Description = request.Description
    movie.Director = request.Director
    movie.Producer = request.Producer
    movie.MovieTypeId = request.MovieTypeId

    // Обновление жанров, категорий и возрастных ограничений
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

    movie.Categories = categories
    movie.Genres = genres
    movie.Ages = ages

    // Сохраняем изменения в базе данных
    err = h.moviesRepo.Update(c, movie)
    if err != nil {
        logger.Error("Failed to update movie", zap.Int("id", movieId), zap.Error(err))
        c.JSON(http.StatusInternalServerError, models.NewApiError("Failed to update movie"))
        return
    }

    logger.Info("Movie updated successfully", zap.Int("id", movieId))
    c.JSON(http.StatusOK, gin.H{"message": "Movie updated successfully"})
}

// @Summary Delete movie
// @Description Delete a movie by its ID
// @Tags Movies
// @Accept json
// @Produce json
// @Param id path int true "Movie ID"
// @Success 200 {object} map[string]string "Movie deleted successfully"
// @Failure 400 {object} models.ApiError
// @Failure 404 {object} models.ApiError
// @Failure 500 {object} models.ApiError
// @Router /admin/movies/{id} [delete]
func (h *MoviesHandler) Delete(c *gin.Context) {
    logger := logger.GetLogger()

    // Получаем ID фильма из параметров URL
    idStr := c.Param("id")
    movieId, err := strconv.Atoi(idStr)
    if err != nil {
        logger.Error("Invalid movie ID", zap.String("id", idStr), zap.Error(err))
        c.JSON(http.StatusBadRequest, models.NewApiError("Invalid movie ID"))
        return
    }

    // Проверяем, существует ли фильм
    movie, err := h.moviesRepo.FindById(c, movieId)
    if err != nil {
        logger.Error("Movie not found", zap.Int("id", movieId), zap.Error(err))
        c.JSON(http.StatusNotFound, models.NewApiError("Movie not found"))
        return
    }

    // Удаляем фильм
    err = h.moviesRepo.Delete(c, movie.Id)
    if err != nil {
        logger.Error("Failed to delete movie", zap.Int("id", movieId), zap.Error(err))
        c.JSON(http.StatusInternalServerError, models.NewApiError("Failed to delete movie"))
        return
    }

    logger.Info("Movie deleted successfully", zap.Int("id", movieId))
    c.JSON(http.StatusOK, gin.H{"message": "Movie deleted successfully"})
}


// @Summary Add media (cover, screenshots)
// @Description Upload cover and screenshots for a movie
// @Tags Movies
// @Accept multipart/form-data
// @Produce json
// @Param movieId path int true "Movie ID"
// @Param cover formData file true "Cover image"
// @Param screenshots formData file true "Screenshots" multiple
// @Success 200 {object} map[string]string "Media added successfully"
// @Failure 400 {object} models.ApiError
// @Failure 404 {object} models.ApiError
// @Failure 500 {object} models.ApiError
// @Router /admin/movies/{movieId}/media [patch]
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