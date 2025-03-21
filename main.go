package main

import (
	"context"
	"ozinshe_production/config"
	"ozinshe_production/docs"
	"ozinshe_production/handlers/admin"
	"ozinshe_production/handlers/public"
	"ozinshe_production/logger"
	"ozinshe_production/middlewares"
	"ozinshe_production/repositories"
	"time"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/spf13/viper"
	"go.uber.org/zap"

	ginzap "github.com/gin-contrib/zap"
	swaggerfiles "github.com/swaggo/files"
	swagger "github.com/swaggo/gin-swagger"
)

// @title           Ozinshe Production
// @version         1.0
// @description     Personal online platform providing information about media content
// @termsOfService  http://swagger.io/terms/
//
// @contact.name   Madiyar Bexultanov
// @contact.url    https://www.linkedin.com/in/madiyar-bexultanov-b21902258/
// @contact.email  bexultanovmadiyar@gmail.com
//
// @license.name  Apache 2.0
// @license.url   http://www.apache.org/licenses/LICENSE-2.0.html
//
// @host      localhost:8081
// @BasePath  /
//
// @securityDefinitions.apikey Bearer
// @in header
// @name Authorization
// @description Type "Bearer" followed by a space and JWT token.
//
// @externalDocs.description  OpenAPI
// @externalDocs.url          https://swagger.io/resources/open-api/
func main() {
	r := gin.New()

	logger := logger.GetLogger()

	defer func() {
		if r := recover(); r != nil {
			logger.Error("Application crashed!", zap.Any("error", r))
		}
	}()

	r.Use(
		ginzap.Ginzap(logger, time.RFC3339, true),
		ginzap.RecoveryWithZap(logger, true),
	)

	corsConfig := cors.Config{
		AllowAllOrigins: true,
		AllowHeaders:    []string{"*"},
		AllowMethods:    []string{"*"},
	}

	r.Use(cors.New(corsConfig))
	gin.SetMode(gin.ReleaseMode)

	logger.Info("Loading configuration...")
	err := loadConfig()
	if err != nil {
		logger.Fatal("Failed to load config", zap.Error(err))
	}

	logger.Info("Connecting to database...")
	conn, err := connectToDb()
	if err != nil {
		logger.Fatal("Database connection failed", zap.Error(err))
	}

	r.Use(func(c *gin.Context) {
		c.Set("db", conn)
		c.Next()
	})

	moviesRepository := repositories.NewMoviesRepository(conn)
	recommendationsRepository := repositories.NewRecommendationsRepository(conn)
	seasonsRepository := repositories.NewSeasonsRepository(conn)
	episodesRepository := repositories.NewEpisodesRepository(conn)
	movieTypesRepository := repositories.NewMovieTypesRepository(conn)
	usersRepository := repositories.NewUsersRepository(conn)
	agesRepository := repositories.NewAgesRepository(conn)
	genresRepository := repositories.NewGenresRepository(conn)
	categoriesRepository := repositories.NewCategoriesRepository(conn)
	rolesRepository := repositories.NewRolesRepository(conn)
	searchRepository  := repositories.NewSearchRepository(conn)
	mediaRepository := repositories.NewMediaRepository(conn)

	homepageRepository := repositories.NewHomepageRepository(conn)

	moviesHandler := admin.NewMoviesHandler(moviesRepository, movieTypesRepository, genresRepository,  agesRepository, categoriesRepository)
	recommendationsHandler := admin.NewRecommendationsHandler(recommendationsRepository)
	contentsHandler := admin.NewContentsHandler(seasonsRepository, episodesRepository)
	usersHandler := admin.NewUsersHandler(usersRepository)
	movieTypesHandler := admin.NewMovieTypesHandler(movieTypesRepository)
	agesHandler := admin.NewAgesHandler(agesRepository)
	genresHandler := admin.NewGenresHandler(genresRepository)
	categoriesHandler := admin.NewCategoriesHandler(categoriesRepository)
	rolesHandler := admin.NewRolesHandler(rolesRepository)
	searchHandler := admin.NewSearchHandler(searchRepository)
	mediaHandler := admin.NewMediaHandler(mediaRepository)

	HomepageHandler := public.NewHomepageHandler(homepageRepository)

	authHandler := public.NewAuthHandlers(usersRepository)
	profilesHandler := public.NewProfilesHandler(usersRepository)
	googleAuthHandler := public.NewAuthHandlers(usersRepository)

	authorized := r.Group("")
	authorized.Use(middlewares.AuthMiddleware)

	authorized.GET("/public/profile/:id", profilesHandler.UserProfile)
	authorized.PUT("/public/profile/:id", profilesHandler.Update)
	authorized.PUT("/public/profile/changepassword/:id", profilesHandler.ChangePassword)

	authorized.GET("/public/homepage", HomepageHandler.GetMainScreen)

	authorized.POST("/public/auth/signOut", authHandler.SignOut)

	permitted := r.Group("")
	permitted.Use(middlewares.AuthMiddleware)
	permitted.Use(middlewares.CheckPermissionMiddleware)
	
	// Фильмы
	movies := permitted.Group("/admin/movies")
	{
		movies.GET("", moviesHandler.FindAll)
		movies.POST("", moviesHandler.Create)
		movies.GET("/:id", moviesHandler.FindById)
		movies.PUT("/:id", moviesHandler.Update)
		movies.DELETE("/:id", moviesHandler.Delete)
	
		seasons := movies.Group("/:id/seasons")
		{
			seasons.POST("", contentsHandler.AddSeasonsAndEpisodes)
			seasons.PUT("/:seasonId/edit", contentsHandler.UpdateSeason)
			seasons.DELETE("/:seasonId", contentsHandler.DeleteSeason)
		}
	}

	// Сезоны и эпизоды
	seasons := permitted.Group("/admin/seasons")
	{
		seasons.PUT("/:seasonId/episodes/:episodeId", contentsHandler.UpdateEpisode)
		seasons.DELETE("/:seasonId/episodes/:episodeId", contentsHandler.DeleteEpisode)
	}

	// Обложка и Скриншоты
	media := permitted.Group("/admin/media")
	{
		media.GET("/movies/:id", mediaHandler.GetMovieMedia)       
		media.PATCH("/movies/:id", mediaHandler.UploadMovieMedia)             
		media.POST("/movies/:id/media", mediaHandler.UploadSingleMovieMedia) 
		media.DELETE("/movies/:id/media", mediaHandler.DeleteMovieMedia)  
	}
	
	
	// Рекомендации
	recommendations := permitted.Group("/admin/recommendations")
	{
		recommendations.GET("", recommendationsHandler.FindAll)
		recommendations.POST("", recommendationsHandler.Create)
		recommendations.GET("/:id", recommendationsHandler.FindById)
		recommendations.DELETE("/:id", recommendationsHandler.Delete)
	}
	
	// Типы фильмов
	movieTypes := permitted.Group("/admin/movieTypes")
	{
		movieTypes.GET("", movieTypesHandler.FindAll)
		movieTypes.POST("", movieTypesHandler.Create)
		movieTypes.GET("/:id", movieTypesHandler.FindById)
		movieTypes.PUT("/:id", movieTypesHandler.Update)
		movieTypes.DELETE("/:id", movieTypesHandler.Delete)
	}
	
	// Категории
	categories := permitted.Group("/admin/categories")
	{
		categories.GET("", categoriesHandler.FindAll)
		categories.POST("", categoriesHandler.Create)
		categories.GET("/:id", categoriesHandler.FindById)
		categories.PUT("/:id", categoriesHandler.Update)
		categories.DELETE("/:id", categoriesHandler.Delete)
	}
	
	// Пользователи
	users := permitted.Group("/admin/users")
	{
		users.GET("", usersHandler.FindAll)
		users.GET("/:id", usersHandler.FindById)
		users.PUT("/:id/role", usersHandler.AssignRole)
		users.DELETE("/:id", usersHandler.Delete)
	}
	
	// Роли
	roles := permitted.Group("/admin/roles")
	{
		roles.GET("", rolesHandler.FindAll)
		roles.POST("", rolesHandler.Create)
		roles.GET("/:id", rolesHandler.FindById)
		roles.PUT("/:id", rolesHandler.Update)
		roles.DELETE("/:id", rolesHandler.Delete)
	}
	
	// Жанры
	genres := permitted.Group("/admin/genres")
	{
		genres.GET("", genresHandler.FindAll)
		genres.POST("", genresHandler.Create)
		genres.GET("/:id", genresHandler.FindById)
		genres.PUT("/:id", genresHandler.Update)
		genres.DELETE("/:id", genresHandler.Delete)
	}
	
	// Возрастные ограничения
	ages := permitted.Group("/admin/ages")
	{
		ages.GET("", agesHandler.FindAll)
		ages.POST("", agesHandler.Create)
		ages.GET("/:id", agesHandler.FindById)
		ages.PUT("/:id", agesHandler.Update)
		ages.DELETE("/:id", agesHandler.Delete)
	}
	
	// Поиск
	permitted.GET("/admin/search", searchHandler.SearchAll)
	

	unauthorized := r.Group("")
	unauthorized.POST("/public/auth/signUp", authHandler.SignUp)
	unauthorized.POST("/public/auth/signIn", authHandler.SignIn)

	unauthorized.GET("/public/auth/google", googleAuthHandler.GoogleLogin)
	unauthorized.GET("/public/auth/google/callback", authHandler.GoogleCallback)

	docs.SwaggerInfo.BasePath = "/"
	unauthorized.GET("/swagger/*any", swagger.WrapHandler(swaggerfiles.Handler))

	logger.Info("Application starting...")
	for _, route := range r.Routes() {
		logger.Info("Registered route", zap.String("method", route.Method), zap.String("path", route.Path))
	}

	if err := r.Run(config.Config.AppHost); err != nil {
		logger.Fatal("Server failed to start", zap.Error(err))
	}
}

func loadConfig() error {
	viper.SetConfigFile(".env")
    viper.AutomaticEnv()

	err := viper.ReadInConfig()
	if err != nil {
		return err
	}

	var mapConfig config.MapConfig
	err = viper.Unmarshal(&mapConfig)
	if err != nil {
		return err
	}

	config.Config = &mapConfig

	return nil
}

func connectToDb() (*pgxpool.Pool, error) {
	conn, err := pgxpool.New(context.Background(), config.Config.DbConnectionString)
	if err != nil {
		return nil, err
	}

	err = conn.Ping(context.Background())
	if err != nil {
		return nil, err
	}
	return conn, nil
}
