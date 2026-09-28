package lichess

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/atye/chessfs/internal/domain"
	oapicodegen "github.com/atye/golichess/oapi-codegen"
)

func newTestOAPICODEGEN(t *testing.T, handler http.Handler) *OAPICODEGEN {
	t.Helper()

	server := httptest.NewServer(handler)
	t.Cleanup(server.Close)

	client, err := oapicodegen.NewClient(server.URL, oapicodegen.WithHTTPClient(server.Client()))
	if err != nil {
		t.Fatalf("new client: %v", err)
	}

	responseClient, err := oapicodegen.NewClientWithResponses(server.URL, oapicodegen.WithHTTPClient(server.Client()))
	if err != nil {
		t.Fatalf("new response client: %v", err)
	}

	return &OAPICODEGEN{Client: client, ResponseClient: responseClient}
}

func TestGetUsername(t *testing.T) {
	client := newTestOAPICODEGEN(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/api/account" {
			http.Error(w, fmt.Sprintf("unexpected path: %s", r.URL.Path), http.StatusNotFound)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"id":"user-1","username":"alice","url":"https://lichess.org/@/alice"}`))
	}))

	got, err := client.GetUsername(context.Background())
	if err != nil {
		t.Fatalf("GetUsername returned error: %v", err)
	}
	if got != "alice" {
		t.Fatalf("GetUsername = %q, want %q", got, "alice")
	}
}

func TestGetCorrespondenceGames(t *testing.T) {
	client := newTestOAPICODEGEN(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/api/account":
			w.Header().Set("Content-Type", "application/json")
			_, _ = w.Write([]byte(`{"id":"user-1","username":"alice","url":"https://lichess.org/@/alice"}`))
		case "/api/account/playing":
			w.Header().Set("Content-Type", "application/json")
			_, _ = w.Write([]byte(`{"nowPlaying":[
				{"gameId":"game-white-user","color":"white","perf":"correspondence","speed":"correspondence","rating":1500,"opponent":{"id":"op-1","username":"bob","rating":1700}},
				{"gameId":"game-black-user","color":"black","perf":"correspondence","speed":"correspondence","rating":1400,"opponent":{"id":"op-2","username":"carol","rating":1600}},
				{"gameId":"game-white-ai","color":"white","perf":"correspondence","speed":"correspondence","rating":1550,"opponent":{"ai":3,"username":"Stockfish"}},
				{"gameId":"game-black-ai","color":"black","perf":"correspondence","speed":"correspondence","rating":1450,"opponent":{"ai":6,"username":"Stockfish"}},
				{"gameId":"game-anon","color":"white","perf":"correspondence","speed":"correspondence","rating":1350,"opponent":{"id":null,"username":"anonymous"}},
				{"gameId":"game-chess960-correspondence","color":"white","perf":"chess960","speed":"correspondence","rating":1600,"opponent":{"id":"op-3","username":"dave","rating":1800}}
			]}`))
		default:
			http.Error(w, fmt.Sprintf("unexpected path: %s", r.URL.Path), http.StatusNotFound)
		}
	}))

	games, err := client.GetCorrespondenceGames(context.Background())
	if err != nil {
		t.Fatalf("GetCorrespondenceGames returned error: %v", err)
	}
	if len(games) != 6 {
		t.Fatalf("len(games) = %d, want 6", len(games))
	}

	got := map[string]domain.Game{}
	for _, game := range games {
		got[game.GameID] = game
	}

	if got["game-white-user"].WhiteUsername != "alice" || got["game-white-user"].BlackUsername != "bob" || got["game-white-user"].WhiteRating != "1500" || got["game-white-user"].BlackRating != "1700" {
		t.Fatalf("white-user game mismatch: %#v", got["game-white-user"])
	}
	if got["game-black-user"].WhiteUsername != "carol" || got["game-black-user"].BlackUsername != "alice" || got["game-black-user"].WhiteRating != "1600" || got["game-black-user"].BlackRating != "1400" {
		t.Fatalf("black-user game mismatch: %#v", got["game-black-user"])
	}
	if got["game-white-ai"].WhiteUsername != "alice" || got["game-white-ai"].BlackUsername != "Stockfish" || got["game-white-ai"].BlackRating != "3" {
		t.Fatalf("white-ai game mismatch: %#v", got["game-white-ai"])
	}
	if got["game-black-ai"].WhiteUsername != "Stockfish" || got["game-black-ai"].BlackUsername != "alice" || got["game-black-ai"].WhiteRating != "6" {
		t.Fatalf("black-ai game mismatch: %#v", got["game-black-ai"])
	}
	if got["game-anon"].BlackUsername != "anonymous" || got["game-anon"].WhiteUsername != "alice" {
		t.Fatalf("anonymous game mismatch: %#v", got["game-anon"])
	}
	if got["game-chess960-correspondence"].WhiteUsername != "alice" || got["game-chess960-correspondence"].BlackUsername != "dave" || got["game-chess960-correspondence"].WhiteRating != "1600" || got["game-chess960-correspondence"].BlackRating != "1800" {
		t.Fatalf("chess960 correspondence game mismatch: %#v", got["game-chess960-correspondence"])
	}
}

func TestGetGameStatus(t *testing.T) {
	t.Run("white-turn styled winner white", func(t *testing.T) {
		client := newTestOAPICODEGEN(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			switch r.URL.Path {
			case "/api/account":
				w.Header().Set("Content-Type", "application/json")
				_, _ = w.Write([]byte(`{"id":"user-1","username":"alice","url":"https://lichess.org/@/alice"}`))
			case "/api/board/game/stream/game-1":
				w.Header().Set("Content-Type", "application/x-ndjson")
				_, _ = w.Write([]byte("{\"type\":\"gameFull\",\"id\":\"game-1\",\"white\":{\"id\":\"w-1\",\"aiLevel\":3},\"black\":{\"id\":\"b-1\",\"name\":\"bob\"},\"state\":{\"moves\":\"e2e4 e7e5\",\"wtime\":120000,\"btime\":240000,\"wdraw\":true,\"wtakeback\":true,\"winner\":\"white\"},\"variant\":{\"key\":\"standard\",\"name\":\"Standard\"},\"rated\":false}\n"))
			default:
				http.Error(w, fmt.Sprintf("unexpected path: %s", r.URL.Path), http.StatusNotFound)
			}
		}))

		status, err := client.GetGameStatus(context.Background(), "game-1", domain.Styled)
		if err != nil {
			t.Fatalf("GetGameStatus returned error: %v", err)
		}
		if status.Turn != "AI (White)" {
			t.Fatalf("Turn = %q, want %q", status.Turn, "AI (White)")
		}
		if status.Clock != 120000*time.Millisecond {
			t.Fatalf("Clock = %v, want %v", status.Clock, 120000*time.Millisecond)
		}
		if !status.WhiteOfferingDraw || !status.WhiteOfferingTakeback {
			t.Fatalf("offer flags mismatch: draw=%v takeback=%v", status.WhiteOfferingDraw, status.WhiteOfferingTakeback)
		}
		if status.Winner != "AI" {
			t.Fatalf("Winner = %q, want %q", status.Winner, "AI")
		}
		if status.Board == "" {
			t.Fatal("Board is empty")
		}
	})

	t.Run("black-turn text winner black", func(t *testing.T) {
		client := newTestOAPICODEGEN(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			switch r.URL.Path {
			case "/api/account":
				w.Header().Set("Content-Type", "application/json")
				_, _ = w.Write([]byte(`{"id":"user-1","username":"alice","url":"https://lichess.org/@/alice"}`))
			case "/api/board/game/stream/game-2":
				w.Header().Set("Content-Type", "application/x-ndjson")
				_, _ = w.Write([]byte("{\"type\":\"gameFull\",\"id\":\"game-2\",\"white\":{\"id\":\"w-1\",\"name\":\"alice\"},\"black\":{\"id\":\"b-1\",\"aiLevel\":4},\"state\":{\"moves\":\"e2e4\",\"wtime\":100000,\"btime\":90000,\"bdraw\":true,\"btakeback\":true,\"winner\":\"black\"},\"variant\":{\"key\":\"standard\",\"name\":\"Standard\"},\"rated\":false}\n"))
			default:
				http.Error(w, fmt.Sprintf("unexpected path: %s", r.URL.Path), http.StatusNotFound)
			}
		}))

		status, err := client.GetGameStatus(context.Background(), "game-2", domain.Text)
		if err != nil {
			t.Fatalf("GetGameStatus returned error: %v", err)
		}
		if status.Turn != "AI (Black)" {
			t.Fatalf("Turn = %q, want %q", status.Turn, "AI (Black)")
		}
		if status.Clock != 90000*time.Millisecond {
			t.Fatalf("Clock = %v, want %v", status.Clock, 90000*time.Millisecond)
		}
		if !status.BlackOfferingDraw || !status.BlackOfferingTakeback {
			t.Fatalf("offer flags mismatch: draw=%v takeback=%v", status.BlackOfferingDraw, status.BlackOfferingTakeback)
		}
		if status.Winner != "AI" {
			t.Fatalf("Winner = %q, want %q", status.Winner, "AI")
		}
		if status.Board == "" {
			t.Fatal("Board is empty")
		}
	})
}

func TestGetFinishedCorrespondenceGames(t *testing.T) {
	client := newTestOAPICODEGEN(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/api/account":
			w.Header().Set("Content-Type", "application/json")
			_, _ = w.Write([]byte(`{"id":"user-1","username":"alice","url":"https://lichess.org/@/alice"}`))
		case "/api/games/user/alice":
			w.Header().Set("Content-Type", "application/x-ndjson")
			_, _ = w.Write([]byte("{\"id\":\"g-white-ai\",\"perf\":\"chess960\",\"speed\":\"correspondence\",\"players\":{\"white\":{\"aiLevel\":3},\"black\":{\"user\":{\"name\":\"bob\",\"id\":\"u2\"},\"rating\":1700}},\"winner\":\"white\",\"pgn\":\"1. e4 e5 1-0\"}\n{\"id\":\"g-black-ai\",\"perf\":\"chess960\",\"speed\":\"correspondence\",\"players\":{\"white\":{\"user\":{\"name\":\"alice\",\"id\":\"u1\"},\"rating\":1600},\"black\":{\"aiLevel\":5}},\"winner\":\"black\",\"pgn\":\"1. e4 e5 1-0\"}\n{\"id\":\"g-user-vs-user\",\"perf\":\"chess960\",\"speed\":\"correspondence\",\"players\":{\"white\":{\"user\":{\"name\":\"alice\",\"id\":\"u1\"},\"rating\":1600},\"black\":{\"user\":{\"name\":\"bob\",\"id\":\"u2\"},\"rating\":1700}},\"winner\":\"black\",\"pgn\":\"1. e4 e5 1-0\"}\n"))
		default:
			http.Error(w, fmt.Sprintf("unexpected path: %s", r.URL.Path), http.StatusNotFound)
		}
	}))

	games, err := client.GetFinishedCorrespondenceGames(context.Background(), 5)
	if err != nil {
		t.Fatalf("GetFinishedCorrespondenceGames returned error: %v", err)
	}
	if len(games) != 3 {
		t.Fatalf("len(games) = %d, want 3", len(games))
	}

	got := map[string]domain.GameResult{}
	for _, game := range games {
		got[game.GameID] = game
	}

	if got["g-white-ai"].White != AI || got["g-white-ai"].Black != "bob" || got["g-white-ai"].WhiteRating != 3 || got["g-white-ai"].Winner != "white" {
		t.Fatalf("white-ai result mismatch: %#v", got["g-white-ai"])
	}
	if got["g-black-ai"].Black != AI || got["g-black-ai"].White != "alice" || got["g-black-ai"].BlackRating != 5 || got["g-black-ai"].Winner != "black" {
		t.Fatalf("black-ai result mismatch: %#v", got["g-black-ai"])
	}
	if got["g-user-vs-user"].White != "alice" || got["g-user-vs-user"].Black != "bob" || got["g-user-vs-user"].WhiteRating != 1600 || got["g-user-vs-user"].BlackRating != 1700 || got["g-user-vs-user"].Winner != "black" {
		t.Fatalf("user-user result mismatch: %#v", got["g-user-vs-user"])
	}
}

func TestGetFinishedCorrespondenceGames_Draw(t *testing.T) {
	t.Run("draw uses Draw when winner is absent", func(t *testing.T) {
		client := newTestOAPICODEGEN(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			switch r.URL.Path {
			case "/api/account":
				w.Header().Set("Content-Type", "application/json")
				_, _ = w.Write([]byte(`{"id":"user-1","username":"alice","url":"https://lichess.org/@/alice"}`))
			case "/api/games/user/alice":
				w.Header().Set("Content-Type", "application/x-ndjson")
				_, _ = w.Write([]byte("{\"id\":\"g-draw\",\"perf\":\"chess960\",\"speed\":\"correspondence\",\"players\":{\"white\":{\"user\":{\"name\":\"alice\",\"id\":\"u1\"},\"rating\":1600},\"black\":{\"user\":{\"name\":\"bob\",\"id\":\"u2\"},\"rating\":1700}},\"pgn\":\"1. e4 e5 1/2-1/2\"}\n"))
			default:
				http.Error(w, fmt.Sprintf("unexpected path: %s", r.URL.Path), http.StatusNotFound)
			}
		}))

		games, err := client.GetFinishedCorrespondenceGames(context.Background(), 5)
		if err != nil {
			t.Fatalf("GetFinishedCorrespondenceGames returned error: %v", err)
		}
		if len(games) != 1 {
			t.Fatalf("len(games) = %d, want 1", len(games))
		}
		if games[0].Winner != "Draw" {
			t.Fatalf("Winner = %q, want %q", games[0].Winner, "Draw")
		}
	})
}

func TestGetPGNUsesCache(t *testing.T) {
	calls := 0
	client := newTestOAPICODEGEN(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/game/export/game-1":
			calls++
			w.Header().Set("Content-Type", "application/json")
			_, _ = w.Write([]byte(`{"id":"game-1","pgn":"1. e4 e5 1-0"}`))
		default:
			http.Error(w, fmt.Sprintf("unexpected path: %s", r.URL.Path), http.StatusNotFound)
		}
	}))

	first, err := client.GetPGN(context.Background(), "game-1")
	if err != nil {
		t.Fatalf("first GetPGN returned error: %v", err)
	}
	second, err := client.GetPGN(context.Background(), "game-1", domain.UseCache())
	if err != nil {
		t.Fatalf("second GetPGN returned error: %v", err)
	}
	if first != second {
		t.Fatalf("GetPGN mismatch: %q != %q", first, second)
	}
	if calls != 1 {
		t.Fatalf("GetPGN call count = %d, want 1 (cache used)", calls)
	}
	if !strings.Contains(first, "1. e4 e5 1-0") {
		t.Fatalf("PGN = %q, want to contain %q", first, "1. e4 e5 1-0")
	}
}

func TestChallengeRandomCorrespondence(t *testing.T) {
	client := newTestOAPICODEGEN(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost || r.URL.Path != "/api/board/seek" {
			http.Error(w, fmt.Sprintf("unexpected path: %s", r.URL.Path), http.StatusNotFound)
			return
		}
		if err := r.ParseForm(); err != nil {
			t.Fatalf("ParseForm: %v", err)
		}
		if got := r.Form.Get("variant"); got != "standard" {
			t.Fatalf("variant = %q, want %q", got, "standard")
		}
		if got := r.Form.Get("days"); got != "3" {
			t.Fatalf("days = %q, want %q", got, "3")
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"id":"seek-1"}`))
	}))

	id, err := client.ChallengeRandomCorrespondence(context.Background(), "standard", 3, true)
	if err != nil {
		t.Fatalf("ChallengeRandomCorrespondence returned error: %v", err)
	}
	if id != "seek-1" {
		t.Fatalf("id = %q, want %q", id, "seek-1")
	}
}

