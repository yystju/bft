package main

import (
	"errors"
	"time"
)

const (
	//BlockSize is the packet size. In UDP case, it's 540 as the internet MAC packet size should be 548 for applications.
	BlockSize = 540 - 16
)

// Packet is the structure for udp packet.
type Packet struct {
	Seq   int32
	Sess  int32
	Stamp int64
	Len   int16
	Data  [BlockSize]byte
}

// Take is the function to reassign the content of data.
func (p *Packet) Take(data []byte) (*Packet, error) {
	if len(data) > BlockSize {
		return nil, errors.New("Too large as a packet")
	}

	p.Len = int16(len(data))
	copy(p.Data[:], data)
	p.Stamp = time.Now().Unix()

	return p, nil
}

// SortPacketBySeq is a sort utility for Packet
type SortPacketBySeq []*Packet

func (a SortPacketBySeq) Len() int           { return len(a) }
func (a SortPacketBySeq) Swap(i, j int)      { a[i], a[j] = a[j], a[i] }
func (a SortPacketBySeq) Less(i, j int) bool { return a[i].Seq < a[j].Seq }

// NewPacket is the constructor of Packet
func NewPacket(seq, sess int32, data []byte) (*Packet, error) {
	p := new(Packet)

	p.Seq = seq
	p.Sess = sess

	return p.Take(data)
}

// NewEOF is the constructor of EOF Packet
func NewEOF(seq, sess int32) (*Packet, error) {
	p := new(Packet)

	p.Seq = seq
	p.Sess = sess
	p.Len = -1
	p.Stamp = time.Now().Unix()

	return p, nil
}
