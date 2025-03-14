package main

import (
	"context"
	"ozinshe_production/config"
	"ozinshe_production/docs"
	"ozinshe_production/handlers/public"
	"ozinshe_production/handlers/admin"
	"ozinshe_production/logger"
	"ozinshe_production/middlewares"
	"ozinshe_production/repositories"
	"time"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/spf13/viper"

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


	

	err := loadConfig()
	if err != nil {
		panic(err)
	}

	conn, err := connectToDb()
	if err != nil {
		panic(err)
	}

	r.Use(func(c *gin.Context) {
		c.Set("db", conn)
		c.Next()
	})

	moviesRepository := repositories.NewMoviesRepository(conn)
	seasonsRepository := repositories.NewSeasonsRepository(conn)
	episodessRepository := repositories.NewEpisodesRepository(conn)
	usersRepository := repositories.NewUsersRepository(conn)
	agesRepository := repositories.NewAgesRepository(conn)
	genresRepository := repositories.NewGenresRepository(conn)
	categoriesRepository := repositories.NewCategoriesRepository(conn)
	rolesRepository := repositories.NewRolesRepository(conn)


	moviesHandler := admin.NewMoviesHandler(moviesRepository, genresRepository,  agesRepository, categoriesRepository, seasonsRepository, episodessRepository)
	usersHandler := admin.NewUsersHandler(usersRepository)
	agesHandler := admin.NewAgesHandler(agesRepository)
	genresHandler := admin.NewGenresHandler(genresRepository)
	categoriesHandler := admin.NewCategoriesHandler(categoriesRepository)
	rolesHandler := admin.NewRolesHandler(rolesRepository)


	authHandler := public.NewAuthHandlers(usersRepository)
	profilesHandler := public.NewProfilesHandler(usersRepository)
	googleAuthHandler := public.NewAuthHandlers(usersRepository)

	authorized := r.Group("")
	authorized.Use(middlewares.AuthMiddleware)

	authorized.GET("/public/profile/:id", profilesHandler.UserProfile)
	authorized.PUT("/public/profile/:id", profilesHandler.Update)
	authorized.PUT("/public/profile/changepassword/:id", profilesHandler.ChangePassword)

	authorized.POST("/public/auth/signOut", authHandler.SignOut)

	permitted := r.Group("")
	permitted.Use(middlewares.AuthMiddleware)
	permitted.Use(middlewares.CheckPermissionMiddleware)

	permitted.GET("/admin/movies", moviesHandler.FindAll)
	permitted.GET("/admin/movies/:id", moviesHandler.FindById)
	permitted.POST("/admin/movies", moviesHandler.Create)
	permitted.POST("/admin/movies/:movieId/seasons", moviesHandler.AddSeasonsAndEpisodes)
	permitted.PATCH("/admin/movies/:movieId/media", moviesHandler.AddMedia)
	
	permitted.GET("/admin/categories", categoriesHandler.FindAll)
	permitted.GET("/admin/categories/:id", categoriesHandler.FindById)
	permitted.POST("/admin/categories", categoriesHandler.Create)
	permitted.PUT("/admin/categories/:id", categoriesHandler.Update)
	permitted.DELETE("/admin/categories/:id", categoriesHandler.Delete)

	permitted.GET("/admin/users", usersHandler.FindAll)
	permitted.GET("/admin/users/:id", usersHandler.FindById)
	permitted.PUT("/admin/users/getRole/:id", usersHandler.AssignRole)
	permitted.DELETE("/admin/users/:id", usersHandler.Delete)

	permitted.GET("/admin/roles", rolesHandler.FindAll)
	permitted.GET("/admin/roles/:id", rolesHandler.FindById)
	permitted.POST("/admin/roles", rolesHandler.Create)
	permitted.PUT("/admin/roles/:id", rolesHandler.Update)
	permitted.DELETE("/admin/roles/:id", rolesHandler.Delete)

	permitted.GET("/admin/genres", genresHandler.FindAll)
	permitted.GET("/admin/genres/:id", genresHandler.FindById)
	permitted.POST("/admin/genres", genresHandler.Create)
	permitted.PUT("/admin/genres/:id", genresHandler.Update)
	permitted.DELETE("/admin/genres/:id", genresHandler.Delete)
	
	permitted.GET("/admin/ages", agesHandler.FindAll)
	permitted.GET("/admin/ages/:id", agesHandler.FindById)
	permitted.POST("/admin/ages", agesHandler.Create)
	permitted.PUT("/admin/ages/:id", agesHandler.Update)
	permitted.DELETE("/admin/ages/:id", agesHandler.Delete)

	unauthorized := r.Group("")
	unauthorized.POST("/public/auth/signUp", authHandler.SignUp)
	unauthorized.POST("/public/auth/signIn", authHandler.SignIn)

	unauthorized.GET("/public/auth/google", googleAuthHandler.GoogleLogin)
	unauthorized.GET("/public/auth/google/callback", authHandler.GoogleCallback)

	docs.SwaggerInfo.BasePath = "/"
	unauthorized.GET("/swagger/*any", swagger.WrapHandler(swaggerfiles.Handler))

	logger.Info("Application starting...")

	r.Run(config.Config.AppHost)
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
