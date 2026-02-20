package types

import "time"

type Key string

type Admin struct {
	Key             string
	KeyExpireLength time.Duration
	IsKeyExpired    bool
	HasKeyBeenUsed  bool
	AdminChan       chan int
}

type SseMsg struct {
	TotalUsers int `json:"total_users"`
	TotalPosts int `json:"total_posts"`
}

type ClientRequest struct {
	Content string `json:"content"`
}

type Post struct {
	PostId     string `json:"post_id"`
	DateOfPost string `json:"date"`
	Username   string `json:"username"`
	Content    string `json:"content"`
}

type ServerInfo struct {
	UniqueUsers       uint64        `json:"unique_users"`
	LastServerRestart time.Time     `json:"last_server_restart"`
	ServerAge         time.Duration `json:"server_age(seconds)"`
}

type IpSlice struct {
	IpHashes []string
}
