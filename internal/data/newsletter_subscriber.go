package data

import (
	"canvas/validator"
	"context"
	"crypto/rand"
	"database/sql"
	"fmt"
	"time"
)

var ErrRecordNotFound = fmt.Errorf("record not found")

type Newsletter_Subscriber struct {
    ID     int  `json:"id"`
    Email  string `json:"email"`
    token  string `json:"token"`
    confirmed bool `json:"confirm"`
    active  bool `json:"active"`
    createdAt string `json:"created_at"`
    updatedAt string `json:"updated_at"`
    IsAdmin bool `json:"is_admin"`
}

type NewsletterSubscriberModel struct {
    DB *sql.DB
}


func (m NewsletterSubscriberModel) Insert(email string) (string,error) {

    token, err := generateToken()
    if err != nil {
        return "", err
    }

    stmt:= `INSERT INTO newsletter_subscribers(email,token) 
            VALUES ($1,$2) ON CONFLICT(email) DO UPDATE SET 
            token = EXCLUDED.token, updated_at = now()`

   ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
   defer cancel()

   _,err = m.DB.ExecContext(ctx,stmt,email,token)

   return token,err

}

func (m NewsletterSubscriberModel) Confirm(token string) (string,error) {
    var email string
    stmt:= `UPDATE newsletter_subscribers SET confirmed = true WHERE token = $1 RETURNING email`

    ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
    defer cancel()

    err := m.DB.QueryRowContext(ctx,stmt,token).Scan(&email)

    if err != nil {
        if err == sql.ErrNoRows {
            return "", ErrRecordNotFound
        }
        return "", err
    }

    return email, nil
}

func generateToken() (string, error) {
    secret:= make([]byte, 32)

    _, err := rand.Read(secret)

    if err != nil {
        return "", err
    }
    token := fmt.Sprintf("%x", secret)
    return token, nil
}

func (m NewsletterSubscriberModel) GetNewsletterSubscribers() ([]*Newsletter_Subscriber, error) {
    stmt := `SELECT email FROM newsletter_subscribers WHERE confirmed and active = true order by created_at`
    ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
    defer cancel()

    rows, err := m.DB.QueryContext(ctx, stmt)
    if err != nil {
        return nil, err
    }
    defer rows.Close()

    var subscribers []*Newsletter_Subscriber

    for rows.Next() {
        var subscriber Newsletter_Subscriber

        err = rows.Scan(&subscriber.Email)
       
        if err != nil {
            return nil, err
        }
        subscribers = append(subscribers, &subscriber)
    }

    return subscribers, nil
}

// get active and is_admin subscribers:

func (m NewsletterSubscriberModel) GetAdminSubscriber()(Newsletter_Subscriber,error) {
    stmt := `SELECT id,email,is_admin FROM newsletter_subscribers WHERE active = true and is_admin = true ORDER BY created_at LIMIT 1`
    ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
    defer cancel()

    var subscriber Newsletter_Subscriber

    err := m.DB.QueryRowContext(ctx, stmt).Scan(&subscriber.ID,&subscriber.Email,&subscriber.IsAdmin)

    if err != nil {
        if err == sql.ErrNoRows {
            return subscriber, ErrRecordNotFound
        }
        return subscriber, err
    }

    return subscriber, nil
}





func ValidateEmail(v *validator.Validator, email string) {
    v.Check(email != "", "email", "must be provided")
    v.Check(validator.Matches(email, validator.EmailRX), "email", "must be a valid email address")
}

// Check that the plaintext token has been provided and is exactly 26 bytes long.
func ValidateTokenPlaintext(v *validator.Validator, tokenPlaintext string) {
    v.Check(tokenPlaintext != "", "token", "must be provided")
}

// func ValidateUser(v *validator.Validator, user *User) {
// 	v.Check(user.Name != "", "name", "must be provided")
// 	v.Check(len(user.Name) <= 500, "name", "must not be more than 500 bytes long")

// 	// Call the standalone ValidateEmail() helper.
// 	ValidateEmail(v, user.Email)

// }