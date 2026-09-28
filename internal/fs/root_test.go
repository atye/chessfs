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

func newTestLogger() *slog.Logger {
	return slog.New(slog.NewTextHandler(io.Discard, nil))
}

type testRootNode struct {
	fs.Inode
}

func newTestRoot(chess domain.Chess) *Root {
	parent := &testRootNode{}
	fs.NewNodeFS(parent, nil)
	root := NewRoot(chess, newTestLogger())
	root.Inode = *parent.NewInode(context.Background(), root, fs.StableAttr{Mode: syscall.S_IFDIR})
	return root
}

func TestRootReaddir(t *testing.T) {
	root := newTestRoot(&mock.Chess{})

	stream, errno := root.Readdir(context.Background())
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
	if !reflect.DeepEqual(names, []string{"correspondence"}) {
		t.Fatalf("Readdir() names = %v, want [correspondence]", names)
	}
}

func TestRootLookup(t *testing.T) {
	root := newTestRoot(&mock.Chess{})

	inode, errno := root.Lookup(context.Background(), "correspondence", &fuse.EntryOut{})
	if errno != 0 {
		t.Fatalf("Lookup() errno = %d, want 0", errno)
	}
	if inode == nil {
		t.Fatal("Lookup() returned nil inode")
	}
	if inode.Mode() != syscall.S_IFDIR {
		t.Fatalf("Lookup() mode = %d, want %d", inode.Mode(), syscall.S_IFDIR)
	}

	inode, errno = root.Lookup(context.Background(), "missing", &fuse.EntryOut{})
	if errno != syscall.ENOENT {
		t.Fatalf("Lookup() missing errno = %d, want %d", errno, syscall.ENOENT)
	}
	if inode != nil {
		t.Fatal("Lookup() missing returned non-nil inode")
	}
}
