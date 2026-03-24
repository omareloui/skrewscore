package handlers

import (
	"log"
	"math"
	"net/http"
	"strings"

	"github.com/a-h/templ"
	"github.com/omareloui/skrewscore/internal/game"
	"github.com/omareloui/skrewscore/views"
)

func renderFull(w http.ResponseWriter, r *http.Request, content templ.Component) {
	w.Header().Set("Content-Type", "text/html")
	if err := views.Layout(content).Render(r.Context(), w); err != nil {
		log.Println("template error:", err)
	}
}

func renderPartial(w http.ResponseWriter, r *http.Request, content templ.Component) {
	w.Header().Set("Content-Type", "text/html")
	if err := content.Render(r.Context(), w); err != nil {
		log.Println("template error:", err)
	}
}

func isHTMX(r *http.Request) bool {
	return r.Header.Get("HX-Request") == "true"
}

// render picks full layout for direct navigation, partial for HTMX swaps.
func render(w http.ResponseWriter, r *http.Request, content templ.Component) {
	if isHTMX(r) {
		renderPartial(w, r, content)
	} else {
		renderFull(w, r, content)
	}
}

func extractGameID(path string) string {
	path = strings.TrimPrefix(path, "/game/")
	if before, _, ok := strings.Cut(path, "/"); ok {
		return before
	}
	return path
}

func findWinners(g *game.Game) []game.Team {
	minTotal := math.Inf(1)
	for i := range g.Teams {
		if t := g.TotalScore(i); t < minTotal {
			minTotal = t
		}
	}
	var winners []game.Team
	for i, team := range g.Teams {
		if g.TotalScore(i) == minTotal {
			winners = append(winners, team)
		}
	}
	return winners
}
