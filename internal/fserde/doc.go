/* SPDX-License-Identifier: BSD-3-Clause
 * Copyright (c) 2023-2025 Intel Corporation.
 */

/*
The FrameSerde (fserde) package provides a framework to deserialize/serialize frames
from text string format to a binary byte slice or binary to text string.

Serialize   - Converts a ethernet binary frame to a text string. ToString
Deserialize - Converts a ethernet text string to a binary ethernet frame. ToBinary
*/
package fserde

/*
# High level view of the FrameSerde package.

# FrameString basic format

The format of the FrameString to encode (text to binary) is a simple text string.
The format of the string is:
<FrameName>:=<protocol-layers>

The <FrameName> is the name of the frame, which can be used to identify the
frame. Must be a string with no white space characters. The FrameName is case
sensitive.

# Protocol-layers format

The <protocol-layers> is a delimiter-separated list of protocol-layers, where
the delimiter is a '/' character. The protocol-layer name can be: Ether, IPv4, IPv6,
ICMP, ICMPv6, UDP, TCP, ... plus a number of other protocol-layer names and information
layers. Each protocol-layer has the following format:
<protocol-name>(<protocol-value>, <protocol-value, ...)

Example:
Frame1 := Ether( dst=01:22:33:44:55:66, proto=0x800 )/
	IPv4(src=192.168.1.2, dst=192.168.1.3)/Payload(len=100)

# Protocol-value format

The protocol-name is case-insensitive along with the protocol-value names. White spaces in
the string(s) will trimmed from the beginning and end of the strings.

The example defines the name of the frame "Frame1". The protocol-layers are:
Ether, IPv4 and Payload. Each protocol-layer has a set of protocol-value pairs in a
key=value format. Look into each protocol file for more information about each protocol.

# Default frame-value format

The API via the serde.Create(cfg FrameSerdeCfg) function is the main entry point to
the package. The FrameSerdeCfg can be used to configure the structure currently
contains and array of strings defining a set of default frames.

The default frames allow a frame to be encoded using default values for a protocol-layer.
If the protocol-layer being encoded does not define all of the fields in the protocol-layer,
a default frame can be specified using the protocol-layer Defaults(<default-frame-name>)
function.

Frame1:=Ether(dst=01:22:33:44:55:66, proto=0x800)/IPv4(src=192.168.1.2, dst=192.168.1.3)/
	Payload(len=100)/Defaults(DefaultFrame-1)

The list of default frames must be specified in the FrameSerdeCfg.DefaultFrames and the
format of the default frames is the same as the format of FrameString
*/
