package main

import (
	"context"
	"database/sql"
	"flag"
	"log"
	"net/http"
	"os"
	"time"

	_ "github.com/lib/pq"
	"github.com/redis/go-redis/v9"
	"mfa.moulay/internal/cache"
	"mfa.moulay/internal/data"
	"mfa.moulay/internal/mailer"
)


type Config struct {
	Port int `json:"port"`
	Env string `json:"evn"`
	Version string `json:"version"`
	db struct {
		dsn string
	}
}

type application struct {
	cfg Config
	logger *log.Logger
	models data.Models
	redisClient *redis.Client
	mailer mailer.Mailer
}


func main() {
	var cfg Config

	// replace later with config.json file
	flag.IntVar(&cfg.Port, "port", 8080, "Server will listen in this port")
	flag.StringVar(&cfg.Env, "env", "developement", "Env")
	flag.StringVar(&cfg.Version, "version", "v1.0", "version")
	flag.Parse()

	// initilize the logger:
	logger := log.New(os.Stdout, "",  log.Ldate | log.Ltime)

	// intilize the connection with the database
	cfg.db.dsn = "postgres://cns_moulay:password@localhost/cns_sec?sslmode=disable"

	// initialize the connection to the database:
	db ,err := openDB(cfg)
	if err != nil {
		logger.Fatal(err)
	}
	defer db.Close()

	logger.Printf("database connection successful")

	// initilize the models:
	models := data.NewModel(db)

	// Initilize the redis cache:
	client, err := cache.NewRedisClient()
	if err != nil {
		logger.Fatal(err)
	}


	// initilize the application struct 
	app := &application {
		cfg: cfg,
		logger: logger,
		models: models,
		redisClient: client,
		mailer: mailer.New("sandbox.smtp.mailtrap.io", 2525, "ce4ff5406e409d", "546853b1550b78", "Moulay <moulay.mohamed.bouabd.elli@ensia.edu.dz>"), //! values are hard coded gonna be replaced mnb3d
	}

	mux := http.NewServeMux()
	mux.HandleFunc("/", app.healthCheckHandler)

	srv := &http.Server{
		// Addr: fmt.Sprintf(":%d", cfg.Port),
		Addr: "0.0.0.0:8080",
		Handler: app.routes(),
		IdleTimeout: time.Minute,
		ReadTimeout: 10 * time.Second,
		WriteTimeout: 30 * time.Second,
	}

	// start the server:
	logger.Printf("starting %s server on %s", cfg.Env, srv.Addr)
	err = srv.ListenAndServe()
	logger.Fatal(err)

}

func openDB(cfg Config) (*sql.DB, error) {
	db, err := sql.Open("postgres", cfg.db.dsn)
	if err != nil {
		return nil, err
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err = db.PingContext(ctx); err != nil {
		return nil, err
	}

	return db, nil

}