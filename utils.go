// Copyright 2013 Vedran Vuk. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package irc

import (
	"strings"
)

func isNum(s string) bool {
	for i := 0; i < len(s); i++ {
		if s[i] < 48 || s[i] > 57 {
			return false
		}
	}
	return true
}

// Strips IRC text formatting codes from "text" and returns the result.
func StripControlCodes(text string) (r string) {
	c := strings.Split(text, "")
	m := 0
	i := 0
	for i < len(c) {
		switch m {
		case 0:
			switch rune(c[i][0]) {
			case 0x02, 0x0F, 0x16, 0x1D, 0x1F:
				i++
				continue
			case 0x03:
				m = 1
			default:
				r += c[i]
			}
		case 1:
			if isNum(c[i]) {
				m = 2
			} else {
				m = 0
				continue
			}
		case 2:
			if isNum(c[i]) {
				m = 3
			} else if c[i] == "," {
				m = 4
			} else {
				m = 0
				continue
			}
		case 3:
			if c[i] == "," {
				if i+1 < len(c) {
					if isNum(c[i+1]) {
						m = 4
					} else {
						m = 0
						continue
					}
				} else {
					m = 0
					continue
				}
				m = 4
			} else {
				m = 0
				continue
			}
		case 4:
			if isNum(c[i]) {
				m = 5
			} else {
				m = 0
				continue
			}
		case 5:
			if isNum(c[i]) {
				m = 0
			} else {
				m = 0
				continue
			}
		}
		i++
	}
	return
}

// ChannelMode defines channel modes of a user on a channel.
type ChannelMode int

// Channel mode of a user on a channel.
const (
	CmVoice  ChannelMode = 1 << iota // Voiced user.
	CmHalfOp                         // Half-operator/Helper.
	CmOp                             // Channel operator.
	CmAdmin                          // Channel admin.
	CmOwner                          // Channel owner.
)

// Parses out the channel modes from a nickname with channel mode prefix(es)
// For example: s with value "&@somenick" will return "somenick" and a
// union of channel modes.
func ParseChannelModes(s string) (nick string, cm ChannelMode) {
	for i := 0; i < len(s); i++ {
		switch string(s[i]) {
		case "~":
			cm |= CmOwner
		case "&":
			cm |= CmAdmin
		case "@":
			cm |= CmOp
		case "%":
			cm |= CmHalfOp
		case "+":
			cm |= CmVoice
		default:
			return s[i:], cm
		}
	}
	return "", 0
}
