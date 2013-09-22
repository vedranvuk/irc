// Copyright 2013 Vedran Vuk. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package irc

import (
	"strings"
)

// Entity is a helper type providing methods to easily parse out details
// about various types of 'entities' on IRC that are the source or destination
// of messages. For example in a OnMsg (PRIVMSG) event handler it can describe
// both the source of the event as a user in the form of hostmask or a nickname,
// and the target of the event as a channel or yourself as a nickname or a
// hostmask. It is also used with IRCMessage.Prefix().
type Entity string

func NewEntity(s string) Entity {
	return Entity(s)
}

// Return Entity as string.
func (e *Entity) String() string {
	return string(*e)
}

// Returns true if Entity is an empty string.
func (e *Entity) IsEmpty() bool {
	return len(string(*e)) == 0
}

// Returns true if Entity is a channel name.
func (e *Entity) IsChan() bool {
	return strings.HasPrefix(string(*e), "#") && len(*e) > 1
}

// Returns true if Entity is a Hostmask.
func (e *Entity) IsHostmask() bool {
	a := strings.Index(e.String(), "!")
	b := strings.Index(e.String(), "@")
	return b > a && a > -1
}

// Returns true if Entity is a server name or a nickname.
func (e *Entity) IsName() bool {
	return !e.IsChan() && !e.IsHostmask()
}

// Returns the Nickname if a Hostmask is contained, empty string otherwise.
func (e *Entity) Nickname() string {
	if i := strings.Index(e.String(), "!"); i > 0 {
		return e.String()[:i]
	} else if len(e.String()) > 0 {
		return e.String()
	}
	return ""
}

// Returns the Username if a Hostmask is contained, empty string otherwise.
func (e *Entity) Username() string {
	a := strings.Index(e.String(), "!")
	b := strings.Index(e.String(), "@")
	if b > a && a > -1 {
		return e.String()[a+1 : b]
	}
	return ""
}

// Returns the Hostname if a Hostmask is contained, empty string otherwise.
func (e *Entity) Hostname() string {
	if i := strings.Index(e.String(), "@"); i > 0 && i+1 < len(e.String()) {
		return e.String()[i+1:]
	}
	return ""
}
