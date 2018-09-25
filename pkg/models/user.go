package models

// User is a user of the blog system
type User struct {
	ID       int64  `json:"id"`
	Username string `json:"username"`
	Name     string `json:"name"`
	Password string `json:"password"`
}

// Validate will validate a user
func (u User) Validate() error {
	if u.Username == "" {
		return ValidationError{"Username", "empty"}
	}

	return nil
}
