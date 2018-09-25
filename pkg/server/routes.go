package server

import (
	"net/http"

	"github.com/gorilla/mux"
)

// Routes sets up and returns a router
// We are using the Gorilla web toolkit for the router (http://www.gorillatoolkit.org/pkg/mux)
func (s *Server) Routes() *mux.Router {
	// Create a new Gorilla router
	r := mux.NewRouter()

	/**** API routes *****/
	// Authentication
	r.HandleFunc("/api/auth", s.userAuthenticateAPIHandler).Methods(http.MethodPost)

	// Create post
	r.HandleFunc("/api/post", s.ReqToken(s.postCreateUpdateAPIHandler)).Methods(http.MethodPost)

	// Read all posts
	r.HandleFunc("/api/post", s.postsGetAPIHandler).Methods(http.MethodGet)
	// Read a single post
	r.HandleFunc("/api/post/{slug}", s.postGetAPIHandler).Methods(http.MethodGet)

	// Update post
	r.HandleFunc("/api/post/{slug}", s.ReqToken(s.postCreateUpdateAPIHandler)).Methods(http.MethodPut)

	/**** Web routes ****/
	// Serve the static files directory (http://www.gorillatoolkit.org/pkg/mux)
	r.PathPrefix("/assets/").Handler(http.StripPrefix("/assets/", http.FileServer(http.Dir("./static"))))

	// Setup the root URL
	r.HandleFunc("/", s.rootHandler("templates/main.html", "templates/root.html"))

	// Setup the URL for registering new users
	r.HandleFunc("/register", s.userCreateHandler("templates/main.html", "templates/register.html")).Methods(http.MethodGet)

	// Setup the URL for saving a new user
	r.HandleFunc("/register", s.userSaveHandler).Methods(http.MethodPost)

	// Setup the URL for registering new users
	r.HandleFunc("/login", s.userLoginHandler("templates/main.html", "templates/login.html")).Methods(http.MethodGet)

	// Setup the URL for authenticating a user
	r.HandleFunc("/login", s.userAuthenticateHandler).Methods(http.MethodPost)

	// Setup the URL for user logout
	r.HandleFunc("/logout", s.ReqAuth(s.userLogoutHandler))

	// Setup the URL for creating a new post
	r.HandleFunc("/new", s.ReqAuth(s.postCreateHandler("templates/main.html", "templates/create.html"))).Methods(http.MethodGet)

	// Setup the URL for saving a post. Should listen only to POST requests, we do so by using Methods
	r.HandleFunc("/new", s.ReqAuth(s.postSaveHandler)).Methods(http.MethodPost)

	// This one needs to be last
	// Setup the URL for getting a single post, takes the slug as a parameter (http://www.gorillatoolkit.org/pkg/mux)
	r.HandleFunc("/{slug}", s.postReadHandler("templates/main.html", "templates/post.html"))

	return r
}
