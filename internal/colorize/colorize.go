// SPDX-License-Identifier: BSD-3-Clause
// Copyright (c) 2023-2025 Intel Corporation

package colorize

import (
	"fmt"
	"strings"

	tcell "github.com/gdamore/tcell/v2"
	"golang.org/x/text/language"
	"golang.org/x/text/message"
)

// colorizeInfo structure
type colorizeInfo struct {
	defWidth       int
	floatPrecision int
	defForeground  string
	defBackground  string
	defFlags       string
}

var colorInfo colorizeInfo

// Default values for width and precision
const (
	defWidth     = int(0)
	defPrecision = int(2)
)

// Color constant names we can use
const (
	NoColor                = ""
	DefaultColor           = "white"
	YellowColor            = "yellow"
	GreenColor             = "green"
	GoldenRodColor         = "goldenrod"
	OrangeColor            = "orange"
	TealColor              = "teal"
	CornSilkColor          = "cornsilk"
	DeepPinkColor          = "deeppink"
	RedColor               = "red"
	BlueColor              = "blue"
	LightBlueColor         = "lightblue"
	LightCoralColor        = "lightcoral"
	LightCyanColor         = "lightcyan"
	LavenderColor          = "lavender"
	LightSalmonColor       = "lightsalmon"
	MediumBlueColor        = "mediumblue"
	MistyRoseColor         = "mistyrose"
	SkyBlueColor           = "skyblue"
	LightSkyBlueColor      = "lightskyblue"
	MediumSpringGreenColor = "mediumspringgreen"
	WheatColor             = "wheat"
	YellowGreenColor       = "yellowgreen"
	LightYellowColor       = "lightyellow"
	DarkOrangeColor        = "darkorange"
	LightGreenColor        = "lightgreen"
	DarkMagentaColor       = "darkmagenta"
	CyanColor              = "aqua"
)

// SetDefault initializes the colorizeInfo struct with default values for color, width, precision, and flags.
// If precision is negative, it is set to the default value.
//
// Parameters:
// - foreground: A string representing the default foreground color.
// - background: A string representing the default background color.
// - width: An integer representing the default width of the field.
// - precision: An integer representing the default precision for float values.
// - flags: A string representing the default flags for color attributes.
func SetDefault(foreground, background string, width, precision int, flags string) {

	// when precision is negative then set to the default value
	if precision < 0 {
		precision = defPrecision
	}

	colorInfo = colorizeInfo{
		defWidth:       width,
		floatPrecision: precision,
		defForeground:  foreground,
		defBackground:  background,
		defFlags:       flags,
	}
}

// DefaultForegroundColor returns the default color
func DefaultForegroundColor() string {
	return colorInfo.defForeground
}

// SetDefaultForegroundColor - Set the default foreground color
func SetDefaultForegroundColor(color string) {
	colorInfo.defForeground = color
}

// DefaultBackgroundColor returns the default color
func DefaultBackgroundColor() string {
	return colorInfo.defBackground
}

// SetDefaultBackgroundColor - Set the default background color
func SetDefaultBackgroundColor(color string) {
	colorInfo.defBackground = color
}

// DefaultWidth returns the default width
func DefaultWidth() int {
	return colorInfo.defWidth
}

// SetDefaultWidth - Set the default width
func SetDefaultWidth(w int) {
	colorInfo.defWidth = w
}

// SetFloatPrecision - Set float precision
func SetFloatPrecision(w int) {
	colorInfo.floatPrecision = w
}

// FloatPrecision returns the default precision
func FloatPrecision() int {
	return colorInfo.floatPrecision
}

// DefaultFlags returns the default flags
func DefaultFlags() string {
	return colorInfo.defFlags
}

// SetDefaultFlags - Set flags
func SetDefaultFlags(f string) {
	colorInfo.defFlags = f
}

