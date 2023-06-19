package models

import (
	"database/sql"
	"errors"
	"github.com/go-sql-driver/mysql"
	"golang.org/x/crypto/bcrypt"
	"strings"
	"time"
)

// User represents a user in the system.
type User struct {
	ID             int       `db:"id"`              // The user ID.
	Name           string    `db:"name"`            // The user name.
	Email          string    `db:"email"`           // The user email address.
	HashedPassword []byte    `db:"hashed_password"` // The user password, hashed with bcrypt.
	Created        time.Time `db:"created"`         // The user creation date.
}

// UserModel is a repository for users.
type UserModel struct {
	DB *sql.DB // The database connection.
}

// Insert inserts a new user into the database.
func (m *UserModel) Insert(name, email, password string) error {
	// Generate a bcrypt hash of the password.
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), 12)
	if err != nil {
		return err
	}

	// Prepare the SQL statement.
	stmt := `INSERT INTO users (name, email, hashed_password, created)
    VALUES(?, ?, ?, UTC_TIMESTAMP())`

	// Execute the SQL statement.
	_, err = m.DB.Exec(stmt, name, email, string(hashedPassword))
	if err != nil {
		// Check if the error is a MySQL error with code 1062 (duplicate key).
		var mySQLError *mysql.MySQLError
		if errors.As(err, &mySQLError) {
			if mySQLError.Number == 1062 && strings.Contains(mySQLError.Message, "users_uc_email") {
				// The email address is already in use.
				return ErrDuplicateEmail
			}
		}

		// Return the error.
		return err
	}

	// No error.
	return nil
}

// Authenticate a user by email and password.
func (m *UserModel) Authenticate(email, password string) (int, error) {
	// Get the user ID and hashed password from the database.
	var id int
	var hashedPassword []byte
	stmt := "SELECT id, hashed_password FROM users WHERE email = ?"
	err := m.DB.QueryRow(stmt, email).Scan(&id, &hashedPassword)
	if err != nil {
		// If there are no rows in the database, the user does not exist.
		if errors.Is(err, sql.ErrNoRows) {
			return 0, ErrInvalidCredentials
		} else {
			// Otherwise, return the error.
			return 0, err
		}
	}

	// Compare the hashed password with the user-supplied password.
	err = bcrypt.CompareHashAndPassword(hashedPassword, []byte(password))
	if err != nil {
		// If the passwords do not match, return an error.
		if errors.Is(err, bcrypt.ErrMismatchedHashAndPassword) {
			return 0, ErrInvalidCredentials
		} else {
			// Otherwise, return the error.
			return 0, err
		}
	}

	// The user has been authenticated successfully.
	return id, nil
}

func (m *UserModel) Get(id int) (*User, error) {
	var user User

	stmt := `SELECT id, name, email, created FROM users WHERE id = ?`

	err := m.DB.QueryRow(stmt, id).Scan(&user.ID, &user.Name, &user.Email, &user.Created)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrNoRecord
		} else {
			return nil, err
		}
	}

	return &user, nil
}

// Check if a user exists with the given ID.
func (m *UserModel) Exists(id int) (bool, error) {
	// Check if the user exists in the database.
	var exists bool
	stmt := "SELECT EXISTS(SELECT true FROM users WHERE id = ?)"
	err := m.DB.QueryRow(stmt, id).Scan(&exists)
	return exists, err
}

// UserModelInterface is an interface that defines the methods for a user model.
type UserModelInterface interface {
	Insert(name, email, password string) error
	Authenticate(email, password string) (int, error)
	Exists(id int) (bool, error)
	Get(id int) (*User, error)
}
