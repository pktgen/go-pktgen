// SPDX-License-Identifier: BSD-3-Clause
// Copyright (c) 2023 Intel Corporation

package pcap

import (
	"bufio"
	"encoding/binary"
	"fmt"
	"os"
	"time"
)

const (
	MicrosecondMagic  = 0xA1B2C3D4
	NanosecondMagic   = 0xA1B23C4D
	NanoToMicroSecond = 1000
	LinkTypeNull      = 0
	LinkTypeEthernet  = 1
	MajorVersion      = 2
	MinorVersion      = 4
	FCSLength         = 4
	DefaultSpanLength = (9 * 1024)
	MaxSpanLength     = 65535
)

type LinkType struct {
	FCSLen        uint8  // high order 0-2 bits
	FCSPresent    bool   // bit 3, bit 4 is reserved must be zero
	LinkLayerType uint16 // 16-bits Link layer type
}

type FileHeader struct {
	Magic    uint32   // value 0xA1B23C4D       - 4 bytes
	Major    uint16   // Major format 2         - 2 bytes
	Minor    uint16   // Minor format 4         - 2 bytes
	ThisZone uint32   // GMT offset in seconds  - 4 bytes
	SigFigs  uint32   // accuracy of timestamps - 4 bytes
	SpanLen  uint32   // max length captured    - 4 bytes
	LinkType LinkType // Link layer type        - 4 bytes
}

type PacketRecord struct {
	Seconds        uint32 // Seconds that have elapsed since 1970-01-01 00:00:00 UTC
	MicroNanoSec   uint32 // fractions of a second in micro-seconds or nano-seconds
	CaptureLength  uint32 // Length of the captured packet
	OriginalLength uint32 // Original length of the packet
	data           []byte // Raw packet data
}

type PacketCapture struct {
	fileHeader FileHeader
	pktRecords []*PacketRecord
}

func (p *PacketCapture) String() string {
	return fmt.Sprintf("PacketCapture: fileHeader: %+v, Count: %d", p.fileHeader, len(p.pktRecords))
}

func New() *PacketCapture {

	return &PacketCapture{
		fileHeader: FileHeader{
			Magic:   MicrosecondMagic,
			Major:   MajorVersion,
			Minor:   MinorVersion,
			SpanLen: DefaultSpanLength,
			LinkType: LinkType{
				FCSLen:        FCSLength,
				FCSPresent:    false,
				LinkLayerType: LinkTypeEthernet,
			},
		},
		pktRecords: []*PacketRecord{},
	}
}

func (p *PacketCapture) SetMagicNanoSeconds() *PacketCapture {
	p.fileHeader.Magic = NanosecondMagic

	return p
}

func (p *PacketCapture) SetSpanLen(spanLen uint32) *PacketCapture {
	p.fileHeader.SpanLen = spanLen

	return p
}

func (p *PacketCapture) SetFCSLength(v uint8) *PacketCapture {
	p.fileHeader.LinkType.FCSLen = v

	return p
}

func (p *PacketCapture) SetFCSPresent(v bool) *PacketCapture {
	p.fileHeader.LinkType.FCSPresent = v

	return p
}

func (p *PacketCapture) SetLinkType(linkType uint16) *PacketCapture {
	p.fileHeader.LinkType.LinkLayerType = linkType

	return p
}

// PacketRecord routines

func (p *PacketCapture) NewPacket() *PacketRecord {
	t := time.Now()

	fraction := t.Nanosecond()
	if p.fileHeader.Magic == MicrosecondMagic {
		fraction /= NanoToMicroSecond
	}

	return &PacketRecord{
		Seconds:        uint32(t.Second()),
		MicroNanoSec:   uint32(fraction),
		CaptureLength:  0,
		OriginalLength: 0,
		data:           nil,
	}
}

func (p *PacketCapture) GetPacketRecords() []*PacketRecord {
	return p.pktRecords
}

func (p *PacketCapture) GetFileHeader() FileHeader {
	return p.fileHeader
}

func (p *PacketCapture) GetPacketRecordsCount() int {
	return len(p.pktRecords)
}

func (p *PacketCapture) GetSpanLen() uint32 {
	return p.fileHeader.SpanLen
}

func (p *PacketCapture) GetFCSLength() uint8 {
	return p.fileHeader.LinkType.FCSLen
}

func (p *PacketCapture) GetFCSPresent() bool {
	return p.fileHeader.LinkType.FCSPresent
}

func (p *PacketCapture) GetLinkType() uint16 {
	return p.fileHeader.LinkType.LinkLayerType
}

func (p *PacketRecord) SetCaptureLength(length uint) *PacketRecord {

	p.CaptureLength = uint32(length)

	return p
}

func (p *PacketRecord) SetOriginalLength(length uint) *PacketRecord {

	p.OriginalLength = uint32(length)

	return p
}

func (p *PacketRecord) SetData(data []byte, dataLen uint) *PacketRecord {

	if dataLen > 0 && len(data) <= int(dataLen) {
		p.data = data
	} else {
		p.data = data[:dataLen]
	}

	return p
}

func (p *PacketCapture) AddPacket(pktData []byte) *PacketCapture {

	pktLen := uint(len(pktData))
	if pktLen > uint(p.fileHeader.SpanLen) {
		pktLen = MaxSpanLength
	}
	record := p.NewPacket().
		SetData(pktData, pktLen).
		SetCaptureLength(pktLen).
		SetOriginalLength(uint(len(pktData)))

	p.pktRecords = append(p.pktRecords, record)

	return p
}

func (p *PacketCapture) FileHeader() []byte {
	data := make([]byte, 0)

	return data
}

func (fh *FileHeader) constructFileHeader() []byte {
	var buf [24]byte

	binary.LittleEndian.PutUint32(buf[0:4], fh.Magic)
	binary.LittleEndian.PutUint16(buf[4:6], fh.Major)
	binary.LittleEndian.PutUint16(buf[6:8], fh.Minor)
	binary.LittleEndian.PutUint32(buf[8:12], fh.ThisZone)
	binary.LittleEndian.PutUint32(buf[12:16], fh.SigFigs)
	binary.LittleEndian.PutUint32(buf[16:20], fh.SpanLen)
	linkType := uint32(fh.LinkType.LinkLayerType)
	linkType |= uint32(fh.LinkType.FCSLen) << 29
	if fh.LinkType.FCSPresent {
		linkType |= uint32(1 << 28)
	}
	binary.LittleEndian.PutUint32(buf[20:24], linkType)
	return buf[:]
}

func (ph *PacketRecord) constructPacketHeader() []byte {
	var buf [16]byte

	binary.LittleEndian.PutUint32(buf[0:4], ph.Seconds)
	binary.LittleEndian.PutUint32(buf[4:8], ph.MicroNanoSec)
	binary.LittleEndian.PutUint32(buf[8:12], ph.CaptureLength)
	binary.LittleEndian.PutUint32(buf[12:16], ph.OriginalLength)

	return buf[:]
}

func (p *PacketCapture) Write(path string) error {

	if file, err := os.Create(path); err != nil {
		return err
	} else {
		defer file.Close()

		w := bufio.NewWriter(file)
		hdr := p.fileHeader.constructFileHeader()
		if _, err := w.Write(hdr); err != nil {
			return err
		}

		for _, r := range p.pktRecords {
			ph := r.constructPacketHeader()
			if _, err := w.Write(ph); err != nil {
				return err
			}
			if _, err := w.Write(r.data); err != nil {
				return err
			}
		}
		w.Flush()
	}

	return nil
}
