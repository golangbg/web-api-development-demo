package server

import (
	"context"
	"encoding/gob"
	"fmt"
	"net/http"
	"time"

	"github.com/golangbg/web-api-development/pkg/database"
	"github.com/golangbg/web-api-development/pkg/models"
	"github.com/gorilla/sessions"
)

// SessionName represents the name under which sessions will be saved
const SessionName = "blog"

// Server is a custom type which contains data required to run the server
type Server struct {
	// Use embedding to make our custom server type compatible with http.Server
	// https://www.ardanlabs.com/blog/2014/05/methods-interfaces-and-embedded-types.html
	// https://www.ardanlabs.com/blog/2015/09/composition-with-go.html
	http.Server

	// store provides cookie and filesystem sessions and infrastructure for custom session backends
	// http://www.gorillatoolkit.org/pkg/sessions
	store *sessions.CookieStore
	db    *database.DB
}

// Close contains all the steps for a graceful shutdown of the server
func (s *Server) Close() {
	// Close the database
	s.db.CloseDB()

	// Shutdown the http server
	s.Shutdown(context.Background())
}

// PrepareData prepares data which is to be send with flashes
func (s *Server) PrepareData(w http.ResponseWriter, r *http.Request, data map[string]interface{}) {
	// Get the session
	session, err := s.store.Get(r, SessionName)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// Check if the session has flash messages, if so pass them via data
	if flashes := session.Flashes(); len(flashes) > 0 {
		data["Flashes"] = flashes
		session.Save(r, w)
	}

	// Check if the session has an active user, if so pass it via data
	if activeUser, ok := session.Values["activeUser"]; ok {
		data["ActiveUser"] = activeUser
	}
}

// New initializes and returns a pointer to a custom server (https://gobyexample.com/pointers)
func New(addr string) (*Server, error) {
	db, err := database.New("goblog.db")
	if err != nil {
		return nil, fmt.Errorf("database error: %v", err)
	}
	// Create custom server
	srv := &Server{
		Server: http.Server{
			Addr:         addr,
			ReadTimeout:  10 * time.Second,
			WriteTimeout: 10 * time.Second,
		},
		store: sessions.NewCookieStore([]byte("something-very-secret")),
		db:    db,
	}

	// Connect the server's handler with the routes
	srv.Handler = srv.Routes()

	return srv, nil
}

// init is used for one-time actions (https://medium.com/golangspec/init-functions-in-go-eac191b3860a)
func init() {
	// Sessions uses gob for encoding/decoring. Therefore we need to register our post and user model once, so that
	// we can use it in sessions (https://golang.org/pkg/encoding/gob/)
	gob.Register(&models.Post{})
	gob.Register(&models.User{})
}
