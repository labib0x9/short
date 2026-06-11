package queue

type Message struct {
	To    string `json:"to"`
	Name  string `json:"name"`
	Token string `json:"token"`
}
