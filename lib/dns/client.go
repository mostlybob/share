// Copyright 2018, Shulhan <ms@kilabit.info>. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package dns

import (
	"time"
)

//
// Client is interface that implement sending and receiving DNS message.
//
type Client interface {
	Close() error
	Lookup(allowRecursion bool, qtype, qclass uint16, qname []byte) (*Message, error)
	RemoteAddr() string
	Query(req *Message) (*Message, error)
	SetTimeout(t time.Duration)
	SetRemoteAddr(addr string) error
}
