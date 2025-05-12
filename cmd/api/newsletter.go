package main

import (
	"canvas/internal/data"
	"canvas/messaging"
	"canvas/validator"
	"net/http"
)


func (app *application) createNewsletterHandler(w http.ResponseWriter, r *http.Request){
	var input struct {
		Title string `json:"title"`
		Body  string `json:"body"`
		Tags  []string `json:"tags"`
	}

	err := app.readJSON(w, r, &input)
	if err != nil {
		app.badRequestResponse(w, r, err)
		return
	}

	v := validator.New()
	data.ValidateNewsletter(v, input.Title, input.Body)
	if !v.Valid() {
		app.failedValidationResponse(w, r, v.Errors)
		return
	}

	adminSubscriber,err:=app.models.NewsletterSubscribers.GetAdminSubscriber()

	if err!=nil{
		app.serverErrorResponse(w,r,err)
		return
		
	}

	

	newsletter, err := app.models.Newsletter.Insert(input.Title, input.Body, input.Tags, adminSubscriber.ID)
	if err != nil {
		app.serverErrorResponse(w, r, err)
		return
	}

	// get the active and confirmed subscribers
	subscribers, err := app.models.NewsletterSubscribers.GetNewsletterSubscribers()
	if err != nil {
		app.serverErrorResponse(w, r, err)
		return
	}
	// send the newsletter to the subscribers using the messaging queue

	for _, subscriber := range subscribers {
		err = app.queue.Send(r.Context(),messaging.Message{
			"job":   "newsletter_email",
			"id":    newsletter.ID,
			"email": subscriber.Email,

		})

		if err != nil {
			app.serverErrorResponse(w, r, err)
			return
		}
	}

	data := envelope{
		"newsletter": newsletter,
	}

	err = app.writeJSON(w, http.StatusCreated, data, nil)
	if err != nil {
		app.serverErrorResponse(w, r, err)
	}
}


func (app *application) getNewletterByIdHandler(w http.ResponseWriter, r *http.Request){
	id, err := app.readIDParam(r)
	if err != nil {
		app.badRequestResponse(w, r, err)
		return
	}

	newsletter, err := app.models.Newsletter.GetNewsletter(id)
	if err != nil {
		if err == data.ErrRecordNotFound {
			app.notFoundResponse(w, r)
		} else {
			app.serverErrorResponse(w, r, err)
		}
		return
	}

	data := envelope{
		"newsletter": newsletter,
	}

	err = app.writeJSON(w, http.StatusOK, data, nil)
	if err != nil {
		app.serverErrorResponse(w, r, err)
	}
}

func (app *application) getNewslettersHandler(w http.ResponseWriter, r *http.Request){
	newsletters, err := app.models.Newsletter.GetNewsletters()
	if err != nil {
		app.serverErrorResponse(w, r, err)
		return
	}

	data := envelope{
		"newsletters": newsletters,
	}

	err = app.writeJSON(w, http.StatusOK, data, nil)
	if err != nil {
		app.serverErrorResponse(w, r, err)
	}
}

func (app *application) searchNewsletterHandler(w http.ResponseWriter,r *http.Request){
	query := r.URL.Query().Get("query")
	if query == "" {
		app.badRequestResponse(w, r, nil)
		return
	}

	newsletters, err := app.models.Newsletter.SearchNewsletter(query)
	if err != nil {
		app.serverErrorResponse(w, r, err)
		return
	}

	data := envelope{
		"newsletters": newsletters,
	}

	err = app.writeJSON(w, http.StatusOK, data, nil)
	if err != nil {
		app.serverErrorResponse(w, r, err)
	}
}