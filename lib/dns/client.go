// Copyright 2018, Shulhan <ms@kilabit.info>. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package dns

import (
	"net"
)

//
// Client is interface that implement sending and receiving DNS message.
//
type Client interface {
	Close() error
	RemoteAddr() net.Addr
	Sender
	Receiver
}