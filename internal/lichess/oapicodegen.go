package lichess

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/atye/chessfs/internal/chess"
	"github.com/atye/chessfs/internal/domain"
	oapicodegen "github.com/atye/golichess/oapi-codegen"
)

var _ domain.Chess = (*OAPICODEGEN)(nil)

const (
	AI = "Stockfish"
)

type OAPICODEGEN struct {
	Client         *oapicodegen.Client
	ResponseClient *oapicodegen.ClientWithResponses
	username       string
	pgnCache       sync.Map
	uciMovesCache  sync.Map
}

func (o *OAPICODEGEN) GetUsername(ctx context.Context) (string, error) {
	if o.username == "" {
		if err := o.setUsername(ctx); err != nil {
			return "", err
		}
	}

	return o.username, nil
}

func (o *OAPICODEGEN) GetCorrespondenceGames(ctx context.Context) ([]domain.Game, error) {
	username, err := o.GetUsername(ctx)
	if err != nil {
		return nil, err
	}

	resp, err := o.ResponseClient.ApiAccountPlayingWithResponse(ctx, &oapicodegen.ApiAccountPlayingParams{})
	if err != nil {
		return nil, err
	}

	if resp.StatusCode() != http.StatusOK {
		return nil, fmt.Errorf("%d: %s", resp.StatusCode(), string(resp.GetBody()))
	}

	ret := []domain.Game{}
	for _, game := range resp.GetJSON200().NowPlaying {
		if game.Speed == oapicodegen.SpeedCorrespondence {
			var whiteUsername, blackUsername, whiteRating, blackRating string
			userWhite := true

			switch game.Color {
			case oapicodegen.GameColorWhite:
				whiteUsername = username
				whiteRating = strconv.Itoa(game.Rating)
			case oapicodegen.GameColorBlack:
				blackUsername = username
				blackRating = strconv.Itoa(game.Rating)
				userWhite = false
			}

			ai, err := game.Opponent.AsApiAccountPlaying200JSONResponseBodyNowPlayingOpponent2()
			if err != nil {
				return nil, err
			}

			if ai.Ai != 0 {
				switch userWhite {
				case true:
					blackUsername = ai.Username
					blackRating = strconv.Itoa(ai.Ai)
				case false:
					whiteUsername = ai.Username
					whiteRating = strconv.Itoa(ai.Ai)
				}
			} else {
				user, err := game.Opponent.AsApiAccountPlaying200JSONResponseBodyNowPlayingOpponent0()
				if err != nil {
					return nil, err
				}

				if user.Rating != nil {
					switch userWhite {
					case true:
						blackUsername = user.Username
						blackRating = strconv.Itoa(*user.Rating)
					case false:
						whiteUsername = user.Username
						whiteRating = strconv.Itoa(*user.Rating)
					}
				} else {
					anonymous, err := game.Opponent.AsApiAccountPlaying200JSONResponseBodyNowPlayingOpponent1()
					if err != nil {
						return nil, err
					}

					switch userWhite {
					case true:
						blackUsername = anonymous.Username
					case false:
						whiteUsername = anonymous.Username
					}
				}
			}
			ret = append(ret, domain.Game{
				WhiteUsername: whiteUsername,
				WhiteRating:   whiteRating,
				BlackUsername: blackUsername,
				BlackRating:   blackRating,
				GameID:        game.GameId,
			})
		}
	}

	return ret, nil
}

