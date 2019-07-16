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

// Server is a ...
func server() {
	pc, err := net.ListenPacket("udp", addr)

	if err != nil {
		log.Fatal(err)
	}

	defer pc.Close()

	// cache := make(map[string][]*Packet)
	cache := new(sync.Map)

	buffer := make([]byte, 1024)

	for {
		n, addr, err := pc.ReadFrom(buffer)

		if err != nil {
			log.Fatal(err)
		}

		r := bytes.NewReader(buffer[0:n])

		p := new(Packet)

		err = binary.Read(r, binary.BigEndian, p)

		if err != nil {
			log.Panic(err)
		}

		// log.Printf("SERVER RECEIVED : %s (%d-%d:%d)", addr.String(), p.Sess, p.Seq, p.Len)

		go process(pc, addr, p, cache)
	}
}

func process(pc net.PacketConn, addr net.Addr, p *Packet, cache *sync.Map) {
	key := fmt.Sprintf("%s_%d", addr, int(p.Sess))

	value, ok := cache.Load(key)

	if p.Len < 0 { //EoF
		log.Printf("SERVER EoF")

		if ok {
			l := value.([]*Packet)

			sort.Sort(SortPacketBySeq(l))

			log.Printf("len(l) : %d", len(l))

			var outs io.Writer
			var f *os.File

			if "" == out {
				outs = bufio.NewWriter(os.Stdout)
			} else {
				fileName := fmt.Sprintf("%s.%d.udp", addr.String(), p.Sess)

				fileName = strings.ReplaceAll(fileName, ":", ".")
				fileName = strings.ReplaceAll(fileName, "[", "")
				fileName = strings.ReplaceAll(fileName, "]", "")

				fullPathName := path.Join(out, fileName)

				log.Println(fullPathName)

				f, err := os.OpenFile(fullPathName, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0666)

				if err != nil {
					log.Panic(err)
				}

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

			if f != nil {
				f.Close()
			}
		}

		ret, _ := NewPacket(p.Sess, p.Seq, []byte("ACCEPTED"))

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

		buff := new(bytes.Buffer)

		err := binary.Write(buff, binary.BigEndian, ret)

		if err != nil {
			log.Panic(err)
		}

		pc.WriteTo(buff.Bytes(), addr)
	}
}
