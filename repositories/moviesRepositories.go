package repositories

import (
	"context"
	"ozinshe_production/models"

	"github.com/jackc/pgx/v5/pgxpool"
)

type MoviesRepository struct {
	db *pgxpool.Pool
}

func NewMoviesRepository(conn *pgxpool.Pool) *MoviesRepository {
	return &MoviesRepository{db: conn}
}

func (r *MoviesRepository) FindById(c context.Context, id int) (models.Movie, error){
	sql := `
	select 
	m.id, m.title, m.release_year, m.runtime, m.keywords, m.description, m.director, 
	m.producer,
	COALESCE(m.cover, '') AS cover, 
    COALESCE(m.screenshots, '{}'::TEXT[]) AS screenshots, 
	g.id, g.title,
	c.id, c.title,
	a.id, a.title,
	s.id, s.number, s.movie_id,
	e.id, e.number, e.video_url, e.season_id
	from movies m
	LEFT JOIN movie_genres mg ON mg.movie_id = m.id
	LEFT JOIN genres g ON mg.genre_id = g.id
	LEFT JOIN movie_categories mc ON mc.movie_id = m.id
	LEFT JOIN categories c ON mc.category_id = c.id
	LEFT JOIN movie_ages ma ON ma.movie_id = m.id
	LEFT JOIN ages a ON ma.age_id = a.id
	LEFT JOIN seasons s ON s.movie_id = m.id
	LEFT  JOIN episodes e ON e.season_id = s.id
	where m.id = $1
	`

	rows, err := r.db.Query(c, sql, id)
	if err != nil {
		return models.Movie{}, err
	}
	defer rows.Close()

	var movie models.Movie
	genresMap := make(map[int]models.Genre)
	categoriesMap := make(map[int]models.Category)
	agesMap := make(map[int]models.Ages)
	seasonsMap := make(map[int]models.Season)
	seasonEpisodesMap := make(map[int][]models.Episode)

	for rows.Next() {
		var g models.Genre
		var c models.Category
		var a models.Ages
		var s models.Season
		var e models.Episode

		err := rows.Scan(
			&movie.Id, &movie.Title, &movie.ReleaseYear, &movie.Runtime, &movie.KeyWords, 
			&movie.Description, &movie.Director, &movie.Producer, &movie.Cover, &movie.Screenshots,
			&g.Id, &g.Title,
			&c.Id, &c.Title,
			&a.Id, &a.Title,
			&s.Id, &s.Number, &s.MovieID,
			&e.Id, &e.Number, &e.VideoURL, &e.SeasonID,
		)
		if err != nil {
			return models.Movie{}, err
		}

		if g.Id != 0 {
			genresMap[g.Id] = g
		}
		if c.Id != 0 {
			categoriesMap[c.Id] = c
		}
		if a.Id != 0 {
			agesMap[a.Id] = a
		}
		if s.Id != 0 {
			if _, exists := seasonsMap[s.Id]; !exists {
				seasonsMap[s.Id] = s
			}
		}
		if e.Id != 0 {
			exists := false
			for _, ep := range seasonEpisodesMap[s.Id] {
				if ep.Id == e.Id {
					exists = true
					break
				}
			}
			if !exists {
				seasonEpisodesMap[s.Id] = append(seasonEpisodesMap[s.Id], e)
			}
		}
	}

	// Добавляем уникальные жанры, категории и возрасты
	for _, genre := range genresMap {
		movie.Genres = append(movie.Genres, genre)
	}
	for _, category := range categoriesMap {
		movie.Categories = append(movie.Categories, category)
	}
	for _, age := range agesMap {
		movie.Ages = append(movie.Ages, age)
	}
	// Добавляем сезоны и их эпизоды
	for _, season := range seasonsMap {
		season.Episodes = seasonEpisodesMap[season.Id]
		movie.Seasons = append(movie.Seasons, season)
	}

	if err = rows.Err(); err != nil {
		return models.Movie{}, err
	}

	return movie, nil
}

