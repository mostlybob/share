// Copyright 2019, Shulhan <ms@kilabit.info>. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package io

import (
	"testing"

	"github.com/shuLhan/share/lib/test"
)

func TestReaderScanInt64(t *testing.T) {
	cases := []struct {
		desc string
		src  []byte
		exp  int64
		expc byte
	}{{
		desc: "With empty input",
	}, {
		desc: "With non digit",
		src:  []byte("a -1"),
		expc: 'a',
	}, {
		desc: "With leading spaces",
		src:  []byte("  +1"),
		exp:  1,
	}, {
		desc: "With -1",
		src:  []byte("-1"),
		exp:  -1,
	}, {
		desc: "With -1",
		src:  []byte("-1x"),
		exp:  -1,
		expc: 'x',
	}, {
		desc: "With +1",
		src:  []byte("+1"),
		exp:  1,
	}, {
		desc: "With 1000",
		src:  []byte("1000"),
		exp:  1000,
	}, {
		desc: "With 9876543210 1",
		src:  []byte("9876543210 1"),
		exp:  9876543210,
		expc: ' ',
	}, {
		desc: "With leading zero 001",
		src:  []byte("-001"),
		exp:  -1,
	}}

	r := &Reader{}

	for _, c := range cases {
		t.Log(c.desc)

		r.Init(c.src)

		got, gotc := r.ScanInt64()

		test.Assert(t, "n", c.exp, got, true)
		test.Assert(t, "c", c.expc, gotc, true)
	}
}
