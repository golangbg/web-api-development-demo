package server

import (
	"database/sql"
	"fmt"
	"html/template"
	"log"
	"net/http"
	"sync"

	"golang.org/x/crypto/bcrypt"

	"github.com/golangbg/web-api-development-demo/pkg/models"
	"github.com/gorilla/mux"
)

// rootHandler gets and displays all posts
func (s *Server) rootHandler(files ...string) http.HandlerFunc {
	// This part is executed only once when we invoke rootHandler in routes.go (so when the server instance is created)
	log.Println("rootHandler initialization")

	// ParseFiles creates a new Template and parses the template definitions from the named files
	// Must will cause the program to panic if template initialization goes wrong
	// https://golang.org/pkg/html/template/#ParseFiles
	tpl := template.Must(template.New("").ParseFiles(files...))

	// We'll return the HandlerFunc which the router will use
	return func(w http.ResponseWriter, r *http.Request) {
		log.Println("rootHandler HandlerFunc")
		data := make(map[string]interface{})

		// Prepare the data which will be sent to the template
		var err error
		data["Posts"], err = s.db.GetAllPosts()
		if err != nil {
			log.Printf("database error: %v", err)
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		// Prepare data
		s.PrepareData(w, r, data)

		// Execute the template (https://golang.org/pkg/text/template/#Template.Execute)
		if err := tpl.ExecuteTemplate(w, "main", data); err != nil {
			// Parsing the template went wrong, let's log and return the error
			log.Printf("template execution error: %v", err)
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
	}
}

// postReadHandler gets and displays a single post
func (s *Server) postReadHandler(files ...string) http.HandlerFunc {
	log.Println("postReadHandler initialization")

	// sync.Once allows to defer expensive transactions until the first time the handlerFunc is called
	// also this will allow to reply with a decent error instead of panicing
	// https://golang.org/pkg/sync/#Once.Do
	var (
		init sync.Once
		tpl  *template.Template
		err  error
	)

	return func(w http.ResponseWriter, r *http.Request) {
		log.Println("postReadHandler HandlerFunc")

		// Execute initialization transactions only once
		init.Do(func() {
			log.Println("postReadHandler init.Do")
			tpl, err = template.New("").ParseFiles(files...)
		})
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		args := mux.Vars(r)

		post, err := s.db.GetPostBySlug(args["slug"])
		if err != nil && err == sql.ErrNoRows {
			http.NotFound(w, r)
			return
		} else if err != nil {
			log.Printf("database error: %v", err)
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		// Prepare the data which will be sent to the template
		data := map[string]interface{}{
			"Post": post,
		}

		// Prepare data
		s.PrepareData(w, r, data)

		// Execute the template (https://golang.org/pkg/text/template/#Template.Execute)
		if err := tpl.ExecuteTemplate(w, "main", data); err != nil {
			// Parsing the template went wrong, let's log and return the error
			log.Printf("template execution error: %v", err)
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
	}
}

// postCreateHandler renders and displays a form for creating a new post
func (s *Server) postCreateHandler(files ...string) http.HandlerFunc {
	var (
		init sync.Once
		tpl  *template.Template
		err  error
	)

	return func(w http.ResponseWriter, r *http.Request) {
		// Execute initialization transactions only once
		init.Do(func() {
			tpl, err = template.New("").ParseFiles(files...)
		})
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		data := make(map[string]interface{})

		// Get the session
		session, err := s.store.Get(r, SessionName)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		// Check if the session has a currentPost, if so pass it via data
		if currentPost, ok := session.Values["currentPost"]; ok {
			data["CurrentPost"] = currentPost
			delete(session.Values, "currentPost")
			session.Save(r, w)
		}

		// Prepare the data
		s.PrepareData(w, r, data)

		// Execute the template (https://golang.org/pkg/text/template/#Template.Execute)
		if err := tpl.ExecuteTemplate(w, "main", data); err != nil {
			// Parsing the template went wrong, let's log and return the error
			log.Printf("template execution error: %v", err)
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
	}
}

func (s *Server) postSaveHandler(w http.ResponseWriter, r *http.Request) {
	// Parse the HTML form
	if err := r.ParseForm(); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// Get the slug from the parsed HTML form
	slug := r.FormValue("slug")

	// Get the session
	session, err := s.store.Get(r, SessionName)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// Create a post
	post := models.Post{
		Slug:  slug,
		Title: r.FormValue("title"),
		Body:  template.HTML(r.FormValue("body")),
		// Link the new post the the logged in user by getting the userID from the session
		// session.Values uses an interface to store data, therefore we need to assert to int64
		// https://tour.golang.org/methods/15
		UserID: session.Values["activeUserID"].(int64),
	}

	// Get the session
	session, err = s.store.Get(r, SessionName)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// Validate the post
	if err := post.Validate(); err != nil {
		// Validation went wrong, we will use the session to pass the validation error in an elegant way

		// Add a flash message and the post to the session. Then save the session.
		session.AddFlash(err.Error())
		session.Values["currentPost"] = post
		session.Save(r, w)

		// Redirect
		http.Redirect(w, r, "/new", http.StatusFound)
		return
	}

	// Create / overwrite the post
	if _, err := s.db.SavePost(post); err != nil {
		// Add a flash message and the post to the session. Then save the session.
		session.AddFlash(fmt.Sprintf("database error: %v", err.Error()))
		session.Values["currentPost"] = post
		session.Save(r, w)

		// Redirect
		http.Redirect(w, r, "/new", http.StatusFound)
		return
	}
	http.Redirect(w, r, "/"+post.Slug, http.StatusFound)
}

// userCreateHandler renders and displays a form for creating a new user
func (s *Server) userCreateHandler(files ...string) http.HandlerFunc {
	var (
		init sync.Once
		tpl  *template.Template
		err  error
	)

	return func(w http.ResponseWriter, r *http.Request) {
		// Execute initialization transactions only once
		init.Do(func() {
			tpl, err = template.New("").ParseFiles(files...)
		})
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		data := make(map[string]interface{})

		// Get the session
		session, err := s.store.Get(r, SessionName)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		// Check if the session has a currentPost, if so pass it via data
		if currentUser, ok := session.Values["currentUser"]; ok {
			data["CurrentUser"] = currentUser
			delete(session.Values, "currentUser")
			session.Save(r, w)
		}

		// Prepare the data
		s.PrepareData(w, r, data)

		// Execute the template (https://golang.org/pkg/text/template/#Template.Execute)
		if err := tpl.ExecuteTemplate(w, "main", data); err != nil {
			// Parsing the template went wrong, let's log and return the error
			log.Printf("template execution error: %v", err)
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
	}
}

func (s *Server) userSaveHandler(w http.ResponseWriter, r *http.Request) {
	// Parse the HTML form
	if err := r.ParseForm(); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// Get the session
	session, err := s.store.Get(r, SessionName)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// Create a user
	user := models.User{
		Username: r.FormValue("username"),
		Name:     r.FormValue("name"),
	}

	// Validate the user
	if err := user.Validate(); err != nil {
		// Validation went wrong, we will use the session to pass the validation error in an elegant way
		// Add a flash message and the post to the session. Then save the session.
		session.AddFlash(err.Error())
		session.Values["currentUser"] = user
		session.Save(r, w)

		// Redirect
		http.Redirect(w, r, "/register", http.StatusFound)
		return
	}

	// Check if password and confirmPassword match
	password := r.FormValue("password")
	confirmPassword := r.FormValue("confirmPassword")

	if password != confirmPassword {
		// Add a flash message and the post to the session. Then save the session.
		session.AddFlash("passwords don't match")
		session.Values["currentUser"] = user
		session.Save(r, w)

		// Redirect
		http.Redirect(w, r, "/register", http.StatusFound)
		return
	}

	// Create / overwrite the user
	if _, err := s.db.SaveUser(user, password); err != nil {
		// Add a flash message and the post to the session. Then save the session.
		session.AddFlash(fmt.Sprintf("database error: %v", err.Error()))
		session.Values["currentUser"] = user
		session.Save(r, w)

		// Redirect
		http.Redirect(w, r, "/register", http.StatusFound)
		return
	}
	http.Redirect(w, r, "/", http.StatusFound)
}

// userLoginHandler renders and displays a form for logging in
func (s *Server) userLoginHandler(files ...string) http.HandlerFunc {
	var (
		init sync.Once
		tpl  *template.Template
		err  error
	)

	return func(w http.ResponseWriter, r *http.Request) {
		// Execute initialization transactions only once
		init.Do(func() {
			tpl, err = template.New("").ParseFiles(files...)
		})
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		data := make(map[string]interface{})
		// Prepare data
		s.PrepareData(w, r, data)

		// Execute the template (https://golang.org/pkg/text/template/#Template.Execute)
		if err := tpl.ExecuteTemplate(w, "main", data); err != nil {
			// Parsing the template went wrong, let's log and return the error
			log.Printf("template execution error: %v", err)
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
	}
}

// userAuthenticateHandler handles user authentication
func (s *Server) userAuthenticateHandler(w http.ResponseWriter, r *http.Request) {
	// Parse the HTML form
	if err := r.ParseForm(); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// Get the session
	session, err := s.store.Get(r, SessionName)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// Get the username and password from the form
	// Here we do that manually, Gorilla offers schemas to make this job easier
	// http://www.gorillatoolkit.org/pkg/schema
	username := r.FormValue("username")
	password := r.FormValue("password")

	// Get the user from the database
	user, err := s.db.GetUserByUsername(username)
	if err != nil {
		// Couldn't get the user from the database
		log.Printf("db error: %v", err)
		session.AddFlash("login failed")
		session.Save(r, w)

		// Redirect
		http.Redirect(w, r, "/login", http.StatusFound)
		return
	}

	// Check if the password matches
	if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(password)); err != nil {
		// Password doesn't match
		session.AddFlash("login failed")
		session.Save(r, w)

		// Redirect
		http.Redirect(w, r, "/login", http.StatusFound)
		return
	}

	session.Values["activeUser"] = user.Username
	session.Values["activeUserID"] = user.ID
	session.Save(r, w)
	http.Redirect(w, r, "/", http.StatusFound)
}

// userLogoutHandler logs out the current user
func (s *Server) userLogoutHandler(w http.ResponseWriter, r *http.Request) {
	// Get the session
	session, err := s.store.Get(r, SessionName)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// Delete the active user from the session
	delete(session.Values, "activeUser")
	delete(session.Values, "activeUserID")

	// Save the session
	session.Save(r, w)

	// Redirect the user
	http.Redirect(w, r, "/", http.StatusFound)
}
