// SPDX-License-Identifier: BSD-3-Clause
// Copyright (c) 2022-2024 Intel Corporation

package fserde

import (
	"bytes"
	"fmt"
	"regexp"
	"strings"
)

type ToBinaryConverter interface {
	// Methods for ToBinary
	Parse(opts string) error
	ApplyDefaults() error
	WriteLayer() error
}

// splitLayerString returns the LayerName and LayerOptions strings by removing ()
// plus removing leading and trailing whitespace.
func (fr *Frame) splitLayerString(layer string) (lType, lOptions string) {

	lType, lOptions = "", ""

	if len(layer) > 0 {
		// Split the frame string into two parts LayerType and LayerOptions strings.
		// i.e., Ether(ethertype=0x800) to lType: "Ether" and lOptions: "ethertype=0x800".
		str := strings.FieldsFunc(layer, func(r rune) bool {
			return r == '(' || r == ')'
		})
		if len(str) == 2 {
			lType = strings.TrimSpace(str[0])
			lOptions = strings.TrimSpace(str[1])
		} else if len(str) == 1 {
			lType = strings.TrimSpace(str[0])
			lOptions = ""
		} else {
			return
		}
		if lType == "" {
			return "", ""
		}
	}
	return
}

func (fr *Frame) toBinaryParse(li *LayerInfo) error {

	if li == nil {
		return fmt.Errorf("layer info is nil")
	}

	switch li.Name {
	case LayerCount:
		l := newFuncs.countNewFn(fr)
		if err := l.Parse(li.Opts); err != nil {
			return err
		}
		fr.layersMap[li.Name] = l
		li.Layer = l
	case LayerDot1AD:
		l := newFuncs.dot1adNewFn(fr)
		if err := l.Parse(li.Opts); err != nil {
			return err
		}
		fr.layersMap[li.Name] = l
		li.Layer = l
	case LayerDot1Q:
		l := newFuncs.dot1qNewFn(fr)
		if err := l.Parse(li.Opts); err != nil {
			return err
		}
		fr.layersMap[li.Name] = l
		li.Layer = l
	case LayerEcho:
		l := newFuncs.echoNewFn(fr)
		if err := l.Parse(li.Opts); err != nil {
			return err
		}
		fr.layersMap[li.Name] = l
		li.Layer = l
	case LayerEther:
		l := newFuncs.etherNewFn(fr)
		if err := l.Parse(li.Opts); err != nil {
			return err
		}
		fr.layersMap[li.Name] = l
		li.Layer = l
	case LayerICMPv4:
		l := newFuncs.icmpv4NewFn(fr)
		if err := l.Parse(li.Opts); err != nil {
			return err
		}
		fr.layersMap[li.Name] = l
		li.Layer = l
	case LayerICMPv6:
		l := newFuncs.icmpv6NewFn(fr)
		if err := l.Parse(li.Opts); err != nil {
			return err
		}
		fr.layersMap[li.Name] = l
		li.Layer = l
	case LayerIPv4:
		l := newFuncs.ipv4NewFn(fr)
		if err := l.Parse(li.Opts); err != nil {
			return err
		}
		fr.layersMap[li.Name] = l
		li.Layer = l
	case LayerIPv6:
		l := newFuncs.ipv6NewFn(fr)
		if err := l.Parse(li.Opts); err != nil {
			return err
		}
		fr.layersMap[li.Name] = l
		li.Layer = l
	case LayerPayload:
		l := newFuncs.payloadNewFn(fr)
		if err := l.Parse(li.Opts); err != nil {
			return err
		}
		fr.layersMap[li.Name] = l
		li.Layer = l
	case LayerQinQ:
		l := newFuncs.qinqNewFn(fr)
		if err := l.Parse(li.Opts); err != nil {
			return err
		}
		fr.layersMap[li.Name] = l
		li.Layer = l
	case LayerSCTP:
		l := newFuncs.sctpNewFn(fr)
		if err := l.Parse(li.Opts); err != nil {
			return err
		}
		fr.layersMap[li.Name] = l
		li.Layer = l
	case LayerTCP:
		l := newFuncs.tcpNewFn(fr)
		if err := l.Parse(li.Opts); err != nil {
			return err
		}
		fr.layersMap[li.Name] = l
		li.Layer = l
	case LayerTSC:
		l := newFuncs.tscNewFn(fr)
		if err := l.Parse(li.Opts); err != nil {
			return err
		}
		fr.layersMap[li.Name] = l
		li.Layer = l
	case LayerUDP:
		l := newFuncs.udpNewFn(fr)
		if err := l.Parse(li.Opts); err != nil {
			return err
		}
		fr.layersMap[li.Name] = l
		li.Layer = l
	case LayerVxLan:
		l := newFuncs.vxlanNewFn(fr)
		if err := l.Parse(li.Opts); err != nil {
			return err
		}
		fr.layersMap[li.Name] = l
		li.Layer = l
	case LayerDefaults:
		l := newFuncs.defaultsNewFn(fr)
		if err := l.Parse(li.Opts); err != nil {
			return err
		}
		fr.layersMap[li.Name] = l
		li.Layer = l
	default:
		return dbug.Errorf("invalid layer type: '%s'", li.Name)
	}

	return nil
}

