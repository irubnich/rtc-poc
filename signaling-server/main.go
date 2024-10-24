package main

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
)

var sessions = map[string]*Session{}

type Candidate struct {
	SDP       string `json:"sdp"`
	OfferType string `json:"type"`
}

type Session struct {
	Offer            Candidate        `json:"offer,omitempty"`
	Answer           *Candidate       `json:"answer,omitempty"`
	OfferCandidates  []map[string]any `json:"offer_candidates"`
	AnswerCandidates []map[string]any `json:"answer_candidates"`
}

func main() {
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

		s := Session{
			Offer:            candidate,
			Answer:           nil,
			OfferCandidates:  []map[string]any{},
			AnswerCandidates: []map[string]any{},
		}

		sessions[id] = &s

		fmt.Printf("new session: %v\n", s)

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

		s := sessions[id]
		s.Answer = &candidate

		sessions[id] = s

		fmt.Printf("set answer on session: %v\n", s)

		w.Header().Set("Access-Control-Allow-Origin", "*")
	})

	http.HandleFunc("/getSession", func(w http.ResponseWriter, r *http.Request) {
		q, err := url.ParseQuery(r.URL.RawQuery)
		if err != nil {
			panic(err)
		}
		id := q.Get("id")

		s := sessions[id]
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

		s := sessions[id]
		s.OfferCandidates = append(s.OfferCandidates, candidate)
		sessions[id] = s

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

		s := sessions[id]
		s.AnswerCandidates = append(s.AnswerCandidates, candidate)
		sessions[id] = s

		fmt.Printf("added answer candidate: %v\n", candidate)

		w.Header().Set("Access-Control-Allow-Origin", "*")
	})

	http.ListenAndServe(":8080", nil)
}