func TestChallengeAiCorrespondence(t *testing.T) {
	client := newTestOAPICODEGEN(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost || r.URL.Path != "/api/challenge/ai" {
			http.Error(w, fmt.Sprintf("unexpected path: %s", r.URL.Path), http.StatusNotFound)
			return
		}
		if got := r.Header.Get("Content-Type"); !strings.Contains(got, "application/x-www-form-urlencoded") {
			t.Fatalf("Content-Type = %q, want application/x-www-form-urlencoded", got)
		}
		if err := r.ParseForm(); err != nil {
			t.Fatalf("ParseForm: %v", err)
		}
		if got := r.Form.Get("variant"); got != "standard" {
			t.Fatalf("variant = %q, want %q", got, "standard")
		}
		if got := r.Form.Get("color"); got != "white" {
			t.Fatalf("color = %q, want %q", got, "white")
		}
		if got := r.Form.Get("level"); got != "5" {
			t.Fatalf("level = %q, want %q", got, "5")
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusCreated)
		_, _ = w.Write([]byte(`{"id":"ai-game-1"}`))
	}))

	id, err := client.ChallengeAiCorrespondence(context.Background(), "standard", "white", 5, 3)
	if err != nil {
		t.Fatalf("ChallengeAiCorrespondence returned error: %v", err)
	}
	if id != "ai-game-1" {
		t.Fatalf("id = %q, want %q", id, "ai-game-1")
	}
}

