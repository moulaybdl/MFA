package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"strings"

	"github.com/julienschmidt/httprouter"
)

type envelope map[string]interface{}


func (app *application) readIDParam(r *http.Request) (int64, error){
	params := httprouter.ParamsFromContext(r.Context())

	id, err := strconv.ParseInt(params.ByName("id"), 10, 64)
	if err != nil || id < 1 {
		return 0, err
	}

	return id, nil
}

func (app *application) writeJSON(w http.ResponseWriter, r *http.Request, status int, data envelope, headers http.Header) error {
	js, err := json.Marshal(data)
	if err != nil {
		return nil
	}

	js = append(js, '\n')

	for key, value := range headers {
		w.Header()[key] = value
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	w.Write(js)

	return nil
}

func (app *application) readJSON(w http.ResponseWriter, r *http.Request, dst interface{}) error{
	err := json.NewDecoder(r.Body).Decode(&dst)
	if err != nil {
		var sytaxError *json.SyntaxError
		var unmarshalTypeError *json.UnmarshalTypeError
		var invalidUnmarshalError *json.InvalidUnmarshalError

		switch {
		case errors.As(err, &sytaxError):
			return fmt.Errorf("body contains badly-formed JSON (at character %d)", sytaxError.Offset)
		

		case errors.Is(err, io.ErrUnexpectedEOF):
			return errors.New("body contains badly-formed JSON")

		case errors.As(err, &unmarshalTypeError):
			return fmt.Errorf("body contains an invalid value for the %q field (at character %d)", unmarshalTypeError.Field, unmarshalTypeError.Offset)

		case errors.Is(err, io.EOF):
			return errors.New("body must not be empty")

		case strings.HasPrefix(err.Error(), "json: unknown field "):
			fieldName := strings.TrimPrefix(err.Error(), "json: unknown field ")
			return fmt.Errorf("body contains unknown key %s", fieldName)

		// uncomment this to set maxBytes for the request body to prevent DoS attacks
		// case err.Error() == "http: request body too large":
		// 	return fmt.Errorf("body must not be larger than %d bytes", maxBytes)
			

		case errors.As(err, &invalidUnmarshalError):
			panic(err)

		default:
			return err
	}

}
return nil
}