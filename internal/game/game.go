package game

import (
	"strings"
	"time"
)

const TotalRounds = 5

type TeamMode string

const (
	ModeSum     TeamMode = "sum"
	ModeAverage TeamMode = "average"
)

type Player struct {
	Name string `bson:"name"`
}

type Team struct {
	Players []Player `bson:"players"`
}

func (t Team) DisplayName() string {
	names := make([]string, len(t.Players))
	for i, p := range t.Players {
		names[i] = p.Name
	}
	return strings.Join(names, " & ")
}

type RoundEntry struct {
	RawScores   []float64 `bson:"raw_scores"`
	TeamScore   float64   `bson:"team_score"`
	CalledSkrew bool      `bson:"called_skrew"`
	Mode        TeamMode  `bson:"mode"`
	Final       float64   `bson:"final"`
}

type Round struct {
	Number      int          `bson:"number"`
	Mode        TeamMode     `bson:"mode"`
	Entries     []RoundEntry `bson:"entries"`
	Locked      bool         `bson:"locked"`
	SkrewCaller int          `bson:"skrew_caller"`
}

type Game struct {
	ID           string    `bson:"_id"`
	Teams        []Team    `bson:"teams"`
	SoloMode     bool      `bson:"solo_mode"`
	Rounds       []Round   `bson:"rounds"`
	CurrentRound int       `bson:"current_round"`
	Done         bool      `bson:"done"`
	CreatedAt    time.Time `bson:"created_at"`
}

func (g *Game) TotalScore(teamIdx int) float64 {
	total := 0.0
	for _, r := range g.Rounds {
		if r.Locked && teamIdx < len(r.Entries) {
			total += r.Entries[teamIdx].Final
		}
	}
	return total
}

func (g *Game) CurrentRoundData() *Round {
	if g.CurrentRound < 1 || g.CurrentRound > len(g.Rounds) {
		return nil
	}
	return &g.Rounds[g.CurrentRound-1]
}
