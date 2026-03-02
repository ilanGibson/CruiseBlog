package types

import "time"

type Key string

type Admin struct {
	Key              string
	Path             string
	IsPathExpired    bool
	HasPathBeenUsed  bool
	PathExpireLength time.Duration
}

type Event struct {
	TotalUsers int
	TotalPosts int
}

type SseMsg struct {
	TotalUsers int `json:"total_users"`
	TotalPosts int `json:"total_posts"`
}

type ClientRequest struct {
	Content string `json:"content"`
}

type Post struct {
	PostId     [32]byte
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
