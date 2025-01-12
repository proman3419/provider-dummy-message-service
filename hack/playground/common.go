package main

type Message struct {
	Id      int    `json:"id"`
	Content string `json:"content"`
}

type Messages struct {
	Messages []Message `json:"messages"`
}
