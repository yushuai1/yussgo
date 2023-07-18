// 实现一个叫做 hellfs 的文件系统
package main

import (
	"bazil.org/fuse"
	"bazil.org/fuse/fs"
	_ "bazil.org/fuse/fs/fstestutil"
	"log"
	"yussgo/myio/fuse/baizi/fs_util"
)

func main() {
	var mountpoint string
	mountpoint = "/home/yu/local/tmp/test1"
	//os.Mkdir(mountpoint, os.ModeDir)
	//flag.StringVar(&mountpoint, "mountpoint", "", "mount point(dir)?")
	//flag.Parse()
	//
	//if mountpoint == "" {
	//	log.Fatal("please input invalid mount point\n")
	//}
	// 建立一个负责解析和封装 FUSE 请求监听通道对象；
	c, err := fuse.Mount(mountpoint, fuse.FSName("helloworld"), fuse.Subtype("hellofs"))
	if err != nil {
		log.Fatal(err)
	}
	defer c.Close()

	// 把 FS 结构体注册到 server，以便可以回调处理请求
	err = fs.Serve(c, fs_util.FS{})
	if err != nil {
		log.Fatal(err)
	}
}