func (r *MoviesRepository) FindAll(c context.Context) ([]models.Movie, error) {
	sql := `
	SELECT 
	m.id, m.title, m.description, m.release_year, m.director, m.producer, 
	m.runtime, m.keywords, 
	COALESCE(m.cover, '') AS cover, 
    COALESCE(m.screenshots, '{}'::TEXT[]) AS screenshots, 
	g.id, g.title, 
	c.id, c.title,
	a.id, a.title,
	s.id, s.number, s.movie_id,
	e.id, e.number, e.video_url, e.season_id
	FROM movies m
	LEFT JOIN movie_genres mg ON mg.movie_id = m.id
	LEFT JOIN genres g ON mg.genre_id = g.id
	LEFT JOIN movie_categories mc ON mc.movie_id = m.id
	LEFT JOIN categories c ON mc.category_id = c.id
	LEFT JOIN movie_ages ma ON ma.movie_id = m.id
	LEFT JOIN ages a ON ma.age_id = a.id
	LEFT JOIN seasons s ON s.movie_id = m.id
	LEFT JOIN episodes e ON e.season_id = s.id
	`

	rows, err := r.db.Query(c, sql)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	
	moviesMap := make(map[int]*models.Movie)

	
	for rows.Next() {
		var m models.Movie
		var g models.Genre
		var c models.Category
		var a models.Ages
		var s models.Season
		var e models.Episode
	
		err := rows.Scan(
			&m.Id, &m.Title, &m.Description, &m.ReleaseYear, &m.Director, 
			&m.Producer, &m.Runtime, &m.KeyWords, &m.Cover, &m.Screenshots,
			&g.Id, &g.Title,
			&c.Id, &c.Title,
			&a.Id, &a.Title,
			&s.Id, &s.Number, &s.MovieID,
			&e.Id, &e.Number, &e.VideoURL,  &e.SeasonID,
		)
		if err != nil {
			return nil, err
		}
	
		// Проверяем, есть ли фильм в map
		existingMovie, exists := moviesMap[m.Id]
		if !exists {
			// Создаем новый фильм в map
			moviesMap[m.Id] = &m
			existingMovie = &m
		}
	
		// Добавляем жанры, категории и возрастные ограничения, избегая дубликатов
		if g.Id != 0 && !containsGenre(existingMovie.Genres, g) {
			existingMovie.Genres = append(existingMovie.Genres, g)
		}
		if c.Id != 0 && !containsCategory(existingMovie.Categories, c) {
			existingMovie.Categories = append(existingMovie.Categories, c)
		}
		if a.Id != 0 && !containsAge(existingMovie.Ages, a) {
			existingMovie.Ages = append(existingMovie.Ages, a)
		}
		if s.Id != 0 && !containsSeason(existingMovie.Seasons, s) {
			existingMovie.Seasons = append(existingMovie.Seasons, s)
		}
		for i, season := range existingMovie.Seasons {
			if season.Id == s.Id {
				// Проверяем, есть ли уже эпизод в сезоне
				if !containsEpisode(existingMovie.Seasons[i].Episodes, e) {
					existingMovie.Seasons[i].Episodes = append(existingMovie.Seasons[i].Episodes, e)
				}
				break
			}
		}
		
	}
	
	if err = rows.Err(); err != nil {
		return nil, err
	}
	
	// Собираем итоговый список фильмов
	movies := make([]models.Movie, 0, len(moviesMap))
	for _, movie := range moviesMap {
		movies = append(movies, *movie)
	}
	
	
	return movies, nil
}

func (r *MoviesRepository) Create(c context.Context, movie models.Movie) (int, error){
	tx, err := r.db.Begin(c)
	if err != nil {
		return 0, err
	}

	// Гарантируем Rollback, если ошибка возникнет до Commit
	defer func() {
		if err != nil {
			tx.Rollback(c)
		}
	}()

	var id int
	row := tx.QueryRow(c, `insert into movies(title, release_year, runtime, keywords, description, director, producer) 
	values($1, $2, $3, $4, $5, $6, $7) returning id`, 
	movie.Title, movie.ReleaseYear, movie.Runtime, movie.KeyWords, movie.Description, movie.Director, movie.Producer)

	err = row.Scan(&id)
	if err != nil {
		return 0, err
	}

	for _, genre := range movie.Genres {
		_, err = tx.Exec(c, "insert into movie_genres(movie_id, genre_id) values($1, $2)", id, genre.Id)
		if err != nil {
			return 0, err
		}
	}
	
	for _, category := range movie.Categories {
		_, err = tx.Exec(c, "insert into movie_categories(movie_id, category_id) values($1, $2)", id, category.Id)
		if err != nil {
			return 0, err
		}
	}
	
	for _, age := range movie.Ages {
		_, err = tx.Exec(c, "insert into movie_ages(movie_id, age_id) values($1, $2)", id, age.Id)
		if err != nil {
			return 0, err
		}
	}
	
	err = tx.Commit(c)
	if err != nil {
		return 0, err
	}

	return id, nil
}

func (r *MoviesRepository) UpdateCoverAndScreenshots(c context.Context, movieID int, cover string, screenshots []string) error {
	// Обновление Cover
	_, err := r.db.Exec(c, `
		UPDATE movies 
		SET cover = $1 
		WHERE id = $2
	`, cover, movieID)
	if err != nil {
		return err
	}

	// Обновление Screenshots (если они переданы)
	if len(screenshots) > 0 {
		// Для простоты, перезаписываем все screenshots
		_, err := r.db.Exec(c, `
			UPDATE movies 
			SET screenshots = $1 
			WHERE id = $2
		`, screenshots, movieID)
		if err != nil {
			return err
		}
	}
	return nil
}



func containsGenre(genres []models.Genre, g models.Genre) bool {
	for _, genre := range genres {
		if genre.Id == g.Id {
			return true
		}
	}
	return false
}

func containsCategory(categories []models.Category, c models.Category) bool {
	for _, category := range categories {
		if category.Id == c.Id {
			return true
		}
	}
	return false
}

func containsAge(ages []models.Ages, a models.Ages) bool {
	for _, age := range ages {
		if age.Id == a.Id {
			return true
		}
	}
	return false
}

func containsSeason(seasons []models.Season, s models.Season) bool {
	for _, season := range seasons {
		if season.Id == s.Id {
			return true
		}
	}
	return false
}

func containsEpisode(episodes []models.Episode, e models.Episode) bool {
	for _, episode := range episodes {
		if episode.Id == e.Id {
			return true
		}
	}
	return false
}