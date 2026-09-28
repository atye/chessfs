package mock

import (
	"context"

	"github.com/atye/chessfs/internal/domain"
)

type Chess struct {
	GetUsernameFn    func(context.Context) (string, error)
	GetUsernameFnErr error

	GetCorrespondenceGamesFn    func(context.Context) ([]domain.Game, error)
	GetCorrespondenceGamesFnErr error

	GetFinishedCorrespondenceGamesFn    func(context.Context, int) ([]domain.GameResult, error)
	GetFinishedCorrespondenceGamesFnErr error

	GetGameStatusFn    func(context.Context, string, domain.BoardStyle) (domain.GameStatus, error)
	GetGameStatusFnErr error

	GetPGNFn    func(context.Context, string, ...domain.Option) (string, error)
	GetPGNFnErr error

	ChallengeRandomCorrespondenceFn    func(context.Context, string, int, bool) (string, error)
	ChallengeRandomCorrespondenceFnErr error

	ChallengeAiCorrespondenceFn    func(context.Context, string, string, int, int) (string, error)
	ChallengeAiCorrespondenceFnErr error

	MoveFn    func(context.Context, string, string) error
	MoveFnErr error

	ResignFn    func(context.Context, string) error
	ResignFnErr error

	DrawFn    func(context.Context, string, bool) error
	DrawFnErr error

	TakebackFn    func(context.Context, string, bool) error
	TakebackFnErr error

	AbortFn    func(context.Context, string) error
	AbortFnErr error

	ClaimVictoryFn    func(context.Context, string) error
	ClaimVictoryFnErr error

	ClaimDrawFn    func(context.Context, string) error
	ClaimDrawFnErr error
}

func (f *Chess) GetUsername(ctx context.Context) (string, error) {
	if f.GetUsernameFnErr != nil {
		return "", f.GetUsernameFnErr
	}
	if f.GetUsernameFn != nil {
		return f.GetUsernameFn(ctx)
	}
	return "me", nil
}

func (f *Chess) GetCorrespondenceGames(ctx context.Context) ([]domain.Game, error) {
	if f.GetCorrespondenceGamesFnErr != nil {
		return nil, f.GetCorrespondenceGamesFnErr
	}
	if f.GetCorrespondenceGamesFn != nil {
		return f.GetCorrespondenceGamesFn(ctx)
	}
	return nil, nil
}

func (f *Chess) GetFinishedCorrespondenceGames(ctx context.Context, max int) ([]domain.GameResult, error) {
	if f.GetFinishedCorrespondenceGamesFnErr != nil {
		return nil, f.GetFinishedCorrespondenceGamesFnErr
	}
	if f.GetFinishedCorrespondenceGamesFn != nil {
		return f.GetFinishedCorrespondenceGamesFn(ctx, max)
	}
	return nil, nil
}

func (f *Chess) GetGameStatus(ctx context.Context, gameID string, style domain.BoardStyle) (domain.GameStatus, error) {
	if f.GetGameStatusFnErr != nil {
		return domain.GameStatus{}, f.GetGameStatusFnErr
	}
	if f.GetGameStatusFn != nil {
		return f.GetGameStatusFn(ctx, gameID, style)
	}
	return domain.GameStatus{}, nil
}

func (f *Chess) GetPGN(ctx context.Context, gameID string, options ...domain.Option) (string, error) {
	if f.GetPGNFnErr != nil {
		return "", f.GetPGNFnErr
	}
	if f.GetPGNFn != nil {
		return f.GetPGNFn(ctx, gameID, options...)
	}
	return "", nil
}

func (f *Chess) ChallengeRandomCorrespondence(ctx context.Context, variant string, days int, rated bool) (string, error) {
	if f.ChallengeRandomCorrespondenceFnErr != nil {
		return "", f.ChallengeRandomCorrespondenceFnErr
	}
	if f.ChallengeRandomCorrespondenceFn != nil {
		return f.ChallengeRandomCorrespondenceFn(ctx, variant, days, rated)
	}
	return "game-id", nil
}

func (f *Chess) ChallengeAiCorrespondence(ctx context.Context, variant string, color string, level int, days int) (string, error) {
	if f.ChallengeAiCorrespondenceFnErr != nil {
		return "", f.ChallengeAiCorrespondenceFnErr
	}
	if f.ChallengeAiCorrespondenceFn != nil {
		return f.ChallengeAiCorrespondenceFn(ctx, variant, color, level, days)
	}
	return "game-id", nil
}

func (f *Chess) Move(ctx context.Context, gameID string, move string) error {
	if f.MoveFnErr != nil {
		return f.MoveFnErr
	}
	if f.MoveFn != nil {
		return f.MoveFn(ctx, gameID, move)
	}
	return nil
}

func (f *Chess) Resign(ctx context.Context, gameID string) error {
	if f.ResignFnErr != nil {
		return f.ResignFnErr
	}
	if f.ResignFn != nil {
		return f.ResignFn(ctx, gameID)
	}
	return nil
}

func (f *Chess) Draw(ctx context.Context, gameID string, accept bool) error {
	if f.DrawFnErr != nil {
		return f.DrawFnErr
	}
	if f.DrawFn != nil {
		return f.DrawFn(ctx, gameID, accept)
	}
	return nil
}

func (f *Chess) Takeback(ctx context.Context, gameID string, accept bool) error {
	if f.TakebackFnErr != nil {
		return f.TakebackFnErr
	}
	if f.TakebackFn != nil {
		return f.TakebackFn(ctx, gameID, accept)
	}
	return nil
}

func (f *Chess) Abort(ctx context.Context, gameID string) error {
	if f.AbortFnErr != nil {
		return f.AbortFnErr
	}
	if f.AbortFn != nil {
		return f.AbortFn(ctx, gameID)
	}
	return nil
}

func (f *Chess) ClaimVictory(ctx context.Context, gameID string) error {
	if f.ClaimVictoryFnErr != nil {
		return f.ClaimVictoryFnErr
	}
	if f.ClaimVictoryFn != nil {
		return f.ClaimVictoryFn(ctx, gameID)
	}
	return nil
}

func (f *Chess) ClaimDraw(ctx context.Context, gameID string) error {
	if f.ClaimDrawFnErr != nil {
		return f.ClaimDrawFnErr
	}
	if f.ClaimDrawFn != nil {
		return f.ClaimDrawFn(ctx, gameID)
	}
	return nil
}
