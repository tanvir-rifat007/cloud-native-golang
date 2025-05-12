package main

import (
	"canvas/internal/data"
	"canvas/messaging"
	"canvas/validator"
	"fmt"
	"net/http"
)

func (app *application) NewsletterSignup(w http.ResponseWriter, r *http.Request){

	var input struct{
		Email string `json:"email"`
	}

	err:= app.readJSON(w,r,&input)

	if err!=nil{
		app.badRequestResponse(w,r,err)
		return
	}

	// validate the email:
	v:= validator.New()
	data.ValidateEmail(v,input.Email)
	if !v.Valid(){
		app.failedValidationResponse(w,r,v.Errors)

		return
	}

	token,err:=app.models.NewsletterSubscribers.Insert(input.Email)

	fmt.Println("token:",token,"email:",input.Email)

	if err!=nil{
		app.serverErrorResponse(w,r,err)
		return
	}
	// send the email to the queue
	err= app.queue.Send(r.Context(),messaging.Message{
		"job":   "confirmation_email",
		"email": input.Email,
		"token": token,

	})


	

	if err!=nil{
		app.serverErrorResponse(w,r,err)
		return
	}

	data:= envelope{
		"message": "you will receive a confirmation email shortly",
		"token":   token,
		"email":  input.Email,
	}
	

	err = app.writeJSON(w,http.StatusCreated,data,nil)

	if err!=nil{
		app.serverErrorResponse(w,r,err)
		return
	}

}


// newsletter confirmation handler

func (app *application) NewsletterConfirmation(w http.ResponseWriter, r *http.Request){

	tokenPlainText := r.URL.Query().Get("token")

	
	// validate the token
	v := validator.New()

	data.ValidateTokenPlaintext(v, tokenPlainText)
	if !v.Valid() {
		app.failedValidationResponse(w, r, v.Errors)
		return
	}

	// confirm the token
	email,err := app.models.NewsletterSubscribers.Confirm(tokenPlainText)
	if err != nil {
		if err == data.ErrRecordNotFound {
			app.notFoundResponse(w, r)
			return
		}
		app.serverErrorResponse(w, r, err)
		return
	}
	// send the email to the queue


	err = app.queue.Send(r.Context(), messaging.Message{
		"job":   "welcome_email",
		"email":email,
	})

	if err != nil {
		app.serverErrorResponse(w, r, err)
		return
	}
	

	redirectURL := fmt.Sprintf("http://localhost:8080/activated?token=%s", tokenPlainText)
http.Redirect(w, r, redirectURL, http.StatusSeeOther)

	


}