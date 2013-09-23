// Copyright 2013 Vedran Vuk. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// Package irc implements an IRC client.
package irc

import (
	"bufio"
	"crypto/tls"
	"errors"
	"fmt"
	stringsex "github.com/vedranvuk/strings"
	"net"
)

var (
	// Thrown if Dial() is called on an already established connection.
	ErrAlreadyConnected = errors.New("already connected")
	// Thrown if Close() is called on an already disconnected client.
	ErrAlreadyDisconnected = errors.New("already disconnected")
	// If a read or write op is issued on a disconnected client.
	ErrNotConnected = errors.New("not connected")
)

type IRC struct {
	Nick string // Nickname.
	User string // Username/identd name.
	Geck string // Geckos/Real Name.
	Mode string // Mode.

	pass string // Server password.

	// Event handlers.
	OnRaw      func(m Message, in bool)
	OnPingPong func()
	OnJoin     func(channel string, user Entity)
	OnPart     func(channel, message string, user Entity)
	OnKick     func(channel, message string, user Entity)
	OnPrivMsg  func(message string, source, target Entity)
	OnNotice   func(message string, source, target Entity)
	OnNick     func(newnick string, user Entity)
	OnQuit     func(message string, user Entity)
	OnError    func(message string)

	OnServerWelcome   func(message string)
	OnMessageOfTheDay func(message []string)

	OnNickError func()

	MaxMsgLen int  // Maximum message length in bytes that will be sent.
	WriteRaw  bool // Write raw commands to stdout.

	conn net.Conn      // TCP connection.
	rbuf *bufio.Reader // Read buffer.

	// Temp variables.
	motd *[]string // MOTD combiner.
}

// Creates a new *IRC instance.
// "nick" is the nickname to be used.
// "user" is the username/ident to be used (optional, "nick" used if empty).
// "geck" is the real name/geckos to be used (optional, "nick" used if empty).
// "mode" is the user mode to be used. (optional, +i if empty).
// Returns an error if "nick" is empty.
func New(nick, user, geck, mode string) (*IRC, error) {
	if nick == "" {
		return nil, errors.New("nickname not specified")
	}
	if user == "" {
		user = nick
	}
	if geck == "" {
		geck = nick
	}
	if mode == "" {
		mode = "+i"
	}
	return &IRC{
		Nick:      nick,
		User:      user,
		Geck:      geck,
		Mode:      mode,
		MaxMsgLen: 400,
	}, nil
}

// Connect to "addr" using optional "password" for IRC server.
// "laddr" is an optional local address to bind to.
// If tlscfg is not nil, a TLS secured connection is estblished.
func (i *IRC) Dial(addr, password string, laddr *net.TCPAddr, tlscfg *tls.Config) error {
	if i.conn != nil {
		return ErrAlreadyConnected
	}

	a, err := net.ResolveTCPAddr("tcp", addr)
	if err != nil {
		fmt.Errorf("error resolving target addr: %s", err)
	}

	c, err := net.DialTCP("tcp", laddr, a)
	if err != nil {
		return err
	}

	if tlscfg != nil {
		i.conn = tls.Client(c, tlscfg)
	} else {
		i.conn = c
	}
	i.rbuf = bufio.NewReader(i.conn)
	i.pass = password
	return nil
}

// Check if client is connected.
func (i *IRC) Connected() bool {
	return i.conn != nil
}

// Return connection for lower level access. Can return nil if disconnected.
func (i *IRC) Conn() net.Conn {
	return i.conn
}

