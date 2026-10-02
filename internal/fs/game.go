package fs

import (
	"bytes"
	"context"
	"fmt"
	"log/slog"
	"syscall"
	"text/tabwriter"

	"github.com/atye/chessfs/internal/domain"
	"github.com/atye/chessfs/internal/env"
	"github.com/hanwen/go-fuse/v2/fs"
	"github.com/hanwen/go-fuse/v2/fuse"
)

var actionFiles = map[string][]byte{
	"move":          moveFileContent,
	"resign":        resignFileContent,
	"draw":          drawFileContent,
	"takeback":      takebackFileContent,
	"abort":         abortFileContent,
	"claim-victory": claimVictoryFileContent,
	"claim-draw":    claimDrawFileContent,
}

type GameDir struct {
	fs.Inode
	chess  domain.Chess
	log    *slog.Logger
	gameID string
}

func (g *GameDir) Readdir(
	ctx context.Context,
) (fs.DirStream, syscall.Errno) {
	var ret []fuse.DirEntry
	for file := range actionFiles {
		ret = append(ret, fuse.DirEntry{
			Name: file, Mode: actionfilePermission,
		})
	}

	ret = append(ret, fuse.DirEntry{
		Name: "status", Mode: readFilePermission,
	})

	return fs.NewListDirStream(ret), 0
}

func (g *GameDir) Lookup(ctx context.Context, name string, out *fuse.EntryOut) (*fs.Inode, syscall.Errno) {
	content, ok := actionFiles[name]
	if ok {
		file := &ActionFile{
			chess:   g.chess,
			log:     g.log,
			gameID:  g.gameID,
			action:  name,
			content: content,
		}
		return g.NewInode(
			ctx,
			file,
			fs.StableAttr{
				Mode: actionfilePermission,
			},
		), 0
	}

	switch name {
	case "status":
		status, err := g.chess.GetGameStatus(ctx, g.gameID, env.GetBoardStyle())
		if err != nil {
			g.log.Error("getting game status", "error", err)
			return nil, syscall.EIO
		}

		var buf bytes.Buffer

		w := tabwriter.NewWriter(&buf, 0, 0, 1, ' ', 0)

		// Use \t to mark where the alignment boundary should happen
		fmt.Fprintf(w, "Turn:\t%s\n", status.Turn)
		fmt.Fprintf(w, "Time:\t%s\n", status.Clock)
		fmt.Fprintf(w, "WhiteOfferingDraw:\t%t\n", status.WhiteOfferingDraw)
		fmt.Fprintf(w, "BlackOfferingDraw:\t%t\n", status.BlackOfferingDraw)
		fmt.Fprintf(w, "WhiteOfferingTakeback:\t%t\n", status.WhiteOfferingTakeback)
		fmt.Fprintf(w, "BlackOfferingTakeback:\t%t\n\n", status.BlackOfferingTakeback)
		fmt.Fprintf(w, "%s", status.Board)

		w.Flush()

		statusFile := &StatusFile{
			status: buf.Bytes(),
		}

		return g.NewInode(
			ctx,
			statusFile,
			fs.StableAttr{
				Mode: readFilePermission,
			},
		), 0
	}
	return nil, syscall.ENOENT
}