func (o *OAPICODEGEN) GetGameStatus(ctx context.Context, gameID string, style domain.BoardStyle) (domain.GameStatus, error) {
	resp, err := o.ResponseClient.BoardGameStream(ctx, gameID)
	if err != nil {
		return domain.GameStatus{}, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		b, err := io.ReadAll(resp.Body)
		if err != nil {
			return domain.GameStatus{}, err
		}
		return domain.GameStatus{}, fmt.Errorf("%d: %s", resp.StatusCode, string(b))
	}

	var event oapicodegen.GameFullEvent
	scanner := bufio.NewScanner(resp.Body)
	for scanner.Scan() {
		line := scanner.Bytes()
		if len(line) == 0 {
			continue
		}

		if err := json.Unmarshal(line, &event); err != nil {
			return domain.GameStatus{}, fmt.Errorf("unmarshaling game meta: %v", err)
		}
		break
	}

	if err := scanner.Err(); err != nil {
		return domain.GameStatus{}, fmt.Errorf("scanning response body: %v", err)
	}

	moves := strings.Fields(event.State.Moves)

	whiteName := event.White.Name
	if whiteName == "" {
		if event.White.AiLevel != nil {
			whiteName = "AI"
		}
	}

	blackName := event.Black.Name
	if blackName == "" {
		if event.Black.AiLevel != nil {
			blackName = "AI"
		}
	}

	turn := ""
	clock := 0
	whiteOfferingDraw := false
	blackOfferingDraw := false
	whiteOfferingTakeback := false
	blackOfferingTakeback := false
	if len(moves)%2 == 0 {
		turn = fmt.Sprintf("%s (%s)", whiteName, "White")
		clock = event.State.Wtime
		if event.State.Wdraw != nil {
			whiteOfferingDraw = *event.State.Wdraw
		}
		if event.State.Wtakeback != nil {
			whiteOfferingTakeback = *event.State.Wtakeback
		}
	} else {
		turn = fmt.Sprintf("%s (%s)", blackName, "Black")
		clock = event.State.Btime
		if event.State.Bdraw != nil {
			blackOfferingDraw = *event.State.Bdraw
		}
		if event.State.Btakeback != nil {
			blackOfferingTakeback = *event.State.Btakeback
		}
	}

	winner := ""
	if event.State.Winner != nil {
		winnerColor := *event.State.Winner
		switch winnerColor {
		case oapicodegen.GameColorWhite:
			winner = whiteName
		case oapicodegen.GameColorBlack:
			winner = blackName
		}
	}

	user, err := o.GetUsername(ctx)
	if err != nil {
		return domain.GameStatus{}, err
	}

	board, err := chess.Draw(event.State.Moves, strings.EqualFold(user, whiteName))
	if err != nil {
		return domain.GameStatus{}, err
	}

	return domain.GameStatus{
		Turn:                  turn,
		Clock:                 time.Duration(clock) * time.Millisecond,
		BlackOfferingDraw:     blackOfferingDraw,
		WhiteOfferingDraw:     whiteOfferingDraw,
		BlackOfferingTakeback: blackOfferingTakeback,
		WhiteOfferingTakeback: whiteOfferingTakeback,
		Board:                 board,
		Winner:                winner,
	}, nil
}

func (o *OAPICODEGEN) GetPGN(ctx context.Context, gameID string, options ...domain.Option) (string, error) {
	opts := &domain.Options{}
	for _, o := range options {
		o(opts)
	}

	if opts.UseCache {
		v, ok := o.pgnCache.Load(gameID)
		if ok {
			if pgn, ok := v.(string); ok {
				return pgn, nil
			}
			return "", fmt.Errorf("pgn for game %s not a string", gameID)
		}
	}

	resp, err := o.ResponseClient.GamePgnWithResponse(ctx, gameID, &oapicodegen.GamePgnParams{
		PgnInJson: func() *bool {
			b := true
			return &b
		}(),
		Accept: func() *oapicodegen.GamePgnParamsAccept {
			a := oapicodegen.GamePgnParamsAcceptApplicationjson
			return &a
		}(),
	})
	if err != nil {
		return "", err
	}

	if resp.StatusCode() != http.StatusOK {
		return "", fmt.Errorf("%d: %s", resp.StatusCode(), string(resp.GetBody()))
	}

	game, err := resp.JSON200.AsGameJson()
	if err != nil {
		return "", fmt.Errorf("converting response to gamepgn: %v", err)
	}

	if game.Pgn == nil {
		return "", fmt.Errorf("game %s pgn is nil", gameID)
	}

	o.pgnCache.Store(gameID, *game.Pgn)

	return *game.Pgn, nil
}

