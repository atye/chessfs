package fs

import (
	"context"
	"syscall"
	"testing"

	"github.com/atye/chessfs/internal/domain/mock"
	"github.com/hanwen/go-fuse/v2/fuse"
)

func TestActionFileMethodsAndExecute(t *testing.T) {
	chess := &mock.Chess{}
	file := &ActionFile{chess: chess, log: newTestLogger(), gameID: "game-1", action: "move", content: []byte("e2e4")}

	attrs := &fuse.AttrOut{}
	if errno := file.Getattr(context.Background(), nil, attrs); errno != 0 {
		t.Fatalf("Getattr() errno = %d, want 0", errno)
	}
	if attrs.Mode != actionfilePermission {
		t.Fatalf("Getattr() mode = %d, want %d", attrs.Mode, actionfilePermission)
	}

	attrs = &fuse.AttrOut{}
	if errno := file.Setattr(context.Background(), nil, &fuse.SetAttrIn{}, attrs); errno != 0 {
		t.Fatalf("Setattr() errno = %d, want 0", errno)
	}
	if attrs.Mode != actionfilePermission {
		t.Fatalf("Setattr() mode = %d, want %d", attrs.Mode, actionfilePermission)
	}

	result, errno := file.Read(context.Background(), nil, make([]byte, 2), 1)
	if errno != 0 {
		t.Fatalf("Read() errno = %d, want 0", errno)
	}
	data, status := result.Bytes(nil)
	if status != fuse.OK {
		t.Fatalf("Read() status = %v, want %v", status, fuse.OK)
	}
	if string(data) != "2e" {
		t.Fatalf("Read() = %q, want %q", string(data), "2e")
	}

	if err := file.execute(context.Background(), "challenge-random", []byte("variant:standard days:2 rated:true")); err != nil {
		t.Fatalf("execute(challenge-random) err = %v, want nil", err)
	}
	if err := file.execute(context.Background(), "challenge-ai", []byte("variant:standard color:white level:7 days:3")); err != nil {
		t.Fatalf("execute(challenge-ai) err = %v, want nil", err)
	}
	if err := file.execute(context.Background(), "move", []byte("e2e4")); err != nil {
		t.Fatalf("execute(move) err = %v, want nil", err)
	}
	if err := file.execute(context.Background(), "draw", []byte("offer")); err != nil {
		t.Fatalf("execute(draw) err = %v, want nil", err)
	}
	if err := file.execute(context.Background(), "takeback", []byte("reject")); err != nil {
		t.Fatalf("execute(takeback) err = %v, want nil", err)
	}
	if err := file.execute(context.Background(), "abort", nil); err != nil {
		t.Fatalf("execute(abort) err = %v, want nil", err)
	}
	if err := file.execute(context.Background(), "claim-victory", nil); err != nil {
		t.Fatalf("execute(claim-victory) err = %v, want nil", err)
	}
	if err := file.execute(context.Background(), "claim-draw", nil); err != nil {
		t.Fatalf("execute(claim-draw) err = %v, want nil", err)
	}
	if err := file.execute(context.Background(), "challenge-random", []byte("days:bad")); err == nil {
		t.Fatal("execute(challenge-random invalid) err = nil, want non-nil")
	}
	if err := file.execute(context.Background(), "draw", []byte("unknown")); err == nil {
		t.Fatal("execute(draw invalid) err = nil, want non-nil")
	}
	if err := file.execute(context.Background(), "bad-action", nil); err == nil {
		t.Fatal("execute(invalid action) err = nil, want non-nil")
	}

	file.action = "bad-action"
	written, errno := file.Write(context.Background(), nil, []byte("bad"), 0)
	if errno != syscall.EIO {
		t.Fatalf("Write() errno = %d, want %d", errno, syscall.EIO)
	}
	if written != 0 {
		t.Fatalf("Write() written = %d, want 0", written)
	}

	file.action = "move"
	written, errno = file.Write(context.Background(), nil, []byte("e2e4"), 0)
	if errno != 0 {
		t.Fatalf("Write(valid) errno = %d, want 0", errno)
	}
	if written != uint32(len("e2e4")) {
		t.Fatalf("Write(valid) written = %d, want %d", written, len("e2e4"))
	}
}

func TestActionFileReadEOF(t *testing.T) {
	f := &ActionFile{content: []byte("abc")}
	result, errno := f.Read(context.Background(), nil, make([]byte, 10), int64(len(f.content)))
	if errno != 0 {
		t.Fatalf("Read() off EOF errno = %d, want 0", errno)
	}
	data, status := result.Bytes(nil)
	if status != fuse.OK {
		t.Fatalf("Read() status = %v, want %v", status, fuse.OK)
	}
	if len(data) != 0 {
		t.Fatalf("Read() = %q, want empty result", string(data))
	}
}

func TestActionFileExecuteChallengeAIError(t *testing.T) {
	file := &ActionFile{chess: &mock.Chess{}, log: newTestLogger(), action: "challenge-ai", content: []byte("level:nope")}
	if err := file.execute(context.Background(), "challenge-ai", []byte("level:nope")); err == nil {
		t.Fatal("execute(challenge-ai invalid) err = nil, want non-nil")
	}
}

func TestActionFileSetattrAndOpen(t *testing.T) {
	file := &ActionFile{content: []byte("abc")}
	if fh, flags, errno := file.Open(context.Background(), 0); fh != nil || flags != 0 || errno != 0 {
		t.Fatalf("Open() = (%v, %d, %d), want (nil, 0, 0)", fh, flags, errno)
	}
}

func TestActionFileExecuteDrawTakebackFormats(t *testing.T) {
	chess := &mock.Chess{}
	file := &ActionFile{chess: chess, log: newTestLogger(), gameID: "g1", action: "draw"}
	if err := file.execute(context.Background(), "draw", []byte("accept")); err != nil {
		t.Fatalf("draw accept err = %v, want nil", err)
	}
	file.action = "takeback"
	if err := file.execute(context.Background(), "takeback", []byte("offer")); err != nil {
		t.Fatalf("takeback offer err = %v, want nil", err)
	}
}
