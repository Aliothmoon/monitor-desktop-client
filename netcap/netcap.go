package netcap

import (
	"bytes"
	"encoding/binary"
	"log"
	"monitor-desktop-client/utils"
	"time"

	"github.com/google/gopacket"
	"github.com/google/gopacket/layers"
	"github.com/google/gopacket/pcap"
	"github.com/google/gopacket/tcpassembly"
	"github.com/google/gopacket/tcpassembly/tcpreader"
)

const (
	Snaplen      = 1600
	Promiscuous  = true
	Timeout      = pcap.BlockForever
	StreamExpiry = time.Minute // TCP流超时时间
)

type SniStreamFactory struct {
	Ch chan string
}
type SniStream struct {
	bytes []byte
	done  bool
}

func OpenLive(device string) *SniStreamFactory {

	handle, err := pcap.OpenLive(
		device,
		Snaplen,
		Promiscuous,
		Timeout,
	)
	if err != nil {
		log.Println("无法打开网络设备:", device, err)
		return nil
	}
	log.Println("成功打开网络监控设备:", device)

	// 注意：不要在这里提前关闭，移到协程内部
	// defer handle.Close()

	// 设置过滤器
	err = handle.SetBPFFilter("tcp port 443")
	if err != nil {
		log.Println("设置BPF过滤器失败:", err)
		handle.Close()
		return nil
	}
	log.Println("成功设置BPF过滤器 'tcp port 443'")

	// 创建TCP流重组器
	streamFactory := &SniStreamFactory{
		Ch: make(chan string, 1024),
	}
	streamPool := tcpassembly.NewStreamPool(streamFactory)
	assembler := tcpassembly.NewAssembler(streamPool)
	assembler.MaxBufferedPagesPerConnection = 100
	assembler.MaxBufferedPagesTotal = 1000

	packetSource := gopacket.NewPacketSource(handle, handle.LinkType())
	packets := packetSource.Packets()

	ticker := time.Tick(time.Minute)
	utils.Go(func() {
		defer handle.Close() // 移动到这里确保协程结束时关闭句柄
		log.Printf("开始处理来自设备 %s 的网络数据包", device)

		for {
			select {
			case packet := <-packets:
				if packet == nil {
					log.Printf("设备 %s 停止提供数据包", device)
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
				log.Printf("清理设备 %s 的过期流", device)
				assembler.FlushOlderThan(time.Now().Add(-StreamExpiry))
			}
		}
	})
	return streamFactory
}

func (s *SniStreamFactory) New(netFlow, _ gopacket.Flow) tcpassembly.Stream {
	stream := &SniStream{}
	r := tcpreader.NewReaderStream()

	// 异步处理重组后的流数据
	utils.Go(func() {
		buf := make([]byte, 8192) // 增加缓冲区大小，从4096增加到8192
		log.Printf("创建新的TCP流: %s", netFlow)
		for {
			n, err := r.Read(buf)
			if err != nil {
				log.Printf("读取TCP流结束: %s, 错误: %v", netFlow, err)
				break
			}
			stream.bytes = append(stream.bytes, buf[:n]...)
			// 尝试解析SNI
			if sni, ok := processDataHead(stream.bytes); ok && !stream.done {
				log.Printf("[SNI] 成功解析域名: %s (来源: %s)", sni, netFlow)
				s.Ch <- sni // 确保SNI数据被发送到通道
				stream.done = true
			}
		}
	})
	return &r
}

func processDataHead(data []byte) (string, bool) {
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
