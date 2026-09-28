package fs

import (
	"context"
	"fmt"
	"log/slog"
	"strings"
	"syscall"

	"github.com/atye/chessfs/internal/domain"
	"github.com/hanwen/go-fuse/v2/fs"
	"github.com/hanwen/go-fuse/v2/fuse"
)

type GamesDir struct {
	fs.Inode
	chess domain.Chess
	log   *slog.Logger
}

func (g *GamesDir) Readdir(ctx context.Context) (fs.DirStream, syscall.Errno) {
	games, err := g.chess.GetCorrespondenceGames(ctx)
	if err != nil {
		g.log.Error("getting games", "error", err.Error())
		return nil, syscall.EIO
	}

	entries := []fuse.DirEntry{}
	for _, game := range games {
		entries = append(entries, fuse.DirEntry{
			Name: fmt.Sprintf("%s(%s).%s(%s).%s", game.WhiteUsername, game.WhiteRating, game.BlackUsername, game.BlackRating, game.GameID),
			Mode: syscall.S_IFDIR,
		})
	}

	return fs.NewListDirStream(entries), 0
}

func (g *GamesDir) Lookup(
	ctx context.Context,
	name string,
	out *fuse.EntryOut,
) (*fs.Inode, syscall.Errno) {
	gameDirName := strings.Split(name, ".")
	if len(gameDirName) != 3 {
		g.log.Error("invalid game directory", "name", name)
		return nil, syscall.ENOENT
	}

	gameDir := &GameDir{
		chess:  g.chess,
		log:    g.log,
		gameID: gameDirName[2],
	}

	return g.NewInode(
		ctx,
		gameDir,
		fs.StableAttr{
			Mode: syscall.S_IFDIR,
		},
	), 0
}
