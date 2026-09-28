package fs

import (
	"context"
	"fmt"
	"log/slog"
	"strings"
	"syscall"

	"github.com/atye/chessfs/internal/domain"
	"github.com/atye/chessfs/internal/env"
	"github.com/hanwen/go-fuse/v2/fs"
	"github.com/hanwen/go-fuse/v2/fuse"
)

type CorrespondenceResultsDir struct {
	fs.Inode
	chess domain.Chess
	log   *slog.Logger
}

func (r *CorrespondenceResultsDir) Readdir(ctx context.Context) (fs.DirStream, syscall.Errno) {
	games, err := r.chess.GetFinishedCorrespondenceGames(ctx, env.GetMaxFinishedCorrespondenceGames())
	if err != nil {
		r.log.Error("getting finished correspondence games", "error", err)
		return nil, syscall.EIO
	}

	entries := make([]fuse.DirEntry, 0, len(games))
	for _, game := range games {
		entries = append(entries, fuse.DirEntry{
			Name: fmt.Sprintf("%s(%d).%s(%d).%s", game.White, game.WhiteRating, game.Black, game.BlackRating, game.GameID),
			Mode: readFilePermission,
		})
	}

	return fs.NewListDirStream(entries), 0
}

func (r *CorrespondenceResultsDir) Lookup(ctx context.Context, name string, out *fuse.EntryOut) (*fs.Inode, syscall.Errno) {
	parts := strings.Split(name, ".")
	if len(parts) != 3 {
		r.log.Error("invalid game result directory", "name", name)
		return nil, syscall.ENOENT
	}

	gameID := parts[2]

	pgn, err := r.chess.GetPGN(ctx, gameID, domain.UseCache())
	if err != nil {
		r.log.Error("getting game pgn", "error", err)
		return nil, syscall.EIO
	}

	resultFile := &ResultFile{
		pgn: pgn,
	}

	return r.NewInode(
		ctx,
		resultFile,
		fs.StableAttr{
			Mode: readFilePermission,
		},
	), 0
}

type ResultFile struct {
	fs.Inode
	pgn string
}

func (f *ResultFile) Open(ctx context.Context, flags uint32) (fh fs.FileHandle, fuseFlags uint32, errno syscall.Errno) {
	return nil, 0, 0
}

func (f *ResultFile) Getattr(
	ctx context.Context,
	fh fs.FileHandle,
	out *fuse.AttrOut,
) syscall.Errno {
	out.Mode = readFilePermission
	out.Size = uint64(len(f.pgn))
	return 0
}

func (f *ResultFile) Read(
	ctx context.Context,
	fh fs.FileHandle,
	dest []byte,
	off int64,
) (fuse.ReadResult, syscall.Errno) {
	pgn := []byte(f.pgn)

	if off >= int64(len(pgn)) {
		return fuse.ReadResultData([]byte{}), 0
	}

	end := int(off) + len(dest)
	if end > len(pgn) {
		end = len(pgn)
	}

	return fuse.ReadResultData(pgn[off:end]), 0
}
