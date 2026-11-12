package mocks

import (
	"snippetbox/internal/models"
	"time"
)

var mockSnippetGet = &models.SnippetGet{
	ID:       1,
	Title:    "An old silent pond",
	Content:  "An old silent pond...",
	Created:  time.Now(),
	Expires:  time.Now(),
	UserName: "Alice Jones",
}

var mockSnippetLatest = &models.SnippetLatest{
	ID:      1,
	Title:   "An old silent pond",
	Content: "An old silent pond...",
	Created: time.Now(),
	Expires: time.Now(),
}

type SnippetModel struct{}

func (m *SnippetModel) Insert(title string, content string, expires int, userID int) (int, error) {
	return 2, nil
}

func (m *SnippetModel) Get(id int, user_id int) (*models.SnippetGet, error) {
	switch id {
	case 1:
		return mockSnippetGet, nil
	default:
		return nil, models.ErrNoRecord
	}
}

func (m *SnippetModel) Latest(user_id int) ([]*models.SnippetLatest, error) {
	return []*models.SnippetLatest{mockSnippetLatest}, nil
}