func TestOMove(t *testing.T) {
	client := newTestOAPICODEGEN(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost || r.URL.Path != "/api/board/game/game-1/move/e4" {
			http.Error(w, fmt.Sprintf("unexpected path: %s", r.URL.Path), http.StatusNotFound)
			return
		}
		w.WriteHeader(http.StatusOK)
	}))

	if err := client.Move(context.Background(), "game-1", "e4"); err != nil {
		t.Fatalf("Move returned error: %v", err)
	}
}

func TestResign(t *testing.T) {
	client := newTestOAPICODEGEN(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost || r.URL.Path != "/api/board/game/game-1/resign" {
			http.Error(w, fmt.Sprintf("unexpected path: %s", r.URL.Path), http.StatusNotFound)
			return
		}
		w.WriteHeader(http.StatusOK)
	}))

	if err := client.Resign(context.Background(), "game-1"); err != nil {
		t.Fatalf("Resign returned error: %v", err)
	}
}

func TestDraw(t *testing.T) {
	client := newTestOAPICODEGEN(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost || r.URL.Path != "/api/board/game/game-1/draw/true" {
			http.Error(w, fmt.Sprintf("unexpected path: %s", r.URL.Path), http.StatusNotFound)
			return
		}
		w.WriteHeader(http.StatusOK)
	}))

	if err := client.Draw(context.Background(), "game-1", true); err != nil {
		t.Fatalf("Draw returned error: %v", err)
	}
}

