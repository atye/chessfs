package fs

import (
	"context"
	"log/slog"
	"syscall"

	"github.com/atye/chessfs/internal/domain"
	"github.com/hanwen/go-fuse/v2/fs"
	"github.com/hanwen/go-fuse/v2/fuse"
)

type CorrespondenceDir struct {
	fs.Inode
	chess domain.Chess
	log   *slog.Logger
}

func (g *CorrespondenceDir) Readdir(ctx context.Context) (fs.DirStream, syscall.Errno) {
	entries := []fuse.DirEntry{}

	entries = append(entries, fuse.DirEntry{
		Name: "games",
		Mode: syscall.S_IFDIR,
	})

	entries = append(entries, fuse.DirEntry{
		Name: "results",
		Mode: syscall.S_IFDIR,
	})

	entries = append(entries, fuse.DirEntry{
		Name: "challenge-random",
		Mode: actionfilePermission,
	})

	entries = append(entries, fuse.DirEntry{
		Name: "challenge-ai",
		Mode: actionfilePermission,
	})

	return fs.NewListDirStream(entries), 0
}

func (g *CorrespondenceDir) Lookup(
	ctx context.Context,
	name string,
	out *fuse.EntryOut,
) (*fs.Inode, syscall.Errno) {
	switch name {
	case "challenge-random":
		challengeFile := &ActionFile{
			chess:   g.chess,
			log:     g.log,
			action:  "challenge-random",
			content: challengeRandomFileContent,
		}
		return g.NewInode(
			ctx,
			challengeFile,
			fs.StableAttr{
				Mode: actionfilePermission,
			},
		), 0
	case "challenge-ai":
		challengeFile := &ActionFile{
			chess:   g.chess,
			log:     g.log,
			action:  "challenge-ai",
			content: challengeAIFileContent,
		}
		return g.NewInode(
			ctx,
			challengeFile,
			fs.StableAttr{
				Mode: actionfilePermission,
			},
		), 0
	case "games":
		gamesDir := &GamesDir{
			chess: g.chess,
			log:   g.log,
		}
		return g.NewInode(
			ctx,
			gamesDir,
			fs.StableAttr{
				Mode: syscall.S_IFDIR,
			},
		), 0
	case "results":
		resultsDir := &CorrespondenceResultsDir{
			chess: g.chess,
			log:   g.log,
		}
		return g.NewInode(
			ctx,
			resultsDir,
			fs.StableAttr{
				Mode: syscall.S_IFDIR,
			},
		), 0
	}
	return nil, syscall.ENOENT
}
