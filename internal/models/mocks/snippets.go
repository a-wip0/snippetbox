package mocks

import (
	"snippetbox/internal/models"
	"time"
)

var mockSnippet = &models.Snippet{
	ID:       1,
	Title:    "An old silent pond",
	Content:  "An old silent pond...",
	Created:  time.Now(),
	Expires:  time.Now(),
	UserName: "Alice Jones",
}

type SnippetModel struct{}

func (m *SnippetModel) Insert(title string, content string, expires int, userID int) (int, error) {
	return 2, nil
}

func (m *SnippetModel) Get(id int) (*models.Snippet, error) {
	switch id {
	case 1:
		return mockSnippet, nil
	default:
		return nil, models.ErrNoRecord
	}
}

func (m *SnippetModel) Latest() ([]*models.Snippet, error) {
	return []*models.Snippet{mockSnippet}, nil
}
