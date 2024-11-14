package main

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"sync"
)

type Sessions struct {
	mu       sync.Mutex
	sessions map[string]*Session
}

func (sessions *Sessions) Create(id string, candidate *Candidate) *Session {
	s := &Session{
		Offer:            candidate,
		Answer:           nil,
		OfferCandidates:  []map[string]any{},
		AnswerCandidates: []map[string]any{},
	}

	sessions.mu.Lock()
	defer sessions.mu.Unlock()
	sessions.sessions[id] = s

	return s
}

func (sessions *Sessions) Get(id string) *Session {
	sessions.mu.Lock()
	defer sessions.mu.Unlock()
	return sessions.sessions[id]
}

func (sessions *Sessions) SetAnswer(id string, answer *Candidate) {
	sessions.mu.Lock()
	defer sessions.mu.Unlock()
	session := sessions.sessions[id]
	session.Answer = answer
	sessions.sessions[id] = session
}

func (sessions *Sessions) AddOfferCandidate(id string, candidate map[string]any) {
	sessions.mu.Lock()
	defer sessions.mu.Unlock()

	s, ok := sessions.sessions[id]
	if !ok {
		panic("no session with ID " + id)
	}
	s.OfferCandidates = append(s.OfferCandidates, candidate)
	sessions.sessions[id] = s
}

func (sessions *Sessions) AddAnswerCandidate(id string, candidate map[string]any) {
	sessions.mu.Lock()
	defer sessions.mu.Unlock()

	s := sessions.sessions[id]
	s.AnswerCandidates = append(s.AnswerCandidates, candidate)
	sessions.sessions[id] = s
}

type Candidate struct {
	SDP       string `json:"sdp"`
	OfferType string `json:"type"`
}

type Session struct {
	Offer            *Candidate       `json:"offer,omitempty"`
	Answer           *Candidate       `json:"answer,omitempty"`
	OfferCandidates  []map[string]any `json:"offer_candidates"`
	AnswerCandidates []map[string]any `json:"answer_candidates"`
}

func main() {
	sessions := Sessions{
		sessions: map[string]*Session{},
	}

	http.HandleFunc("/createSession", func(w http.ResponseWriter, r *http.Request) {
		q, err := url.ParseQuery(r.URL.RawQuery)
		if err != nil {
			panic(err)
		}
		id := q.Get("id")

		body, err := io.ReadAll(r.Body)
		if err != nil {
			panic(err)
		}

		var candidate Candidate
		err = json.Unmarshal(body, &candidate)
		if err != nil {
			panic(err)
		}

		newSession := sessions.Create(id, &candidate)

		fmt.Printf("new session: %v\n", newSession)

		w.Header().Set("Access-Control-Allow-Origin", "*")
	})

	http.HandleFunc("/setAnswerOnSession", func(w http.ResponseWriter, r *http.Request) {
		q, err := url.ParseQuery(r.URL.RawQuery)
		if err != nil {
			panic(err)
		}
		id := q.Get("id")

		body, err := io.ReadAll(r.Body)
		if err != nil {
			panic(err)
		}

		var candidate Candidate
		err = json.Unmarshal(body, &candidate)
		if err != nil {
			panic(err)
		}

		sessions.SetAnswer(id, &candidate)

		fmt.Println("set answer on session")

		w.Header().Set("Access-Control-Allow-Origin", "*")
	})

	http.HandleFunc("/getSession", func(w http.ResponseWriter, r *http.Request) {
		q, err := url.ParseQuery(r.URL.RawQuery)
		if err != nil {
			panic(err)
		}
		id := q.Get("id")

		s := sessions.Get(id)
		bytes, err := json.Marshal(s)
		if err != nil {
			panic(err)
		}

		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Write(bytes)
	})

	http.HandleFunc("/addOfferCandidate", func(w http.ResponseWriter, r *http.Request) {
		q, err := url.ParseQuery(r.URL.RawQuery)
		if err != nil {
			panic(err)
		}
		id := q.Get("id")

		body, err := io.ReadAll(r.Body)
		if err != nil {
			panic(err)
		}

		var candidate map[string]any
		err = json.Unmarshal(body, &candidate)
		if err != nil {
			panic(err)
		}

		sessions.AddOfferCandidate(id, candidate)

		fmt.Printf("added offer candidate: %v\n", candidate)

		w.Header().Set("Access-Control-Allow-Origin", "*")
	})

	http.HandleFunc("/addAnswerCandidate", func(w http.ResponseWriter, r *http.Request) {
		q, err := url.ParseQuery(r.URL.RawQuery)
		if err != nil {
			panic(err)
		}
		id := q.Get("id")

		body, err := io.ReadAll(r.Body)
		if err != nil {
			panic(err)
		}

		var candidate map[string]any
		err = json.Unmarshal(body, &candidate)
		if err != nil {
			panic(err)
		}

		sessions.AddAnswerCandidate(id, candidate)

		fmt.Printf("added answer candidate: %v\n", candidate)

		w.Header().Set("Access-Control-Allow-Origin", "*")
	})

	http.ListenAndServe(":8080", nil)
}
