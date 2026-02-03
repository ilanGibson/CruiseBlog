package types

type Key string

type Request struct {
	Content string `json:"content"`
}

type Post struct {
	DateOfPost string `json:"date"`
	Username   string `json:"username"`
	Content    string `json:"content"`
}
