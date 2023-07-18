package main

import (
	"fmt"
	"os"
	"time"
	"yussgo/fuse/gofuse/mem/util"

	"github.com/hanwen/go-fuse/v2/fuse"
	"github.com/hanwen/go-fuse/v2/fuse/nodefs"
)

func main() {
	// Scans the arg list and sets up flags
	//debug := flag.Bool("debug", false, "print debugging messages.")
	//flag.Parse()
	//if flag.NArg() < 2 {
	//	// TODO - where to get program name?
	//	fmt.Println("usage: main MOUNTPOINT BACKING-PREFIX")
	//	os.Exit(2)
	//}
	//prefix := flag.Arg(1)
	//mountPoint := flag.Arg(0)
	debug := new(bool)
	*debug = true
	mountPoint := "fuse/gofuse/mem/YU2"
	prefix := "He"
	root := util.NewMemNodeFSRootYU(prefix)
	conn := nodefs.NewFileSystemConnector(root, nil)
	server, err := fuse.NewServer(conn.RawFS(), mountPoint, &fuse.MountOptions{
		Debug: *debug,
	})
	if err != nil {
		fmt.Printf("Mount fail: %v\n", err)
		os.Exit(1)
	}
	fmt.Println("Mounted!")
	go func(s *fuse.Server) {
		time.Sleep(60 * time.Millisecond * 1000)
		server.Unmount()
		fmt.Println("un Mounted!")
		os.Exit(0)
	}(server)

	server.Serve()

}
