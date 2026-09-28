package fs

import (
	"context"
	"log/slog"
	"syscall"

	"github.com/atye/chessfs/internal/domain"
	"github.com/hanwen/go-fuse/v2/fs"
	"github.com/hanwen/go-fuse/v2/fuse"
)

const (
	actionfilePermission = syscall.S_IFREG | 0644
	readFilePermission   = syscall.S_IFREG | 0444
)

type Root struct {
	fs.Inode
	chess domain.Chess
	log   *slog.Logger
}

func NewRoot(chess domain.Chess, log *slog.Logger) *Root {
	return &Root{
		chess: chess,
		log:   log,
	}
}

func (r *Root) Readdir(ctx context.Context) (fs.DirStream, syscall.Errno) {
	return fs.NewListDirStream([]fuse.DirEntry{
		{
			Name: "correspondence",
			Mode: syscall.S_IFDIR,
		},
	}), 0
}

func (r *Root) Lookup(
	ctx context.Context,
	name string,
	out *fuse.EntryOut,
) (*fs.Inode, syscall.Errno) {
	switch name {
	case "correspondence":
		gamesDir := &CorrespondenceDir{
			chess: r.chess,
			log:   r.log,
		}

		return r.NewInode(
			ctx,
			gamesDir,
			fs.StableAttr{
				Mode: syscall.S_IFDIR,
			},
		), 0
	}
	return nil, syscall.ENOENT
}
