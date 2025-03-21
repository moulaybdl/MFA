package main

import "net/http"


func (app *applciation) healthCheckHandler(w http.ResponseWriter, r *http.Request){
	w.Write([]byte("this is the health check handler"))
}