func (fr *Frame) toBinaryApplyDefaults() error {

	for _, s := range fr.layerInfo {
		if layer, ok := fr.layersMap[s.Name]; ok {
			switch d := layer.(type) {
			case *CountLayer:
				if err := d.ApplyDefaults(); err != nil {
					return err
				}
			case *Dot1adLayer:
				if err := d.ApplyDefaults(); err != nil {
					return err
				}
			case *Dot1qLayer:
				if err := d.ApplyDefaults(); err != nil {
					return err
				}
			case *EchoLayer:
				if err := d.ApplyDefaults(); err != nil {
					return err
				}
			case *EtherLayer:
				if err := d.ApplyDefaults(); err != nil {
					return err
				}
			case *ICMPv4Layer:
				if err := d.ApplyDefaults(); err != nil {
					return err
				}
			case *ICMPv6Layer:
				if err := d.ApplyDefaults(); err != nil {
					return err
				}
			case *IPv4Layer:
				if err := d.ApplyDefaults(); err != nil {
					return err
				}
			case *IPv6Layer:
				if err := d.ApplyDefaults(); err != nil {
					return err
				}
			case *PayloadLayer:
				if err := d.ApplyDefaults(); err != nil {
					return err
				}
			case *QinQLayer:
				if err := d.ApplyDefaults(); err != nil {
					return err
				}
			case *SCTPLayer:
				if err := d.ApplyDefaults(); err != nil {
					return err
				}
			case *TCPLayer:
				if err := d.ApplyDefaults(); err != nil {
					return err
				}
			case *TSCLayer:
				if err := d.ApplyDefaults(); err != nil {
					return err
				}
			case *UDPLayer:
				if err := d.ApplyDefaults(); err != nil {
					return err
				}
			case *VxLanLayer:
				if err := d.ApplyDefaults(); err != nil {
					return err
				}
			case *DefaultsLayer:
				if err := d.ApplyDefaults(); err != nil {
					return err
				}
			default:
				return fmt.Errorf("Frame.toBinaryApplyDefaults:invalid layer type: '%T'", layer)
			}
		} else {
			return fmt.Errorf("unknown layer: %s", s.Name)
		}
	}

	return nil
}

