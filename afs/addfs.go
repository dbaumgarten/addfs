package afs

import (
	"flag"
	"os"
	"path/filepath"
	"regexp"
	"time"

	"github.com/hanwen/go-fuse/fuse"
	"github.com/hanwen/go-fuse/fuse/nodefs"
	"github.com/hanwen/go-fuse/fuse/pathfs"
)

type AddFS struct {
	sourcedir string
	pathfs.FileSystem
	opts                AddFSOpts
	mutableFilesRegexes []*regexp.Regexp
}

type AddFSOpts struct {
	AllowRootMutation bool
	MutableFiles      []string
}

func NewAddFS(source string, opts AddFSOpts) (*AddFS, error) {
	fs := &AddFS{
		source,
		pathfs.NewLoopbackFileSystem(source),
		opts,
		make([]*regexp.Regexp, len(opts.MutableFiles)),
	}
	for i, v := range opts.MutableFiles {
		var err error
		fs.mutableFilesRegexes[i], err = regexp.Compile(v)
		if err != nil {
			return nil, err
		}
	}
	return fs, nil
}

func (fs *AddFS) Run() error {
	origAbs, _ := filepath.Abs(flag.Arg(0))
	mOpts := &fuse.MountOptions{
		Options:    []string{"default_permissions"},
		AllowOther: true,
		Name:       "AddFS",
		FsName:     origAbs,
		Debug:      false,
	}
	opts := &nodefs.Options{
		NegativeTimeout: time.Second,
		AttrTimeout:     time.Second,
		EntryTimeout:    time.Second,
	}
	pathNode := pathfs.NewPathNodeFs(fs, &pathfs.PathNodeFsOptions{})
	conn := nodefs.NewFileSystemConnector(pathNode.Root(), opts)
	server, err := fuse.NewServer(conn.RawFS(), flag.Arg(1), mOpts)
	if err != nil {
		return err
	}
	server.Serve()
	return nil
}

func (fs *AddFS) rootUserPermit(context *fuse.Context) bool {
	return fs.opts.AllowRootMutation && context.Uid == 0
}

func (fs *AddFS) fileAlreadyExists(file string, context *fuse.Context) bool {
	_, status := fs.GetAttr(file, context)
	return status.Ok()
}

func (fs *AddFS) isMutable(file string) bool {
	for _, r := range fs.mutableFilesRegexes {
		if r.MatchString(file) {
			return true
		}
	}
	return false
}

func (fs *AddFS) containsDangerousFlags(flags uint32) bool {
	mask := uint32(os.O_APPEND | os.O_WRONLY | os.O_TRUNC | os.O_RDWR)
	return (flags & mask) != 0
}

func (fs *AddFS) Truncate(name string, size uint64, context *fuse.Context) (code fuse.Status) {
	if fs.rootUserPermit(context) || fs.isMutable(name) {
		return fs.FileSystem.Truncate(name, size, context)
	}
	return fuse.EACCES
}

func (fs *AddFS) Rename(oldName string, newName string, context *fuse.Context) (code fuse.Status) {
	if fs.rootUserPermit(context) || fs.isMutable(oldName) {
		return fs.FileSystem.Rename(oldName, newName, context)
	}
	return fuse.EACCES
}

func (fs *AddFS) Rmdir(name string, context *fuse.Context) (code fuse.Status) {
	if fs.rootUserPermit(context) || fs.isMutable(name) {
		return fs.FileSystem.Rmdir(name, context)
	}
	return fuse.EACCES
}

func (fs *AddFS) Unlink(name string, context *fuse.Context) (code fuse.Status) {
	if fs.rootUserPermit(context) || fs.isMutable(name) {
		return fs.FileSystem.Unlink(name, context)
	}
	return fuse.EACCES
}

func (fs *AddFS) Mkdir(name string, mode uint32, context *fuse.Context) fuse.Status {
	err := fs.FileSystem.Mkdir(name, mode, context)
	if !err.Ok() {
		return err
	}
	return fs.FileSystem.Chown(name, context.Uid, context.Gid, context)
}

func (fs *AddFS) Open(name string, flags uint32, context *fuse.Context) (file nodefs.File, code fuse.Status) {
	// Error when opening already existing file for writing
	if fs.fileAlreadyExists(name, context) && fs.containsDangerousFlags(flags) && !(fs.rootUserPermit(context) || fs.isMutable(name)) {
		return nil, fuse.EACCES
	}
	return fs.FileSystem.Open(name, flags, context)
}

func (fs *AddFS) Create(name string, flags uint32, mode uint32, context *fuse.Context) (file nodefs.File, code fuse.Status) {
	// Error when creating already existing file
	if fs.fileAlreadyExists(name, context) && !(fs.rootUserPermit(context) || fs.isMutable(name)) {
		return nil, fuse.EACCES
	}
	file, status := fs.FileSystem.Create(name, flags, mode, context)
	if !status.Ok() {
		return nil, status
	}
	mode = mode&uint32(os.ModeSetuid.Perm()) ^ 0
	return file, fs.FileSystem.Chown(name, context.Uid, context.Gid, context)
}
