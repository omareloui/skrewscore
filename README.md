# Skrew Scorer

A web-based score tracker for **سكرو (Skrew)**, the card game by [Yahya Azzam](https://www.youtube.com/channel/UC40wQE39COKAZV0eMrr0TEA).

## Stack

- **Backend:** Go (stdlib `net/http`)
- **Frontend:** HTMX 1.9 + vanilla CSS
- **Database:** MongoDB
- **Templating:** [Templ](https://templ.guide/)

## Getting Started

```bash
cp .env.example .env


air         # Run in dev mode (watches .templ and .go files)
# or
make templ  # Generate templates
make        # Run the app
```

Both env vars are optional and fall back to defaults.

## URL Structure

| Method | Path                          | Description                        |
| ------ | ----------------------------- | ---------------------------------- |
| GET    | `/`                           | Setup page (new game form)         |
| POST   | `/start`                      | Create game → redirect to game URL |
| POST   | `/start-new`                  | HTMX: return setup partial         |
| GET    | `/game/<uuid>`                | View/edit game in progress or done |
| POST   | `/game/<uuid>/set-round-mode` | Toggle sum/average for teams       |
| POST   | `/game/<uuid>/submit-round`   | Lock a round with scores           |

## Game Rules

### Overview

Skrew is a strategic card game where players compete to end each round with the
**lowest possible score**. The player who thinks they have the lowest score
calls **"Skrew"** to end the round — but if they're wrong, they're penalized.

### Scoring

```plain
base = raw_score

if round == 4:
    base = base * 2

if called_skrew AND raw_score != min_score:
    base = base * 2   # penalty

if raw_score == min_score:
    base = 0          # winner(s)
```

## Design Notes

- **No auth** — the UUID in the URL is the only access control.
- **Expression eval** — scores support `+`, `-`, `*`, `/`, parentheses, and unary minus, parsed via Go's `go/parser` AST.
- **HTMX partials** — handlers check the `HX-Request` header; HTMX gets a named partial, direct navigation gets the full layout.
