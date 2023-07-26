package main

import (
	"bytes"
	"fmt"
	"net"
	"time"

	"github.com/google/gopacket"
	"github.com/google/gopacket/layers"
	"github.com/google/gopacket/packetio"
	"github.com/google/gopacket/pcap"
)

func main() {
	// 获取当前设备的IP和MAC地址
	interfaces, err := net.Interfaces()
	if err != nil {
		panic(err)
	}
	var localIP net.IP
	var localMAC net.HardwareAddr
	for _, iface := range interfaces {
		if iface.Flags&net.FlagUp != 0 && iface.Flags&net.FlagLoopback == 0 {
			addrs, err := iface.Addrs()
			if err != nil {
				panic(err)
			}
			for _, addr := range addrs {
				switch addr := addr.(type) {
				case *net.IPNet:
					if addr.IP.To4() != nil {
						localIP = addr.IP
					}
				case *net.IPAddr:
					if addr.IP.To4() != nil {
						localIP = addr.IP
					}
				}
			}
			localMAC = iface.HardwareAddr
			break
		}
	}
	if localIP == nil || localMAC == nil {
		panic("Could not find local IP and MAC addresses")
	}
	fmt.Println("Local IP:", localIP)
	fmt.Println("Local MAC:", localMAC)

	// 目标IP地址
	targetIP := net.ParseIP("192.168.1.1")
	if targetIP == nil {
		panic("Invalid target IP address")
	}
	fmt.Println("Target IP:", targetIP)

	// 使用gopacket构建ARP请求包
	var buf gopacket.SerializeBuffer
	opts := gopacket.SerializeOptions{}
	eth := layers.Ethernet{
		SrcMAC:       localMAC,
		DstMAC:       net.HardwareAddr{0xff, 0xff, 0xff, 0xff, 0xff, 0xff},
		EtherNettype: layers.EthernetTypeARP,
	}
	arp := layers.ARP{
		AddrType:          layers.LinkTypeEthernet,
		Protocol:          layers.EthernetTypeIPv4,
		HwAddressSize:     6,
		ProtAddressSize:   4,
		Operation:         layers.ARPRequest,
		SourceHwAddress:   []byte(localMAC),
		SourceProtAddress: []byte(localIP.To4()),
		DstHwAddress:      []byte{0x00, 0x00, 0x00, 0x00, 0x00, 0x00},
		DstProtAddress:    []byte(targetIP.To4()),
	}
	if err := gopacket.SerializeLayers(&buf, opts, &eth, &arp); err != nil {
		panic(err)
	}

	// 打开网络接口并发送ARP请求包
	handle, err := pcap.OpenLive("eth0", 65535, true, pcap.BlockForever)
	if err != nil {
		panic(err)
	}
	defer handle.Close()
	if err := handle.WritePacketData(buf.Bytes()); err != nil {
		panic(err)
	}

	// 等待一段时间以获得目标设备的MAC地址
	time.Sleep(time.Second)
	packetSource := packetio.NewPacketSource(handle, handle.LinkType())
	for packet := range packetSource.Packets() {
		arpLayer := packet.Layer(layers.LayerTypeARP)
		if arpLayer != nil {
			arpPacket, _ := arpLayer.(*layers.ARP)
			if arpPacket.Operation == layers.ARPReply && bytes.Equal(arpPacket.SourceProtAddress, targetIP.To4()) {
				fmt.Println("Target MAC:", net.HardwareAddr(arpPacket.SourceHwAddress))
				return
			}
		}
	}
	panic("Could not resolve target MAC address")
}