// Colorize function is used to add color to the value passed, with optional width, precision, foreground color, background color, and flags.
//
// Parameters:
// - color: A string representing the color to be applied. If empty, the default foreground color will be used.
// - v: The value to be colorized. It can be of type string, integer, float, or any other type.
// - w: A variadic parameter that can contain optional width, precision, foreground color, background color, and flags.
//
// Returns:
// - A string with the colorized value.
//
// Usage:
//
//	colorize.Colorize("red", "Hello, World!", 20, 2, "yellow", "blue")
//	colorize.Colorize("", 123.456, 10, 2)
//	colorize.Colorize("green", "Goodbye", -10) // Left alignment
func Colorize(color string, v interface{}, w ...interface{}) string {
	// Set default foreground color if not provided
	if colorInfo.defForeground == "" {
		colorInfo.defForeground = "ivory"
	}

	// Initialize variables
	width := int(0)
	precision := colorInfo.floatPrecision
	foreground := colorInfo.defForeground
	if len(color) > 0 {
		foreground = color
	}
	background := colorInfo.defBackground
	flags := colorInfo.defFlags

	// Process optional parameters
	for i, v := range w {
		switch i {
		case 0: // Width of the field
			p := v.(int)
			width = p
		case 1: // Precision of the float value
			p := v.(int)
			if p >= 0 {
				precision = p
			}
		case 2: // foreground color
			s := v.(string)
			if len(s) > 0 {
				foreground = s
			}
		case 3: // background color
			s := v.(string)
			if len(s) > 0 {
				background = s
			}
		case 4: // flags used for color attributes
			s := v.(string)
			if len(s) > 0 {
				flags = s
			}
		}
	}

	// Build up the color tag strings for begin and end of the field to be printed
	str := fmt.Sprintf("[%s:%s:%s]", foreground, background, flags)
	def := fmt.Sprintf("[%s:%s:%s]", colorInfo.defForeground, colorInfo.defBackground, colorInfo.defFlags)

	p := message.NewPrinter(language.English)

	// Switch case to handle different types of values
	switch v.(type) {
	case string:
		return fmt.Sprintf("%[1]s%[3]*[2]s%[4]s", str, v, width, def)
	case uint64, uint32, uint16, uint8:
		return p.Sprintf("%[1]s%[3]*[2]d%[4]s", str, v, width, def)
	case int, int64, int32, int16, int8:
		return p.Sprintf("%[1]s%[3]*[2]d%[4]s", str, v, width, def)
	case float64, float32:
		return p.Sprintf("%[1]s%[3]*.[4]*[2]f%[5]s", str, v, width, precision, def)
	default:
		return fmt.Sprintf("%[1]s%[2]v%s", str, v, def)
	}
}

// ColorWithName - Find and set the color by name.
// This function takes a color name as a string, converts it to lowercase,
// and checks if it exists in the tcell.ColorNames map. If it does, the color is used.
// If it doesn't, the color is set to the default color, OrangeColor.
// Then, it calls the Colorize function with the determined color, the provided value,
// and any additional parameters.
//
// Parameters:
// - color: A string representing the color name.
// - a: The value to be colorized. It can be of type string, integer, float, or any other type.
// - w: A variadic parameter that can contain optional width, precision, foreground color, background color, and flags.
//
// Returns:
// - A string with the colorized value.
//
// Usage:
//
//	colorize.ColorWithName("red", "Hello, World!", 20, 2, "yellow", "blue")
//	colorize.ColorWithName("", 123.456, 10, 2)
//	colorize.ColorWithName("green", "Goodbye", -10) // Left alignment
func ColorWithName(color string, a interface{}, w ...interface{}) string {

	color = strings.ToLower(color)

	_, ok := tcell.ColorNames[color]

	if !ok {
		color = OrangeColor
	}
	return Colorize(color, a, w...)
}

// Default - The default color for the text
func Default(a interface{}, w ...interface{}) string {

	return ColorWithName(DefaultForegroundColor(), a, w...)
}

// Yellow - return string based on the color given
func Yellow(a interface{}, w ...interface{}) string {

	return ColorWithName(YellowColor, a, w...)
}

