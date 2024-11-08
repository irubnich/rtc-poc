package main

import (
	"bufio"
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"slices"
	"time"

	"github.com/pion/webrtc/v4"
)

type Candidate struct {
	SDP       string `json:"sdp"`
	OfferType string `json:"type"`
}

type Session struct {
	Offer            Candidate             `json:"offer,omitempty"`
	Answer           *Candidate            `json:"answer,omitempty"`
	OfferCandidates  []webrtc.ICECandidate `json:"offer_candidates"`
	AnswerCandidates []webrtc.ICECandidate `json:"answer_candidates"`
}

var addedOfferCandidates = []webrtc.ICECandidate{}

// var signalingServerURL = "https://autodiscovery-signaling.app-builder-on-prem.net"
var signalingServerURL = "http://localhost:8080"

func main() {
	config := webrtc.Configuration{
		ICEServers: []webrtc.ICEServer{
			{
				URLs: []string{"stun:stun1.l.google.com:19302", "stun:stun2.l.google.com:19302"},
			},
		},
		ICECandidatePoolSize: 10,
	}
	pc, err := webrtc.NewPeerConnection(config)
	if err != nil {
		panic(err)
	}
	defer func() {
		if cErr := pc.Close(); cErr != nil {
			panic("cannot close peer connection")
		}
	}()

	pc.OnConnectionStateChange(func(pcs webrtc.PeerConnectionState) {
		if pcs == webrtc.PeerConnectionStateConnected {
			fmt.Println("CONNECTED")
		}
	})

	pc.OnDataChannel(func(dc *webrtc.DataChannel) {
		dc.OnMessage(func(msg webrtc.DataChannelMessage) {
			fmt.Printf("got msg: %s\n", msg.Data)
		})
	})

	fmt.Print("Press 'Enter' to signal")
	if _, err = bufio.NewReader(os.Stdin).ReadBytes('\n'); err != nil {
		panic(err)
	}

	sessionID := "a"

	pc.OnICECandidate(func(i *webrtc.ICECandidate) {
		if i != nil {
			addAnswerCandidate(sessionID, i)
		}
	})

	session := getSession(sessionID)
	err = pc.SetRemoteDescription(webrtc.SessionDescription{
		Type: webrtc.NewSDPType(session.Offer.OfferType),
		SDP:  session.Offer.SDP,
	})
	if err != nil {
		panic(err)
	}

	answer, err := pc.CreateAnswer(&webrtc.AnswerOptions{})
	if err != nil {
		panic(err)
	}
	err = pc.SetLocalDescription(answer)
	if err != nil {
		panic(err)
	}
	setAnswerOnSession(sessionID, answer)

	// poll for further offer candidates
	for {
		session = getSession(sessionID)
		if len(session.OfferCandidates) != len(addedOfferCandidates) {
			for _, c := range session.OfferCandidates {
				if !slices.Contains(addedOfferCandidates, c) {
					fmt.Println("adding new offer candidate")
					addedOfferCandidates = append(addedOfferCandidates, c)
					err = pc.AddICECandidate(c.ToJSON())
				}
			}
		}
		time.Sleep(1000)
	}
}

func createSession(id string, offer webrtc.SessionDescription) {
	offerJSON, err := json.Marshal(offer)
	if err != nil {
		panic(err)
	}

	body := bytes.NewBuffer(offerJSON)
	_, err = http.Post(signalingServerURL+"/createSession?id="+id, "", body)
	if err != nil {
		panic(err)
	}
}

func getSession(id string) Session {
	res, err := http.Get(signalingServerURL + "/getSession?id=" + id)
	resStr, err := io.ReadAll(res.Body)
	if err != nil {
		panic(err)
	}

	var s Session
	err = json.Unmarshal(resStr, &s)
	if err != nil {
		panic(err)
	}

	return s
}

func addAnswerCandidate(id string, candidate *webrtc.ICECandidate) {
	fmt.Println("adding candidate")
	candidateJSON, err := json.Marshal(candidate.ToJSON())
	if err != nil {
		panic(err)
	}

	body := bytes.NewBuffer(candidateJSON)
	_, err = http.Post(signalingServerURL+"/addAnswerCandidate?id="+id, "", body)
	if err != nil {
		panic(err)
	}
}

func setAnswerOnSession(id string, answer webrtc.SessionDescription) {
	answerJSON, err := json.Marshal(answer)
	if err != nil {
		panic(err)
	}

	body := bytes.NewBuffer(answerJSON)
	_, err = http.Post(signalingServerURL+"/setAnswerOnSession?id="+id, "", body)
	if err != nil {
		panic(err)
	}
}
