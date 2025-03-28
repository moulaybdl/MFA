package main

import (
	"errors"
	"net/http"

	"mfa.moulay/internal/data"
	"mfa.moulay/internal/validator"
)

func (app *application) createUser(w http.ResponseWriter, r *http.Request){
	var input struct {
		Name string `json:"name"`
		Email string `json:"email"`
		Password string `json:"password"`
	}

	// decode the struct:
	err := app.readJSON(w, r, &input)
	if err != nil {
		app.badRequestResponse(w , r, err)
		return
	}

	// populate the User struct:
	var user data.User

	user.Name = input.Name
	user.Email = input.Email
	err = user.Password.CreatePassowrd(&input.Password)
	if err != nil {
		app.serverErrorResponse(w, r, err)
		return
	}

	// validate the user struct
	v := validator.New()


	if data.ValidateUser(v, &user); !v.Valid() {
		app.failedValidationResponse(w, r, v.Errors)
		return
	}


	// user model is safe, we can insert it:
	err = app.models.Users.Insert(&user)
	if err != nil {
		switch {
			case errors.Is(err, data.ErrDuplicateEmail):
				v.AddError("email", "a user with this email address already exists")
				app.failedValidationResponse(w, r, v.Errors)
				return
			case errors.Is(err, data.ErrDuplicatePhoneNumber):
				v.AddError("phone number", "a user with this phone number already exists")
				app.failedValidationResponse(w, r, v.Errors)
				return
			default:
				app.serverErrorResponse(w, r, err)
				return
		}
	}

	err = app.writeJSON(w, r, http.StatusCreated, envelope{"user": user}, nil)
	if err != nil {
		app.serverErrorResponse(w, r, err)
		return
	}

	
}


