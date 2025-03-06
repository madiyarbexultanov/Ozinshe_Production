package repositories

import (
	"context"
	"ozinshe_production/models"

	"github.com/jackc/pgx/v5/pgxpool"
)

type RolesRepository struct {
	db *pgxpool.Pool
}

func NewRolesRepository(conn *pgxpool.Pool) *RolesRepository {
	return &RolesRepository{db: conn}
}

func (r *RolesRepository) FindAll(c context.Context) ([]models.Role, error) {
	rows, err := r.db.Query(c, "SELECT id, name, can_edit_projects, can_edit_categories, can_edit_users, can_edit_roles, can_edit_genres, can_edit_ages FROM roles")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var roles []models.Role
	for rows.Next() {
		var role models.Role
		err := rows.Scan(&role.Id, &role.Name, &role.CanEditProjects, &role.CanEditCategories, &role.CanEditUsers, &role.CanEditRoles, &role.CanEditGenres, &role.CanEditAges)
		if err != nil {
			return nil, err
		}
		roles = append(roles, role)
	}

	return roles, nil
}

func (r *RolesRepository) FindById(c context.Context, id int) (models.Role, error) {
	var role models.Role
	row := r.db.QueryRow(c, "SELECT id, name, can_edit_projects, can_edit_categories, can_edit_users, can_edit_roles, can_edit_genres, can_edit_ages FROM roles WHERE id = $1", id)
	err := row.Scan(&role.Id, &role.Name, &role.CanEditProjects, &role.CanEditCategories, &role.CanEditUsers, &role.CanEditRoles, &role.CanEditGenres, &role.CanEditAges)
	if err != nil {
		return models.Role{}, err
	}
	return role, nil
}

func (r *RolesRepository) Create(c context.Context, role models.Role) (int, error) {
	var id int
	row := r.db.QueryRow(c, `
        INSERT INTO roles (name, can_edit_projects, can_edit_categories, can_edit_users, can_edit_roles, can_edit_genres, can_edit_ages) 
        VALUES ($1, $2, $3, $4, $5, $6, $7) RETURNING id`,
		role.Name, role.CanEditProjects, role.CanEditCategories, role.CanEditUsers, role.CanEditRoles, role.CanEditGenres, role.CanEditAges)

	err := row.Scan(&id)
	if err != nil {
		return 0, err
	}
	return id, nil
}

func (r *RolesRepository) Update(c context.Context, id int, role models.Role) error {
	_, err := r.db.Exec(c, `
        UPDATE roles SET name=$1, can_edit_projects=$2, can_edit_categories=$3, can_edit_users=$4, can_edit_roles=$5, can_edit_genres=$6, can_edit_ages=$7
        WHERE id=$8`,
		role.Name, role.CanEditProjects, role.CanEditCategories, role.CanEditUsers, role.CanEditRoles, role.CanEditGenres, role.CanEditAges, id)

	if err != nil {
		return err
	}

	return nil
}

func (r *RolesRepository) Delete(c context.Context, id int) error {
	_, err := r.db.Exec(c, "DELETE FROM roles WHERE id=$1", id)
	if err != nil {
		return err
	}
	return nil
}