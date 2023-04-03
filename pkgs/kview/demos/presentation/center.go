package main

import "github.com/pktgen/go-pktgen/pkgs/kview"

// Center returns a new primitive which shows the provided primitive in its
// center, given the provided primitive's size.
func Center(width, height int, p kview.Primitive) kview.Primitive {
	subFlex := kview.NewFlex()
	subFlex.SetDirection(kview.FlexRow)
	subFlex.AddItem(kview.NewBox(), 0, 1, false)
	subFlex.AddItem(p, height, 1, true)
	subFlex.AddItem(kview.NewBox(), 0, 1, false)

	flex := kview.NewFlex()
	flex.AddItem(kview.NewBox(), 0, 1, false)
	flex.AddItem(subFlex, width, 1, true)
	flex.AddItem(kview.NewBox(), 0, 1, false)

	return flex
}
