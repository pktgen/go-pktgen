/* SPDX-License-Identifier: BSD-3-Clause
 * Copyright (c) 2023-2025 Intel Corporation.
 */

package fserde

import (
	"fmt"
	"net"
	"strconv"
	"strings"

	"golang.org/x/net/ipv4"
)

const (
	IPv4MinLen    = 20
	IPv4DefaultID = 1234
	DefaultTTL    = 64
)

type IPv4Layer struct {
	hdr   *LayerHdr
	ipHdr ipv4.Header
}

func (ip *IPv4Layer) String() string {
	return fmt.Sprintf("%s(version=%d, len=%d, tos=%d, totallen=%d, id=%d, "+
		"flags=%#x, frag=%#x, ttl=%d, protocol=%d, checksum=%#x, src=%v, dst=%v)",
		ip.Name(),
		ip.ipHdr.Version, ip.ipHdr.Len, ip.ipHdr.TOS, ip.ipHdr.TotalLen, ip.ipHdr.ID,
		ip.ipHdr.Flags, ip.ipHdr.FragOff, ip.ipHdr.TTL, ip.ipHdr.Protocol, ip.ipHdr.Checksum,
		ip.ipHdr.Src, ip.ipHdr.Dst)
}

func (e *IPv4Layer) Name() LayerName {
	return e.hdr.layerName
}

func IPv4New(fr *Frame) *IPv4Layer {
	return &IPv4Layer{
		hdr: LayerConstructor(fr, LayerIPv4, LayerIPv4Type),
	}
}

func isIPZero(ip net.IP) bool {
	return len(ip) == 0 || net.IP.Equal(net.IPv4zero, ip.To4())
}

func (ip *IPv4Layer) Parse(opts string) error {

	options := strings.Split(opts, ",")

	for _, opt := range options {
		opt = strings.TrimSpace(opt)

		kvp := strings.Split(opt, "=")
		if len(kvp) != 2 {
			return fmt.Errorf("invalid options string")
		}
		key := strings.ToLower(strings.TrimSpace(kvp[0]))
		val := strings.ToLower(strings.TrimSpace(kvp[1]))

		switch key {
		case "ver":
			if ver, err := strconv.ParseInt(val, 0, 0); err != nil {
				return err
			} else if ver != 4 {
				return fmt.Errorf("invalid version: %v", ver)
			} else {
				ip.ipHdr.Version = int(ver)
			}

		case "tos":
			if tos, err := strconv.ParseInt(val, 0, 0); err != nil {
				return err
			} else if tos == 0 {
				return fmt.Errorf("invalid TOS")
			} else {
				ip.ipHdr.TOS = int(tos)
			}

		case "id":
			if id, err := strconv.ParseInt(val, 0, 0); err != nil {
				return err
			} else {
				ip.ipHdr.ID = int(id)
			}

		case "flags":
			if flags, err := strconv.ParseInt(val, 0, 0); err != nil {
				return err
			} else {
				ip.ipHdr.Flags = ipv4.HeaderFlags(flags)
			}

		case "fragOffset":
			if fragOffset, err := strconv.ParseInt(val, 0, 0); err != nil {
				return err
			} else {
				ip.ipHdr.FragOff = int(fragOffset)
			}

		case "ttl":
			if ttl, err := strconv.ParseInt(val, 0, 0); err != nil {
				return err
			} else {
				ip.ipHdr.TTL = int(ttl)
			}

		case "protocol":
			if protocol, err := strconv.ParseInt(val, 0, 0); err != nil {
				return err
			} else {
				ip.ipHdr.Protocol = int(protocol)
			}

		case "src":
			ip.ipHdr.Src = net.ParseIP(val)

		case "dst":
			ip.ipHdr.Dst = net.ParseIP(val)

		default:
			return fmt.Errorf("unknown ether option: [%s]", opt)
		}
	}

	if ip.ipHdr.ID == 0 {
		ip.ipHdr.ID = IPv4DefaultID
	}
	if ip.ipHdr.TTL == 0 {
		ip.ipHdr.TTL = DefaultTTL
	}
	if isIPZero(ip.ipHdr.Dst) {
		ip.ipHdr.Dst = net.IPv4zero
	}
	if isIPZero(ip.ipHdr.Src) {
		ip.ipHdr.Src = net.IPv4zero
	}

	// Set the protocol header length include option bytes and the TotalLen
	// with the header length. The TotalLen will be updated after the other
	// layers are parsed.
	ip.ipHdr.Len = IPv4MinLen
	ip.ipHdr.TotalLen = ip.ipHdr.Len

	ip.hdr.proto.name = ip.Name()
	ip.hdr.proto.offset = ip.hdr.fr.GetOffset(ip.Name())
	ip.hdr.proto.length = uint16(ip.ipHdr.Len & 0xFF)

	ip.hdr.fr.AddProtocol(&ip.hdr.proto)

	return nil
}

