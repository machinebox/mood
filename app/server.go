package app

import (
	"net/http"
	"path/filepath"

	"github.com/machinebox/mb/internal/way"
	"github.com/machinebox/sdk-go/textbox"
	"github.com/machinebox/twitterfeed"
	"golang.org/x/net/context"
)

// Server is the app server.
type Server struct {
	assets      string
	tweetReader *twitterfeed.TweetReader
	textbox     *textbox.Client
	router      *way.Router
}

// New makes a new Server.
func New(assets string, tweetReader *twitterfeed.TweetReader, textbox *textbox.Client) *Server {
	srv := &Server{
		assets:      assets,
		tweetReader: tweetReader,
		textbox:     textbox,
		router:      way.NewRouter(),
	}
	srv.router.HandleFunc(http.MethodGet, "/analysis", srv.handleAnalysis)
	srv.router.Handle(http.MethodGet, "/assets/", way.Static("/assets/", assets))
	srv.router.HandleFunc(http.MethodGet, "/", srv.handleIndex)
	return srv
}

func (s *Server) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	s.router.ServeHTTP(w, r)
}

func (s *Server) handleIndex(ctx context.Context, w http.ResponseWriter, r *http.Request) error {
	http.ServeFile(w, r, filepath.Join(s.assets, "index.html"))
	return nil
}
