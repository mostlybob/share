// Copyright 2019, Shulhan <ms@kilabit.info>. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package spf

//
// List of modifiers.
//
const (
	modifierRedirect int = iota
	modifierExp
	modifierUnknown
)

type modifier struct {
	kind  byte
	name  string
	value string
}
