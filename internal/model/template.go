package model

import (
	"time"
)

// Template represents a nuclei template
type Template struct {
	ID          string    `json:"id"`
	Name        string    `json:"name"`
	Author      string    `json:"author"`
	Tags        []string  `json:"tags"`
	Severity    string    `json:"severity"`
	Type        string    `json:"type"`
	Description string    `json:"description"`
	Path        string    `json:"path"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}
