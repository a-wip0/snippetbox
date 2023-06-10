package models

import (
	"database/sql"
	"errors"
	"time"
)

type Snippet struct {
	ID      int
	Title   string
	Content string
	Created time.Time
	Expires time.Time
}

type SnippetModel struct {
	// DB is the database connection used to store snippets.
	DB *sql.DB
}

// Insert a new snippet into the database.
func (m *SnippetModel) Insert(title string, content string, expires int) (int, error) {
	// stmt is the SQL statement to be executed.
	stmt := `INSERT INTO snippets (title, content, created, expires)
	VALUES(?, ?, UTC_TIMESTAMP(), DATE_ADD(UTC_TIMESTAMP(), INTERVAL ? DAY))`

	// result is the result of executing the SQL statement.
	result, err := m.DB.Exec(stmt, title, content, expires)
	if err != nil {
		// Error if there is an error executing the SQL statement.
		return 0, err
	}

	// id is the ID of the newly inserted snippet.
	id, err := result.LastInsertId()
	if err != nil {
		// Error if there is an error getting the ID of the newly inserted snippet.
		return 0, err
	}

	// Return the ID of the newly inserted snippet.
	return int(id), nil
}

// Get a snippet from the database by its ID.
func (m *SnippetModel) Get(id int) (*Snippet, error) {
	// stmt is the SQL statement to be executed.
	stmt := `SELECT id, title, content, created, expires FROM snippets
	WHERE expires > UTC_TIMESTAMP() AND id = ?`

	// row is the result of executing the SQL statement.
	row := m.DB.QueryRow(stmt, id)

	// s is a new Snippet struct.
	s := &Snippet{}

	// err is the error of scanning the row into the Snippet struct.
	err := row.Scan(&s.ID, &s.Title, &s.Content, &s.Created, &s.Expires)
	if err != nil {
		// If the error is sql.ErrNoRows, then there is no snippet with the specified ID.
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrNoRecord
		} else {
			// Otherwise, there was an error scanning the row into the Snippet struct.
			return nil, err
		}
	}

	// Return the Snippet struct.
	return s, nil
}

// Latest Gets the latest 10 snippets from the database.
func (m *SnippetModel) Latest() ([]*Snippet, error) {

	// stmt is the SQL statement to be executed.
	stmt := `SELECT id, title, content, created, expires FROM snippets
	WHERE expires > UTC_TIMESTAMP() ORDER BY id DESC LIMIT 10`

	// rows is the result of executing the SQL statement.
	rows, err := m.DB.Query(stmt)
	if err != nil {
		// Error if there is an error executing the SQL statement.
		return nil, err
	}

	// defer is used to close the rows object when the function exits.
	defer func(rows *sql.Rows) {
		// Close the rows object.
		err := rows.Close()
		if err != nil {
			// Panic if there is an error closing the rows object.
			panic(err)
		}
	}(rows)

	// snippets is a slice of Snippet structs.
	snippets := []*Snippet{}

	// For each row in the result set, create a new Snippet struct and append it to the snippets slice.
	for rows.Next() {
		s := &Snippet{}
		err := rows.Scan(&s.ID, &s.Title, &s.Content, &s.Created, &s.Expires)
		if err != nil {
			// Error if there is an error scanning the row into the Snippet struct.
			return nil, err
		}
		snippets = append(snippets, s)
	}

	// If there is an error with the rows object, return an error.
	if err = rows.Err(); err != nil {
		return nil, err
	}

	// Return the snippets slice.
	return snippets, nil
}

// SnippetModelInterface is an interface that defines the methods that can be used to interact with a SnippetModel.
type SnippetModelInterface interface {
	Insert(title string, content string, expires int) (int, error)
	Get(id int) (*Snippet, error)
	Latest() ([]*Snippet, error)
}
