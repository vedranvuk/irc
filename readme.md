# Description

Package irc implements an IRC client. Another one, yeah.

Still tweaking and adding, first commit basically. Will add extra event handlers and maybe change some 
internals but no breaking changes planned.


# Example

	// Most basic client that connects to a server and joins a channel on 
	// server welcome. Writes everything to stdout. Run() blocks, you can stuff 
	// Run() into a goroutine if you need async.

	package main
	
	import (
		"github.com/vedranvuk/irc"
		"log"
	)
	
	func main() {
		cli, err := irc.New("MyNick", "MyUsername", "MyRealName", "+ix")
		if err != nil {
			log.Fatalf("error creating new irc client: %s", err)
		}
		cli.WriteRaw = true
		cli.OnServerWelcome = func(message string) {
			cli.CmdJoin("#somechannel", "")
		}
		
		err = cli.Dial("irc.myircnetwork.com:6667", "", nil, nil)
		if err != nil {
			log.Fatalf("connect error: %s", err)
		}
	
		err = cli.Run()
		if err != nil {
			log.Fatalf("Run() error: %s", err)
		}
	}
	
	
# Installation

Add to import clause for auto install on compile or

`go get -u github.com/vedranvuk/irc` into a terminal.

# Licence

Copyright (c) 2013 Vedran Vuk. All rights reserved. Use of this source code is 
governed by a BSD-style license that can be found in the LICENSE file.




