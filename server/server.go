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
	"sync/atomic"
	"time"

	"CruiseBlog/types"
	"crypto/rand"
)

var userID types.Key

type Server struct {
	blog        []types.Post
	usernames   map[string]string
	uniqueUsers atomic.Uint64
	blogMu      sync.Mutex
	lastLoaded  time.Time
}

func NewServer() *Server {
	return &Server{blog: make([]types.Post, 0), usernames: make(map[string]string)}
}

func (s *Server) JoinServer(w http.ResponseWriter, req *http.Request) {
	if req.URL.Path != "/" {
		http.NotFound(w, req)
		return
	}

	cookie, err := req.Cookie("session")
	// aka if cookie does not exist
	if err != nil {
		// IpIsUnique returns true is ip in unique
		if uniqueIP := utils.IpIsUnique(strings.Split(req.RemoteAddr, ":")[0]); uniqueIP == true {
			fmt.Println("adding user")
			s.uniqueUsers.Add(1)
		}

		var UUID [16]byte
		_, err := rand.Read(UUID[:])
		if err != nil {
			log.Fatal(err)
		}
		userUUID := hex.EncodeToString(UUID[:])
		s.blogMu.Lock()
		s.usernames[userUUID] = utils.GetRandValue()
		s.blogMu.Unlock()
		cookie := new(http.Cookie)
		cookie.Name = "session"
		cookie.Value = userUUID
		http.SetCookie(w, cookie)
		http.Redirect(w, req, "/home/about.html", http.StatusSeeOther)
	} else {
		// aka if cookie is not mapped to username in server
		// happens when user has correct cookie and server is taken down then restarted
		// user still has cookie in browser to pass initial check but their cookie
		// is not mapped to username because username map[string]string is not persistant
		// between server restarts
		s.blogMu.Lock()
		if _, ok := s.usernames[cookie.Value]; !ok {
			s.usernames[cookie.Value] = utils.GetRandValue()
		}
		s.blogMu.Unlock()
		s.LoadPosts()
		http.Redirect(w, req, "/home", http.StatusSeeOther)
	}
}

func (s *Server) AddPost(w http.ResponseWriter, req *http.Request) {
	s.blogMu.Lock()
	if _, ok := s.usernames[(req.Context().Value(userID)).(*http.Cookie).Value]; !ok {
		s.usernames[(req.Context().Value(userID)).(*http.Cookie).Value] = utils.GetRandValue()
	}
	s.blogMu.Unlock()

	// get post content
	body, _ := (io.ReadAll(req.Body))
	defer req.Body.Close()

	var content types.Request
	err := json.Unmarshal(body, &content)
	if err != nil {
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	s.blogMu.Lock()
	// get userID cookie value
	username := s.usernames[(req.Context().Value(userID)).(*http.Cookie).Value]
	s.blogMu.Unlock()

	if utils.CleanPost(content.Content) {

		date := strings.Split(fmt.Sprint(time.Now()), ".")[0]
		newPost := types.Post{DateOfPost: fmt.Sprint(date), Username: fmt.Sprint(username), Content: fmt.Sprint(content.Content)}

		s.blogMu.Lock()
		// add post to []Post
		s.blog = append(s.blog, newPost)
		s.blogMu.Unlock()

		f, _ := json.Marshal(newPost)
		err := utils.SavePost(newPost)

		if err != nil {
			http.Error(w, "Internal Server Error", http.StatusInternalServerError)
			return
		}
		w.Write(f)
	} else {
		w.WriteHeader(422)
	}
}

func (s *Server) GetPosts(w http.ResponseWriter, req *http.Request) {
	s.blogMu.Lock()
	if _, ok := s.usernames[(req.Context().Value(userID)).(*http.Cookie).Value]; !ok {
		s.usernames[(req.Context().Value(userID)).(*http.Cookie).Value] = utils.GetRandValue()
	}

	f, err := json.Marshal(s.blog)
	s.blogMu.Unlock()
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

func (s *Server) ServerInfo(w http.ResponseWriter, req *http.Request) {
	ServerInfo := types.ServerInfo{UniqueUsers: s.uniqueUsers.Load(), LastServerRestart: s.lastLoaded, ServerAge: time.Duration(time.Since(s.lastLoaded).Seconds())}

	f, err := json.Marshal(ServerInfo)
	if err != nil {
		fmt.Println(err)
	}

	w.Write(f)
}

func (s *Server) RequireAuth() http.HandlerFunc {
	return func(w http.ResponseWriter, req *http.Request) {
		username, err := req.Cookie("session")
		if err != nil {
			http.Error(w, "Unauthorized", http.StatusUnauthorized)
			return
		}

		ctx := context.WithValue(req.Context(), userID, username)
		switch req.Method {
		case "POST":
			// ctx := context.WithValue(req.Context(), userID, username)
			s.AddPost(w, req.WithContext(ctx))
		case "GET":
			s.GetPosts(w, req.WithContext(ctx))
		}
	}
}
