package main

import (
	"net/http"
	"path/filepath"

	"github.com/machinebox/sdk-go/textbox"
	"github.com/machinebox/twitterfeed"
	"github.com/matryer/way"
)

// Server is the app server.
type Server struct {
	assets      string
	tweetReader *twitterfeed.TweetReader
	textbox     *textbox.Client
	router      *way.Router
}

// New makes a new Server.
func NewServer(assets string, tweetReader *twitterfeed.TweetReader, textbox *textbox.Client) *Server {
	srv := &Server{
		assets:      assets,
		tweetReader: tweetReader,
		textbox:     textbox,
		router:      way.NewRouter(),
	}
	srv.router.HandleFunc(http.MethodGet, "/analysis", srv.handleAnalysis)
	srv.router.Handle(http.MethodGet, "/assets/", Static("/assets/", assets))
	srv.router.HandleFunc(http.MethodGet, "/", srv.handleIndex)
	return srv
}

func (s *Server) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	s.router.ServeHTTP(w, r)
}

func (s *Server) handleIndex(w http.ResponseWriter, r *http.Request) {
	http.ServeFile(w, r, filepath.Join(s.assets, "index.html"))
}

// Static gets a static file server for the specified path.
func Static(stripPrefix, dir string) http.Handler {
	h := http.StripPrefix(stripPrefix, http.FileServer(http.Dir(dir)))
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		h.ServeHTTP(w, r)
	})
}