func (l *IPv4Layer) ApplyDefaults() error {

	d := l.hdr.fr.defaultsFrame
	if d == nil {
		return nil
	}

	dl, ok := d.GetLayer(LayerIPv4).(*IPv4Layer)
	if !ok {
		return nil
	}
	if l.ipHdr.Version == 0 && dl.ipHdr.Version != 0 {
		l.ipHdr.Version = dl.ipHdr.Version
	}
	if l.ipHdr.TOS == 0 && dl.ipHdr.TOS != 0 {
		l.ipHdr.TOS = dl.ipHdr.TOS
	}
	if l.ipHdr.ID == 0 && dl.ipHdr.ID != 0 {
		l.ipHdr.ID = dl.ipHdr.ID
	}
	if l.ipHdr.Flags == 0 && dl.ipHdr.Flags != 0 {
		l.ipHdr.Flags = dl.ipHdr.Flags
	}
	if l.ipHdr.FragOff == 0 && dl.ipHdr.FragOff != 0 {
		l.ipHdr.FragOff = dl.ipHdr.FragOff
	}
	if l.ipHdr.TTL == 0 && dl.ipHdr.TTL != 0 {
		l.ipHdr.TTL = dl.ipHdr.TTL
	}
	if l.ipHdr.Protocol == 0 && dl.ipHdr.Protocol != 0 {
		l.ipHdr.Protocol = dl.ipHdr.Protocol
	}
	if isIPZero(l.ipHdr.Dst) && !isIPZero(dl.ipHdr.Dst) {
		l.ipHdr.Dst = dl.ipHdr.Dst
	}
	if isIPZero(l.ipHdr.Src) && !isIPZero(dl.ipHdr.Src) {
		l.ipHdr.Src = dl.ipHdr.Src
	}

	// Set the protocol header length include option bytes and the TotalLen
	// with the header length. The TotalLen will be updated after the other
	// layers are parsed.
	l.ipHdr.Len = IPv4MinLen
	l.ipHdr.TotalLen = l.ipHdr.Len

	return nil
}

func (l *IPv4Layer) WriteLayer() error {

	ip := &l.ipHdr
	fr := l.hdr.fr
	frame := fr.frame

	frame.WriteByte(uint8(ip.Version<<4) | (uint8(ip.Len&0xFF) >> 2))
	frame.Append(uint8(ip.TOS))

	frame.Append(uint16(ip.TotalLen))
	frame.Append(uint16(ip.ID))
	frame.Append(uint16(ip.Flags<<12) | (uint16(ip.FragOff)))
	frame.Append(uint8(ip.TTL))

	ip.Protocol = fr.GetProtocolID()
	frame.Append(uint8(ip.Protocol))

	cksum := IPv4HeaderChecksum(ip)
	frame.Append(cksum)

	frame.Append(ip.Src)
	frame.Append(ip.Dst)
	frame.Append(ip.Options)

	return nil
}
