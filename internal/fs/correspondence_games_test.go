package fs

import (
	"context"
	"io"
	"log/slog"
	"reflect"
	"syscall"
	"testing"

	"github.com/atye/chessfs/internal/domain"
	"github.com/atye/chessfs/internal/domain/mock"
	"github.com/hanwen/go-fuse/v2/fs"
	"github.com/hanwen/go-fuse/v2/fuse"
)

type testRoot struct {
	fs.Inode
}

func newTestCorrespondenceDir(t *testing.T, chess domain.Chess) *CorrespondenceDir {
	t.Helper()
	root := &testRoot{}
	fs.NewNodeFS(root, nil)
	dir := &CorrespondenceDir{
		chess: chess,
		log:   slog.New(slog.NewTextHandler(io.Discard, nil)),
	}
	dir.Inode = *root.NewInode(context.Background(), dir, fs.StableAttr{Mode: syscall.S_IFDIR})
	return dir
}

func TestCorrespondenceDirReaddir(t *testing.T) {
	root := newTestCorrespondenceDir(t, &mock.Chess{})

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

	want := []string{
		"games",
		"results",
		"challenge-random",
		"challenge-ai",
	}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("Readdir() = %v, want %v", got, want)
	}
}

func TestCorrespondenceDirLookup(t *testing.T) {
	dir := newTestCorrespondenceDir(t, &mock.Chess{})

	t.Run("games", func(t *testing.T) {
		inode, errno := dir.Lookup(context.Background(), "games", &fuse.EntryOut{})
		if errno != 0 {
			t.Fatalf("Lookup() errno = %d, want 0", errno)
		}
		if inode == nil {
			t.Fatal("Lookup() returned nil inode")
		}
		if inode.Mode() != syscall.S_IFDIR {
			t.Fatalf("Lookup() mode = %d, want %d", inode.Mode(), syscall.S_IFDIR)
		}
	})

	t.Run("challenge-random", func(t *testing.T) {
		inode, errno := dir.Lookup(context.Background(), "challenge-random", &fuse.EntryOut{})
		if errno != 0 {
			t.Fatalf("Lookup() errno = %d, want 0", errno)
		}
		if inode == nil {
			t.Fatal("Lookup() returned nil inode")
		}
		if inode.Mode()&syscall.S_IFMT != syscall.S_IFREG {
			t.Fatalf("Lookup() mode = %d, want regular file type %d", inode.Mode(), syscall.S_IFREG)
		}
	})

	t.Run("challenge-ai", func(t *testing.T) {
		inode, errno := dir.Lookup(context.Background(), "challenge-ai", &fuse.EntryOut{})
		if errno != 0 {
			t.Fatalf("Lookup() errno = %d, want 0", errno)
		}
		if inode == nil {
			t.Fatal("Lookup() returned nil inode")
		}
		if inode.Mode()&syscall.S_IFMT != syscall.S_IFREG {
			t.Fatalf("Lookup() mode = %d, want regular file type %d", inode.Mode(), syscall.S_IFREG)
		}
	})

	t.Run("results", func(t *testing.T) {
		inode, errno := dir.Lookup(context.Background(), "results", &fuse.EntryOut{})
		if errno != 0 {
			t.Fatalf("Lookup() errno = %d, want 0", errno)
		}
		if inode == nil {
			t.Fatal("Lookup() returned nil inode")
		}
		if inode.Mode() != syscall.S_IFDIR {
			t.Fatalf("Lookup() mode = %d, want %d", inode.Mode(), syscall.S_IFDIR)
		}
	})

	t.Run("invalid-name", func(t *testing.T) {
		inode, errno := dir.Lookup(context.Background(), "bad-name", &fuse.EntryOut{})
		if errno != syscall.ENOENT {
			t.Fatalf("Lookup() errno = %d, want %d", errno, syscall.ENOENT)
		}
		if inode != nil {
			t.Fatal("Lookup() returned non-nil inode for invalid name")
		}
	})
}
