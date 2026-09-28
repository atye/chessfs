package fs

import (
	"context"
	"strings"
	"testing"

	"github.com/atye/chessfs/internal/domain"
	"github.com/hanwen/go-fuse/v2/fuse"
)

func TestStatusFileMethods(t *testing.T) {
	file := &StatusFile{status: []byte("hello world")}

	attrs := &fuse.AttrOut{}
	if errno := file.Getattr(context.Background(), nil, attrs); errno != 0 {
		t.Fatalf("Getattr() errno = %d, want 0", errno)
	}
	if attrs.Mode != readFilePermission {
		t.Fatalf("Getattr() mode = %d, want %d", attrs.Mode, readFilePermission)
	}
	if attrs.Size != uint64(len(file.status)) {
		t.Fatalf("Getattr() size = %d, want %d", attrs.Size, len(file.status))
	}

	attrs = &fuse.AttrOut{}
	if errno := file.Setattr(context.Background(), nil, &fuse.SetAttrIn{}, attrs); errno != 0 {
		t.Fatalf("Setattr() errno = %d, want 0", errno)
	}
	if attrs.Mode != readFilePermission {
		t.Fatalf("Setattr() mode = %d, want %d", attrs.Mode, readFilePermission)
	}

	result, errno := file.Read(context.Background(), nil, make([]byte, 5), 0)
	if errno != 0 {
		t.Fatalf("Read() errno = %d, want 0", errno)
	}
	data, status := result.Bytes(nil)
	if status != fuse.OK {
		t.Fatalf("Read() status = %v, want %v", status, fuse.OK)
	}
	if string(data) != "hello" {
		t.Fatalf("Read() = %q, want %q", string(data), "hello")
	}

	result, errno = file.Read(context.Background(), nil, make([]byte, 10), int64(len(file.status)))
	if errno != 0 {
		t.Fatalf("Read() eof errno = %d, want 0", errno)
	}
	data, status = result.Bytes(nil)
	if status != fuse.OK {
		t.Fatalf("Read() eof status = %v, want %v", status, fuse.OK)
	}
	if len(data) != 0 {
		t.Fatalf("Read() beyond EOF = %q, want empty data", string(data))
	}
}

func TestGameDirStatusFileRead(t *testing.T) {
	status := domain.GameStatus{Turn: "black", Board: "board"}
	statusFile := &StatusFile{status: []byte("Turn:\tblack\nBoard:\tboard\n")}
	result, errno := statusFile.Read(context.Background(), nil, make([]byte, 5), 0)
	if errno != 0 {
		t.Fatalf("Read() errno = %d, want 0", errno)
	}
	data, statusCode := result.Bytes(nil)
	if statusCode != fuse.OK {
		t.Fatalf("Read() status = %v, want %v", statusCode, fuse.OK)
	}
	if !strings.Contains(string(data), "Turn") {
		t.Fatalf("Read() = %q, want Turn prefix", string(data))
	}
	if status.Board != "board" {
		t.Fatalf("status.Board = %q, want %q", status.Board, "board")
	}
}
