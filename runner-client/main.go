package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"slices"
	"time"

	"github.com/pion/webrtc/v4"
)

// note that this code is very similar to the web app since it also uses the WebRTC API

type Candidate struct {
	SDP       string `json:"sdp"`
	OfferType string `json:"type"`
}

type Session struct {
	Offer            *Candidate            `json:"offer,omitempty"`
	Answer           *Candidate            `json:"answer,omitempty"`
	OfferCandidates  []webrtc.ICECandidate `json:"offer_candidates"`
	AnswerCandidates []webrtc.ICECandidate `json:"answer_candidates"`
}

var addedAnswerCandidates = []webrtc.ICECandidate{}
var runnerID = "runner_abc" // this must match what's in the "session ID" web UI input

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
		if err := pc.Close(); err != nil {
			panic(err)
		}
	}()

	pc.AddTransceiverFromKind(webrtc.RTPCodecTypeAudio)

	channel, err := pc.CreateDataChannel("test", nil)
	if err != nil {
		panic(err)
	}
	channel.OnOpen(func() {
		fmt.Println("data channel opened, sending test message...")
		channel.Send([]byte("test"))
	})
	channel.OnMessage(func(msg webrtc.DataChannelMessage) {
		fmt.Printf("got message from browser: %s\n", msg.Data)
	})

	pc.OnConnectionStateChange(func(pcs webrtc.PeerConnectionState) {
		fmt.Printf("new connection state: %s\n", pcs)
	})

	offer, err := pc.CreateOffer(&webrtc.OfferOptions{})
	if err != nil {
		panic(err)
	}

	// we must create a session before SetLocalDescription is called since ICE candidates start coming in right after that
	// and we need a session that we can attach these ICE candidates to
	createSession(runnerID, offer)

	err = pc.SetLocalDescription(offer)
	if err != nil {
		panic(err)
	}

	pc.OnICECandidate(func(i *webrtc.ICECandidate) {
		if i != nil {
			addOfferCandidate(runnerID, i)
		}
	})

	// poll until we have an answer from the web UI
	for {
		session := getSession(runnerID)
		if session.Answer != nil {
			err := pc.SetRemoteDescription(webrtc.SessionDescription{
				Type: webrtc.NewSDPType(session.Answer.OfferType),
				SDP:  session.Answer.SDP,
			})
			if err != nil {
				panic(err)
			}
			break
		}
		time.Sleep(1000)
	}

	// this makes the server run forever, polling for further ICE candidates as they come in from the web UI
	for {
		session := getSession(runnerID)
		if len(session.AnswerCandidates) != len(addedAnswerCandidates) {
			for _, c := range session.OfferCandidates {
				if !slices.Contains(addedAnswerCandidates, c) {
					fmt.Println("received new answer candidate")
					addedAnswerCandidates = append(addedAnswerCandidates, c)
					err = pc.AddICECandidate(c.ToJSON())
					if err != nil {
						panic(err)
					}
				}
			}
		}
		time.Sleep(1000)
	}
}

func createSession(id string, localSessionDesc webrtc.SessionDescription) {
	localSessionJSON, err := json.Marshal(localSessionDesc)
	if err != nil {
		panic(err)
	}

	body := bytes.NewBuffer(localSessionJSON)
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

func addOfferCandidate(id string, candidate *webrtc.ICECandidate) {
	candidateJSON, err := json.Marshal(candidate.ToJSON())
	if err != nil {
		panic(err)
	}

	body := bytes.NewBuffer(candidateJSON)
	_, err = http.Post(signalingServerURL+"/addOfferCandidate?id="+id, "", body)
	if err != nil {
		panic(err)
	}
}
