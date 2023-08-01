package main

import (
	"fmt"
	"github.com/google/gopacket"
	"github.com/google/gopacket/pcap"
)

func main() {
	path := "C:/Users/myAdmin/Desktop/tcpdemo.pcap"
	// 使用pcap子包读取xxx.pcap数据包
	handle, err := pcap.OpenOffline(path)
	if err != nil {
		panic(err)
	}

	// 使用gopacket创建数据包源对象
	packetSource := gopacket.NewPacketSource(handle, handle.LinkType())
	// 使用通道传送一个packet数据包
	packet := <-packetSource.Packets()
	// 查看数据包
	fmt.Println(packet)
}
