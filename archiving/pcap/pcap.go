package main

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"log"
	"time"

	"github.com/google/gopacket"
	"github.com/google/gopacket/layers"
	"github.com/google/gopacket/pcap"
	"github.com/google/gopacket/tcpassembly"
	"github.com/google/gopacket/tcpassembly/tcpreader"
)

const (
	snaplen      = 1600
	promiscuous  = true
	timeout      = pcap.BlockForever
	streamExpiry = time.Minute // TCP流超时时间
)

type sniStreamFactory struct{}
type sniStream struct {
	bytes []byte
	done  bool
}

func (s *sniStreamFactory) New(netFlow, tcpFlow gopacket.Flow) tcpassembly.Stream {
	stream := &sniStream{}
	r := tcpreader.NewReaderStream()

	// 异步处理重组后的流数据
	go func() {
		buf := make([]byte, 4096)
		for {
			n, err := r.Read(buf)
			if err != nil {
				break
			}
			stream.bytes = append(stream.bytes, buf[:n]...)
			// 尝试解析SNI
			if sni, ok := extractSNI(stream.bytes); ok && !stream.done {
				fmt.Printf("[SNI] %s (From: %s)\n", sni, netFlow)
				stream.done = true
			}
		}
	}()
	return &r
}

const device = `\Device\NPF_{28B33438-DE72-48EF-98F9-8791DFFE27A9}`

func main() {
	// 选择网卡（可手动指定）

	ifs, err := pcap.FindAllDevs()
	if err != nil {
		log.Fatal(err)
		return
	}
	for _, inter := range ifs {
		for _, ip := range inter.Addresses {
			log.Println(inter.Name, ip.IP.String())
		}
	}

	handle, err := pcap.OpenLive(
		device,
		snaplen,
		promiscuous,
		timeout,
	)
	if err != nil {
		log.Fatal(err)
	}
	defer handle.Close()

	// 设置过滤器
	err = handle.SetBPFFilter("tcp port 443")
	if err != nil {
		log.Fatal(err)
	}

	// 创建TCP流重组器
	streamFactory := &sniStreamFactory{}
	streamPool := tcpassembly.NewStreamPool(streamFactory)
	assembler := tcpassembly.NewAssembler(streamPool)
	assembler.MaxBufferedPagesPerConnection = 100
	assembler.MaxBufferedPagesTotal = 1000

	packetSource := gopacket.NewPacketSource(handle, handle.LinkType())
	packets := packetSource.Packets()

	ticker := time.Tick(time.Minute)
	for {
		select {
		case packet := <-packets:
			if packet == nil {
				return
			}
			if packet.NetworkLayer() == nil || packet.TransportLayer() == nil {
				continue
			}
			tcp, ok := packet.TransportLayer().(*layers.TCP)
			if !ok {
				continue
			}

			// 将数据包交给重组器处理
			assembler.AssembleWithTimestamp(
				packet.NetworkLayer().NetworkFlow(),
				tcp,
				packet.Metadata().Timestamp,
			)

		case <-ticker:
			// 定期清理过期的流
			assembler.FlushOlderThan(time.Now().Add(-streamExpiry))
		}
	}
}

// 改进的SNI解析函数
func extractSNI(data []byte) (string, bool) {
	r := bytes.NewReader(data)

	// TLS记录头
	var (
		contentType uint8
		version     uint16
		length      uint16
	)
	if err := binary.Read(r, binary.BigEndian, &contentType); err != nil {
		return "", false
	}
	if contentType != 0x16 { // Handshake
		return "", false
	}

	if err := binary.Read(r, binary.BigEndian, &version); err != nil {
		return "", false
	}
	if version < 0x0301 { // TLS 1.0+
		return "", false
	}

	if err := binary.Read(r, binary.BigEndian, &length); err != nil {
		return "", false
	}
	if int(length) > r.Len() {
		return "", false
	}

	// 握手协议
	var handshakeType uint8
	if err := binary.Read(r, binary.BigEndian, &handshakeType); err != nil {
		return "", false
	}
	if handshakeType != 0x01 { // ClientHello
		return "", false
	}

	// 跳过握手长度、版本、随机数等字段
	_, _ = r.Seek(3+2+32, 1) // 3字节长度 + 2字节版本 + 32字节随机数

	// Session ID
	var sessionIDLen uint8
	if err := binary.Read(r, binary.BigEndian, &sessionIDLen); err != nil {
		return "", false
	}
	_, _ = r.Seek(int64(sessionIDLen), 1)

	// 密码套件
	var cipherSuitesLen uint16
	if err := binary.Read(r, binary.BigEndian, &cipherSuitesLen); err != nil {
		return "", false
	}
	_, _ = r.Seek(int64(cipherSuitesLen), 1)

	// 压缩方法
	var compressionMethodsLen uint8
	if err := binary.Read(r, binary.BigEndian, &compressionMethodsLen); err != nil {
		return "", false
	}
	_, _ = r.Seek(int64(compressionMethodsLen), 1)

	// 扩展列表
	if r.Len() == 0 {
		return "", false
	}

	var extensionsLen uint16
	if err := binary.Read(r, binary.BigEndian, &extensionsLen); err != nil {
		return "", false
	}

	remaining := int(extensionsLen)
	for remaining > 0 {
		var extType uint16
		var extLen uint16
		if err := binary.Read(r, binary.BigEndian, &extType); err != nil {
			break
		}
		if err := binary.Read(r, binary.BigEndian, &extLen); err != nil {
			break
		}

		remaining -= 4         // 已读取4字节（类型+长度）
		if extType == 0x0000 { // SNI扩展
			var nameListLen uint16
			if err := binary.Read(r, binary.BigEndian, &nameListLen); err != nil {
				return "", false
			}

			nameData := make([]byte, nameListLen)
			if _, err := r.Read(nameData); err != nil {
				return "", false
			}

			nameReader := bytes.NewReader(nameData)
			for nameReader.Len() > 0 {
				var nameType uint8
				if err := binary.Read(nameReader, binary.BigEndian, &nameType); err != nil {
					break
				}

				var nameLen uint16
				if err := binary.Read(nameReader, binary.BigEndian, &nameLen); err != nil {
					break
				}

				if nameType == 0x00 { // host_name
					name := make([]byte, nameLen)
					if _, err := nameReader.Read(name); err != nil {
						break
					}
					return string(name), true
				}

				// 跳过其他类型
				_, _ = nameReader.Seek(int64(nameLen), 1)
			}
		} else {
			// 跳过其他扩展
			_, _ = r.Seek(int64(extLen), 1)
		}
		remaining -= int(extLen)
	}

	return "", false
}

// 辅助函数：选择网卡
func selectDevice() string {
	devices, err := pcap.FindAllDevs()
	if err != nil {
		log.Fatal(err)
	}

	// 打印可用网卡
	fmt.Println("Available interfaces:")
	for i, dev := range devices {
		fmt.Printf("[%d] %s", i+1, dev.Name)
		if len(dev.Description) > 0 {
			fmt.Printf(" - %s", dev.Description)
		}
		fmt.Println()
	}

	// 选择第一个有IP地址的网卡
	for _, dev := range devices {
		if len(dev.Addresses) > 0 {
			fmt.Printf("\nAuto-selected interface: %s\n", dev.Name)
			return dev.Name
		}
	}

	log.Fatal("No available interfaces with IP addresses")
	return ""
}
