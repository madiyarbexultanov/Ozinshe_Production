package repositories

import (
	"context"
	"ozinshe_production/models"

	"github.com/jackc/pgx/v5/pgxpool"
)

type UsersRepository struct {
	db *pgxpool.Pool
}

func NewUsersRepository(conn *pgxpool.Pool) *UsersRepository {
	return &UsersRepository{db: conn}
}

func (r *UsersRepository) FindAll(c context.Context) ([]models.User, error)  {
	rows, err := r.db.Query(c, "select id, name, email from users")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var users []models.User
	for rows.Next() {
		var user models.User
		err := rows.Scan(&user.Id, &user.Name, &user.Email)
		if err != nil {
			return nil, err
		}
		users = append(users, user)
	}

	return users, nil
}

func (r *UsersRepository) FindById(c context.Context, id int) (models.User, error)  {
	var user models.User
	row := r.db.QueryRow(c, "select id, name, email, role_id from users where id = $1", id)
	err := row.Scan(&user.Id, &user.Name, &user.Email, &user.RoleID)
	if err != nil {
		return models.User{}, err
	}
	return user, nil
}


func (r *UsersRepository) SignUp(c context.Context, user models.User) (int, error) {
	var id int
	err := r.db.QueryRow(c, "insert into users(email, password) values($1, $2) returning id", 
		user.Email, user.PasswordHash).Scan(&id)

	if err != nil {
		return 0, err
	}

	return id, nil
}

func (r *UsersRepository) Update(c context.Context, id int, user models.User) error {
	_, err := r.db.Exec(c, `
        UPDATE users SET email=$1, name=$2 WHERE id=$3`,
		 user.Email, user.Name, id)

	if err != nil {
		return err
	}
	return nil
}

func (r *UsersRepository) Delete(c context.Context, id int) error {
	_, err := r.db.Exec(c, "delete from users where id=$1", id)
	if err != nil {
		return err
	}
	return nil
}


func (r *UsersRepository) FindByEmail(c context.Context, email string) (models.User, error) {
	var user models.User
	row := r.db.QueryRow(c, "select id, email, password from users where email = $1", email)
	if err := row.Scan(&user.Id, &user.Email, &user.PasswordHash); err != nil {
		return models.User{}, err
	}

	return user, nil
}

func (r *UsersRepository) AssignRole(c context.Context, userID int, roleID int) error {
	_, err := r.db.Exec(c, "UPDATE users SET role_id = $1 WHERE id = $2", roleID, userID)
	return err
}