func TestTakeback(t *testing.T) {
	client := newTestOAPICODEGEN(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost || r.URL.Path != "/api/board/game/game-1/takeback/false" {
			http.Error(w, fmt.Sprintf("unexpected path: %s", r.URL.Path), http.StatusNotFound)
			return
		}
		w.WriteHeader(http.StatusOK)
	}))

	if err := client.Takeback(context.Background(), "game-1", false); err != nil {
		t.Fatalf("Takeback returned error: %v", err)
	}
}

func TestAbort(t *testing.T) {
	client := newTestOAPICODEGEN(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost || r.URL.Path != "/api/board/game/game-1/abort" {
			http.Error(w, fmt.Sprintf("unexpected path: %s", r.URL.Path), http.StatusNotFound)
			return
		}
		w.WriteHeader(http.StatusOK)
	}))

	if err := client.Abort(context.Background(), "game-1"); err != nil {
		t.Fatalf("Abort returned error: %v", err)
	}
}

func TestClaimVictory(t *testing.T) {
	client := newTestOAPICODEGEN(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost || r.URL.Path != "/api/board/game/game-1/claim-victory" {
			http.Error(w, fmt.Sprintf("unexpected path: %s", r.URL.Path), http.StatusNotFound)
			return
		}
		w.WriteHeader(http.StatusOK)
	}))

	if err := client.ClaimVictory(context.Background(), "game-1"); err != nil {
		t.Fatalf("ClaimVictory returned error: %v", err)
	}
}

func TestClaimDraw(t *testing.T) {
	client := newTestOAPICODEGEN(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost || r.URL.Path != "/api/board/game/game-1/claim-draw" {
			http.Error(w, fmt.Sprintf("unexpected path: %s", r.URL.Path), http.StatusNotFound)
			return
		}
		w.WriteHeader(http.StatusOK)
	}))

	if err := client.ClaimDraw(context.Background(), "game-1"); err != nil {
		t.Fatalf("ClaimDraw returned error: %v", err)
	}
}
