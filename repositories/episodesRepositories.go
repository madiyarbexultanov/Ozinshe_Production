package repositories

import (
	"context"
	"fmt"
	"ozinshe_production/models"

	"github.com/jackc/pgx/v5/pgxpool"
)

type EpisodesRepository struct {
	db *pgxpool.Pool
}

func NewEpisodesRepository(conn *pgxpool.Pool) *EpisodesRepository {
	return &EpisodesRepository{db: conn}
}

// Проверка, существует ли эпизод в данном сезоне
func (r *EpisodesRepository) Exists(c context.Context, seasonID, episodeNumber int) (bool, error) {
	var exists bool
	query := `SELECT EXISTS(SELECT 1 FROM episodes WHERE season_id=$1 AND number=$2)`
	err := r.db.QueryRow(c, query, seasonID, episodeNumber).Scan(&exists)
	if err != nil {
		return false, fmt.Errorf("failed to check if episode exists: %w", err)
	}
	return exists, nil
}

func (r *EpisodesRepository) FindById(c context.Context, id int) (models.Episode, error) {
	var episode models.Episode
	row := r.db.QueryRow(c, "SELECT id, number, season_id, video_url FROM episodes WHERE id = $1", id)
	err := row.Scan(&episode.Id, &episode.Number, &episode.SeasonID, &episode.VideoURL)
	if err != nil {
		return models.Episode{}, err
	}
	return episode, nil
}

// Создание нового эпизода (с проверкой наличия)
func (r *EpisodesRepository) Create(c context.Context, episode models.Episode) (int, error) {
	exists, err := r.Exists(c, episode.SeasonID, episode.Number)
	if err != nil {
		return 0, err
	}
	if exists {
		return 0, fmt.Errorf("episode already exists in season %d", episode.SeasonID)
	}

	var episodeID int
	query := `INSERT INTO episodes (season_id, number, video_url) VALUES ($1, $2, $3) RETURNING id`
	err = r.db.QueryRow(c, query, episode.SeasonID, episode.Number, episode.VideoURL).Scan(&episodeID)
	if err != nil {
		return 0, fmt.Errorf("failed to create episode: %w", err)
	}
	return episodeID, nil
}

// Получение всех эпизодов сезона
func (r *EpisodesRepository) FindAllBySeasonID(c context.Context, seasonID int) ([]models.Episode, error) {
	rows, err := r.db.Query(c, `SELECT id, season_id, number, video_url FROM episodes WHERE season_id = $1`, seasonID)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch episodes: %w", err)
	}
	defer rows.Close()

	var episodes []models.Episode
	for rows.Next() {
		var episode models.Episode
		if err := rows.Scan(&episode.Id, &episode.SeasonID, &episode.Number, &episode.VideoURL); err != nil {
			return nil, fmt.Errorf("failed to scan episode: %w", err)
		}
		episodes = append(episodes, episode)
	}
	return episodes, rows.Err()
}

// Обновление эпизода
func (r *EpisodesRepository) Update(c context.Context, episode models.Episode) error {
	_, err := r.db.Exec(c, `UPDATE episodes SET number = $1, video_url = $2 WHERE id = $3`,
		episode.Number, episode.VideoURL, episode.Id)
	return err
}

// Удаление эпизода
func (r *EpisodesRepository) Delete(c context.Context, episodeID int) error {
	_, err := r.db.Exec(c, `DELETE FROM episodes WHERE id = $1`, episodeID)
	return err
}