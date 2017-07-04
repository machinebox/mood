package main

import (
	"log"
	"net/http"
	"strings"
	"sync"
	"time"

	"context"

	"github.com/gorilla/websocket"
	"github.com/machinebox/mood/textboxtally"
)

const (
	socketBufferSize  = 1024
	messageBufferSize = 256
)

var upgrader = &websocket.Upgrader{ReadBufferSize: socketBufferSize, WriteBufferSize: socketBufferSize}

type request struct {
	Terms []string
}

type response struct {
	Error string                `json:"error,omitempty"`
	Tally map[string]*tallyData `json:"tally,omitempty"`
}

type tallyData struct {
	Count            int                              `json:"count"`
	TopKeywords      []textboxtally.Keyword           `json:"top_keywords"`
	SentimentAverage float64                          `json:"sentiment_average"`
	TopEntities      map[string][]textboxtally.Entity `json:"top_entities"`
}

func (s *Server) handleAnalysis(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	socket, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Fatal("upgrader.Upgrade:", err)
	}

	var talliesLock sync.RWMutex
	tallies := make(map[string]*textboxtally.Tally)

	var socketLock sync.Mutex
	var tweetCtx context.Context
	tweetCtxCancel := context.CancelFunc(func() {})
	var stop bool

	defer func() {
		stop = true
		tweetCtxCancel()
	}()

	// receive updates
	var req request
	go func() {
		for {
			if err := socket.ReadJSON(&req); err != nil {
				log.Println("ReadJSON:", err)
				return
			}
			// remove the data from any terms no longer being tracked
			talliesLock.Lock()
			for k := range tallies {
				found := false
				for _, term := range req.Terms {
					if term == k {
						found = true
						break
					}
				}
				if !found {
					delete(tallies, k)
				}
			}
			talliesLock.Unlock()
			log.Println("updated terms to", req.Terms)
			tweetCtxCancel()
		}
	}()

	// read tweets and maintain tally
	go func() {
		defer func() {
			tweetCtxCancel()
		}()
		for {
			tweetCtx, tweetCtxCancel = context.WithCancel(ctx)
			if len(req.Terms) == 0 {
				// no request yet, so wait for context to be cancelled
				// and we'll loop back round to try again after.
				<-tweetCtx.Done()
				continue
			}
			for tweet := range s.tweetReader.Run(tweetCtx, req.Terms...) {
				if len(tweet.Text) == 0 {
					// skip empty tweets
					continue
				}
				analysis, err := s.textbox.Check(strings.NewReader(tweet.Text))
				if err != nil {
					log.Println("textbox.Check:", err)
					res := response{
						Error: err.Error(),
					}
					socketLock.Lock()
					if err := socket.WriteJSON(res); err != nil {
						if err == websocket.ErrCloseSent {
							socketLock.Unlock()
							return
						}
						log.Println("WriteJSON:", err)
					}
					socketLock.Unlock()
					continue
				}
				// update the tallies for each term
				for _, term := range tweet.Terms {
					talliesLock.Lock()
					tally, ok := tallies[term]
					if !ok {
						tally = textboxtally.New()
						tallies[term] = tally
					}
					talliesLock.Unlock()
					tally.Add(analysis)
				}
			}
			if stop {
				return
			}
		}
	}()

	// send responses
	for {
		select {
		case <-time.After(1 * time.Second):
			res := response{
				Tally: make(map[string]*tallyData),
			}
			talliesLock.RLock()
			for term, tally := range tallies {
				res.Tally[term] = &tallyData{
					TopKeywords:      tally.TopKeywords(),
					SentimentAverage: tally.SentimentAverage(),
					TopEntities:      tally.TopEntities(),
					Count:            tally.Count(),
				}
			}
			talliesLock.RUnlock()
			socketLock.Lock()
			if err := socket.WriteJSON(res); err != nil {
				if err == websocket.ErrCloseSent {
					socketLock.Unlock()
					return
				}
				log.Println("WriteJSON:", err)
				socketLock.Unlock()
				return
			}
			socketLock.Unlock()
		}
	}

}
