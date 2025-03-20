package repositories

import (
	"context"
	"fmt"
	"ozinshe_production/models"

	"github.com/jackc/pgx/v5/pgxpool"
)

type UsersRepository struct {
	db *pgxpool.Pool
}

func NewUsersRepository(conn *pgxpool.Pool) *UsersRepository {
	return &UsersRepository{db: conn}
}

func (r *UsersRepository) FindAll(c context.Context, filters models.Userfilters) ([]models.User, error)  {
	sql := "select id, name, email, phone_number, birth_date from users where 1=1"

	if filters.Sort != "" {
		sql = fmt.Sprintf("%s and created_at ilike `%%%s%%`", sql, filters.Sort)
	}

	rows, err := r.db.Query(c, sql, filters)
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

func (r *UsersRepository) UserProfile(c context.Context, id int) (models.User, error)  {
	var user models.User
	row := r.db.QueryRow(c, "select id, name, email, phone_number, birth_date from users where id = $1", id)
	err := row.Scan(&user.Id, &user.Name, &user.Email, &user.Phone, &user.Birthday)
	if err != nil {
		return models.User{}, err
	}
	return user, nil
}


func (r *UsersRepository) Update(c context.Context, id int, user models.User) error {
	_, err := r.db.Exec(c, `
        UPDATE users SET email=$1, name=$2, phone_number=$3, birth_date=$4 WHERE id=$5`,
		 user.Email, user.Name, user.Phone, user.Birthday, id)

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

func (r *UsersRepository) ChangePasswordHash(c context.Context, id int, password string) error {
	_, err := r.db.Exec(c, "update users set password=$1 where id=$2", password, id)
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