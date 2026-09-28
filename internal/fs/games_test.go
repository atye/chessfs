package fs

import (
	"context"
	"reflect"
	"syscall"
	"testing"

	"github.com/atye/chessfs/internal/domain"
	"github.com/atye/chessfs/internal/domain/mock"
	"github.com/hanwen/go-fuse/v2/fs"
	"github.com/hanwen/go-fuse/v2/fuse"
)

func newTestGamesDir(t *testing.T, chess domain.Chess) *GamesDir {
	t.Helper()
	root := &testRoot{}
	fs.NewNodeFS(root, nil)
	dir := &GamesDir{chess: chess, log: newTestLogger()}
	dir.Inode = *root.NewInode(context.Background(), dir, fs.StableAttr{Mode: syscall.S_IFDIR})
	return dir
}

func TestGamesDirReaddir(t *testing.T) {
	games := []domain.Game{{
		WhiteUsername: "stockfish",
		WhiteRating:   "3000",
		BlackUsername: "lichess",
		BlackRating:   "3000",
		GameID:        "Qa7FJNk2",
	}}
	root := newTestGamesDir(t, &mock.Chess{
		GetCorrespondenceGamesFn: func(ctx context.Context) ([]domain.Game, error) {
			return games, nil
		},
	})

	stream, errno := root.Readdir(context.Background())
	if errno != 0 {
		t.Fatalf("Readdir returned errno %d", errno)
	}
	if stream == nil {
		t.Fatal("Readdir returned nil stream")
	}
	defer stream.Close()

	var got []string
	for stream.HasNext() {
		entry, errno := stream.Next()
		if errno != 0 {
			t.Fatalf("stream.Next returned errno %d", errno)
		}
		got = append(got, entry.Name)
	}

	want := []string{"stockfish(3000).lichess(3000).Qa7FJNk2"}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("Readdir() = %v, want %v", got, want)
	}
}

func TestGamesDirLookup(t *testing.T) {
	games := []domain.Game{{
		WhiteUsername: "stockfish",
		WhiteRating:   "3000",
		BlackUsername: "lichess",
		BlackRating:   "3000",
		GameID:        "Qa7FJNk2",
	}}
	dir := newTestGamesDir(t, &mock.Chess{
		GetCorrespondenceGamesFn: func(ctx context.Context) ([]domain.Game, error) {
			return games, nil
		},
	})

	inode, errno := dir.Lookup(context.Background(), "stockfish(3000).lichess(3000).Qa7FJNk2", &fuse.EntryOut{})
	if errno != 0 {
		t.Fatalf("Lookup() errno = %d, want 0", errno)
	}
	if inode == nil {
		t.Fatal("Lookup() returned nil inode")
	}
	if inode.Mode() != syscall.S_IFDIR {
		t.Fatalf("Lookup() mode = %d, want %d", inode.Mode(), syscall.S_IFDIR)
	}

	inode, errno = dir.Lookup(context.Background(), "bad-name", &fuse.EntryOut{})
	if errno != syscall.ENOENT {
		t.Fatalf("Lookup() errno = %d, want %d", errno, syscall.ENOENT)
	}
	if inode != nil {
		t.Fatal("Lookup() returned non-nil inode for invalid name")
	}
}
