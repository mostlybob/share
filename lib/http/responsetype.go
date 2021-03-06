// Copyright 2018, Shulhan <ms@kilabit.info>. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package http

// ResponseType define type for HTTP response
type ResponseType int

// List of valid response type.
const (
	ResponseTypeNone   ResponseType = 0
	ResponseTypeBinary              = 1 << iota
	ResponseTypePlain
	ResponseTypeJSON
)
