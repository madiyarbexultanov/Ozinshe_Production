package repositories

import (
	"context"
	"ozinshe_production/models"

	"github.com/jackc/pgx/v5/pgxpool"
)

type MediaRepository struct {
    db *pgxpool.Pool
}

func NewMediaRepository(db *pgxpool.Pool) *MediaRepository {
    return &MediaRepository{db: db}
}

// Получение медиафайлов фильма
func (r *MediaRepository) GetMovieMedia(c context.Context, movieID int) (*models.MovieMedia, error) {
    var media models.MovieMedia

    err := r.db.QueryRow(c, "SELECT cover, screenshots FROM movies WHERE id = $1", movieID).
        Scan(&media.Cover, &media.Screenshots)
    if err != nil {
        return nil, err
    }

    return &media, nil
}

// Обновление обложки и скриншотов фильма
func (r *MediaRepository) UpdateMovieMedia(c context.Context, movieID int, cover *string, screenshots []string) error {
    if cover != nil {
        _, err := r.db.Exec(c, "UPDATE movies SET cover = $1 WHERE id = $2", *cover, movieID)
        if err != nil {
            return err
        }
    }

    if len(screenshots) > 0 {
        _, err := r.db.Exec(c, "UPDATE movies SET screenshots = array_cat(screenshots, $1) WHERE id = $2", screenshots, movieID)
        if err != nil {
            return err
        }
    }

    return nil
}

// Обновление отдельного медиафайла (обложка или один скриншот)
func (r *MediaRepository) UpdateSingleMovieMedia(c context.Context, movieID int, mediaType string, filename string) error {
    if mediaType == "cover" {
        _, err := r.db.Exec(c, "UPDATE movies SET cover = $1 WHERE id = $2", filename, movieID)
        return err
    }

    if mediaType == "screenshot" {
        _, err := r.db.Exec(c, "UPDATE movies SET screenshots = array_append(screenshots, $1) WHERE id = $2", filename, movieID)
        return err
    }

    return nil
}

// Удаление обложки или конкретного скриншота фильма
func (r *MediaRepository) DeleteMovieMedia(c context.Context, movieID int, imageURL string) error {
    var media models.MovieMedia

    // Получаем текущие данные
    err := r.db.QueryRow(c, "SELECT cover, screenshots FROM movies WHERE id = $1", movieID).
        Scan(&media.Cover, &media.Screenshots)
    if err != nil {
        return err
    }

    // Проверяем, удаляем обложку или скриншот
    if media.Cover != nil && *media.Cover == imageURL {
        _, err := r.db.Exec(c, "UPDATE movies SET cover = NULL WHERE id = $1", movieID)
        return err
    }

    // Удаляем скриншот из массива
    updatedScreenshots := []string{}
    for _, s := range media.Screenshots {
        if s != imageURL {
            updatedScreenshots = append(updatedScreenshots, s)
        }
    }

    // Обновляем список скриншотов
    _, err = r.db.Exec(c, "UPDATE movies SET screenshots = $1 WHERE id = $2", updatedScreenshots, movieID)
    return err
}
