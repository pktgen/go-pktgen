/* SPDX-License-Identifier: BSD-3-Clause
 * Copyright (c) 2022-2024 Intel Corporation.
 */

package fserde

import (
	"fmt"
	"regexp"
	"strings"
)

type QinQLayer struct {
	hdr *LayerHdr
	q   [2]*Dot1qLayer
}

func (e *QinQLayer) String() string {

	q1 := fmt.Sprintf("%v", e.q[0])
	q2 := fmt.Sprintf("%v", e.q[1])

	r := regexp.MustCompile(`\s*(?i)dot1q\(`)

	q1 = r.ReplaceAllLiteralString(q1, "Dot1q{")
	q1 = strings.Replace(q1, ")", "}", -1)

	q2 = r.ReplaceAllLiteralString(q2, "Dot1q{")
	q2 = strings.Replace(q2, ")", "}", -1)

	str := fmt.Sprintf("QinQ(%v, %v)", q1, q2)

	return str
}

func QinQNew(fr *Frame) *QinQLayer {
	qinq := &QinQLayer{
		hdr: LayerConstructor(fr, LayerQinQ, LayerQinQType),
	}
	qinq.q[0] = Dot1qNew(fr)
	qinq.q[1] = Dot1qNew(fr)
	return qinq
}

func (l *QinQLayer) Name() LayerName {
	return l.hdr.layerName
}

func (l *QinQLayer) Parse(opts string) error {

	options := strings.Split(opts, ",")

	for i, opt := range options {
		opt = strings.TrimSpace(opt)

		str := strings.FieldsFunc(opt, func(r rune) bool {
			return r == '{' || r == '}'
		})
		if len(str) == 2 {
			str[0] = strings.TrimSpace(str[0])
			str[1] = strings.TrimSpace(str[1])
		} else if len(str) == 1 {
			str[0] = strings.TrimSpace(str[0])
			str = append(str, "")
		} else {
			return fmt.Errorf("invalid QinQ options: %v", str)
		}

		if !strings.EqualFold(str[0], "Dot1q") {
			return fmt.Errorf("invalid QinQ should be Dot1q{...}: %s", str[0])
		}

		if len(str[1]) > 0 {
			if err := l.q[i].ParseDot1q(str[1]); err != nil {
				return err
			}
		}
	}
	l.q[0].dot1q.tPid = QinQID

	dbug.Printf("%v\n", l)

	l.hdr.proto.name = l.Name()
	l.hdr.proto.offset = l.hdr.fr.GetOffset(l.Name())
	l.hdr.proto.length = 8

	l.hdr.fr.AddProtocol(&l.hdr.proto)

	return nil
}

func (l *QinQLayer) ApplyDefaults() error {

	d := l.hdr.fr.defaultsFrame
	if d == nil {
		return nil
	}

	dl, ok := d.GetLayer(LayerQinQ).(*QinQLayer)
	if !ok {
		return nil
	}

	dbug.Printf("%v\n", dl)

	return nil
}

func (l *QinQLayer) WriteLayer() error {

	data := l.hdr.fr.frame

	dstSrcMAC := make([]byte, 12)
	copy(dstSrcMAC, data.Bytes()[:12])
	restData := make([]byte, data.Len()-12)
	copy(restData, data.Bytes()[12:])

	data.Reset()
	data.Append(dstSrcMAC)
	data.Append(l.q[0].dot1q.tPid)
	data.Append(l.q[0].dot1q.tci)
	data.Append(l.q[1].dot1q.tPid)
	data.Append(l.q[1].dot1q.tci)
	data.Append(restData)

	dbug.Printf("%v\n", data.Bytes())

	return nil
}
