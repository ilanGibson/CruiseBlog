package server

import (
	"CruiseBlog/utils"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"time"

	"CruiseBlog/types"
)

var userID types.Key

type Server struct {
	blog []types.Post
}

func NewServer() *Server {
	return &Server{blog: make([]types.Post, 0)}
}

func (s *Server) JoinServer(w http.ResponseWriter, req *http.Request) {
	if req.URL.Path != "/" {
		http.NotFound(w, req)
		return
	}

	_, err := req.Cookie("username")
	// aka if cookie does not exist
	if err != nil {
		cookie := new(http.Cookie)
		cookie.Name = "username"
		tempRandomVal := utils.GetRandValue()
		cookie.Value = tempRandomVal
		http.SetCookie(w, cookie)
		http.Redirect(w, req, "/home/about.html", http.StatusSeeOther)
	} else {
		http.Redirect(w, req, "/home", http.StatusSeeOther)
	}
}

func (s *Server) AddPost(w http.ResponseWriter, req *http.Request) {
	// get post content
	body, _ := (io.ReadAll(req.Body))
	defer req.Body.Close()

	var content types.Request
	err := json.Unmarshal(body, &content)
	if err != nil {
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	// get userID cookie value
	username := (req.Context().Value(userID)).(*http.Cookie).Value

	// TODO fix cleanpost
	if utils.CleanPost("") {
		date := time.Now()
		newPost := types.Post{DateOfPost: fmt.Sprint(date), Username: fmt.Sprint(username), Content: fmt.Sprint(content.Content)}

		// add post to []Post
		s.blog = append(s.blog, newPost)
		f, _ := json.Marshal(newPost)
		// TODO swap response write and file write and handle err
		w.Write(f)
	} else {
		f, _ := json.Marshal("against cruise blog policy")
		w.Write(f)
	}
}

func (s *Server) GetPosts(w http.ResponseWriter, _ *http.Request) {
	f, err := json.Marshal(s.blog)
	if err != nil {
		log.Fatal("get post err", err)
	}
	_, err = w.Write(f)
	if err != nil {
		fmt.Println("get posts err", err)
	}
}

func RequireAuth(s *Server) http.HandlerFunc {
	return func(w http.ResponseWriter, req *http.Request) {
		username, err := req.Cookie("username")
		if err != nil {
			http.Error(w, "Unauthorized", http.StatusUnauthorized)
			return
		}

		switch req.Method {
		case "POST":
			ctx := context.WithValue(req.Context(), userID, username)
			s.AddPost(w, req.WithContext(ctx))
		case "GET":
			s.GetPosts(w, req)
		}
	}
}
