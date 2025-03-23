package main

import (
	"net/http"

	"github.com/julienschmidt/httprouter"
)

func (app *application) routes() *httprouter.Router {
	router := httprouter.New()

	router.HandlerFunc(http.MethodGet,"/v1/healthcheck" ,app.healthCheckHandler)

	// user routes:
	router.HandlerFunc(http.MethodPost, "/v1/users", app.createUser)


	return router
}