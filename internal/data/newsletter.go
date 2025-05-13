package data

import (
	"canvas/validator"
	"context"
	"database/sql"
	"time"

	"github.com/lib/pq"
)


type Newsletter struct {
	ID      string
	Title   string
	Body    string
	CreatedAT time.Time
	UpdatedAT time.Time
	Tags    []string
	CreatedBy int
	FileURLs  []string 
	Owner     string

}


type NewsletterModel struct {
	DB *sql.DB
}


func (m NewsletterModel) Insert(title, body string, tags []string, createdBy int) (*Newsletter, error) {
	stmt := `
		INSERT INTO newsletters (title, body, tags, created_by)
		VALUES ($1, $2, $3, $4)
		RETURNING id, title, body, created_at, updated_at, tags, created_by
	`

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	var newsletter Newsletter
	err := m.DB.QueryRowContext(ctx, stmt, title, body, pq.Array(tags), createdBy).Scan(
		&newsletter.ID,
		&newsletter.Title,
		&newsletter.Body,
		&newsletter.CreatedAT,
		&newsletter.UpdatedAT,
		pq.Array(&newsletter.Tags),
		&newsletter.CreatedBy,
	)

	if err != nil {
		return nil, err
	}

	return &newsletter, nil
}



func (m NewsletterModel) GetNewsletter(id int) (*Newsletter, error) {
	stmt := `
		SELECT 
			n.id, n.title, n.body, n.created_at, n.updated_at, n.tags, n.created_by,
			COALESCE(array_agg(nf.file_url) FILTER (WHERE nf.file_url IS NOT NULL), '{}') AS file_urls
		FROM newsletters n
		LEFT JOIN newsletter_files nf ON n.id = nf.newsletter_id
		WHERE n.id = $1
		GROUP BY n.id
	`

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	var newsletter Newsletter
	err := m.DB.QueryRowContext(ctx, stmt, id).Scan(
		&newsletter.ID,
		&newsletter.Title,
		&newsletter.Body,
		&newsletter.CreatedAT,
		&newsletter.UpdatedAT,
		pq.Array(&newsletter.Tags),
		&newsletter.CreatedBy,
		pq.Array(&newsletter.FileURLs), // aggregate result
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, ErrRecordNotFound
		}
		return nil, err
	}

	return &newsletter, nil
}


func (m NewsletterModel) GetNewsletters()([]*Newsletter, error) {
	 stmt := `SELECT id,title,body,created_at,updated_at FROM newsletters`
 ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
 defer cancel()

 rows, err := m.DB.QueryContext(ctx, stmt)
 if err != nil {
	return nil, err
 }
 defer rows.Close()

 newsletters := []*Newsletter{}

 for rows.Next() {
	var newsletter Newsletter
	err = rows.Scan(&newsletter.ID, &newsletter.Title, &newsletter.Body, &newsletter.CreatedAT, &newsletter.UpdatedAT)
	if err != nil {
		return nil, err
	}
	newsletters = append(newsletters, &newsletter)
 }

 if err = rows.Err(); err != nil {
	return nil, err
 }

 return newsletters, nil
}

func (m NewsletterModel) UpdateNewsletter(id int, title, body string) (*Newsletter, error) {
	 stmt := `UPDATE newsletters SET title = $1, body = $2, updated_at = now() WHERE id = $3 RETURNING *`
 ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
 defer cancel()

 var newsletter Newsletter
 err := m.DB.QueryRowContext(ctx, stmt, title, body, id).Scan(&newsletter.ID, &newsletter.Title, &newsletter.Body, &newsletter.CreatedAT, &newsletter.UpdatedAT)

 if err != nil {
	if err == sql.ErrNoRows {
		return nil, ErrRecordNotFound
	}
	return nil, err
 }
 return &newsletter, nil
}

func (m NewsletterModel) DeleteNewsletter(id int) error {
	 stmt := `DELETE FROM newsletters WHERE id = $1`
 ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
 defer cancel()

 _, err := m.DB.ExecContext(ctx, stmt, id)

 if err != nil {
	if err == sql.ErrNoRows {
		return ErrRecordNotFound
	}
	return err
 }
 return nil
}


func (m NewsletterModel) SearchNewsletter(text string) ([]*Newsletter, error) {
	 stmt := `SELECT id,title,body,created_at,updated_at FROM newsletters
	          WHERE to_tsvector(title || '' || body) @@ websearch_to_tsquery($1)`
 ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
 defer cancel()
 rows, err := m.DB.QueryContext(ctx, stmt, text)
 if err != nil {
	return nil, err
 }
 defer rows.Close()
 newsletters := []*Newsletter{}
 for rows.Next() {
	var newsletter Newsletter
	err = rows.Scan(&newsletter.ID, &newsletter.Title, &newsletter.Body, &newsletter.CreatedAT, &newsletter.UpdatedAT)
	if err != nil {
		return nil, err
	}
	newsletters = append(newsletters, &newsletter)


 }
 if err = rows.Err(); err != nil {
	return nil, err
 }
 return newsletters, nil



 
}


func ValidateNewsletter(v *validator.Validator, title, body string) {
	v.Check(title != "", "title", "must be provided")
	v.Check(len(title) <= 100, "title", "must not be more than 100 bytes long")
	v.Check(body != "", "body", "must be provided")
	v.Check(len(body) <= 5000, "body", "must not be more than 5000 bytes long")
}



func (n NewsletterModel) InsertFile(newsletterID int, url string, filename string) error {
	query := `
		INSERT INTO newsletter_files (newsletter_id, file_url, filename)
		VALUES ($1, $2, $3)
	`

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
 defer cancel()

	_, err := n.DB.ExecContext(ctx,query, newsletterID, url, filename)
	return err
}