// DarkMagenta - return string based on the color given
func DarkMagenta(a interface{}, w ...interface{}) string {

	return ColorWithName(DarkMagentaColor, a, w...)
}

// Green - return string based on the color given
func Green(a interface{}, w ...interface{}) string {

	return ColorWithName(GreenColor, a, w...)
}

// GoldenRod - return string based on the color given
func GoldenRod(a interface{}, w ...interface{}) string {

	return ColorWithName(GoldenRodColor, a, w...)
}

// Orange - return string based on the color given
func Orange(a interface{}, w ...interface{}) string {

	return ColorWithName(OrangeColor, a, w...)
}

// Teal - return string based on the color given
func Teal(a interface{}, w ...interface{}) string {

	return ColorWithName(TealColor, a, w...)
}

// CornSilk - return string based on the color given
func CornSilk(a interface{}, w ...interface{}) string {

	return ColorWithName(CornSilkColor, a, w...)
}

// DeepPink - return string based on the color given
func DeepPink(a interface{}, w ...interface{}) string {

	return ColorWithName(DeepPinkColor, a, w...)
}

// Red - return string based on the color given
func Red(a interface{}, w ...interface{}) string {

	return ColorWithName(RedColor, a, w...)
}

// Blue - return string based on the color given
func Blue(a interface{}, w ...interface{}) string {

	return ColorWithName(BlueColor, a, w...)
}

// LightBlue - return string based on the color given
func LightBlue(a interface{}, w ...interface{}) string {

	return ColorWithName(LightBlueColor, a, w...)
}

// LightCoral - return string based on the color given
func LightCoral(a interface{}, w ...interface{}) string {

	return ColorWithName(LightCoralColor, a, w...)
}

// LightCyan - return string based on the color given
func LightCyan(a interface{}, w ...interface{}) string {

	return ColorWithName(LightCyanColor, a, w...)
}

// Cyan - return string based on the color given
func Cyan(a interface{}, w ...interface{}) string {

	return ColorWithName(CyanColor, a, w...)
}

// Lavender - return string based on the color given
func Lavender(a interface{}, w ...interface{}) string {

	return ColorWithName(LavenderColor, a, w...)
}

// LightSalmon - return string based on the color given
func LightSalmon(a interface{}, w ...interface{}) string {

	return ColorWithName(LightSalmonColor, a, w...)
}

// MediumBlue - return string based on the color given
func MediumBlue(a interface{}, w ...interface{}) string {

	return ColorWithName(MediumBlueColor, a, w...)
}

// MistyRose - return string based on the color given
func MistyRose(a interface{}, w ...interface{}) string {

	return ColorWithName(MistyRoseColor, a, w...)
}

// SkyBlue - return string based on the color given
func SkyBlue(a interface{}, w ...interface{}) string {

	return ColorWithName(SkyBlueColor, a, w...)
}

// LightSkyBlue - return string based on the color given
func LightSkyBlue(a interface{}, w ...interface{}) string {

	return ColorWithName(LightSkyBlueColor, a, w...)
}

// MediumSpringGreen - return string based on the color given
func MediumSpringGreen(a interface{}, w ...interface{}) string {

	return ColorWithName(MediumSpringGreenColor, a, w...)
}

// Wheat - return string based on the color given
func Wheat(a interface{}, w ...interface{}) string {

	return ColorWithName(WheatColor, a, w...)
}

// YellowGreen - return string based on the color given
func YellowGreen(a interface{}, w ...interface{}) string {

	return ColorWithName(YellowGreenColor, a, w...)
}

// LightYellow - return string based on the color given
func LightYellow(a interface{}, w ...interface{}) string {

	return ColorWithName(LightYellowColor, a, w...)
}

// DarkOrange - return string based on the color given
func DarkOrange(a interface{}, w ...interface{}) string {

	return ColorWithName(DarkOrangeColor, a, w...)
}

// LightGreen - return string based on the color given
func LightGreen(a interface{}, w ...interface{}) string {

	return ColorWithName(LightGreenColor, a, w...)
}
