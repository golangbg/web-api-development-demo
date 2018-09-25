package server

import (
	"fmt"
	"net/http"
	"strings"
)

// ReqAuth is a middleware function to ensure that a route can only be accessed by an authenticated user
func (s *Server) ReqAuth(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Get the session
		session, err := s.store.Get(r, SessionName)
		if err != nil {
			http.Redirect(w, r, "/", http.StatusFound)
			return
		}

		// Get the active user from the session
		au, ok := session.Values["activeUser"]
		if !ok {
			// We didn't get an active user, so nobody is logged in
			http.Redirect(w, r, "/", http.StatusFound)
			return
		}

		// Check if this is a valid user
		if _, err := s.db.GetUserByUsername(au.(string)); err != nil {
			// We didn't get a valid user from the db, so we'll deny access
			http.Redirect(w, r, "/", http.StatusFound)
			return
		}

		// Everything went well, let's invoke the next HandlerFunc
		next(w, r)
	}
}

// getUserFromToken will extract the active user from the request headers
func getUserFromToken(r *http.Request) (string, error) {
	// Get the authorization header
	// The header is expected to be formatted as: Authorization: BEARER <token>
	authData := r.Header.Get("Authorization")

	// Split the data on the space
	parts := strings.Split(authData, " ")
	// Check whether it's a bearer token
	if len(parts) != 2 || strings.ToLower(parts[0]) != "bearer" {
		return "", fmt.Errorf("invalid header")
	}

	// Parse the token to get the claims
	claims, err := ParseToken(parts[1])
	if err != nil {
		return "", fmt.Errorf("invalid token")
	}

	// Get the active user from the claims
	au, ok := claims["activeUser"].(string)
	if !ok {
		// We didn't get an active user, so nobody is logged in
		return "", fmt.Errorf("invalid value")
	}

	return au, nil
}

// ReqToken is a middleware function to ensure that a route can only be accessed by an authenticated user
func (s *Server) ReqToken(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		au, err := getUserFromToken(r)
		if err != nil {
			answer(w, http.StatusBadRequest, err.Error())
			return
		}

		// Check if this is a valid user
		if _, err := s.db.GetUserByUsername(au); err != nil {
			// We didn't get a valid user from the db, so we'll deny access
			answer(w, http.StatusUnauthorized, nil)
			return
		}

		// Everything went well, let's invoke the next HandlerFunc
		next(w, r)
	}
}
