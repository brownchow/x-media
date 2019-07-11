package mysql

import (
	"database/sql"
	"encoding/json"

	"github.com/0x113/x-media/video"
	log "github.com/sirupsen/logrus"
)

type videoRepository struct {
	db *sql.DB
}

func NewMySQLVideoRepository(db *sql.DB) video.VideoRepository {
	return &videoRepository{
		db,
	}
}

func (r *videoRepository) SaveMovie(movie *video.Movie) error {
	query := "INSERT INTO movie (title, description, director, genre, duration, rate, release_date, file_name, poster_path, cast) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?) ON DUPLICATE KEY UPDATE title=?, description=?, director=?, genre=?, duration=?, rate=?, release_date=?, file_name=?, poster_path=?, cast=?"
	stmt, err := r.db.Prepare(query)
	if err != nil {
		log.Errorf("Error while preparing statement: %s", err.Error())
		return err
	}

	// convert cast to JSON and then to string TODO: change it
	jsonBytes, err := json.Marshal(movie.Cast)
	if err != nil {
		log.Errorf("Error while converting struct to JSON: %s", err.Error())
		return err
	}
	movieCastString := string(jsonBytes)

	_, err = stmt.Exec(movie.Title, movie.Description, movie.Director, movie.Genre, movie.Duration, movie.Rate, movie.ReleaseDate, movie.FileName, movie.PosterPath, movieCastString, movie.Title, movie.Description, movie.Director, movie.Genre, movie.Duration, movie.Rate, movie.ReleaseDate, movie.FileName, movie.PosterPath, movieCastString)
	if err != nil {
		log.Errorf("Error while executing statement: %s", err.Error())
		return err
	}
	log.Infof("Updated movie with title: %s", movie.Title)
	return nil

}

func (r *videoRepository) FindAllMovies() ([]*video.Movie, error) {
	rows, err := r.db.Query("SELECT * FROM movie")
	if err != nil {
		log.Errorf("Error while selecting all movies: %s", err.Error())
		return nil, err
	}
	defer rows.Close()

	var movies []*video.Movie
	for rows.Next() {
		movie := new(video.Movie)
		var movieCastString string
		if err := rows.Scan(&movie.MovieID, &movie.Title, &movie.Description, &movie.Director, &movie.Genre, &movie.Duration, &movie.Rate, &movie.ReleaseDate, &movie.FileName, &movie.PosterPath, &movieCastString); err != nil {
			log.Errorf("Error while scanning for movie: %s", err.Error())
			return nil, err
		}
		err := json.Unmarshal([]byte(movieCastString), &movie.Cast)
		if err != nil {
			log.Errorf("Error while unmarshalling struct: %s", err.Error())
			return nil, err
		}
		movies = append(movies, movie)
	}

	return movies, nil
}

func (r *videoRepository) SaveTvSeries(tvSeries *video.TVSeries) error {
	query := "INSERT INTO series (title, description, director, genre, episode_duration, rate, release_date, dir_name, poster_path) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?) ON DUPLICATE KEY UPDATE title=?, description=?, director=?, genre=?, episode_duration=?, rate=?, release_date=?, dir_name=?, poster_path=?"
	stmt, err := r.db.Prepare(query)
	if err != nil {
		log.Errorf("Error while preparing statement: %s", err.Error())
		return err
	}

	_, err = stmt.Exec(tvSeries.Title, tvSeries.Description, tvSeries.Director, tvSeries.Genre, tvSeries.EpisodeDuration, tvSeries.Rate, tvSeries.ReleaseDate, tvSeries.DirName, tvSeries.PosterPath, tvSeries.Title, tvSeries.Description, tvSeries.Director, tvSeries.Genre, tvSeries.EpisodeDuration, tvSeries.Rate, tvSeries.ReleaseDate, tvSeries.DirName, tvSeries.PosterPath)
	if err != nil {
		log.Errorf("Error while executing statement: %s", err.Error())
		return err
	}
	log.Infof("Updated serial with title: %s", tvSeries.Title)
	return nil
}

func (r *videoRepository) FindAllTvSeries() ([]*video.TVSeries, error) {
	rows, err := r.db.Query("SELECT * FROM series")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var tvSeries []*video.TVSeries
	for rows.Next() {
		series := new(video.TVSeries)
		if err := rows.Scan(&series.SeriesID, &series.Title, &series.Description, &series.Director, &series.Genre, &series.EpisodeDuration, &series.Rate, &series.ReleaseDate, &series.DirName, &series.PosterPath); err != nil {
			return nil, err
		}
		tvSeries = append(tvSeries, series)
	}

	return tvSeries, nil
}