func (o *OAPICODEGEN) GetFinishedCorrespondenceGames(ctx context.Context, max int) ([]domain.GameResult, error) {
	username, err := o.GetUsername(ctx)
	if err != nil {
		return nil, err
	}

	resp, err := o.Client.ApiGamesUser(ctx, username, &oapicodegen.ApiGamesUserParams{
		Finished: func() *bool {
			b := true
			return &b
		}(),
		PerfType: func() *oapicodegen.PerfType {
			p := oapicodegen.PerfTypeCorrespondence
			return &p
		}(),
		Max: func() *int {
			if max == 0 {
				return nil
			}
			return &max
		}(),
		PgnInJson: func() *bool {
			b := true
			return &b
		}(),
		Accept: func() *oapicodegen.ApiGamesUserParamsAccept {
			a := oapicodegen.ApiGamesUserParamsAcceptApplicationxNdjson
			return &a
		}(),
	})
	if err != nil {
		return nil, err
	}

	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		b, err := io.ReadAll(resp.Body)
		if err != nil {
			return nil, err
		}
		return nil, fmt.Errorf("%d: %s", resp.StatusCode, string(b))
	}

	scanner := bufio.NewScanner(resp.Body)
	var games []oapicodegen.GameJson
	for scanner.Scan() {
		line := scanner.Bytes()
		if len(line) == 0 {
			continue
		}

		var game oapicodegen.GameJson
		if err := json.Unmarshal(line, &game); err != nil {
			return nil, fmt.Errorf("unmarshaling game JSON: %v", err)
		}

		games = append(games, game)
	}

	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("scanning response body: %v", err)
	}

	ret := []domain.GameResult{}
	for _, game := range games {
		if game.Speed == oapicodegen.SpeedCorrespondence {
			var white, black string
			var whiteRating, blackRating int

			whiteAI, err := game.Players.White.AsGamePlayerAi()
			if err != nil {
				return nil, fmt.Errorf("converting white player to AI: %v", err)
			}

			blackAI, err := game.Players.Black.AsGamePlayerAi()
			if err != nil {
				return nil, fmt.Errorf("converting black player to AI: %v", err)
			}

			if whiteAI.AiLevel > 0 {
				white = AI
				whiteRating = whiteAI.AiLevel

				blackUser, err := game.Players.Black.AsGamePlayerUser()
				if err != nil {
					return nil, fmt.Errorf("converting black player: %v", err)
				}

				black = blackUser.User.Name
				blackRating = blackUser.Rating
			} else if blackAI.AiLevel > 0 {
				black = AI
				blackRating = blackAI.AiLevel

				whiteUser, err := game.Players.White.AsGamePlayerUser()
				if err != nil {
					return nil, fmt.Errorf("converting white player: %v", err)
				}

				white = whiteUser.User.Name
				whiteRating = whiteUser.Rating
			} else {
				whiteUser, err := game.Players.White.AsGamePlayerUser()
				if err != nil {
					return nil, fmt.Errorf("converting white player: %v", err)
				}
				blackUser, err := game.Players.Black.AsGamePlayerUser()
				if err != nil {
					return nil, fmt.Errorf("converting black player: %v", err)
				}

				white = whiteUser.User.Name
				whiteRating = whiteUser.Rating
				black = blackUser.User.Name
				blackRating = blackUser.Rating
			}

			if game.Pgn == nil {
				return []domain.GameResult{}, fmt.Errorf("game %s pgn is nil", game.Id)
			}

			winner := "Draw"
			if game.Winner != nil {
				winner = string(*game.Winner)
			}

			ret = append(ret, domain.GameResult{
				GameID:      game.Id,
				PGN:         *game.Pgn,
				White:       white,
				WhiteRating: whiteRating,
				BlackRating: blackRating,
				Black:       black,
				Winner:      winner,
			})
		}
	}
	return ret, nil
}

func (o *OAPICODEGEN) ChallengeRandomCorrespondence(ctx context.Context, variant string, days int, rated bool) (string, error) {
	if d := oapicodegen.ApiBoardSeekFormdataBody1Days(days); !d.Valid() {
		return "", fmt.Errorf("invalid days: %d", days)
	}

	form := url.Values{}
	form.Set("variant", variant)
	form.Set("days", strconv.Itoa(days))
	form.Set("rated", strconv.FormatBool(rated))

	resp, err := o.ResponseClient.ApiBoardSeekWithBodyWithResponse(ctx, "application/x-www-form-urlencoded", strings.NewReader(form.Encode()))
	if err != nil {
		return "", err
	}

	if resp.StatusCode() != http.StatusOK {
		return "", fmt.Errorf("%s", string(resp.GetBody()))
	}

	return resp.JSON200.Id, nil
}

