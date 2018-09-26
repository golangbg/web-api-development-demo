package server

import (
	"database/sql"
	"encoding/json"
	"net/http"
	"time"

	"github.com/gorilla/mux"
	"golang.org/x/crypto/bcrypt"

	"github.com/golangbg/web-api-development/pkg/models"
)

// answer send a json formatted answer to the ResponseWriter
// status represents an HTTP status
func answer(w http.ResponseWriter, status int, data interface{}) {
	// Set the content type we're going to respond with
	w.Header().Set("Content-Type", "application/json; charset=utf-8")

	// Set the status
	w.WriteHeader(status)

	// Encode the data if it isn't nil
	if data != nil {
		// Encode the data to json (https://golang.org/pkg/encoding/json/#Encoder.Encode)
		err := json.NewEncoder(w).Encode(data)
		if err != nil {
			// Encoding failed, so answer with an error
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
	}
}

// postResponse can be used to send a response with a single post item
type postResponse struct {
	Error string      `json:"error"`
	Post  models.Post `json:"post"`
}

// postCreateUpdateAPIHandler creates or updates a single post
func (s *Server) postCreateUpdateAPIHandler(w http.ResponseWriter, r *http.Request) {
	// Decode the request (https://golang.org/pkg/encoding/json/#Decoder.Decode)
	req := models.Post{}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		answer(w, http.StatusBadRequest, postResponse{Error: err.Error()})
		return
	}

	// Get the active user
	au, err := getUserFromToken(r)
	if err != nil {
		answer(w, http.StatusBadRequest, err.Error())
		return
	}

	// Get the user details
	user, err := s.db.GetUserByUsername(au)
	if err != nil {
		// We didn't get a valid user from the db, so we'll deny access
		answer(w, http.StatusBadRequest, err.Error())
		return
	}

	// Set the user ID, because the request doesn't contain this field
	req.UserID = user.ID

	// Perform validation
	if err := req.Validate(); err != nil {
		// Validation failed
		answer(w, http.StatusBadRequest, postResponse{Error: err.Error()})
		return
	}

	// Save the post
	post, err := s.db.SavePost(req)
	if err != nil {
		// Saving went wrong, reply with an error
		answer(w, http.StatusBadRequest, postResponse{Error: err.Error()})
		return
	}

	answer(w, http.StatusCreated, postResponse{Post: post})
}

// postGetAPIHandler gets a single post from the database
func (s *Server) postGetAPIHandler(w http.ResponseWriter, r *http.Request) {
	args := mux.Vars(r)

	// Get the post from the DB
	post, err := s.db.GetPostBySlug(args["slug"])
	if err != nil {
		if err == sql.ErrNoRows {
			answer(w, http.StatusNotFound, postResponse{Error: err.Error()})
			return
		}

		answer(w, http.StatusBadRequest, postResponse{Error: err.Error()})
		return
	}

	answer(w, http.StatusOK, postResponse{Post: post})
}

// postsResponse can be used to send a response with posts items
type postsResponse struct {
	Error string        `json:"error"`
	Posts []models.Post `json:"posts"`
}

// postGetAPIHandler gets all posts from the database
func (s *Server) postsGetAPIHandler(w http.ResponseWriter, r *http.Request) {
	// Get the posts from the DB
	posts, err := s.db.GetAllPosts()
	if err != nil {
		if err == sql.ErrNoRows {
			answer(w, http.StatusNotFound, postResponse{Error: err.Error()})
			return
		}

		answer(w, http.StatusBadRequest, postResponse{Error: err.Error()})
		return
	}

	answer(w, http.StatusOK, postsResponse{Posts: posts})
}

type authenticationRequest struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

type authenticationResponse struct {
	Error string `json:"error"`
	Token string `json:"token"`
}

func (s *Server) userAuthenticateAPIHandler(w http.ResponseWriter, r *http.Request) {
	// Decode the request (https://golang.org/pkg/encoding/json/#Decoder.Decode)
	req := authenticationRequest{}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		answer(w, http.StatusBadRequest, authenticationResponse{Error: err.Error()})
		return
	}

	// Get the user from the database
	user, err := s.db.GetUserByUsername(req.Username)
	if err != nil {
		if err == sql.ErrNoRows {
			answer(w, http.StatusBadRequest, authenticationResponse{Error: "login failed"})
			return
		}

		answer(w, http.StatusBadRequest, postResponse{Error: err.Error()})
		return
	}

	// Check if the password matches
	if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(req.Password)); err != nil {
		// Password doesn't match
		answer(w, http.StatusBadRequest, authenticationResponse{Error: "login failed"})
		return
	}

	// Create a jwt token which is valid for a month
	token, err := CreateToken(map[string]interface{}{"activeUser": user.Username}, time.Now().AddDate(0, 1, 0).Unix())
	if err != nil {
		answer(w, http.StatusBadRequest, authenticationResponse{Error: err.Error()})
		return
	}

	answer(w, http.StatusOK, authenticationResponse{Token: token})
}
