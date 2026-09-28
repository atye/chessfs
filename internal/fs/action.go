package fs

import (
	"context"
	"log/slog"
	"strconv"
	"strings"
	"syscall"

	"github.com/atye/chessfs/internal/domain"
	"github.com/hanwen/go-fuse/v2/fs"
	"github.com/hanwen/go-fuse/v2/fuse"
)

type ActionFile struct {
	fs.Inode
	chess   domain.Chess
	log     *slog.Logger
	gameID  string
	action  string
	content []byte
}

func (f *ActionFile) Open(ctx context.Context, flags uint32) (fh fs.FileHandle, fuseFlags uint32, errno syscall.Errno) {
	return nil, 0, 0
}

func (f *ActionFile) Getattr(
	ctx context.Context,
	fh fs.FileHandle,
	out *fuse.AttrOut,
) syscall.Errno {
	out.Mode = actionfilePermission
	out.Size = uint64(len(f.content))
	return 0
}

func (f *ActionFile) Setattr(
	ctx context.Context,
	fh fs.FileHandle,
	in *fuse.SetAttrIn,
	out *fuse.AttrOut,
) syscall.Errno {
	out.Mode = actionfilePermission
	out.Size = uint64(len(f.content))
	return 0
}

func (f *ActionFile) Read(ctx context.Context, fh fs.FileHandle, dest []byte, off int64) (fuse.ReadResult, syscall.Errno) {
	if off >= int64(len(f.content)) {
		return fuse.ReadResultData([]byte{}), 0
	}

	end := int(off) + len(dest)
	if end > len(f.content) {
		end = len(f.content)
	}

	return fuse.ReadResultData(f.content[off:end]), 0
}

func (f *ActionFile) Write(
	ctx context.Context,
	fh fs.FileHandle,
	data []byte,
	off int64,
) (uint32, syscall.Errno) {
	err := f.execute(ctx, f.action, data)
	if err != nil {
		return 0, syscall.EIO
	}

	return uint32(len(data)), 0
}

func (f *ActionFile) execute(ctx context.Context, action string, data []byte) error {
	variant := domain.DefaultVariant
	color := domain.DefaultColor
	level := domain.DefaultAiLevel
	days := domain.DefaultDays
	rated := domain.DefaultRated
	fields := strings.Fields(string(data))

	var err error
	switch action {
	case "challenge-random":
		for _, value := range fields {
			split := strings.Split(value, ":")
			if len(split) == 2 {
				switch split[0] {
				case "variant":
					variant = split[1]
				case "days":
					days, err = strconv.Atoi(split[1])
					if err != nil {
						f.log.Error("parsing days", "error", err)
						return err
					}
				case "rated":
					rated, err = strconv.ParseBool(split[1])
					if err != nil {
						f.log.Error("parsing rated", "error", err)
						return err
					}
				}
			}
		}
		_, err = f.chess.ChallengeRandomCorrespondence(ctx, variant, days, rated)
	case "challenge-ai":
		for _, value := range fields {
			split := strings.Split(value, ":")
			if len(split) == 2 {
				switch split[0] {
				case "variant":
					variant = split[1]
				case "color":
					color = split[1]
				case "level":
					level, err = strconv.Atoi(split[1])
					if err != nil {
						f.log.Error("parsing level", "error", err)
						return syscall.EINVAL
					}
				case "days":
					days, err = strconv.Atoi(split[1])
					if err != nil {
						f.log.Error("parsing days", "error", err)
						return syscall.EINVAL
					}
				}
			} else {
				f.log.Error("invalid format", "input", value)
				return syscall.EINVAL
			}
		}
		_, err = f.chess.ChallengeAiCorrespondence(ctx, variant, color, level, days)
		if err != nil {
			f.log.Error("challenging ai", "error", err)
			return syscall.EINVAL
		}
	case "move":
		err = f.chess.Move(ctx, f.gameID, strings.TrimSpace(string(data)))

	case "resign":
		err = f.chess.Resign(ctx, f.gameID)

	case "draw":
		if len(fields) < 1 {
			f.log.Error("invalid input to draw")
			return syscall.EINVAL
		}
		switch fields[0] {
		case "offer":
			err = f.chess.Draw(ctx, f.gameID, true)
		case "accept":
			err = f.chess.Draw(ctx, f.gameID, true)
		case "reject":
			err = f.chess.Draw(ctx, f.gameID, false)
		default:
			f.log.Error("invalid draw action", "action", fields[0])
			return syscall.EINVAL
		}

	case "takeback":
		if len(fields) < 1 {
			f.log.Error("invalid input to takeback")
			return syscall.EINVAL
		}
		switch fields[0] {
		case "offer":
			err = f.chess.Takeback(ctx, f.gameID, true)
		case "accept":
			err = f.chess.Takeback(ctx, f.gameID, true)
		case "reject":
			err = f.chess.Takeback(ctx, f.gameID, false)
		default:
			f.log.Error("invalid takeback action", "action", fields[0])
			return syscall.EINVAL
		}

	case "abort":
		err = f.chess.Abort(ctx, f.gameID)

	case "claim-victory":
		err = f.chess.ClaimVictory(ctx, f.gameID)

	case "claim-draw":
		err = f.chess.ClaimDraw(ctx, f.gameID)

	default:
		f.log.Error("invalid action", "action", action)
		return syscall.EINVAL
	}

	if err != nil {
		f.log.Error("executing action", "action", action, "error", err)
		return syscall.EIO
	}

	return nil
}
