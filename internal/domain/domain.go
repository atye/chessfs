package domain

import (
	"context"
	"time"
)

const (
	DefaultDays                           = 3
	DefaultAiLevel                        = 6
	DefaultColor                          = "random"
	DefaultVariant                        = "standard"
	DefaultMaxFinishedCorrespondenceGames = 10
	DefaultRated                          = true
	DefaultStyle                          = Styled
	DefaultLightSquareHex                 = "#9AAABD"
	DefaultDarkSquareHex                  = "#4D5D70"
	DefaultDarkMode                       = true

	LichessTokenEnv                   = "CHESSFS_LICHESS_ACCESS_TOKEN"
	MaxFinishedCorrespondenceGamesEnv = "CHESSFS_MAX_FINISHED_CORRESPONDENCE_GAMES"
	BoardStyleEnv                     = "CHESSFS_BOARD_STYLE"
	LightSquareColorEnv               = "CHESSFS_LIGHT_SQUARE_HEX"
	DarkSquareColorEnv                = "CHESSFS_DARK_SQUARE_HEX"
	DarkModeEnv                       = "CHESSFS_DARK_MODE"

	Lichess = "https://lichess.org"

	Styled BoardStyle = "styled"
	Text   BoardStyle = "text"
)

type Chess interface {
	GetUsername(ctx context.Context) (string, error)
	GetCorrespondenceGames(ctx context.Context) ([]Game, error)
	GetFinishedCorrespondenceGames(ctx context.Context, max int) ([]GameResult, error)
	GetGameStatus(ctx context.Context, gameID string, style BoardStyle) (GameStatus, error)
	GetPGN(ctx context.Context, gameID string, options ...Option) (string, error)
	ChallengeRandomCorrespondence(ctx context.Context, variant string, days int, rated bool) (string, error)
	ChallengeAiCorrespondence(ctx context.Context, variant string, color string, level int, days int) (string, error)
	Move(ctx context.Context, gameID string, move string) error
	Resign(ctx context.Context, gameID string) error
	Draw(ctx context.Context, gameID string, accept bool) error
	Takeback(ctx context.Context, gameID string, accept bool) error
	Abort(ctx context.Context, gameID string) error
	ClaimVictory(ctx context.Context, gameID string) error
	ClaimDraw(ctx context.Context, gameID string) error
}

type Options struct {
	UseCache bool
}

type Option func(o *Options)

func UseCache() Option {
	return func(o *Options) {
		o.UseCache = true
	}
}

type Game struct {
	WhiteUsername string
	WhiteRating   string
	BlackUsername string
	BlackRating   string
	GameID        string
}

type GameStatus struct {
	Turn                  string
	Clock                 time.Duration
	BlackOfferingDraw     bool
	WhiteOfferingDraw     bool
	BlackOfferingTakeback bool
	WhiteOfferingTakeback bool
	Winner                string
	Board                 string
}

type GameResult struct {
	GameID      string
	PGN         string
	White       string
	WhiteRating int
	Black       string
	BlackRating int
	Winner      string
}

type BoardStyle string

func (s BoardStyle) Valid() bool {
	switch s {
	case Styled, Text:
		return true
	}
	return false
}
