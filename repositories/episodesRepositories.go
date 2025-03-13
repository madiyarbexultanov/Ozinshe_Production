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

func (r *EpisodesRepository) Create(c context.Context, episode models.Episode) (int, error) {
	var episodeID int
	query := `INSERT INTO episodes (season_id, number, video_url) 
	          VALUES ($1, $2, $3) RETURNING id`
	err := r.db.QueryRow(c, query, episode.SeasonID, episode.Number, episode.VideoURL).Scan(&episodeID)
	if err != nil {
		return 0, err
	}
	return episodeID, nil
}


func (r *EpisodesRepository) FindAllBySeasonID(c context.Context, seasonID int) ([]models.Episode, error) {
	rows, err := r.db.Query(c, `SELECT id, season_id, number, video_url FROM episodes WHERE season_id = $1`, seasonID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var episodes []models.Episode
	for rows.Next() {
		var episode models.Episode
		if err := rows.Scan(&episode.Id, &episode.SeasonID, &episode.Number, &episode.VideoURL); err != nil {
			return nil, err
		}
		episodes = append(episodes, episode)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}
	return episodes, nil
}

func (r *EpisodesRepository) Update(c context.Context, episode models.Episode) error {
	_, err := r.db.Exec(c, `UPDATE episodes SET number = $1, video_url = $2 WHERE id = $3`,
		episode.Number, episode.VideoURL, episode.Id)
	if err != nil {
		return err
	}
	return nil
}

func (r *EpisodesRepository) Delete(c context.Context, episodeID int) error {
	_, err := r.db.Exec(c, `DELETE FROM episodes WHERE id = $1`, episodeID)
	if err != nil {
		return fmt.Errorf("failed to delete episode: %w", err)
	}
	return nil
}