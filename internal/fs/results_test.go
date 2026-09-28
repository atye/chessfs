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

func newTestResultsDir(chess domain.Chess) *CorrespondenceResultsDir {
	parent := &testRootNode{}
	fs.NewNodeFS(parent, nil)
	dir := &CorrespondenceResultsDir{chess: chess, log: newTestLogger()}
	dir.Inode = *parent.NewInode(context.Background(), dir, fs.StableAttr{Mode: syscall.S_IFDIR})
	return dir
}

func TestCorrespondenceResultsDirReaddir(t *testing.T) {
	chess := &mock.Chess{
		GetFinishedCorrespondenceGamesFn: func(ctx context.Context, max int) ([]domain.GameResult, error) {
			return []domain.GameResult{{
				GameID:      "game-1",
				White:       "Alice",
				WhiteRating: 1500,
				Black:       "Bob",
				BlackRating: 1700,
			}}, nil
		},
	}
	dir := newTestResultsDir(chess)

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
	want := []string{"Alice(1500).Bob(1700).game-1"}
	if !reflect.DeepEqual(names, want) {
		t.Fatalf("Readdir() names = %v, want %v", names, want)
	}
}

func TestCorrespondenceResultsDirLookup(t *testing.T) {
	chess := &mock.Chess{
		GetFinishedCorrespondenceGamesFn: func(ctx context.Context, max int) ([]domain.GameResult, error) {
			return []domain.GameResult{{
				GameID:      "game-1",
				White:       "Alice",
				WhiteRating: 1500,
				Black:       "Bob",
				BlackRating: 1700,
			}}, nil
		},
		GetPGNFn: func(ctx context.Context, gameID string, options ...domain.Option) (string, error) {
			return "[Event \"test\"]\n", nil
		},
	}
	dir := newTestResultsDir(chess)

	inode, errno := dir.Lookup(context.Background(), "Alice(1500).Bob(1700).game-1", &fuse.EntryOut{})
	if errno != 0 {
		t.Fatalf("Lookup() errno = %d, want 0", errno)
	}
	if inode == nil {
		t.Fatal("Lookup() returned nil inode")
	}
	if inode.Mode()&syscall.S_IFMT != syscall.S_IFREG {
		t.Fatalf("Lookup() mode = %d, want regular file type %d", inode.Mode(), syscall.S_IFREG)
	}
	resultFile, ok := inode.Operations().(*ResultFile)
	if !ok {
		t.Fatalf("Lookup() returned %T, want *ResultFile", inode.Operations())
	}
	if resultFile.pgn != "[Event \"test\"]\n" {
		t.Fatalf("ResultFile.pgn = %q, want %q", resultFile.pgn, "[Event \"test\"]\n")
	}

	readResult, errno := resultFile.Read(context.Background(), nil, make([]byte, len("[Event \"test\"]\n")), 0)
	if errno != 0 {
		t.Fatalf("Read() errno = %d, want 0", errno)
	}
	data, status := readResult.Bytes(nil)
	if status != fuse.OK {
		t.Fatalf("Read() status = %v, want %v", status, fuse.OK)
	}
	if string(data) != "[Event \"test\"]\n" {
		t.Fatalf("Read() data = %q, want %q", string(data), "[Event \"test\"]\n")
	}

	_, errno = dir.Lookup(context.Background(), "bad-name", &fuse.EntryOut{})
	if errno != syscall.ENOENT {
		t.Fatalf("Lookup() invalid errno = %d, want %d", errno, syscall.ENOENT)
	}
}

func TestResultFileReadAndAttr(t *testing.T) {
	f := &ResultFile{pgn: "abcde"}
	attrs := &fuse.AttrOut{}
	if errno := f.Getattr(context.Background(), nil, attrs); errno != 0 {
		t.Fatalf("Getattr() errno = %d, want 0", errno)
	}
	if attrs.Mode != readFilePermission {
		t.Fatalf("Getattr() mode = %d, want %d", attrs.Mode, readFilePermission)
	}
	if attrs.Size != uint64(len(f.pgn)) {
		t.Fatalf("Getattr() size = %d, want %d", attrs.Size, len(f.pgn))
	}

	result, errno := f.Read(context.Background(), nil, make([]byte, 2), 1)
	if errno != 0 {
		t.Fatalf("Read() errno = %d, want 0", errno)
	}
	data, status := result.Bytes(nil)
	if status != fuse.OK {
		t.Fatalf("Read() status = %v, want %v", status, fuse.OK)
	}
	if string(data) != "bc" {
		t.Fatalf("Read() = %q, want %q", string(data), "bc")
	}
}

func TestResultFileLookupInvalidName(t *testing.T) {
	dir := newTestResultsDir(&mock.Chess{})
	_, errno := dir.Lookup(context.Background(), "bad", &fuse.EntryOut{})
	if errno != syscall.ENOENT {
		t.Fatalf("Lookup(bad) errno = %d, want %d", errno, syscall.ENOENT)
	}
}
