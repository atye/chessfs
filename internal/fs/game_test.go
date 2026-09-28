package fs

import (
	"context"
	"syscall"
	"testing"
	"time"

	"github.com/atye/chessfs/internal/domain"
	"github.com/atye/chessfs/internal/domain/mock"
	"github.com/hanwen/go-fuse/v2/fs"
	"github.com/hanwen/go-fuse/v2/fuse"
)

func newTestGameDir(chess domain.Chess, gameID string) *GameDir {
	parent := &testRootNode{}
	fs.NewNodeFS(parent, nil)
	dir := &GameDir{chess: chess, log: newTestLogger(), gameID: gameID}
	dir.Inode = *parent.NewInode(context.Background(), dir, fs.StableAttr{Mode: syscall.S_IFDIR})
	return dir
}

func TestGameDirReaddir(t *testing.T) {
	status := domain.GameStatus{
		Turn:                  "white",
		Clock:                 5 * time.Minute,
		WhiteOfferingDraw:     true,
		BlackOfferingDraw:     false,
		WhiteOfferingTakeback: true,
		BlackOfferingTakeback: false,
		Board:                 "8/8/8/8/8/8/8/8 w - - 0 1",
	}
	chess := &mock.Chess{
		GetGameStatusFn: func(ctx context.Context, gameID string, style domain.BoardStyle) (domain.GameStatus, error) {
			return status, nil
		},
	}
	dir := newTestGameDir(chess, "game-1")

	stream, errno := dir.Readdir(context.Background())
	if errno != 0 {
		t.Fatalf("Readdir() errno = %d, want 0", errno)
	}
	if stream == nil {
		t.Fatal("Readdir() returned nil stream")
	}
	defer stream.Close()

	var names []string
	for stream.HasNext() {
		entry, errno := stream.Next()
		if errno != 0 {
			t.Fatalf("Next() errno = %d", errno)
		}
		names = append(names, entry.Name)
	}
	if len(names) != len(actionFiles)+1 {
		t.Fatalf("Readdir() len(names) = %d, want %d", len(names), len(actionFiles)+1)
	}
	seen := map[string]bool{}
	for _, name := range names {
		seen[name] = true
	}
	if !seen["status"] {
		t.Fatal("Readdir() missing status entry")
	}
	for name := range actionFiles {
		if !seen[name] {
			t.Fatalf("Readdir() missing action entry %q", name)
		}
	}
}

func TestGameDirLookup(t *testing.T) {
	status := domain.GameStatus{
		Turn:                  "white",
		Clock:                 5 * time.Minute,
		WhiteOfferingDraw:     true,
		BlackOfferingDraw:     false,
		WhiteOfferingTakeback: true,
		BlackOfferingTakeback: false,
		Board:                 "8/8/8/8/8/8/8/8 w - - 0 1",
	}
	chess := &mock.Chess{
		GetGameStatusFn: func(ctx context.Context, gameID string, style domain.BoardStyle) (domain.GameStatus, error) {
			return status, nil
		},
	}
	dir := newTestGameDir(chess, "game-1")

	inode, errno := dir.Lookup(context.Background(), "move", &fuse.EntryOut{})
	if errno != 0 {
		t.Fatalf("Lookup(move) errno = %d, want 0", errno)
	}
	if inode == nil {
		t.Fatal("Lookup(move) returned nil inode")
	}
	if inode.Mode()&syscall.S_IFMT != syscall.S_IFREG {
		t.Fatalf("Lookup(move) mode = %d, want regular file type %d", inode.Mode(), syscall.S_IFREG)
	}

	statusInode, errno := dir.Lookup(context.Background(), "status", &fuse.EntryOut{})
	if errno != 0 {
		t.Fatalf("Lookup(status) errno = %d, want 0", errno)
	}
	if statusInode == nil {
		t.Fatal("Lookup(status) returned nil inode")
	}
	if statusInode.Mode()&syscall.S_IFMT != syscall.S_IFREG {
		t.Fatalf("Lookup(status) mode = %d, want regular file type %d", statusInode.Mode(), syscall.S_IFREG)
	}

	statusFile, ok := statusInode.Operations().(*StatusFile)
	if !ok {
		t.Fatalf("Lookup(status) returned %T, want *StatusFile", statusInode.Operations())
	}
	if statusFile == nil || statusFile.status == nil {
		t.Fatal("Lookup(status) returned a nil StatusFile or empty status")
	}
	if statusFile.status[0] == 0 {
		t.Fatal("status file content should not be empty")
	}

	inode, errno = dir.Lookup(context.Background(), "missing", &fuse.EntryOut{})
	if errno != syscall.ENOENT {
		t.Fatalf("Lookup(missing) errno = %d, want %d", errno, syscall.ENOENT)
	}
	if inode != nil {
		t.Fatal("Lookup(missing) returned non-nil inode")
	}
}