func (o *OAPICODEGEN) ChallengeAiCorrespondence(ctx context.Context, variant string, color string, level int, days int) (string, error) {
	varientKey := oapicodegen.VariantKey(variant)
	if !varientKey.Valid() {
		return "", fmt.Errorf("invalid variant: %s", variant)
	}

	challengeColor := oapicodegen.ChallengeColor(color)
	if !challengeColor.Valid() {
		return "", fmt.Errorf("invalid color: %s", color)
	}

	if level < 1 || level > 12 {
		return "", fmt.Errorf("invalid level: %d", level)
	}

	daysParam := oapicodegen.ChallengeAiFormdataBodyDays(days)
	if !daysParam.Valid() {
		return "", fmt.Errorf("invalid days: %d", days)
	}

	resp, err := o.ResponseClient.ChallengeAiWithFormdataBodyWithResponse(ctx, oapicodegen.ChallengeAiFormdataRequestBody{
		Variant: &varientKey,
		Color:   &challengeColor,
		Days:    &daysParam,
		Level:   level,
	})
	if err != nil {
		return "", err
	}

	if resp.StatusCode() != http.StatusCreated {
		return "", fmt.Errorf("%d: %s", resp.StatusCode(), string(resp.GetBody()))
	}

	if resp.JSON201 == nil {
		return "", fmt.Errorf("challenge AI response body is nil")
	}

	if resp.JSON201.Id == nil {
		return "", fmt.Errorf("challenge AI game ID is nil")
	}

	return *resp.JSON201.Id, nil
}

func (o *OAPICODEGEN) Move(ctx context.Context, gameID string, move string) error {
	resp, err := o.ResponseClient.BoardGameMoveWithResponse(ctx, gameID, move, &oapicodegen.BoardGameMoveParams{})
	if err != nil {
		return err
	}

	if resp.StatusCode() != http.StatusOK {
		return fmt.Errorf("%d: %s", resp.StatusCode(), string(resp.GetBody()))
	}

	return nil
}

func (o *OAPICODEGEN) Resign(ctx context.Context, gameID string) error {
	resp, err := o.ResponseClient.BoardGameResignWithResponse(ctx, gameID)
	if err != nil {
		return err
	}

	if resp.StatusCode() != http.StatusOK {
		return fmt.Errorf("%d: %s", resp.StatusCode(), string(resp.GetBody()))
	}

	return nil
}

func (o *OAPICODEGEN) Draw(ctx context.Context, gameID string, draw bool) error {
	param := oapicodegen.BoardGameDrawParamsAccept{}
	err := param.UnmarshalText([]byte(strconv.FormatBool(draw)))
	if err != nil {
		return err
	}

	resp, err := o.ResponseClient.BoardGameDrawWithResponse(ctx, gameID, param)
	if err != nil {
		return err
	}

	if resp.StatusCode() != http.StatusOK {
		return fmt.Errorf("%d: %s", resp.StatusCode(), string(resp.GetBody()))
	}

	return nil
}

func (o *OAPICODEGEN) Takeback(ctx context.Context, gameID string, takeback bool) error {
	param := oapicodegen.BoardGameTakebackParamsAccept{}
	err := param.UnmarshalText([]byte(strconv.FormatBool(takeback)))
	if err != nil {
		return err
	}

	resp, err := o.ResponseClient.BoardGameTakebackWithResponse(ctx, gameID, param)
	if err != nil {
		return err
	}

	if resp.StatusCode() != http.StatusOK {
		return fmt.Errorf("%d: %s", resp.StatusCode(), string(resp.GetBody()))
	}

	return nil
}

func (o *OAPICODEGEN) Abort(ctx context.Context, gameID string) error {
	resp, err := o.ResponseClient.BoardGameAbortWithResponse(ctx, gameID)
	if err != nil {
		return err
	}

	if resp.StatusCode() != http.StatusOK {
		return fmt.Errorf("%d: %s", resp.StatusCode(), string(resp.GetBody()))
	}

	return nil
}

func (o *OAPICODEGEN) ClaimVictory(ctx context.Context, gameID string) error {
	resp, err := o.ResponseClient.BoardGameClaimVictoryWithResponse(ctx, gameID)
	if err != nil {
		return err
	}

	if resp.StatusCode() != http.StatusOK {
		return fmt.Errorf("%d: %s", resp.StatusCode(), string(resp.GetBody()))
	}

	return nil
}

func (o *OAPICODEGEN) ClaimDraw(ctx context.Context, gameID string) error {
	resp, err := o.ResponseClient.BoardGameClaimDrawWithResponse(ctx, gameID)
	if err != nil {
		return err
	}

	if resp.StatusCode() != http.StatusOK {
		return fmt.Errorf("%d: %s", resp.StatusCode(), string(resp.GetBody()))
	}

	return nil
}

func (o *OAPICODEGEN) setUsername(ctx context.Context) error {
	resp, err := o.ResponseClient.AccountMeWithResponse(ctx)
	if err != nil {
		return err
	}

	if resp.StatusCode() != http.StatusOK {
		return fmt.Errorf("%d: %s", resp.StatusCode(), string(resp.GetBody()))
	}

	o.username = resp.JSON200.Username
	return nil
}
