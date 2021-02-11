package main

// Response response from backend, if any
type Response struct {
	Detail []struct {
		Loc  []string `json:"loc"`  // location
		Msg  string   `json:"msg"`  // Message
		Type string   `json:"type"` // Error type
	} `json:"detail"`
} // response ...
