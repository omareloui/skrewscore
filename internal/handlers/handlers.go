package handlers

import (
	"fmt"
	"log"
	"math"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/omareloui/skrewscore/internal/eval"
	"github.com/omareloui/skrewscore/internal/game"
	"github.com/omareloui/skrewscore/internal/mongodb"
	"github.com/omareloui/skrewscore/views"
)

func Index(w http.ResponseWriter, r *http.Request) {
	renderFull(w, r, views.Setup())
}

func Start(w http.ResponseWriter, r *http.Request) {
	r.ParseForm()
	mode := r.FormValue("mode")

	g := &game.Game{
		ID:        uuid.New().String(),
		CreatedAt: time.Now(),
	}

	if mode == "pairs" {
		g.SoloMode = false
		p1s := r.Form["team_p1"]
		p2s := r.Form["team_p2"]
		for i := range p1s {
			n1 := strings.TrimSpace(p1s[i])
			n2 := ""
			if i < len(p2s) {
				n2 = strings.TrimSpace(p2s[i])
			}
			if n1 == "" && n2 == "" {
				continue
			}
			team := game.Team{}
			if n1 != "" {
				team.Players = append(team.Players, game.Player{Name: n1})
			}
			if n2 != "" {
				team.Players = append(team.Players, game.Player{Name: n2})
			}
			g.Teams = append(g.Teams, team)
		}
	} else {
		g.SoloMode = true
		for _, name := range r.Form["players"] {
			name = strings.TrimSpace(name)
			if name == "" {
				continue
			}
			g.Teams = append(g.Teams, game.Team{Players: []game.Player{{Name: name}}})
		}
	}

	if len(g.Teams) < 2 {
		render(w, r, views.Setup())
		return
	}

	g.Rounds = make([]game.Round, game.TotalRounds)
	for i := range g.Rounds {
		g.Rounds[i] = game.Round{
			Number:      i + 1,
			Mode:        game.ModeSum,
			Entries:     make([]game.RoundEntry, len(g.Teams)),
			SkrewCaller: -1,
		}
	}
	g.CurrentRound = 1

	if err := mongodb.SaveGame(g); err != nil {
		log.Println("mongodb.SaveGame:", err)
		http.Error(w, "Failed to save game", http.StatusInternalServerError)
		return
	}

	http.Redirect(w, r, "/game/"+g.ID, http.StatusSeeOther)
}

func StartNew(w http.ResponseWriter, r *http.Request) {
	render(w, r, views.Setup())
}

func Game(w http.ResponseWriter, r *http.Request) {
	id := extractGameID(r.URL.Path)
	g, err := mongodb.LoadGame(id)
	if err != nil {
		log.Println("mongodb.LoadGame:", err)
		http.Error(w, "DB error", http.StatusInternalServerError)
		return
	}
	if g == nil {
		renderFull(w, r, views.NotFound())
		return
	}

	if g.Done {
		renderFull(w, r, views.Done(g, g.ID, findWinners(g)))
		return
	}
	renderFull(w, r, views.Round(g, g.CurrentRoundData(), g.ID, ""))
}

func SetRoundMode(w http.ResponseWriter, r *http.Request) {
	id := extractGameID(r.URL.Path)
	r.ParseForm()

	g, err := mongodb.LoadGame(id)
	if err != nil || g == nil {
		render(w, r, views.NotFound())
		return
	}

	cur := g.CurrentRoundData()
	if cur != nil {
		if r.FormValue("mode") == "average" {
			cur.Mode = game.ModeAverage
		} else {
			cur.Mode = game.ModeSum
		}
		g.Rounds[g.CurrentRound-1] = *cur
		mongodb.SaveGame(g)
	}

	render(w, r, views.Round(g, g.CurrentRoundData(), g.ID, ""))
}

func SubmitRound(w http.ResponseWriter, r *http.Request) {
	id := extractGameID(r.URL.Path)
	r.ParseForm()

	g, err := mongodb.LoadGame(id)
	if err != nil || g == nil {
		render(w, r, views.NotFound())
		return
	}

	cur := g.CurrentRoundData()
	if cur == nil || cur.Locked {
		render(w, r, views.Round(g, cur, g.ID, "Round already locked"))
		return
	}

	// Validate skrew caller (required)
	skrewStr := strings.TrimSpace(r.FormValue("skrew_caller"))
	if skrewStr == "" {
		render(w, r, views.Round(g, cur, g.ID, "You must select who called Skrew"))
		return
	}
	skrewCaller, convErr := strconv.Atoi(skrewStr)
	if convErr != nil || skrewCaller < 0 || skrewCaller >= len(g.Teams) {
		render(w, r, views.Round(g, cur, g.ID, "Invalid Skrew caller selection"))
		return
	}

	// Parse scores
	entries := make([]game.RoundEntry, len(g.Teams))
	for i, team := range g.Teams {
		playerCount := len(team.Players)
		rawScores := make([]float64, playerCount)
		for pi := 0; pi < playerCount; pi++ {
			key := fmt.Sprintf("score_%d_%d", i, pi)
			val := strings.TrimSpace(r.FormValue(key))
			v, evalErr := eval.Expr(val)
			if evalErr != nil {
				render(w, r, views.Round(g, cur, g.ID,
					fmt.Sprintf("Invalid score for %s: \"%s\"", team.Players[pi].Name, val)))
				return
			}
			rawScores[pi] = v
		}

		var teamScore float64
		if g.SoloMode || len(rawScores) == 1 {
			teamScore = rawScores[0]
		} else if cur.Mode == game.ModeAverage {
			sum := 0.0
			for _, s := range rawScores {
				sum += s
			}
			teamScore = sum / float64(len(rawScores))
		} else {
			for _, s := range rawScores {
				teamScore += s
			}
		}

		entries[i] = game.RoundEntry{
			RawScores:   rawScores,
			TeamScore:   teamScore,
			CalledSkrew: i == skrewCaller,
			Mode:        cur.Mode,
		}
	}

	// Find minimum team score
	minScore := math.Inf(1)
	for _, e := range entries {
		if e.TeamScore < minScore {
			minScore = e.TeamScore
		}
	}

	// Compute final scores
	for i := range entries {
		e := &entries[i]
		base := e.TeamScore

		// Round 4: double the base first
		if cur.Number == 4 {
			base *= 2
		}

		// Skrew penalty: called skrew but NOT lowest → double (after round-4 multiplier)
		if e.CalledSkrew && e.TeamScore != minScore {
			base *= 2
		}

		// Lowest score(s) → 0 (including skrew caller if they won)
		if e.TeamScore == minScore {
			base = 0
		}

		e.Final = base
	}

	cur.Entries = entries
	cur.Locked = true
	cur.SkrewCaller = skrewCaller
	g.Rounds[g.CurrentRound-1] = *cur

	if g.CurrentRound < game.TotalRounds {
		g.CurrentRound++
		if err := mongodb.SaveGame(g); err != nil {
			log.Println("mongodb.SaveGame:", err)
		}
		render(w, r, views.Round(g, g.CurrentRoundData(), g.ID, ""))
	} else {
		g.Done = true
		if err := mongodb.SaveGame(g); err != nil {
			log.Println("mongodb.SaveGame:", err)
		}
		render(w, r, views.Done(g, g.ID, findWinners(g)))
	}
}
