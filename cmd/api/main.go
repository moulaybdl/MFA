package main

import (
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
	"time"
)


type Config struct {
	Port int `json:"port"`
	Env string `json:"evn"`
	Version string `json:"version"`
}

type applciation struct {
	cfg Config
	logger *log.Logger
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


	// initilize the application struct 
	app := &applciation {
		cfg: cfg,
		logger: logger,
	}

	mux := http.NewServeMux()
	mux.HandleFunc("/", app.healthCheckHandler)

	srv := &http.Server{
		Addr: fmt.Sprintf(":%d", cfg.Port),
		Handler: mux,
		IdleTimeout: time.Minute,
		ReadTimeout: 10 * time.Second,
		WriteTimeout: 30 * time.Second,
	}

	// start the server:
	logger.Printf("starting %s server on %s", cfg.Env, srv.Addr)
	err := srv.ListenAndServe()
	logger.Fatal(err)

}