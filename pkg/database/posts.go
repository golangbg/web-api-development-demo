package database

import (
	"time"

	"github.com/golangbg/web-api-development-demo/pkg/models"
)

// SavePost saves a post to the database
func (db *DB) SavePost(post models.Post) (models.Post, error) {
	// Get the current time
	now := time.Now()

	// If post.Created has a zero value, this is a new post, so let's set the creation date and time
	if post.Created.IsZero() {
		post.Created = now
	}
	post.Modified = now

	// Prepare the query
	q := `INSERT OR REPLACE INTO posts(slug, user_id, title, body, created, modified)
	values(?, ?, ?, ?, ?, ?)`
	stmt, err := db.conn.Prepare(q)
	if err != nil {
		// Preparing the query went wrong, so we'll return an empty post and the error
		return models.Post{}, err
	}
	// Make sure stmt gets closed
	defer stmt.Close()

	// Ececute the query
	if _, err := stmt.Exec(post.Slug, post.UserID, post.Title, post.Body, post.Created, post.Modified); err != nil {
		// Execution went wrong, so we'll return an empty post and the error
		return models.Post{}, err
	}

	// Everything went well, let's return the post and nil for the error
	return post, nil
}

// GetPostBySlug gets a post by it's slug
func (db *DB) GetPostBySlug(slug string) (post models.Post, err error) {
	// Prepare the query
	q := "SELECT posts.slug, posts.user_id, users.name, posts.title, posts.body, posts.created, posts.modified FROM posts LEFT JOIN users ON posts.user_id = users.id WHERE slug=?"
	stmt, err := db.conn.Prepare(q)
	if err != nil {
		// Preparing the query went wrong, so we'll return an empty post and the error
		return post, err
	}
	// Make sure stmt gets closed
	defer stmt.Close()

	// Get the post
	if err := stmt.QueryRow(slug).Scan(&post.Slug, &post.UserID, &post.Author, &post.Title, &post.Body, &post.Created, &post.Modified); err != nil {
		return post, err
	}

	return post, err
}

// GetAllPosts gets all posts from the database
func (db *DB) GetAllPosts() (posts []models.Post, err error) {
	// Prepare the query
	q := "SELECT posts.slug, posts.user_id, users.name, posts.title, posts.body, posts.created, posts.modified FROM posts LEFT JOIN users ON posts.user_id = users.id ORDER BY datetime(created) DESC"
	rows, err := db.conn.Query(q)
	if err != nil {
		// Query preparation went wrong
		return posts, err
	}
	// Make sure the rows iterator gets closed
	defer rows.Close()

	// Loop over the received rows and store them in posts
	for rows.Next() {
		// Initialize an empty post
		post := models.Post{}
		// Fill the post through the row
		if err := rows.Scan(&post.Slug, &post.UserID, &post.Author, &post.Title, &post.Body, &post.Created, &post.Modified); err != nil {
			return posts, err
		}

		// Append the post to posts (https://gobyexample.com/slices)
		posts = append(posts, post)
	}

	return posts, err
}
