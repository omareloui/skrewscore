package views

import (
	"fmt"
	"math"

	"github.com/omareloui/skrewscore/internal/game"
)

func FormatScore(f float64) string {
	if f == math.Trunc(f) {
		return fmt.Sprintf("%.0f", f)
	}
	return fmt.Sprintf("%.1f", f)
}

func IsRound4(n int) bool { return n == 4 }

func FormatRoundNumber(n int) string {
	return fmt.Sprintf("%d", n)
}

func HasLockedRounds(rounds []game.Round) bool {
	for _, r := range rounds {
		if r.Locked {
			return true
		}
	}
	return false
}

func IsWinner(team game.Team, winners []game.Team) bool {
	for _, w := range winners {
		if w.DisplayName() == team.DisplayName() {
			return true
		}
	}
	return false
}
