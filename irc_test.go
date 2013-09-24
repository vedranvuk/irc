package irc

import (
	"fmt"
	"testing"
)

func TestParseChannelModes(t *testing.T) {
	n, m := ParseChannelModes("&@SomeNick")
	if n != "SomeNick" || m != 12 {
		t.Error("ParseChannelModes() failed.")
	}
}

func TestIRCEntity(t *testing.T) {
	const (
		testNick     = "SomeNick"
		testUser     = "SomeUser"
		testHost     = "SomeHost"
		testChan     = "#SomeChan"
		testHostMask = testNick + "!" + testUser + "@" + testHost
	)

	h := Entity(testHostMask)
	v := h.Nickname()
	if v != testNick {
		t.Error("Entity.Nick() failed.")
	}
	v = h.Username()
	if h.Username() != "SomeUser" {
		t.Error("Entity.User() failed.")
	}
	v = h.Hostname()
	if h.Hostname() != "SomeHost" {
		t.Error("Entity.Host() failed.")
	}
	h = Entity(testChan)
	if h.IsChan() != true {
		t.Error("Entity.IsChan() failed.")
	}
	h = Entity(testNick)
	if h.IsName() != true {
		t.Error("Entity.IsName() failed.")
	}
	h = Entity(testHostMask)
	if h.IsName() != false {
		t.Error("Entity.IsName() failed.")
	}
}

func TestIRCMessage(t *testing.T) {
	const (
		testMsgPfx  = ":SomeNick!SomeUser@SomeHost"
		testMsgCmd  = "671"
		testMsgMid  = "yournick othernick"
		testMsgTra  = ":is using a secure connection"
		testMessage = testMsgPfx + " " + testMsgCmd + " " + testMsgMid + " " + testMsgTra
	)

	m := NewMessage(testMessage)
	if m.Prefix().String() != testMsgPfx[1:] {
		t.Error("IRCMessage.Prefix() failed.")
	}
	if m.Command() != testMsgCmd {
		t.Error("IRCMessage.Command() failed.")
	}
	if m.Command() != testMsgCmd {
		t.Error("IRCMessage.Command() failed.")
	}
	if m.IsNumeric() != true {
		t.Error("IRCMessage.IsNumeric() failed.")
	}
	if m.Numeric() != 671 {
		t.Error("IRCMessage.Numeric() failed.")
	}
	if m.Middle() != testMsgMid {
		t.Error("IRCMessage.Middle() failed.")
	}
	if m.Trailing() != testMsgTra[1:] {
		t.Error("IRCMessage.Trailing() failed.")
	}
}