func (fr *Frame) toBinaryUpdateLengths() error {

	if payload, ok := fr.GetLayer(LayerPayload).(*PayloadLayer); ok {
		dbug.Printf("%v\n", payload)

		if ip, ok := fr.GetLayer(LayerIPv4).(*IPv4Layer); ok {
			if udp, ok := fr.GetLayer(LayerUDP).(*UDPLayer); ok {
				udp.udpHdr.Length += uint16(payload.length)
				ip.ipHdr.TotalLen += int(udp.udpHdr.Length)
				dbug.Printf("%v\n", udp)
			} else if tcp, ok := fr.GetLayer(LayerTCP).(*TCPLayer); ok {
				ip.ipHdr.TotalLen += int(tcp.tcpHdr.HdrLen) + int(payload.length)
				dbug.Printf("%v\n", tcp)
			}
			dbug.Printf("%v\n", ip)
		} else if ip, ok := fr.GetLayer(LayerIPv6).(*IPv6Layer); ok {
			if udp, ok := fr.GetLayer(LayerUDP).(*UDPLayer); ok {
				udp.udpHdr.Length += uint16(payload.length)
				ip.ip6Hdr.PayloadLen += int(udp.udpHdr.Length)
				dbug.Printf("%v\n", udp)
			} else if tcp, ok := fr.GetLayer(LayerTCP).(*TCPLayer); ok {
				ip.ip6Hdr.PayloadLen += int(tcp.tcpHdr.HdrLen)
				dbug.Printf("%v\n", tcp)
			}
			dbug.Printf("%v\n", ip)
		}
	}

	return nil
}

func (fr *Frame) toBinaryWriteLayer() error {

	for _, s := range fr.layerInfo {
		if layer, ok := fr.layersMap[s.Name]; ok {
			switch d := layer.(type) {
			case *CountLayer:
				if err := d.WriteLayer(); err != nil {
					return err
				}
			case *Dot1adLayer:
				if err := d.WriteLayer(); err != nil {
					return err
				}
			case *Dot1qLayer:
				if err := d.WriteLayer(); err != nil {
					return err
				}
			case *EchoLayer:
				if err := d.WriteLayer(); err != nil {
					return err
				}
			case *EtherLayer:
				if err := d.WriteLayer(); err != nil {
					return err
				}
			case *ICMPv4Layer:
				if err := d.WriteLayer(); err != nil {
					return err
				}
			case *ICMPv6Layer:
				if err := d.WriteLayer(); err != nil {
					return err
				}
			case *IPv4Layer:
				if err := d.WriteLayer(); err != nil {
					return err
				}
			case *IPv6Layer:
				if err := d.WriteLayer(); err != nil {
					return err
				}
			case *PayloadLayer:
				if err := d.WriteLayer(); err != nil {
					return err
				}
			case *QinQLayer:
				if err := d.WriteLayer(); err != nil {
					return err
				}
			case *SCTPLayer:
				if err := d.WriteLayer(); err != nil {
					return err
				}
			case *TCPLayer:
				if err := d.WriteLayer(); err != nil {
					return err
				}
			case *TSCLayer:
				if err := d.WriteLayer(); err != nil {
					return err
				}
			case *UDPLayer:
				if err := d.WriteLayer(); err != nil {
					return err
				}
			case *VxLanLayer:
				if err := d.WriteLayer(); err != nil {
					return err
				}
			case *DefaultsLayer:
				if err := d.WriteLayer(); err != nil {
					return err
				}
			default:
				return fmt.Errorf("Frame.toBinaryWrite:invalid layer type: '%T'", layer)
			}
		} else {
			return fmt.Errorf("unknown layer: %s", s.Name)
		}
	}

	return nil
}

