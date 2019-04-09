package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"strings"

	"github.com/0x113/x-media/auth"
	"github.com/0x113/x-media/database/mysql"
	"github.com/0x113/x-media/env"
	"github.com/0x113/x-media/video"
	jwt "github.com/dgrijalva/jwt-go"
	_ "github.com/go-sql-driver/mysql"
	"github.com/gorilla/mux"
	log "github.com/sirupsen/logrus"
)

func init() {
	log.SetFormatter(&log.TextFormatter{
		TimestampFormat: "2006-01-02 15:04:05.000",
		FullTimestamp:   true,
	})
}

func main() {
	/* Get env variables */
	dbUser := os.Getenv("db_user")
	dbPass := os.Getenv("db_pass")
	dbHost := os.Getenv("db_host")
	dbPort := os.Getenv("db_port")
	dbName := os.Getenv("db_name")
	jwtSecret := os.Getenv("jwt_secret")

	/* Check env variables */
	if dbUser == "" || dbPass == "" || dbHost == "" || dbPort == "" || dbName == "" || jwtSecret == "" || env.EnvString("video_dir") == "" {
		log.Error("Environment variables can not be empty.")
		log.Println("List of environment variables: ")
		log.Printf("db_user: %s\n", dbUser)
		log.Printf("db_pass: %s\n", dbPass)
		log.Printf("db_host: %s\n", dbHost)
		log.Printf("db_port: %s\n", dbPort)
		log.Printf("db_name: %s\n", dbName)
		log.Printf("jwt_secret: %s", jwtSecret)
		log.Printf("video_dir: %s", env.EnvString("video_dir"))
		os.Exit(0)
	}

	conn := mysqlConnection(dbUser, dbPass, dbHost, dbPort, dbName)
	defer conn.Close()

	/* authentication */
	authRepo := mysql.NewMySQLAuthRepository(conn)
	authService := auth.NewAuthService(authRepo)
	authHandler := auth.NewAuthHandler(authService)

	/* video */
	videoRepo := mysql.NewMySQLVideoRepository(conn)
	videoService := video.NewVideoService(videoRepo)
	videoHandler := video.NewVideoHandler(videoService)

	router := mux.NewRouter().StrictSlash(true)

	/* authentication routes */
	router.HandleFunc("/user/create", authHandler.Create).Methods("POST", "GET")
	router.HandleFunc("/user/token/generate", authHandler.GenerateJWT).Methods("POST")

	/* video routes */
	router.HandleFunc("/api/movies/update", videoHandler.UpdateMovies).Methods("GET")
	router.HandleFunc("/api/movies", videoHandler.AllMovies).Methods("GET")
	router.HandleFunc("/api/tvseries/update", videoHandler.UpdateTvSeries).Methods("GET")
	router.HandleFunc("/api/tvseries", videoHandler.AllTvSeries).Methods("GET")
	router.HandleFunc("/api/episodes/{name}", videoHandler.AllTvSeriesEpisodes).Methods("GET")

	http.Handle("/", accessControl(router))
	http.Handle("/api/", authRequired(router))

	log.Infoln("Launching server on port :8000")
	if err := http.ListenAndServe(":8000", nil); err != nil {
		log.Errorf("ListenAndServe: %s", err)
	}
}

func mysqlConnection(username, password, host, port, dbname string) *sql.DB {
	sqlStmt := fmt.Sprintf("%s:%s@tcp(%s:%s)/%s", username, password, host, port, dbname)
	db, err := sql.Open("mysql", sqlStmt)
	if err != nil {
		log.Errorf("Error while connection to the database: %s", err)
	}
	return db
}

func accessControl(h http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Origin, Content-Type")

		if r.Method == "OPTIONS" {
			return
		}

		h.ServeHTTP(w, r)
	})
}

func authRequired(h http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		response := make(map[string]string)

		tokenHeader := r.Header.Get("Authorization")
		if tokenHeader == "" {
			response["error"] = "Missing authorization token"
			json.NewEncoder(w).Encode(response)
			return
		}

		splitted := strings.Split(tokenHeader, " ")
		if len(splitted) != 2 {
			response["error"] = "Invalid/Malformed authorization token"
			json.NewEncoder(w).Encode(response)
			return
		}

		token, err := jwt.Parse(splitted[1], func(token *jwt.Token) (interface{}, error) {
			if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
				return nil, nil
			}
			return []byte(env.EnvString("jwt_key")), nil
		})
		if err != nil {
			response["error"] = err.Error()
			json.NewEncoder(w).Encode(response)
			return
		}

		if !token.Valid {
			response["error"] = "Invalid authorization token"
			json.NewEncoder(w).Encode(response)
			return
		}

		h.ServeHTTP(w, r)
	})
}
