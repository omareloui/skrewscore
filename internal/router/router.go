package router

import (
	"net/http"
	"strings"

	"github.com/omareloui/skrewscore/internal/handlers"
)

func Router(w http.ResponseWriter, r *http.Request) {
	p, m := r.URL.Path, r.Method

	switch {
	case p == "/" && m == "GET":
		handlers.Index(w, r)
	case p == "/start" && m == "POST":
		handlers.Start(w, r)
	case p == "/start-new" && m == "POST":
		handlers.StartNew(w, r)
	case strings.HasPrefix(p, "/game/") && strings.HasSuffix(p, "/set-round-mode") && m == "POST":
		handlers.SetRoundMode(w, r)
	case strings.HasPrefix(p, "/game/") && strings.HasSuffix(p, "/toggle-loser-double") && m == "POST":
		handlers.ToggleLoserDouble(w, r)
	case strings.HasPrefix(p, "/game/") && strings.HasSuffix(p, "/submit-round") && m == "POST":
		handlers.SubmitRound(w, r)
	case strings.HasPrefix(p, "/game/") && m == "GET":
		handlers.Game(w, r)
	default:
		http.NotFound(w, r)
	}
}
