package fs

import (
	"context"
	"syscall"

	"github.com/hanwen/go-fuse/v2/fs"
	"github.com/hanwen/go-fuse/v2/fuse"
)

type StatusFile struct {
	fs.Inode
	status []byte
}

func (f *StatusFile) Open(ctx context.Context, flags uint32) (fh fs.FileHandle, fuseFlags uint32, errno syscall.Errno) {
	return nil, 0, 0
}

func (f *StatusFile) Getattr(
	ctx context.Context,
	fh fs.FileHandle,
	out *fuse.AttrOut,
) syscall.Errno {
	out.Mode = readFilePermission
	out.Size = uint64(len(f.status))
	return 0
}

func (f *StatusFile) Setattr(
	ctx context.Context,
	fh fs.FileHandle,
	in *fuse.SetAttrIn,
	out *fuse.AttrOut,
) syscall.Errno {
	out.Mode = readFilePermission
	out.Size = uint64(len(f.status))
	return 0
}

func (f *StatusFile) Read(
	ctx context.Context,
	fh fs.FileHandle,
	dest []byte,
	off int64,
) (fuse.ReadResult, syscall.Errno) {
	if off >= int64(len(f.status)) {
		return fuse.ReadResultData([]byte{}), 0
	}

	end := int(off) + len(dest)
	if end > len(f.status) {
		end = len(f.status)
	}

	return fuse.ReadResultData(f.status[off:end]), 0
}
