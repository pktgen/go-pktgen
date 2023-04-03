// SPDX-License-Identifier: BSD-3-Clause
// Copyright (c) 2022-2024 Intel Corporation

package meter

import (
	"fmt"
	"math"
	"strings"
)

type LabelInfo struct {
	Val string
	Fn  func(a interface{}, w ...interface{}) string
}

type Info struct {
	Labels []*LabelInfo
	Bar    *LabelInfo
}

type WidthFunc func() int
type DrawFunc func(mi *Info) string

type Meter struct {
	width             WidthFunc
	draw              DrawFunc
	rateLow, rateHigh float64
}

func New() *Meter {

	return &Meter{}
}

func (m *Meter) SetWidth(width WidthFunc) *Meter {
	m.width = width
	return m
}

func (m *Meter) SetDraw(draw DrawFunc) *Meter {
	m.draw = draw
	return m
}

func (m *Meter) SetRateLimits(rateLow, rateHigh float64) *Meter {
	m.rateLow = rateLow
	m.rateHigh = rateHigh
	return m
}

func (m *Meter) clamp(x, low, high float64) float64 {

	if x > m.rateHigh {
		return m.rateHigh
	}
	if x < m.rateLow {
		return m.rateLow
	}
	return x
}

func (m *Meter) Draw(rate float64, mi *Info) string {

	width := m.width()

	var labelLen int = 0
	for _, l := range mi.Labels {
		labelLen += len(l.Val)
	}
	labelLen += 2 // Account for the brackets

	width -= labelLen

	if width <= 0 {
		return fmt.Sprintf("Invalid rate %v\n", rate)
	}
	rate = m.clamp(rate, m.rateLow, m.rateHigh)
	if rate > 0 {
		rate = math.Ceil((rate / m.rateHigh) * float64(width))
	}

	mi.Bar.Val = strings.Repeat("|", int(rate)) + strings.Repeat(" ", width-int(rate))

	return m.draw(mi)
}
