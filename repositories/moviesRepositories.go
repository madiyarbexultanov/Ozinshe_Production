package repositories

import (
	"context"
	"fmt"
	"ozinshe_production/models"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type MoviesRepository struct {
	db *pgxpool.Pool
}

func NewMoviesRepository(conn *pgxpool.Pool) *MoviesRepository {
	return &MoviesRepository{db: conn}
}

func (r *MoviesRepository) FindById(c context.Context, id int) (models.Movie, error) {
	sql := `
	SELECT 
	m.id, m.title, m.release_year, m.runtime, m.keywords, m.description, m.director, 
	m.producer,
	COALESCE(m.cover, '') AS cover, 
	COALESCE(m.screenshots, '{}'::TEXT[]) AS screenshots,
	COALESCE(mt.title, '') AS movie_type_title,
	g.id, COALESCE(g.title, '') AS genre_title,
	c.id, COALESCE(c.title, '') AS category_title,
	a.id, COALESCE(a.title, '') AS age_title,
	COALESCE(s.id, 0) AS season_id, COALESCE(s.number, 0) AS season_number, COALESCE(s.movie_id, 0) AS season_movie_id,
	COALESCE(e.id, 0) AS episode_id, COALESCE(e.number, 0) AS episode_number, COALESCE(e.video_url, '') AS episode_video_url, COALESCE(e.season_id, 0) AS episode_season_id
	FROM movies m
	LEFT JOIN movie_types mt ON mt.id = m.movie_type_id 
	LEFT JOIN movie_genres mg ON mg.movie_id = m.id
	LEFT JOIN genres g ON mg.genre_id = g.id
	LEFT JOIN movie_categories mc ON mc.movie_id = m.id
	LEFT JOIN categories c ON mc.category_id = c.id
	LEFT JOIN movie_ages ma ON ma.movie_id = m.id
	LEFT JOIN ages a ON ma.age_id = a.id
	LEFT JOIN seasons s ON s.movie_id = m.id
	LEFT JOIN episodes e ON e.season_id = s.id
	WHERE m.id = $1
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
		var g 	models.Genre
		var c 	models.Category
		var a 	models.Ages
		var s 	models.Season
		var e 	models.Episode
		var mt 	models.MovieType

		err := rows.Scan(
			&movie.Id, &movie.Title, &movie.ReleaseYear, &movie.Runtime, &movie.KeyWords,
			&movie.Description, &movie.Director, &movie.Producer, &movie.Media.Cover, &movie.Media.Screenshots,
			&mt.Title,
			&g.Id, &g.Title,
			&c.Id, &c.Title,
			&a.Id, &a.Title,
			&s.Id, &s.Number, &s.MovieID,
			&e.Id, &e.Number, &e.VideoURL, &e.SeasonID,
		)
		if err != nil {
			return models.Movie{}, err
		}

		movie.MovieType = mt.Title
		movie.MovieTypeId = mt.Id

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

	// Добавляем уникальные жанры, категории и возрастные ограничения
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


func (r *MoviesRepository) FindAll(c context.Context, filters models.Moviesfilters) ([]models.Movie, error) {
	sql := `
	SELECT 
	m.id, m.title, m.description, m.release_year, m.director, m.producer, 
	m.runtime, m.keywords, 
	COALESCE(m.cover, '') AS cover, 
	COALESCE(m.screenshots, '{}'::TEXT[]) AS screenshots, 
	mt.id, COALESCE(mt.title, '') AS movie_type_title,
	g.id, COALESCE(g.title, '') AS genre_title, 
	c.id, COALESCE(c.title, '') AS category_title,
	a.id, COALESCE(a.title, '') AS age_title,
	COALESCE(s.id, 0) AS season_id, COALESCE(s.number, 0) AS season_number, COALESCE(s.movie_id, 0) AS season_movie_id,
	COALESCE(e.id, 0) AS episode_id, COALESCE(e.number, 0) AS episode_number, COALESCE(e.video_url, '') AS episode_video_url, COALESCE(e.season_id, 0) AS episode_season_id
	FROM movies m
	LEFT JOIN movie_types mt ON mt.id = m.movie_type_id 
	LEFT JOIN movie_genres mg ON mg.movie_id = m.id
	LEFT JOIN genres g ON mg.genre_id = g.id
	LEFT JOIN movie_categories mc ON mc.movie_id = m.id
	LEFT JOIN categories c ON mc.category_id = c.id
	LEFT JOIN movie_ages ma ON ma.movie_id = m.id
	LEFT JOIN ages a ON ma.age_id = a.id
	LEFT JOIN seasons s ON s.movie_id = m.id
	LEFT JOIN episodes e ON e.season_id = s.id
	where 1=1
	`

	params := pgx.NamedArgs{}

	if filters.GenreIds != "" {
		sql = fmt.Sprintf("%s and g.id = @genreId", sql)
		params["genreId"] = filters.GenreIds
	}

	if filters.CategoryIds != "" {
		sql = fmt.Sprintf("%s and c.id = @categoryId", sql)
		params["categoryId"] = filters.CategoryIds
	}

	if filters.TypeIds != "" {
		sql = fmt.Sprintf("%s and mt.id = @typeid", sql)
		params["typeid"] = filters.TypeIds
	}

	if filters.AgeIds != "" {
		sql = fmt.Sprintf("%s and a.id = @ageId", sql)
		params["ageId"] = filters.AgeIds
	}
	
	rows, err := r.db.Query(c, sql, params)
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
		var mt models.MovieType

		err := rows.Scan(
			&m.Id, &m.Title, &m.Description, &m.ReleaseYear, &m.Director,
			&m.Producer, &m.Runtime, &m.KeyWords, &m.Media.Cover, &m.Media.Screenshots,
			&mt.Id, &mt.Title,
			&g.Id, &g.Title,
			&c.Id, &c.Title,
			&a.Id, &a.Title,
			&s.Id, &s.Number, &s.MovieID,
			&e.Id, &e.Number, &e.VideoURL, &e.SeasonID,
		)
		
		if err != nil {
			return nil, err
		}

		existingMovie, exists := moviesMap[m.Id]
		if !exists {
			m.MovieType = mt.Title
			m.MovieTypeId = mt.Id
			moviesMap[m.Id] = &m
			existingMovie = &m
		}

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
			if season.Id == s.Id && !containsEpisode(existingMovie.Seasons[i].Episodes, e) {
				existingMovie.Seasons[i].Episodes = append(existingMovie.Seasons[i].Episodes, e)
			}
		}
	}

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
	row := tx.QueryRow(c, `insert into movies(title, release_year, runtime, keywords, description, director, producer, movie_type_id) 
	values($1, $2, $3, $4, $5, $6, $7, $8) returning id`, 
	movie.Title, movie.ReleaseYear, movie.Runtime, movie.KeyWords, movie.Description, movie.Director, movie.Producer, movie.MovieTypeId)

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

func (r *MoviesRepository) Update(c context.Context, movie models.Movie) error {
	tx, err := r.db.Begin(c)
	if err != nil {
		return err
	}

	// Гарантируем Rollback, если ошибка возникнет до Commit
	defer func() {
		if err != nil {
			tx.Rollback(c)
		}
	}()

	// Обновление данных фильма
	_, err = tx.Exec(c, `
		UPDATE movies
		SET title = $1, release_year = $2, runtime = $3, keywords = $4, description = $5, 
			director = $6, producer = $7, movie_type_id = $8
		WHERE id = $9
	`, movie.Title, movie.ReleaseYear, movie.Runtime, movie.KeyWords, movie.Description, 
		movie.Director, movie.Producer, movie.MovieTypeId, movie.Id)
	if err != nil {
		return err
	}

	// Обновление жанров
	_, err = tx.Exec(c, `DELETE FROM movie_genres WHERE movie_id = $1`, movie.Id)
	if err != nil {
		return err
	}
	for _, genre := range movie.Genres {
		_, err = tx.Exec(c, "INSERT INTO movie_genres(movie_id, genre_id) VALUES($1, $2)", movie.Id, genre.Id)
		if err != nil {
			return err
		}
	}

	// Обновление категорий
	_, err = tx.Exec(c, `DELETE FROM movie_categories WHERE movie_id = $1`, movie.Id)
	if err != nil {
		return err
	}
	for _, category := range movie.Categories {
		_, err = tx.Exec(c, "INSERT INTO movie_categories(movie_id, category_id) VALUES($1, $2)", movie.Id, category.Id)
		if err != nil {
			return err
		}
	}

	// Обновление возрастных ограничений
	_, err = tx.Exec(c, `DELETE FROM movie_ages WHERE movie_id = $1`, movie.Id)
	if err != nil {
		return err
	}
	for _, age := range movie.Ages {
		_, err = tx.Exec(c, "INSERT INTO movie_ages(movie_id, age_id) VALUES($1, $2)", movie.Id, age.Id)
		if err != nil {
			return err
		}
	}

	err = tx.Commit(c)
	if err != nil {
		return err
	}

	return nil
}

func (r *MoviesRepository) Delete(c context.Context, movieID int) error {
	tx, err := r.db.Begin(c)
	if err != nil {
		return err
	}

	// Гарантируем Rollback, если ошибка возникнет до Commit
	defer func() {
		if err != nil {
			tx.Rollback(c)
		}
	}()

	// Удаление связей с жанрами, категориями и возрастными ограничениями
	_, err = tx.Exec(c, `DELETE FROM movie_genres WHERE movie_id = $1`, movieID)
	if err != nil {
		return err
	}
	_, err = tx.Exec(c, `DELETE FROM movie_categories WHERE movie_id = $1`, movieID)
	if err != nil {
		return err
	}
	_, err = tx.Exec(c, `DELETE FROM movie_ages WHERE movie_id = $1`, movieID)
	if err != nil {
		return err
	}

	// Удаление самого фильма
	_, err = tx.Exec(c, `DELETE FROM movies WHERE id = $1`, movieID)
	if err != nil {
		return err
	}

	err = tx.Commit(c)
	if err != nil {
		return err
	}

	return nil
}

func (r *MoviesRepository) SearchMovies(c context.Context, query string) ([]models.Movie, error) {
	rows, err := r.db.Query(c, `
        SELECT id, title, release_year, runtime, 
               keywords, description, director, 
               producer, cover, screenshots, movie_type_id
        FROM movies
        WHERE title ILIKE '%' || $1 || '%'
        ORDER BY release_year DESC
    `, query)

	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var movies []models.Movie
	for rows.Next() {
		var movie models.Movie
		if err := rows.Scan(
			&movie.Id, &movie.Title, &movie.ReleaseYear, &movie.Runtime,
			&movie.KeyWords, &movie.Description, &movie.Director,
			&movie.Producer, &movie.Media.Cover, &movie.Media.Screenshots, &movie.MovieTypeId,
		); err != nil {
			return nil, err
		}
		movies = append(movies, movie)
	}
	return movies, nil
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