// Runs the I/O loop.
// This blocking function does the initial registration then runs the read loop.
// It should be run immediately if Dial() returns successfully to avoid
// registration timeout. It may return an error at some point, either after
// Close() is called from other goroutine or if a link error occurs.
// Expect a disconnect error, io.EOF.
func (i *IRC) Run() error {
	if i.pass != "" {
		if err := i.SendRaw(fmt.Sprintf("PASS %s", i.pass)); err != nil {
			return err
		}
		i.pass = ""
	}
	if err := i.SendRaw(fmt.Sprintf("NICK %s", i.Nick)); err != nil {
		return err
	}
	if err := i.SendRaw(fmt.Sprintf("USER %s %s * :%s", i.User, i.Mode, i.Geck)); err != nil {
		return err
	}
	for {
		s, err := i.rbuf.ReadString('\n')
		if err != nil {
			return err
		}
		if i.WriteRaw {
			fmt.Printf("-> %s", s)
		}
		m := NewMessage(s)
		if i.OnRaw != nil {
			i.OnRaw(m, true)
		}

		if !m.IsNumeric() {
			switch m.Command() {
			case "PING":
				if m.HasTrailing() {
					i.SendRaw(fmt.Sprintf("PONG :%s", m.Trailing()))
				} else {
					i.SendRaw("PONG")
				}
				if i.OnPingPong != nil {
					i.OnPingPong()
				}
			case "JOIN":
				if i.OnJoin != nil {
					i.OnJoin(m.Trailing(), *m.Prefix())
				}
			case "PART":
				if i.OnPart != nil {
					i.OnPart(m.Middle(), m.Trailing(), *m.Prefix())
				}
			case "KICK":
				if i.OnKick != nil {
					i.OnKick(m.Middle(), m.Trailing(), *m.Prefix())
				}
			case "PRIVMSG":
				if i.OnPrivMsg != nil {
					i.OnPrivMsg(m.Trailing(), *m.Prefix(), NewEntity(m.Middles()[0]))
				}
			case "NOTICE":
				if i.OnNotice != nil {
					i.OnNotice(m.Trailing(), *m.Prefix(), NewEntity(m.Middles()[0]))
				}
			case "NICK":
				if i.OnNick != nil {
					i.OnNick(m.Trailing(), *m.Prefix())
				}
			case "QUIT":
				if i.OnQuit != nil {
					i.OnQuit(m.Trailing(), *m.Prefix())
				}
			case "ERROR":
				if i.OnError != nil {
					i.OnError(m.Trailing())
				}
			}
		} else {
			switch m.Numeric() {
			case 1: // Server welcome.
				if i.OnServerWelcome != nil {
					i.OnServerWelcome(m.Trailing())
				}
			case 372: // MOTD line.
				if i.OnMessageOfTheDay != nil && i.motd != nil {
					*i.motd = append(*i.motd, m.Trailing())
				}
			case 375: // MOTD start.
				if i.OnMessageOfTheDay != nil {
					a := make([]string, 1)
					i.motd = &a
				}
			case 376: // MOTD end.
				if i.OnMessageOfTheDay != nil && i.motd != nil {
					i.OnMessageOfTheDay(*i.motd)
				}
				i.motd = nil
			case 433: // Nick error.
				if i.OnNickError != nil {
					i.OnNickError()
				}
			}
		}
	}
}

// Closes the IRC connection.
func (i *IRC) Close() error {
	if i.conn == nil {
		return ErrAlreadyDisconnected
	}
	err := i.conn.Close()
	i.conn = nil
	i.rbuf = nil
	return err
}

// Send a "raw" command to server.
func (i *IRC) SendRaw(raw string) error {
	if i.conn == nil {
		return ErrNotConnected
	}
	if raw == "" {
		return nil
	}
	s := fmt.Sprintf("%s\r\n", stringsex.LenLimitByRune(raw, i.MaxMsgLen))
	if i.WriteRaw {
		fmt.Printf(" <- %s", s)
	}
	if i.OnRaw != nil {
		i.OnRaw(NewMessage(s), false)
	}
	b := []byte(s)
	n, err := i.conn.Write(b)
	if err != nil {
		return err
	}
	// TODO: make this better, considering Run() is probably a goroutine.
	if n != len(b) {
		i.Close()
		return errors.New("short write")
	}
	return nil
}

// Join a "channel" with an optional channel "key".
func (i *IRC) CmdJoin(channel, key string) error {
	if key != "" {
		return i.SendRaw(fmt.Sprintf("JOIN %s %s", channel, key))
	} else {
		return i.SendRaw(fmt.Sprintf("JOIN %s", channel))
	}
}

// Part the "channel" with an optional part "message".
func (i *IRC) CmdPart(channel, message string) error {
	if message != "" {
		return i.SendRaw(fmt.Sprintf("PART %s :%s", channel, message))
	} else {
		return i.SendRaw(fmt.Sprintf("PART %s", channel))
	}
}

// Send "message" to "target" where target can be channel name or nickname.
func (i *IRC) CmdPrivMsg(message, target string) error {
	max := i.MaxMsgLen - len(fmt.Sprintf("PRIVMSG %s :", target))
	out := stringsex.LenSplitByRune(message, max)
	for _, v := range out {
		return i.SendRaw(fmt.Sprintf("PRIVMSG %s :%s", target, v))
	}
	return nil
}

// Send "notice" to "target" where target can be channel name or nickname.
func (i *IRC) CmdNotice(message, target string) error {
	max := i.MaxMsgLen - len(fmt.Sprintf("NOTICE %s :", target))
	out := stringsex.LenSplitByRune(message, max)
	for _, v := range out {
		return i.SendRaw(fmt.Sprintf("NOTICE %s :%s", target, v))
	}
	return nil
}

// Change your nickname to "newnick".
func (i *IRC) CmdNickname(newnick string) error {
	return i.SendRaw(fmt.Sprintf("NICK :%s", newnick))
}

// Quit IRC with an optional quit "message".
func (i *IRC) CmdQuit(message string) error {
	if message != "" {
		return i.SendRaw(fmt.Sprintf("QUIT :%s", message))
	} else {
		return i.SendRaw("QUIT")
	}
}
