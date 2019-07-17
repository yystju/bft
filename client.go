package main

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"io/ioutil"
	"log"
	"net"
	"os"
	"time"
)

func client() {
	timestamp := time.Now().Unix()

	sess := int(timestamp % 65535)

	conn, err := net.Dial("udp", addr)

	if err != nil {
		log.Panic(err)
	}

	defer conn.Close()

	log.Printf("LOCAL : %s <-> %s\n", conn.RemoteAddr().String(), conn.LocalAddr().String())

	go func() {
		writeData(sess, conn)
	}()

	readData(sess, conn)
}

func readData(sessionID int, conn net.Conn) {
	for {
		buffer := make([]byte, BufferSize)

		n, err := conn.Read(buffer)

		if err != nil {
			log.Printf("CLIENT ERROR ON READ : %s", err.Error())
			break
		}

		p := new(Packet)

		err = binary.Read(bytes.NewReader(buffer[0:n]), binary.BigEndian, p)

		if err != nil {
			log.Panic(err)
		}

		// log.Printf("CLIENT RECEIVED (%d-%d) : %d", p.Sess, p.Seq, time.Now().Unix()-p.Stamp)
	}
}

func writeData(sessionID int, conn net.Conn) {
	var data []byte
	var err error

	if "" != file {
		data, err = ioutil.ReadFile(file)
	} else {
		data, err = ioutil.ReadAll(os.Stdin)
	}

	if err != nil {
		log.Panic(err)
	}

	// log.Printf("data : %s\n", string(data))

	n := len(data)

	if n%BlockSize == 0 {
		n = n / BlockSize
	} else {
		n = n/BlockSize + 1
	}

	i := 0

	for ; i < n; i++ {
		s := data[i*BlockSize : min((i+1)*BlockSize, len(data))]

		h, err := NewPacket(int32(i), int32(sessionID), s)

		if err != nil {
			log.Panic(err)
		}

		var buff bytes.Buffer

		err = binary.Write(&buff, binary.BigEndian, h)

		if err != nil {
			log.Panic(err)
		}

		conn.Write(buff.Bytes())

		if interval > 0 {
			time.Sleep(time.Duration(interval))
		}
	}

	if interval > 0 {
		time.Sleep(time.Duration(interval))
	}

	eof, err := NewEOF(int32(i), int32(sessionID))

	if err != nil {
		log.Panic(err)
	}

	var buff bytes.Buffer

	err = binary.Write(&buff, binary.BigEndian, eof)

	if err != nil {
		log.Panic(err)
	}

	conn.Write(buff.Bytes())

	fmt.Print("Press 'Enter' to continue...")
	fmt.Scanln()

	conn.Close()

	log.Println("CLIENT CLOSED")
}
