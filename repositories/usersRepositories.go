package repositories

import (
	"context"
	"ozinshe_production/logger"
	"ozinshe_production/models"

	"github.com/jackc/pgx/v5/pgxpool"
	"go.uber.org/zap"
)

type UsersRepository struct {
	db *pgxpool.Pool
}

func NewUsersRepository(conn *pgxpool.Pool) *UsersRepository {
	return &UsersRepository{db: conn}
}


func (r *UsersRepository) SignUp(c context.Context, user models.User) (int, error) {
	logger := logger.GetLogger()
	logger.Info("Creating new user", zap.String("email", user.Email))

	var id int
	err := r.db.QueryRow(c, "insert into users(email, password) values($1, $2) returning id", 
		user.Email, user.PasswordHash).Scan(&id)

	if err != nil {
		logger.Error("Could not create user", zap.Error(err))
		return 0, err
	}

	logger.Info("Successfully created user", zap.Int("user_id", id))
	return id, nil
}

func (r *UsersRepository) FindByEmail(c context.Context, email string) (models.User, error) {
	logger := logger.GetLogger()
	logger.Info("Fetching user by email", zap.String("email", email))

	var user models.User
	row := r.db.QueryRow(c, "select id, email, password from users where email = $1", email)
	if err := row.Scan(&user.Id, &user.Email, &user.PasswordHash); err != nil {
		logger.Error("Could not fetch user by email", zap.Error(err))
		return models.User{}, err
	}

	logger.Info("Successfully fetched user by email", zap.String("email", email))
	return user, nil
}


