package main

import (
	"canvas/internal/data"
	"canvas/messaging"
	"canvas/validator"
	"fmt"
	"net/http"
	"strconv"
)

// func (app *application) createNewsletterHandler(w http.ResponseWriter, r *http.Request){
// 	var input struct {
// 		Title string `json:"title"`
// 		Body  string `json:"body"`
// 		Tags  []string `json:"tags"`
// 		Token string `json:"token"`
// 	}

// 	err := app.readJSON(w, r, &input)
// 	if err != nil {
// 		app.badRequestResponse(w, r, err)
// 		return
// 	}

// 	fmt.Println("input",input)

// 	v := validator.New()
// 	data.ValidateNewsletter(v, input.Title, input.Body)
// 	if !v.Valid() {
// 		app.failedValidationResponse(w, r, v.Errors)
// 		return
// 	}

// 	adminSubscriber,err:=app.models.NewsletterSubscribers.GetAdminSubscriber(input.Token)

// 	if err!=nil{
// 		app.serverErrorResponse(w,r,err)
// 		return

// 	}

// 	newsletter, err := app.models.Newsletter.Insert(input.Title, input.Body, input.Tags, adminSubscriber.ID)
// 	if err != nil {
// 		app.serverErrorResponse(w, r, err)
// 		return
// 	}

// 	// get the active and confirmed subscribers
// 	subscribers, err := app.models.NewsletterSubscribers.GetNewsletterSubscribers()
// 	if err != nil {
// 		app.serverErrorResponse(w, r, err)
// 		return
// 	}
// 	// send the newsletter to the subscribers using the messaging queue

// 	for _, subscriber := range subscribers {
// 		err = app.queue.Send(r.Context(),messaging.Message{
// 			"job":   "newsletter_email",
// 			"id":    newsletter.ID,
// 			"email": subscriber.Email,

// 		})

// 		if err != nil {
// 			app.serverErrorResponse(w, r, err)
// 			return
// 		}
// 	}

// 	data := envelope{
// 		"newsletter": newsletter,
// 	}

// 	err = app.writeJSON(w, http.StatusCreated, data, nil)
// 	if err != nil {
// 		app.serverErrorResponse(w, r, err)
// 	}
// }


func (app *application) createNewsletterHandler(w http.ResponseWriter, r *http.Request) {
	// Parse multipart form (max 20MB)
	err := r.ParseMultipartForm(20 << 20)
	if err != nil {
		app.badRequestResponse(w, r, fmt.Errorf("failed to parse form: %w", err))
		return
	}

	title := r.FormValue("title")
	body := r.FormValue("body")
	token := r.FormValue("token")
	tags := r.MultipartForm.Value["tags[]"] // because array inputs are serialized like tags[]=a&tags[]=b

	v := validator.New()
	data.ValidateNewsletter(v, title, body)
	if !v.Valid() {
		app.failedValidationResponse(w, r, v.Errors)
		return
	}

	adminSubscriber, err := app.models.NewsletterSubscribers.GetAdminSubscriber(token)
	if err != nil {
		app.serverErrorResponse(w, r, err)
		return
	}

	newsletter, err := app.models.Newsletter.Insert(title, body, tags, adminSubscriber.ID)
	if err != nil {
		app.serverErrorResponse(w, r, err)
		return
	}

	// Optional file upload
	file, header, err := r.FormFile("file")
	if err == nil {
		defer file.Close()


		// convert newsletter id string to int
		newsletterId,err:=strconv.Atoi(newsletter.ID)
		if err!=nil{
			app.serverErrorResponse(w, r, err)
			return
		}

		// Generate a key for the file
		key := fmt.Sprintf("newsletters/%d/%s",newsletterId, header.Filename)

		// Upload to S3
		err = app.blobstore.Put(r.Context(), app.blobstore.Bucket(), key, header.Header.Get("Content-Type"), file)
		if err != nil {
			app.serverErrorResponse(w, r, fmt.Errorf("upload failed: %w", err))
			return
		}

		


		// Save file record to DB
		err = app.models.Newsletter.InsertFile(newsletterId, key, header.Filename)
		if err != nil {
			app.serverErrorResponse(w, r, err)
			return
		}
	} else if err != http.ErrMissingFile {
		// Only error if it's not a "missing file" case
		app.serverErrorResponse(w, r, err)
		return
	}

	// Send newsletter to subscribers
	subscribers, err := app.models.NewsletterSubscribers.GetNewsletterSubscribers()
	if err != nil {
		app.serverErrorResponse(w, r, err)
		return
	}

	for _, s := range subscribers {
		err = app.queue.Send(r.Context(), messaging.Message{
			"job":   "newsletter_email",
			"id":    newsletter.ID,
			"email": s.Email,
		})
		if err != nil {
			app.serverErrorResponse(w, r, err)
			return
		}
	}

	newsletter.Owner = adminSubscriber.Email

	data := envelope{"newsletter": newsletter}
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