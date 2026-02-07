package server

import (
	"CruiseBlog/utils"
	"context"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"strings"
	"sync"
	"time"

	"CruiseBlog/types"
	"crypto/rand"
)

var userID types.Key

type Server struct {
	blog       []types.Post
	usernames  map[string]string
	blogMu     sync.RWMutex
	lastLoaded time.Time
}

func NewServer() *Server {
	return &Server{blog: make([]types.Post, 0), usernames: make(map[string]string)}
}

func (s *Server) JoinServer(w http.ResponseWriter, req *http.Request) {
	if req.URL.Path != "/" {
		http.NotFound(w, req)
		return
	}

	_, err := req.Cookie("username")
	// aka if cookie does not exist
	if err != nil {
		var UUID [16]byte
		_, err := rand.Read(UUID[:])
		if err != nil {
			log.Fatal(err)
		}
		userUUID := hex.EncodeToString(UUID[:])
		s.usernames[userUUID] = utils.GetRandValue()
		cookie := new(http.Cookie)
		cookie.Name = "session"
		cookie.Value = userUUID
		http.SetCookie(w, cookie)
		http.Redirect(w, req, "/home/about.html", http.StatusSeeOther)
	} else {
		s.LoadPosts()
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
	username := s.usernames[(req.Context().Value(userID)).(*http.Cookie).Value]

	// TODO fix cleanpost
	if utils.CleanPost("") {
		// date := time.Now()
		date := strings.Split(fmt.Sprint(time.Now()), ".")[0]
		newPost := types.Post{DateOfPost: fmt.Sprint(date), Username: fmt.Sprint(username), Content: fmt.Sprint(content.Content)}

		// add post to []Post
		s.blog = append(s.blog, newPost)
		f, _ := json.Marshal(newPost)
		err := utils.SavePost(newPost)

		if err != nil {
			http.Error(w, "Internal Server Error", http.StatusInternalServerError)
			return
		}
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

func (s *Server) LoadPosts() {
	fmt.Println("reading from disk")
	posts, err := utils.GetPostsFromDisk()
	if err != nil {
		fmt.Println("loading posts from disk err", err)
		return
	}

	s.blogMu.Lock()
	s.blog = posts
	s.lastLoaded = time.Now()
	s.blogMu.Unlock()

	go func() {
		ticker := time.NewTicker(3 * time.Minute)
		defer ticker.Stop()
		for range ticker.C {
			fmt.Println("reading from disk")
			posts, err := utils.GetPostsFromDisk()
			if err != nil {
				fmt.Println("loading posts from disk err", err)
				continue
			}

			s.blogMu.Lock()
			s.blog = posts
			s.lastLoaded = time.Now()
			s.blogMu.Unlock()
		}
	}()
}

func (s *Server) RequireAuth() http.HandlerFunc {
	return func(w http.ResponseWriter, req *http.Request) {
		username, err := req.Cookie("session")
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
