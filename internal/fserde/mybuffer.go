/* SPDX-License-Identifier: BSD-3-Clause
 * Copyright (c) 2022-2024 Intel Corporation.
 */

package fserde

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"net"
)

type MyBuffer struct {
	Buf bytes.Buffer
}

func (b *MyBuffer) Bytes() []byte {

	return b.Buf.Bytes()
}

func (b *MyBuffer) Len() int {

	return b.Buf.Len()
}

func (b *MyBuffer) Cap() int {

	return b.Buf.Cap()
}

func (b *MyBuffer) Reset() {

	b.Buf.Reset()
}

func (b *MyBuffer) WriteByte(p byte) error {

	return b.Buf.WriteByte(p)
}

func (b *MyBuffer) Write(p []byte) (n int, err error) {

	return b.Buf.Write(p)
}

func (b *MyBuffer) Read(p []byte) (n int, err error) {

	return b.Buf.Read(p)
}

func (b *MyBuffer) WriteAt(index int, val []byte) error {

	if index+len(val) > b.Buf.Len() {
		return fmt.Errorf("index + value length exceeds buffer size")
	}

	copy(b.Buf.Bytes()[index:], val)

	return nil
}

func (b *MyBuffer) WriteValueAt(index int, val interface{}) error {

	var byteLen int

	switch v := val.(type) {
	case int8, uint8:
		byteLen = 1
	case int16, uint16:
		byteLen = 2
	case int32, uint32:
		byteLen = 4
	case int64, uint64:
		byteLen = 8
	case []byte:
		byteLen = len(v)
	case string:
		byteLen = len(v)
	case net.IP:
		byteLen = net.IPv4len
	case net.HardwareAddr:
		byteLen = HardwareAddrLen
	default:
		return fmt.Errorf("unsupported value type %T", val)
	}
	if index+byteLen > b.Buf.Len() {
		return fmt.Errorf("index + value length exceeds buffer size")
	}

	buf := make([]byte, byteLen)

	switch v := val.(type) {
	case int8:
		buf[0] = byte(v)
	case uint8:
		buf[0] = byte(v)
	case int16:
		buf = binary.BigEndian.AppendUint16(buf, uint16(v))
	case uint16:
		buf = binary.BigEndian.AppendUint16(buf, v)
	case int32:
		buf = binary.BigEndian.AppendUint32(buf, uint32(v))
	case uint32:
		buf = binary.BigEndian.AppendUint32(buf, v)
	case int64:
		buf = binary.BigEndian.AppendUint64(buf, uint64(v))
	case uint64:
		buf = binary.BigEndian.AppendUint64(buf, v)
	case []byte:
		buf = v
	case string:
		buf = []byte(v)
	case net.IP:
		buf = v.To4()
	case net.HardwareAddr:
		buf = v
	default:
		fmt.Printf("MyBuffer.WriteValueAt:unsupported type: %T\n", v)
	}

	dbug.Printf("MyBuffer.WriteValueAt: %v @ %v\n", buf, index)
	loc := b.Buf.Bytes()
	copy(loc[index:], buf)

	return nil
}

func (b *MyBuffer) Append(val interface{}) {

	switch v := val.(type) {
	case int8:
		b.Buf.WriteByte(uint8(v))
	case uint8:
		b.Buf.WriteByte(v)
	case int16:
		b.Buf.Write(binary.BigEndian.AppendUint16([]byte{}, uint16(v)))
	case uint16:
		b.Buf.Write(binary.BigEndian.AppendUint16([]byte{}, v))
	case int32:
		b.Buf.Write(binary.BigEndian.AppendUint32([]byte{}, uint32(v)))
	case uint32:
		b.Buf.Write(binary.BigEndian.AppendUint32([]byte{}, v))
	case int64:
		b.Buf.Write(binary.BigEndian.AppendUint64([]byte{}, uint64(v)))
	case uint64:
		b.Buf.Write(binary.BigEndian.AppendUint64([]byte{}, v))
	case string:
		b.Buf.WriteString(v)
	case []byte:
		b.Buf.Write(v)
	case net.IP:
		b.Buf.Write(v.To4())
	case net.HardwareAddr:
		b.Buf.Write(v)
	case *MyBuffer:
		b.Buf.Write(v.Buf.Bytes())
	default:
		fmt.Printf("MyBuffer.Append:unsupported type: %T\n", v)
	}
}

// BufferDump is a helper function that returns a string representation of the
// given frame.
func (b *MyBuffer) BufferDump() string {

	s := fmt.Sprintf("len %d:\n  ", b.Len())
	for _, v := range b.Bytes() {
		s += fmt.Sprintf("%02x ", v)
	}
	s += "\n"

	return s
}
