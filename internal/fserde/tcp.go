/* SPDX-License-Identifier: BSD-3-Clause
 * Copyright (c) 2022-2024 Intel Corporation.
 */

package fserde

import (
	"fmt"
	"strconv"
	"strings"
)

const (
	TCPMinLen         = 20
	TCPHeaderLen      = 20
	TCPDefaultLen     = 20
	TCPChecksumOffset = 16
)

// Flags that may be set in a TCP segment.
const (
	TCPFinFlag = 1 << iota
	TCPSynFlag
	TCPRstFlag
	TCPPshFlag
	TCPAckFlag
	TCPUrgFlag
	TCPEceFlag
	TCPCwrFlag
)

type TCPOptions struct {
	kind   uint8
	length uint8
	data   []byte
}

type TCPHdr struct {
	SrcPort    uint16       // TCP source port
	DstPort    uint16       // TCP destination port
	SeqNum     uint32       // TCP sequence number
	AckNum     uint32       // TCP acknowledgment number
	Flags      uint16       // TCP flags
	Window     uint16       // TCP window size
	Checksum   uint16       // Always 0 and not configurable
	Urgent     uint16       // TCP urgent pointer
	HdrLen     uint16       // TCP header length
	Options    []byte       // TCP options bytes total length multiple of 4, max 40 bytes
	tcpOptions []TCPOptions // TCP options array
}

type TCPLayer struct {
	hdr    *LayerHdr
	tcpHdr TCPHdr
}

func (t *TCPLayer) String() string {
	h := t.tcpHdr
	s := fmt.Sprintf("TCP(sport=%v, dport=%v, seq=%v, ack=%v, len=%v flags=%v, window=%v, checksum=%v, urgent=%v, ",
		h.SrcPort, h.DstPort, h.SeqNum, h.AckNum, h.HdrLen, h.Flags, h.Window, h.Checksum, h.Urgent)
	if len(h.tcpOptions) > 0 {
		s += "options="
		for i, o := range h.tcpOptions {
			s += fmt.Sprintf("<%d,%d,%v>", o.kind, o.length, o.data)
			if i < len(h.tcpOptions)-1 {
				s += ","
			}
		}
	}
	return s + ")"
}

func TCPNew(fr *Frame) *TCPLayer {
	return &TCPLayer{
		hdr:    LayerConstructor(fr, LayerTCP, LayerTCPType),
		tcpHdr: TCPHdr{Options: []byte{}, tcpOptions: []TCPOptions{}},
	}
}

func (l *TCPLayer) Name() LayerName {
	return l.hdr.layerName
}

func parseTCPFlags(f string) uint16 {

	switch strings.TrimSpace(f) {
	case "fin":
		return TCPFinFlag
	case "syn":
		return TCPSynFlag
	case "rst":
		return TCPRstFlag
	case "psh":
		return TCPPshFlag
	case "ack":
		return TCPAckFlag
	case "urg":
		return TCPUrgFlag
	case "ecn", "ece":
		return TCPEceFlag
	case "cwr":
		return TCPCwrFlag
	default:
		return 0
	}
}

