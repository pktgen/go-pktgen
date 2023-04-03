/* SPDX-License-Identifier: BSD-3-Clause
 * Copyright (c) 2022-2024 Intel Corporation.
 */

package fserde

import (
	"encoding/binary"
	"fmt"
	"strconv"
	"strings"
)

type fillType uint8

const (
	fillTypeNone fillType = iota
	fill8Type
	fill16Type
	fill32Type
	fill64Type
	fillStringType
)

type PayloadLayer struct {
	hdr    *LayerHdr
	length uint16
	fill   fillType
	data   []byte
}

func (pl *PayloadLayer) String() string {
	s := fmt.Sprintf("Payload(length=%d, ", pl.length)
	switch pl.fill {
	case fillStringType:
		s += fmt.Sprintf("string=%q)", pl.data)
	case fill8Type:
		s += fmt.Sprintf("fill=%#x)", pl.data[0])
	case fill16Type:
		s += fmt.Sprintf("fill16=%#x)", binary.BigEndian.Uint16(pl.data))
	case fill32Type:
		s += fmt.Sprintf("fill32=%#x)", binary.BigEndian.Uint32(pl.data))
	case fill64Type:
		s += fmt.Sprintf("fill64=%#x)", binary.BigEndian.Uint64(pl.data))
	default:
		s += "fill=Unknown"
	}
	return s + ")"
}

func PayloadNew(fr *Frame) *PayloadLayer {
	return &PayloadLayer{
		hdr:    LayerConstructor(fr, LayerPayload, LayerPayloadType),
		length: 0,
		fill:   fillTypeNone,
		data:   []byte{},
	}
}

func (l *PayloadLayer) Name() LayerName {
	return l.hdr.layerName
}

func (l *PayloadLayer) Parse(opts string) error {

	if len(opts) > 0 {
		options := strings.Split(opts, ",")

		for _, opt := range options {
			opt = strings.TrimSpace(opt)

			kvp := strings.Split(opt, "=")
			if len(kvp) != 2 {
				return fmt.Errorf("invalid option: %s", opt)
			}

			key := strings.ToLower(strings.TrimSpace(kvp[0]))
			val := strings.TrimSpace(kvp[1])

			switch key {
			case "size", "length", "len":
				if v, err := strconv.ParseInt(val, 0, 0); err != nil {
					return err
				} else {
					l.length = uint16(v)
				}
			case "fill", "fill8":
				if v, err := strconv.ParseInt(val, 0, 0); err != nil {
					return err
				} else {
					l.fill = fill8Type
					l.data = []byte{byte(uint8(v))}
				}
			case "fill16":
				if v, err := strconv.ParseInt(val, 0, 0); err != nil {
					return err
				} else {
					l.fill = fill16Type
					l.data = binary.BigEndian.AppendUint16([]byte{}, uint16(v))
				}
			case "fill32":
				if v, err := strconv.ParseInt(val, 0, 0); err != nil {
					return err
				} else {
					l.fill = fill32Type
					l.data = binary.BigEndian.AppendUint32([]byte{}, uint32(v))
				}
			case "fill64":
				if v, err := strconv.ParseInt(val, 0, 0); err != nil {
					return err
				} else {
					l.fill = fill64Type
					l.data = binary.BigEndian.AppendUint64([]byte{}, uint64(v))
				}
			case "string":
				// trim the quotes from the string
				val = strings.TrimLeft(val, "'")
				val = strings.TrimRight(val, "'")

				l.fill = fillStringType
				for i := 0; i < len(val); i++ {
					l.data = append(l.data, val[i])
				}
			}
		}
	}

	// The size is not given use the length of the data byte slice
	if l.length == 0 && len(l.data) > 0 {
		l.length = uint16(len(l.data))
	}

	l.hdr.proto.name = l.Name()
	l.hdr.proto.offset = l.hdr.fr.GetOffset(l.Name())
	l.hdr.proto.length = l.length

	l.hdr.fr.AddProtocol(&l.hdr.proto)

	dbug.Printf("%v\n", l)
	return nil
}

func (l *PayloadLayer) ApplyDefaults() error {

	d := l.hdr.fr.defaultsFrame
	if d == nil {
		return nil
	}

	if dl, ok := d.GetLayer(LayerPayload).(*PayloadLayer); !ok {
		return nil
	} else {
		if l.length == 0 && dl.length != 0 {
			l.length = dl.length
		}
		if l.fill == fillTypeNone && dl.fill != fillTypeNone {
			l.fill = dl.fill
			l.data = dl.data
		}
		l.hdr.fr.GetProtocol(l.Name()).length = l.length
	}
	dbug.Printf("%v\n", l)

	return nil
}

func (l *PayloadLayer) WriteLayer() error {

	fr := l.hdr.fr
	data := fr.frame

	dbug.Printf("Data Length: %d\n", len(l.data))
	k := 0
	for i := uint16(0); i < l.length; i++ {
		data.Append(l.data[k])
		k++
		if k >= len(l.data) {
			k = 0
		}
	}

	dbug.Printf("Frame Length: %d\n", l.length)

	return nil
}
