package models

import (
	"html/template"
	"regexp"
	"time"
)

// Post is a blog post using tags for (un)marshalling to/from json (https://golang.org/pkg/encoding/json/)
type Post struct {
	Slug     string        `json:"slug"`
	UserID   int64         `json:"-"`
	Author   string        `json:"author"`
	Title    string        `json:"title"`
	Body     template.HTML `json:"body"` // Prevents escaping of HTML (https://golang.org/pkg/html/template/#HTML)
	Created  time.Time     `json:"created"`
	Modified time.Time     `json:"modified"`
}

// Preview returns strips Body of all HTML tags and returns the first 100 characters
func (p Post) Preview() string {
	re := regexp.MustCompile("<.*?>")
	txt := re.ReplaceAllString(string(p.Body), "")

	chars := 100
	if l := len(txt); l < chars {
		chars = l
	}
	return txt[:chars]
}

// Validate performs a validation check on the post's data.
// Returns an error in case of a validation error
// Returns nil if validation passed
func (p Post) Validate() error {
	if p.Slug == "" {
		return ValidationError{"Slug", "invalid value"}
	}

	if p.Title == "" {
		return ValidationError{"Title", "empty"}
	}

	if p.UserID <= 0 {
		return ValidationError{"UserID", "invalid value"}
	}

	return nil
}
