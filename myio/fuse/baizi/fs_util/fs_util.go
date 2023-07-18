package fs_util

import (
	"bazil.org/fuse"
	"bazil.org/fuse/fs"
	"context"
	"os"
	"syscall"
)

// hellofs 文件系统的主体
type FS struct{}

func (FS) Root() (fs.Node, error) {
	return Dir{}, nil
}

// hellofs 文件系统中，Dir 是目录操作的主体
type Dir struct{}

func (Dir) Attr(ctx context.Context, a *fuse.Attr) error {
	println("Attr")
	a.Inode = 20210601
	a.Mode = os.ModeDir | 0555
	return nil
}

// 当 ls 目录的时候，触发的是 ReadDirAll 调用，这里返回指定内容，表明只有一个 hello 的文件；
func (Dir) Lookup(ctx context.Context, name string) (fs.Node, error) {
	println("Lookup")
	// 只处理一个叫做 hello 的 entry 文件，其他的统统返回 not exist
	if name == "hello" {
		return File{}, nil
	}
	return nil, syscall.ENOENT
}

// 定义 Readdir 的行为，固定返回了一个 inode:2 name 叫做 hello 的文件。对应用户的行为一般是 ls 这个目录。
func (Dir) ReadDirAll(ctx context.Context) ([]fuse.Dirent, error) {
	println("ReadDirAll")
	var dirDirs = []fuse.Dirent{{Inode: 2, Name: "hello", Type: fuse.DT_File}}
	return dirDirs, nil
}

func (Dir) Mkdir(ctx context.Context, req *fuse.MkdirRequest) (fs.Node, error) {
	os.Mkdir(req.Name, req.Mode)
	println("Mkdir")

	return nil, nil
}

// hellofs 文件系统中，File 结构体实现了文件系统中关于文件的调用实现
type File struct{}

const fileContent = "hello, world\n"

// 当 stat 这个文件的时候，返回 inode 为 2，mode 为 444
func (File) Attr(ctx context.Context, a *fuse.Attr) error {
	println("Attr File")
	a.Inode = 20210606
	a.Mode = 0444
	a.Size = uint64(len(fileContent))
	return nil
}

// 当 cat 这个文件的时候，文件内容返回 hello，world
func (File) ReadAll(ctx context.Context) ([]byte, error) {
	println("ReadAll File")
	return []byte(fileContent), nil
}

func (File) Remove(ctx context.Context, req *fuse.RemoveRequest) error {

	println("remove File")
	return nil
}

func (File) Write(ctx context.Context, req *fuse.WriteRequest, resp *fuse.WriteResponse) error {
	println("Write File")
	return nil
}

func (File) Read(ctx context.Context, req *fuse.ReadRequest, resp *fuse.ReadResponse) error {
	println("Read File")
	return nil
}

func (File) Rename(ctx context.Context, req *fuse.RenameRequest, newDir fs.Node) error {
	println("Rename File")
	return nil
}
func (File) Create(ctx context.Context, req *fuse.CreateRequest, resp *fuse.CreateResponse) (fs.Node, fs.Handle, error) {
	println("create File")
	return nil, nil, nil
}
func (File) Open(ctx context.Context, req *fuse.OpenRequest, resp *fuse.OpenResponse) (fs.Handle, error) {
	println("Open File")
	return nil, nil
}

func (File) Mknod(ctx context.Context, req *fuse.MknodRequest) (fs.Node, error) {
	println("Mknod File")
	return nil, nil
}

func (File) Statfs(ctx context.Context, req *fuse.StatfsRequest, resp *fuse.StatfsResponse) error {
	println("Statfs File")
	return nil
}
func (File) Access(ctx context.Context, req *fuse.AccessRequest) error {
	println("Access File")
	return nil
}