func (fr *Frame) toBinaryUpdateL4Checksum() error {

	if payload, ok := fr.GetLayer(LayerPayload).(*PayloadLayer); ok {
		dbug.Printf("%v\n", payload)

		if ip, ok := fr.GetLayer(LayerIPv4).(*IPv4Layer); ok {
			if udp, ok := fr.GetLayer(LayerUDP).(*UDPLayer); ok {
				dbug.Printf("**** UDP Checksum: %v\n", udp.udpHdr.Checksum)
				if udp.udpHdr.Checksum {
					data := fr.frame
					off := fr.GetOffset(LayerUDP)
					len := fr.GetLength(LayerUDP)
					len -= UDPHeaderLen

					var d []byte
					if len > 0 {
						d = data.Bytes()[off+UDPHeaderLen : (off + UDPHeaderLen + len)]
					} else {
						d = data.Bytes()[off+UDPHeaderLen:]
					}
					cksum := IPv4UDPChecksum(&ip.ipHdr, &udp.udpHdr, d)

					dbug.Printf("data length: %v, UDP checksum at %d, %#x\n",
						data.Len(), off+uint16(UDPChecksumOffset), cksum)
					data.WriteValueAt(int(off+uint16(UDPChecksumOffset)), cksum)
				}
			} else if tcp, ok := fr.GetLayer(LayerTCP).(*TCPLayer); ok {
				data := fr.frame
				off := fr.GetOffset(LayerTCP)
				len := fr.GetLength(LayerTCP)
				len -= tcp.tcpHdr.HdrLen

				var d []byte
				if len > 0 {
					d = data.Bytes()[off+TCPHeaderLen : (off + TCPHeaderLen + len)]
				} else {
					d = data.Bytes()[off+TCPHeaderLen:]
				}
				cksum := IPv4TCPChecksum(&ip.ipHdr, &tcp.tcpHdr, d)

				dbug.Printf("data length: %v, TCP checksum at %d, %#x\n",
					data.Len(), off+uint16(TCPChecksumOffset), cksum)
				data.WriteValueAt(int(off+uint16(TCPChecksumOffset)), cksum)
			}
			dbug.Printf("%v\n", ip)
		} else if ip, ok := fr.GetLayer(LayerIPv6).(*IPv6Layer); ok {
			if udp, ok := fr.GetLayer(LayerUDP).(*UDPLayer); ok {
				udp.udpHdr.Length += uint16(payload.length)
				ip.ip6Hdr.PayloadLen += int(udp.udpHdr.Length)
				dbug.Printf("%v\n", udp)
			} else if tcp, ok := fr.GetLayer(LayerTCP).(*TCPLayer); ok {
				ip.ip6Hdr.PayloadLen += int(tcp.tcpHdr.HdrLen)
				dbug.Printf("%v\n", tcp)
			}
			dbug.Printf("%v\n", ip)
		}
	}

	return nil
}

func (fr *Frame) layerInfoNew(lName, lOptions string) (*LayerInfo, error) {

	lType := layerTypeFromName(lName)
	if lType == MaxLayerType {
		return nil, fmt.Errorf("unknown layer type: '%s'", lName)
	}

	li := &LayerInfo{
		Name: findLayerName(lName),
		Opts: lOptions,
		Type: lType,
	}
	fr.layerInfo = append(fr.layerInfo, li)

	return li, nil
}

func (fr *Frame) toBinaryAddDefaultLayers() error {

	// Add a payload layer to the frame if one doesn't already exist
	if _, ok := fr.GetLayer(LayerPayload).(*PayloadLayer); !ok {
		opts := ""
		l := newFuncs.payloadNewFn(fr)
		if err := l.Parse(opts); err != nil {
			return err
		}
		fr.layersMap[LayerPayload] = l

		li := &LayerInfo{
			Name:  LayerPayload,
			Opts:  opts,
			Type:  LayerPayloadType,
			Layer: l,
		}
		fr.layerInfo = append(fr.layerInfo, li)
	}

	// Add a payload layer to the frame if one doesn't already exist
	if _, ok := fr.GetLayer(LayerCount).(*CountLayer); !ok {
		opts := "1"
		l := newFuncs.countNewFn(fr)
		if err := l.Parse(opts); err != nil {
			return err
		}
		fr.layersMap[LayerCount] = l

		li := &LayerInfo{
			Name:  LayerCount,
			Opts:  opts,
			Type:  LayerCountType,
			Layer: l,
		}
		fr.layerInfo = append(fr.layerInfo, li)
	}

	return nil
}

