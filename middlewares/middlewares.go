package middlewares

import (
	"net/http"
	"ozinshe_production/config"
	"ozinshe_production/logger"
	"ozinshe_production/models"
	"ozinshe_production/repositories"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"go.uber.org/zap"
)

func AuthMiddleware(c *gin.Context) {
	logger := logger.GetLogger()

	authHeader := c.GetHeader("Authorization")
	if authHeader == "" {
		logger.Warn("Authorization header missing")
		c.JSON(http.StatusUnauthorized, models.NewApiError("authorization header required"))
		c.Abort()
		return
	}

	tokenString := strings.Split(authHeader, "Bearer ")[1]
	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		return []byte(config.Config.JwtSecretKey), nil
	})

	if err != nil || !token.Valid {
		logger.Error("Invalid token", zap.String("error", err.Error()))
		c.JSON(http.StatusUnauthorized, models.NewApiError("invalid token"))
		c.Abort()
		return
	}

	subject, err := token.Claims.GetSubject()
	if err != nil {
		logger.Error("Error getting subject from token", zap.String("error", err.Error()))
		c.JSON(http.StatusUnauthorized, models.NewApiError("error while getting subject"))
		c.Abort()
		return
	}

	userId, _ := strconv.Atoi(subject)
	logger.Info("Token validated", zap.Int("userId", userId))
	c.Set("userId", userId)
	c.Next()
}

func CheckPermissionMiddleware(c *gin.Context) {
	logger := logger.GetLogger()
	userId, exists := c.Get("userId")
	if !exists {
		c.JSON(http.StatusUnauthorized, models.NewApiError("user not found"))
		c.Abort()
		return
	}
	logger.Info("User ID found in context", zap.Int("userId", userId.(int)))

	// Получаем соединение с базой из контекста
	db, exists := c.Get("db")
	if !exists {
		c.JSON(http.StatusInternalServerError, models.NewApiError("database connection not found"))
		c.Abort()
		return
	}

	pool, ok := db.(*pgxpool.Pool)
	if !ok {
		c.JSON(http.StatusInternalServerError, models.NewApiError("invalid database connection"))
		c.Abort()
		return
	}

	userRepo := repositories.NewUsersRepository(pool)
	user, err := userRepo.FindById(c, userId.(int))
	if err != nil {
		logger.Error("User not found in database", zap.String("error", err.Error()))
		c.JSON(http.StatusUnauthorized, models.NewApiError("user not found"))
		c.Abort()
		return
	}
	logger.Info("User found", zap.Int("userId", user.Id), zap.String("userName", user.Name))

	roleRepo := repositories.NewRolesRepository(pool)
	role, err := roleRepo.FindById(c, user.RoleID)
	if err != nil {
		logger.Error("Role not found in database", zap.Int("roleId", user.RoleID), zap.String("error", err.Error()))
		c.JSON(http.StatusUnauthorized, models.NewApiError("role not found"))
		c.Abort()
		return
	}
	logger.Info("Role found", zap.Int("roleId", role.Id), zap.String("roleName", role.Name))

	// Карта для проверки прав на редактирование
	permissions := map[string]bool{
		"/admin/categories": role.CanEditCategories,
		"/admin/projects":   role.CanEditProjects,
		"/admin/users":      role.CanEditUsers,
		"/admin/roles":      role.CanEditRoles,
		"/admin/genres":     role.CanEditGenres,
		"/admin/ages":       role.CanEditAges,
	}

	// Проверяем, есть ли разрешение на текущий маршрут
	if permission, exists := permissions[c.Request.URL.Path]; exists && !permission {
		c.JSON(http.StatusForbidden, models.NewApiError("user does not have permission to edit "+c.Request.URL.Path))
		c.Abort()
		return
	}

	// Если роль имеет доступ, продолжаем выполнение
	logger.Info("User has appropriate permissions")
	c.Next()
}