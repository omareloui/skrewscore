# Skrew Scorer — CLAUDE.md

## Project Overview

A web-based score tracker for the **سكرو (Skrew)** card game, created by Egyptian YouTuber Yahya Azzam.
Built with **Go + HTMX**, persisted in **MongoDB**.

### Stack

- **Backend:** Go (stdlib `net/http`, `go/parser` for expression eval)
- **Frontend:** HTMX 1.9 + vanilla CSS (no JS framework)
- **Database:** MongoDB — one collection `skrew.games`
- **Routing:** Manual router in `router()`, no external router library

### Running

```bash
cp .env.example .env
make templ # Generates the templates

make # Runs the app
```

Both env vars are optional and default to the values above.

### URL Structure

| Method | Path                          | Description                        |
| ------ | ----------------------------- | ---------------------------------- |
| GET    | `/`                           | Setup page (new game form)         |
| POST   | `/start`                      | Create game → redirect to game URL |
| POST   | `/start-new`                  | HTMX: return setup partial         |
| GET    | `/game/<uuid>`                | View/edit game in progress or done |
| POST   | `/game/<id>/toggle-loser-double` | Toggle the loser's double for the round |
| POST   | `/game/<id>/submit-round`        | Lock a round with scores                |

Each game gets a nano ID on creation. Anyone with the `/game/<id>` link has full edit access.

### Key Design Decisions

- **No sessions/auth** — the UUID in the URL is the only access control.
- **Full upsert on every mutation** — `saveGame()` does a MongoDB `ReplaceOne` with upsert. Simple, no partial updates.
- **Expression eval** — scores are parsed using Go's `go/parser` AST, supporting `+`, `-`, `*`, `/`, parentheses, and unary minus. No `eval` or unsafe exec.
- **HTMX partials vs full render** — handlers check `HX-Request` header. HTMX requests get the named partial template; direct browser navigation gets the full layout wrapper.

---

## Game Rules

### Overview

Skrew is a strategic card game where players compete to end each round with the
**lowest possible score**. The player who thinks they have the lowest score
calls **"Skrew"** to end the round — but if they're wrong, they're penalized.

---

### Setup

- Each player (or team) starts with **4 face-down cards**.
- At the very beginning, each player may **peek at exactly 2 of their own cards** — they cannot swap or rearrange them after peeking.
- A draw pile is placed face-down in the center of the table.

---

### Game Modes

The game is played either **individually** (each person for themselves) or in **teams of 2**.

**Team scoring:** the team's round score is the sum of both players' card values.

---

### Turn Structure

Players take turns in order (right to left). On each turn a player must choose one of three actions:

1. **Draw from the pile** — take the top card, then either:
   - **Keep it**: swap it with one of your face-down cards (the replaced card goes to the discard pile face-up).
   - **Discard it**: place it face-up on the discard pile without keeping it.
2. _(Special card actions — see below)_

---

### Calling Skrew

At the start of **their turn**, instead of drawing, a player may call
**"Skrew"** if they believe they hold the lowest total score at the table.

- The player who calls Skrew **skips their own turn**.
- All remaining players each get **one final turn**.
- After that, all cards are **revealed** and scores are tallied.

**Exactly one player must call Skrew each round** — the round does not end without it.

---

### Scoring a Round

1. **Compute each player/team's raw score** (sum of their card values, or sum/average for teams).
2. **Find the minimum raw score** across all players/teams.
3. **Apply round modifiers (in order):**
   - **Double round** (configurable, default round 4) — multiply the raw score by 2.
   - **Loser's Double** (available from round 3) — the player/team with the highest score from the previous round may choose to double the current round. Stacks with the double round (→ ×4).
4. **Apply the Skrew penalty:**
   - If the Skrew caller's raw score **equals the minimum** → they score **0** (they win the round).
   - If the Skrew caller's raw score is **not the minimum** → their score is **doubled** (applied after the round-4 multiplier if applicable).
5. **All players/teams tied at the minimum score → 0 points** for that round (multiple winners allowed).
6. Everyone else scores their (possibly doubled) value.

#### Scoring Formula

```text
base = raw_score

if round == double_round:
    base = base * 2

if loser_doubled:
    base = base * 2       ← stacks with double round → ×4

if called_skrew AND raw_score != min_score:
    base = base * 2       ← penalty (applied after multipliers)

if raw_score == min_score:
    base = 0              ← winner(s), regardless of skrew status
```

---

### Special Cards

| Card            | Arabic      | Effect                                                                                                                                                                      |
| --------------- | ----------- | --------------------------------------------------------------------------------------------------------------------------------------------------------------------------- |
| **Basra**       | بصرة        | Discard one of your own cards of your choice.                                                                                                                               |
| **Take & Give** | خذ وهات     | Swap one of your cards blindly with an opponent's. If you peek at their card before deciding, you must keep both.                                                           |
| **Give Only**   | خذ بس       | Give one of your cards to any other player.                                                                                                                                 |
| **Peek & Swap** | عجب ما عجب  | Look at an opponent's card. If you like it, swap it with one of yours. If not, give it to another player instead.                                                           |
| **Ping / Pong** | بينج / بونج | Prevents the opposing team from playing their next turn. Can be countered by the matching card (Ping counters Pong and vice versa).                                         |
| **Harami**      | الحرامي     | Cannot be discarded if drawn from the pile — must be kept. If a player calls Skrew and holds the Harami at reveal, the **Harami holder becomes the one penalized** instead. |

> Special cards can only be activated when drawn from the draw pile — not when swapped from another player.

---

### Game Modes (Card Variants)

| Mode             | Cards Included                                    |
| ---------------- | ------------------------------------------------- |
| **Classic**      | All cards except Harami, Ping, and Pong           |
| **Harami Mode**  | All cards except Ping and Pong (Harami is active) |
| **Doubles Mode** | All cards except Harami (Ping/Pong are active)    |

---

### Winning the Game

- The game is played over **5 rounds**.
- The overall score is the **sum of all round scores**.
- The player/team with the **lowest total score** at the end of 5 rounds wins.
- Ties for lowest total are allowed — multiple winners possible.
