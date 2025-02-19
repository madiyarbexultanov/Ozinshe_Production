package main

import (
	"context"
	"ozinshe_production/config"
	"ozinshe_production/docs"
	"ozinshe_production/handlers/public"
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

	usersRepository := repositories.NewUsersRepository(conn)

	// usersHandler := handlers.NewUsersHandler(usersRepository)
	authHandler := public.NewAuthHandlers(usersRepository)
	googleAuthHandler := public.NewAuthHandlers(usersRepository)

	authorized := r.Group("")
	authorized.Use(middlewares.AuthMiddleware)
	authorized.POST("/auth/signOut", authHandler.SignOut)

	unauthorized := r.Group("")
	unauthorized.POST("/auth/signUp", authHandler.SignUp)
	unauthorized.POST("/auth/signIn", authHandler.SignIn)

	unauthorized.GET("/auth/google", googleAuthHandler.GoogleLogin)
	unauthorized.GET("/auth/google/callback", authHandler.GoogleCallback)

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
