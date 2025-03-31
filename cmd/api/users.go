package main

import (
	"errors"
	"log"
	"net/http"

	"mfa.moulay/internal/data"
	"mfa.moulay/internal/tokens"
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

	// generate the OTP:
	otp, err := tokens.GenerateOTP() 
	if err != nil {
		app.serverErrorResponse(w, r, err)
		return
	}

	log.Print("otp generated")

	// Store the activation code:
	err = tokens.SetOTPCache(otp, user.ID, app.redisClient, r)
	if err != nil {
		app.serverErrorResponse(w , r, err)
		return
	}
	log.Print("otp stored")

	// generate the Activation code:
	activation_code, err := tokens.GenerateActivationCode()
	if err != nil {
		app.serverErrorResponse(w, r, err)
		return
	}
	log.Print("activation code generated")

	// Store the activate code
	err = tokens.SetActivationCache(r, user.ID, activation_code, app.redisClient)
	if err != nil {
		app.serverErrorResponse(w, r, err)
		return
	}
	log.Print("activation code stored")

	var templateInput struct {
		data.User
		ActivationCode string
	}

	templateInput.User = user
	templateInput.ActivationCode = activation_code

	// send the activation code email
	err = app.mailer.Send(user.Email, "user_welcome.tmpl", templateInput)
	if err != nil {
		app.serverErrorResponse(w, r, err)
		return
	}


	// send response:
	var output struct{
		data.User 
		Otp string `json:"otp"`
	}

	output.User = user
	output.Otp = otp

	
	err = app.writeJSON(w, r, http.StatusCreated, envelope{"user": output}, nil)
	if err != nil {
		app.serverErrorResponse(w, r, err)
		return
	}

	
}

func (app *application) verifyOTP(w http.ResponseWriter, r *http.Request) {
	var input struct {
		UserID int `json:"user_id"`
		OTP string `json:"otp"`
		ActivationCode string `json:"activation_code"`
	}

	// decode the struct:
	err := app.readJSON(w, r, &input)
	if err != nil {
		app.serverErrorResponse(w, r, err)
		return
	}

	//! change this operation so that it is done a middleware
	// validate the otp code
	otp_cache, err := tokens.GetOTPCache(input.UserID, app.redisClient, r)
	if err != nil {
		app.serverErrorResponse(w, r, err)
		return
	}
	log.Print("otp code retrieved")

	log.Printf("otp from cache: ", otp_cache)
	log.Printf("otp from user: ", input.OTP)


	err = tokens.VerifyOTPMatch(otp_cache, input.OTP)
	if err != nil {
		app.failedValidationResponse(w, r, map[string]string{
			"error": "incorrect otp code",
		})
		return
	}
	log.Print("otp code passed validation")

	// validate the activation code:
	activation_cache, err := tokens.GetActivationCode(r, input.UserID, app.redisClient)
	if err != nil {
		app.serverErrorResponse(w, r, err)
		return
	}
	log.Print("activation code retrieved")

	err = tokens.VerifyActivationCode(activation_cache, input.ActivationCode)
	if err != nil {
		app.failedValidationResponse(w, r, map[string]string{
			"error": "incorrect activation code",
		})
		return
	}
	log.Print("activation code verified")

	// alter the state in the user table
	err = app.models.Users.ChangeOTPSate(input.UserID)
	if err != nil {
		app.serverErrorResponse(w, r, err)
		return
	}

	// write response:
	err = app.writeJSON(w, r, http.StatusOK, envelope{"state":"sucess"}, nil)
	if err != nil {
		app.serverErrorResponse(w, r, err)
		return
	}
}