func (f *FrameSerde) toBinaryFrame(frameString string, frameType FrameType) (*Frame, error) {

	// split the string into slices of strings for the layers then trim those layer strings
	splitLayers := func(str string) []string {
		var layers []string

		for _, v := range strings.Split(str, "/") {
			layers = append(layers, strings.TrimSpace(v))
		}
		return layers
	}

	// Split frame string into slices of strings name and frame text
	s := strings.Split(frameString, ":=")
	if len(s) != 2 {
		return nil, fmt.Errorf("missing ':=' <FrameName>:=<Layers>: %v", frameString)
	}
	k := strings.TrimSpace(s[0]) // Frame name or key
	v := strings.TrimSpace(s[1]) // Frame string to be encoded

	key := FrameKey{name: k, ftype: NormalFrameType}
	if _, ok := f.frames[key]; ok {
		return nil, fmt.Errorf("duplicate frame name: %v", k)
	}

	// split up the frame string into the layers for encoding later
	fr := &Frame{
		serde:     f,
		frameType: frameType,
		name:      k,
		layersMap: make(LayerMap, 0),
		protocols: make([]*ProtoInfo, 0),
		frame:     &MyBuffer{Buf: bytes.Buffer{}},
	}

	layers := splitLayers(v)
	if len(layers) == 0 {
		return nil, fmt.Errorf("empty frame string: %v", v)
	}

	dbug.DoPrintf("\n%s\n", k)

	// Build up the layer information and parse each layer
	for _, str := range layers {
		lName, lOptions := fr.splitLayerString(str)
		if len(lName) == 0 {
			return nil, fmt.Errorf("invalid layer string: '%s'", str)
		}

		if li, err := fr.layerInfoNew(lName, lOptions); err != nil {
			return nil, err
		} else {
			// process the layers into the encoded frame data
			if err := fr.toBinaryParse(li); err != nil {
				return nil, err
			}
		}
	}

	if frameType == NormalFrameType {
		if err := fr.toBinaryAddDefaultLayers(); err != nil {
			return nil, err
		}

		if err := fr.toBinaryApplyDefaults(); err != nil {
			return nil, err
		}

		if err := fr.toBinaryUpdateLengths(); err != nil {
			return nil, err
		}

		if err := fr.toBinaryWriteLayer(); err != nil {
			return nil, err
		}

		if err := fr.toBinaryUpdateL4Checksum(); err != nil {
			return nil, err
		}
	}

	return fr, nil
}

// StringToBinary converts a string formatted frame to a binary frame
func (f *FrameSerde) StringToBinary(frameString string) error {

	// Trim off any whitespace 'foo:=Ether()/Payload()' from string
	frameString = strings.TrimSpace(frameString)
	if len(frameString) == 0 {
		return fmt.Errorf("empty frame string")
	}

	// Trim off any trailing '/' in the frame string
	frameString = strings.TrimRight(frameString, "/")

	// deserialize the frame string into each layer
	if frame, err := f.toBinaryFrame(frameString, NormalFrameType); err != nil {
		return err
	} else {
		key := FrameKey{name: frame.name, ftype: NormalFrameType}
		f.frames[key] = frame
		f.frameNames = append(f.frameNames, frame.name)
	}
	return nil
}

// StringsToBinary converts a slice of string frames to a binary formatted frames
func (f *FrameSerde) StringsToBinary(frameStrings []string) error {

	if len(frameStrings) == 0 {
		return fmt.Errorf("empty slice of frame strings")
	}

	for _, frameString := range frameStrings {
		if err := f.StringToBinary(frameString); err != nil {
			return err
		}
	}

	return nil
}

func (f *FrameSerde) DefaultToBinary(frameString string) error {

	r, err := regexp.Compile(`\s*(?i)defaults\s*\(`)
	if err != nil {
		return err
	}

	// Do not allow a Defaults() option in the default frames
	if r.MatchString(frameString) {
		return fmt.Errorf("cannot encode Defaults() option in default frames\n%v", frameString)
	}

	if frame, err := f.toBinaryFrame(frameString, DefaultFrameType); err != nil {
		return err
	} else {
		key := FrameKey{name: frame.name, ftype: DefaultFrameType}
		f.frames[key] = frame
		f.frameNames = append(f.frameNames, frame.name)
	}
	return nil
}

func (f *FrameSerde) DefaultsToBinary(frameStrings []string) error {

	if len(frameStrings) == 0 {
		return fmt.Errorf("empty frame string slice")
	}

	for _, frameString := range frameStrings {
		if err := f.DefaultToBinary(frameString); err != nil {
			return err
		}
	}

	return nil
}
