// Copyright 2013 Vedran Vuk. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package irc

import (
	"strconv"
	"strings"
)

// Message is a message received from the IRC server. It contains helper
// methods to parse out the message details.
// See http://www.irchelp.org/irchelp/rfc/chapter2.html#c2_3 for details.
type Message string

// Creates a new Message from a string and sanitizes it if invalid.
func NewMessage(s string) Message {
	r := strings.Trim(s, string([]rune{0, 10, 13}))
	if strings.IndexAny(r, string(rune(0))) != -1 {
		return ""
	}
	return Message(r)
}

// Return as string.
func (m *Message) String() string {
	return string(*m)
}

// True if empty.
func (m *Message) IsEmpty() bool {
	return len(m.String()) == 0
}

// Returns true if message has prefix, false otherwise or if invalid.
func (m *Message) HasPrefix() bool {
	if m.IsEmpty() {
		return false
	}
	return string(m.String()[0]) == ":"
}

// Returns true if message has the trailing parameters token ":", regardless of
// the number of trailings parameters, which can be 0.
func (m *Message) HasTrailing() bool {
	if m.IsEmpty() {
		return false
	}
	if m.HasPrefix() {
		return strings.Index(string(*m)[1:], ":") > -1
	}
	return strings.Index(string(*m), ":") > -1
}

// Returns true if message command is a numeric.
func (m *Message) IsNumeric() bool {
	_, err := strconv.Atoi(m.Command())
	return err == nil
}

// Returns message prefix, or empty string otherwise or if malformed.
func (m *Message) Prefix() *Entity {
	c := new(Entity)
	if !m.HasPrefix() {
		*c = ""
	}
	if i := strings.Index(m.String(), " "); i > -1 {
		*c = Entity(*m)[1:i]
	}
	return c
}

// Returns the command or numeric in string form, always uppercase.
func (m *Message) Command() string {
	if m.IsEmpty() {
		return ""
	}
	s := m.String()
	if m.HasPrefix() {
		if i := strings.Index(s, " "); i > -1 && i+1 < len(m.String()) {
			s = s[strings.Index(s, " ")+1:]
		} else {
			return ""
		}
	}
	return strings.ToUpper(s[:strings.Index(s, " ")])
}

// Returns the command as integer if command is a numeric, -1 otherwise.
func (m *Message) Numeric() int {
	v, err := strconv.Atoi(m.Command())
	if err != nil {
		return -1
	}
	return v
}

// Returns middle parameters as a string.
func (m *Message) Middle() string {
	if m.IsEmpty() {
		return ""
	}
	s := m.String()
	// Skip prefix.
	if m.HasPrefix() {
		s = s[1:]
		if i := strings.Index(s, " "); i > -1 && i+1 < len(s) {
			s = s[i+1:]
		} else {
			return ""
		}
	}
	// Skip command.
	if i := strings.Index(s, " "); i > -1 && i < len(s) {
		s = s[i+1:]
	} else {
		return ""
	}
	// Remove trailing if present.
	if i := strings.Index(s, ":"); i > -1 {
		s = strings.TrimRight(s[0:i], " ")
	}
	return s
}

// Returns an array of middle parameters.
func (m *Message) Middles() []string {
	return strings.Split(m.Middle(), " ")
}

// Returns trailing parameters as a string.
func (m *Message) Trailing() string {
	if m.IsEmpty() {
		return ""
	}
	s := m.String()
	// Skip prefix.
	if m.HasPrefix() {
		s = s[1:]
		if i := strings.Index(s, " "); i > -1 && i+1 < len(s) {
			s = s[i+1:]
		} else {
			return ""
		}
	}
	if i := strings.Index(s, ":"); i > -1 && i+1 < len(s) {
		s = s[i+1:]
	} else {
		return ""
	}
	return s
}

// Returns an array of trailing parameters.
func (m *Message) Trailings() []string {
	return strings.Split(m.Trailing(), " ")
}
