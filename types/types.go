package types

import "time"

type Key string

type ClientRequest struct {
	Content string `json:"content"`
}

type Post struct {
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
