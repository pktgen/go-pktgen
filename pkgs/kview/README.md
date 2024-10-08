# cview - Terminal-based user interface toolkit

[![GoDoc](https://code.rocketnine.space/tslocum/godoc-static/raw/branch/master/badge.svg)](https://docs.rocketnine.space/code.rocketnine.space/tslocum/cview)
[![Donate via LiberaPay](https://img.shields.io/liberapay/receives/rocketnine.space.svg?logo=liberapay)](https://liberapay.com/rocketnine.space)
[![Donate via Patreon](https://img.shields.io/badge/dynamic/json?color=%23e85b46&label=Patreon&query=data.attributes.patron_count&suffix=%20patrons&url=https%3A%2F%2Fwww.patreon.com%2Fapi%2Fcampaigns%2F5252223)](https://www.patreon.com/rocketnine)

This package is a fork of [tview](https://github.com/rivo/tview).
See [FORK.md](https://code.rocketnine.space/tslocum/cview/src/branch/master/FORK.md)
for more information.

## Demo

`ssh kview.rocketnine.space -p 20000`

[![Recording of presentation demo](https://code.rocketnine.space/tslocum/cview/raw/branch/master/kview.svg)](https://code.rocketnine.space/tslocum/cview/src/branch/master/demos/presentation)

## Features

Available widgets:

- **Input forms** (including **input/password fields**, **drop-down selections**
, **checkboxes**, and **buttons**)
- Navigable multi-color **text views**
- Selectable **lists** with **context menus**
- Modal **dialogs**
- Horizontal and vertical **progress bars**
- **Grid**, **Flexbox** and **tabbed panel layouts**
- Sophisticated navigable **table views**
- Flexible **tree views**
- Draggable and resizable **windows**
- An **application** wrapper

Widgets may be customized and extended to suit any application.

[Mouse support](https://docs.rocketnine.space/code.rocketnine.space/tslocum/cview#hdr-Mouse_Support)
is available.

## Applications

A list of applications powered by cview is available via [pkg.go.dev](https://pkg.go.dev/code.rocketnine.space/tslocum/cview?tab=importedby).

## Installation

```bash
go get code.rocketnine.space/tslocum/cview
```

## Hello World

This basic example creates a TextView titled "Hello, World!" and displays it in
your terminal:

```go
package main

import (
 "github.com/pktgen/go-pktgen/pkgs/kview"
)

func main() {
 app := kview.NewApplication()

 tv := kview.NewTextView()
 tv.SetBorder(true)
 tv.SetTitle("Hello, world!")
 tv.SetText("Lorem ipsum dolor sit amet")

 app.SetRoot(tv, true)
 if err := app.Run(); err != nil {
  panic(err)
 }
}
```

Examples are available via
[godoc](https://docs.rocketnine.space/code.rocketnine.space/tslocum/cview#pkg-examples)
and in the [demos](https://code.rocketnine.space/tslocum/cview/src/branch/master/demos)
directory.

For a presentation highlighting the features of this package, compile and run
the program in the [demos/presentation](https://code.rocketnine.space/tslocum/cview/src/branch/master/demos/presentation)
directory.

## Documentation

Package documentation is available via [godoc](https://docs.rocketnine.space/code.rocketnine.space/tslocum/cview).

An [introduction tutorial](https://rocketnine.space/post/tview-and-you/) is also
available.

## Dependencies

This package is based on [github.com/gdamore/tcell](https://github.com/gdamore/tcell)
(and its dependencies) and [github.com/rivo/uniseg](https://github.com/rivo/uniseg).

## Support

[CONTRIBUTING.md](https://code.rocketnine.space/tslocum/cview/src/branch/master/CONTRIBUTING.md)
describes how to share issues, suggestions and patches (pull requests).