func (l *TCPLayer) Parse(opts string) error {

	options := strings.Split(opts, ",")

	for _, opt := range options {

		opt = strings.TrimSpace(opt)
		if len(opt) == 0 {
			continue
		}

		kvp := strings.Split(opt, "=")
		if len(kvp) != 2 {
			return fmt.Errorf("option needs a key/value pair")
		}
		key := strings.ToLower(strings.TrimSpace(kvp[0]))
		val := strings.ToLower(strings.TrimSpace(kvp[1]))

		switch key {
		case "sport":
			if v, err := strconv.ParseUint(val, 0, 16); err != nil {
				return err
			} else {
				l.tcpHdr.SrcPort = uint16(v)
			}
		case "dport":
			if v, err := strconv.ParseUint(val, 0, 16); err != nil {
				return err
			} else {
				l.tcpHdr.DstPort = uint16(v)
			}
		case "seq", "sequence", "seqnum":
			if v, err := strconv.ParseUint(val, 0, 32); err != nil {
				return err
			} else {
				l.tcpHdr.SeqNum = uint32(v)
			}
		case "ack", "acknowledgements", "acknum":
			if v, err := strconv.ParseUint(val, 0, 32); err != nil {
				return err
			} else {
				l.tcpHdr.AckNum = uint32(v)
			}
		case "len", "length":
			if v, err := strconv.ParseUint(val, 0, 16); err != nil {
				return err
			} else {
				l.tcpHdr.HdrLen = uint16(v)
			}
		case "flags":
			if strings.HasPrefix(val, "0x") {
				if v, err := strconv.ParseUint(val, 0, 16); err != nil {
					return err
				} else {
					l.tcpHdr.Flags = uint16(v)
				}
			} else {
				str := strings.FieldsFunc(val, func(r rune) bool {
					return r == '[' || r == ']'
				})
				flags := strings.Split(str[0], "|")
				for i := range flags {
					l.tcpHdr.Flags |= parseTCPFlags(flags[i])
				}
			}
		case "window":
			if v, err := strconv.ParseUint(val, 0, 16); err != nil {
				return err
			} else {
				l.tcpHdr.Window = uint16(v)
			}
		case "urgent":
			if v, err := strconv.ParseUint(val, 0, 16); err != nil {
				return err
			} else {
				l.tcpHdr.Urgent = uint16(v)
			}
		case "options": // TODO: convert this string to an array of TCPOptions structures.
			// trim the quotes from the string
			val = strings.TrimLeft(val, "'")
			val = strings.TrimRight(val, "'")

			str := fmt.Sprint(val)

			l.tcpHdr.Options = []byte{}

			for i := 0; i < len(str); i++ {
				l.tcpHdr.Options = append(l.tcpHdr.Options, str[i])
			}

			if len(str)%4 != 0 {
				n := 4 - len(str)%4
				dbug.Printf("options pad length: %d\n", n)
				for i := 0; i < n; i++ {
					l.tcpHdr.Options = append(l.tcpHdr.Options, 0)
				}
			}
			dbug.Printf("options: %v\n", l.tcpHdr.Options)

		default:
			return fmt.Errorf("unknown tcp option: [%s]", opt)
		}
	}
	dbug.Printf("**** tcp header %04x\n", uint16(TCPDefaultLen+len(l.tcpHdr.Options)))
	l.tcpHdr.HdrLen = uint16(TCPDefaultLen + len(l.tcpHdr.Options))

	l.hdr.proto.name = l.Name()
	l.hdr.proto.offset = l.hdr.fr.GetOffset(l.Name())
	l.hdr.proto.length = uint16(l.tcpHdr.HdrLen & 0xFF)

	l.hdr.fr.AddProtocol(&l.hdr.proto)

	dbug.Printf("%v\n", l)
	return nil
}

func (l *TCPLayer) ApplyDefaults() error {

	d := l.hdr.fr.defaultsFrame
	if d == nil {
		return nil
	}

	dl, ok := d.GetLayer(LayerTCP).(*TCPLayer)
	if !ok {
		return nil
	}

	if ip, ok := l.hdr.fr.layersMap[LayerIPv4].(*IPv4Layer); ok {
		ip.ipHdr.Protocol = ProtocolTCP
	} else {
		return dbug.Errorf("not support ipv4")
	}

	if l.tcpHdr.SrcPort == 0 && dl.tcpHdr.SrcPort != 0 {
		l.tcpHdr.SrcPort = dl.tcpHdr.SrcPort
	}
	if l.tcpHdr.DstPort == 0 && dl.tcpHdr.DstPort != 0 {
		l.tcpHdr.DstPort = dl.tcpHdr.DstPort
	}
	if l.tcpHdr.SeqNum == 0 && dl.tcpHdr.SeqNum != 0 {
		l.tcpHdr.SeqNum = dl.tcpHdr.SeqNum
	}
	if l.tcpHdr.AckNum == 0 && dl.tcpHdr.AckNum != 0 {
		l.tcpHdr.AckNum = dl.tcpHdr.AckNum
	}
	if l.tcpHdr.Flags == 0 && dl.tcpHdr.Flags != 0 {
		l.tcpHdr.Flags = dl.tcpHdr.Flags
	}
	if l.tcpHdr.Window == 0 && dl.tcpHdr.Window != 0 {
		l.tcpHdr.Window = dl.tcpHdr.Window
	}
	if l.tcpHdr.Urgent == 0 && dl.tcpHdr.Urgent != 0 {
		l.tcpHdr.Urgent = dl.tcpHdr.Urgent
	}
	if l.tcpHdr.HdrLen == 0 && dl.tcpHdr.HdrLen != 0 {
		l.tcpHdr.HdrLen = dl.tcpHdr.HdrLen
	}
	if (l.tcpHdr.Options == nil || len(l.tcpHdr.Options) == 0) && dl.tcpHdr.Options != nil {
		l.tcpHdr.Options = dl.tcpHdr.Options
	}

	dbug.Printf("%v\n", dl)

	return nil
}

func (l *TCPLayer) WriteLayer() error {

	frame := l.hdr.fr.frame
	tcp := l.tcpHdr

	frame.Append(tcp.SrcPort)
	frame.Append(tcp.DstPort)
	frame.Append(tcp.SeqNum)
	frame.Append(tcp.AckNum)
	flags := ((tcp.HdrLen >> 2) << 12) | tcp.Flags
	frame.Append(flags)
	frame.Append(tcp.Window)
	frame.Append(uint16(0)) // Set to zero and updated later
	frame.Append(tcp.Urgent)
	frame.Append(tcp.Options)

	dbug.Printf("%v\n", l)

	return nil
}
