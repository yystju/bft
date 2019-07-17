package main

import (
	"bufio"
	"bytes"
	"encoding/binary"
	"fmt"
	"io"
	"log"
	"net"
	"os"
	"path"
	"sort"
	"strings"
	"sync"
)

var (
	pChan chan *rawPacket
	cache sync.Map
)

// Server is a ...
func server() {
	pc, err := net.ListenPacket("udp", addr)

	if err != nil {
		log.Fatal(err)
	}

	defer pc.Close()

	pChan = make(chan *rawPacket, ChannelSize)

	go consumer(pc)

	for {
		buffer := make([]byte, BufferSize)

		n, addr, err := pc.ReadFrom(buffer)

		if err != nil {
			log.Fatal(err)
		}

		pChan <- newRawPacket(addr, buffer, n)
	}
}

func consumer(pc net.PacketConn) {
	for {
		raw := <-pChan

		r := bytes.NewReader(raw.Buffer[0:raw.N])

		p := new(Packet)

		err := binary.Read(r, binary.BigEndian, p)

		if err != nil {
			log.Panic(err)
		}

		// log.Printf("SERVER RECEIVED : %s (%d-%d:%d)", addr.String(), p.Sess, p.Seq, p.Len)

		go process(pc, raw.Addr, p, &cache)
	}
}

func process(pc net.PacketConn, addr net.Addr, p *Packet, cache *sync.Map) {
	key := fmt.Sprintf("%s_%d", addr.String(), int(p.Sess))

	value, ok := cache.Load(key)

	if p.Len < 0 { //EoF
		log.Printf("SERVER EoF")

		if ok {
			l := value.([]*Packet)

			sort.Sort(SortPacketBySeq(l))

			log.Printf("len(l) : %d", len(l))

			go writeToFile(addr, l, p.Sess)
		}

		ret, _ := NewPacket(p.Sess, p.Seq, []byte("ACCEPTED"))

		ret.Stamp = p.Stamp

		buff := new(bytes.Buffer)

		err := binary.Write(buff, binary.BigEndian, ret)

		if err != nil {
			log.Panic(err)
		}

		pc.WriteTo(buff.Bytes(), addr)

		cache.Delete(key)
	} else {
		if !ok {
			value = make([]*Packet, 0, 1024)
		}

		l := value.([]*Packet)

		cache.Store(key, append(l, p))

		ret, _ := NewPacket(p.Sess, p.Seq, []byte("ACCEPTED"))

		ret.Stamp = p.Stamp

		buff := new(bytes.Buffer)

		err := binary.Write(buff, binary.BigEndian, ret)

		if err != nil {
			log.Panic(err)
		}

		pc.WriteTo(buff.Bytes(), addr)
	}
}

func writeToFile(addr net.Addr, l []*Packet, sess int32) {
	var outs io.Writer

	if "" == out {
		outs = bufio.NewWriter(os.Stdout)
	} else {
		fileName := fmt.Sprintf("%s.%d.udp", addr.String(), sess)

		fileName = strings.ReplaceAll(fileName, ":", ".")
		fileName = strings.ReplaceAll(fileName, "[", "")
		fileName = strings.ReplaceAll(fileName, "]", "")

		fullPathName := path.Join(out, fileName)

		log.Printf("fullPathName : %s", fullPathName)

		f, err := os.OpenFile(fullPathName, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0666)

		if err != nil {
			log.Panic(err)
		}

		defer f.Close()

		outs = f
	}

	for _, p := range l {
		n, err := outs.Write(p.Data[0:p.Len])

		if n != int(p.Len) {
			log.Printf("[WARNING] p.Len : %d, n : %d", p.Len, n)
		}

		if err != nil {
			log.Panic(err)
		}
	}
}
