package data

import "database/sql"

type Model struct {
	NewsletterSubscribers NewsletterSubscriberModel
	Newsletter					 NewsletterModel

}


func NewModel(db *sql.DB) Model {
	return Model{
		NewsletterSubscribers: NewsletterSubscriberModel{DB: db},
		Newsletter:    NewsletterModel{DB: db},
	}
}