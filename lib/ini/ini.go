// Copyright 2018, Shulhan <ms@kilabit.info>. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package ini

import (
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/shuLhan/share/lib/debug"
	libstrings "github.com/shuLhan/share/lib/strings"
)

//
// Ini contains the parsed file.
//
type Ini struct {
	secs []*Section
}

//
// Open and parse INI formatted file.
//
// On fail it will return incomplete instance of Ini with error.
//
func Open(filename string) (in *Ini, err error) {
	reader := newReader()

	in, err = reader.parseFile(filename)

	if debug.Value >= 1 && err == nil {
		err = in.Write(os.Stdout)
	}

	return
}

//
// Parse INI format from text.
//
func Parse(text []byte) (in *Ini, err error) {
	reader := newReader()

	return reader.Parse(text)
}

//
// Add the new key and value to the last item in section and/or subsection.
//
// If section or subsection is not exist it will create a new one.
// If section or key is empty, or value already exist it will not modify the
// INI object.
//
// It will return true if new variable is added, otherwise it will return
// false.
//
func (in *Ini) Add(secName, subName, key, value string) bool {
	if len(secName) == 0 || len(key) == 0 {
		return false
	}

	secName = strings.ToLower(secName)

	sec := in.getSection(secName, subName)
	if sec != nil {
		return sec.add(key, value)
	}

	if len(value) == 0 {
		value = varValueTrue
	}

	sec = newSection(secName, subName)
	v := &variable{
		mode:     lineModeValue,
		key:      key,
		keyLower: strings.ToLower(key),
		value:    value,
	}
	sec.vars = append(sec.vars, v)
	in.secs = append(in.secs, sec)

	return true
}

//
// Section given section and/or subsection name, return the Section object
// that match with it.
// If section name is empty, it will return nil.
// If ini contains duplicate section (or subsection) it will merge all
// of its variables into one section.
//
func (in *Ini) Section(secName, subName string) (sec *Section) {
	if len(secName) == 0 {
		return nil
	}

	sec = &Section{
		name: secName,
		sub:  subName,
	}

	secName = strings.ToLower(secName)
	for x := 0; x < len(in.secs); x++ {
		if secName != in.secs[x].nameLower {
			continue
		}
		if subName != in.secs[x].sub {
			continue
		}

		sec.merge(in.secs[x])
	}
	return
}

//
// Set the last variable's value in section-subsection that match with the
// key.
// If key found it will return true; otherwise it will return false.
//
func (in *Ini) Set(secName, subName, key, value string) bool {
	if len(secName) == 0 || len(key) == 0 {
		return false
	}

	secName = strings.ToLower(secName)

	sec := in.getSection(secName, subName)
	if sec == nil {
		return false
	}

	key = strings.ToLower(key)

	return sec.set(key, value)
}

//
// Unset remove the last variable's in section and/or subsection that match
// with the key.
// If key found it will return true, otherwise it will return false.
//
func (in *Ini) Unset(secName, subName, key string) bool {
	if len(secName) == 0 || len(key) == 0 {
		return false
	}

	secName = strings.ToLower(secName)

	sec := in.getSection(secName, subName)
	if sec == nil {
		return false
	}

	sec.unset(key)

	return true
}

//
// addSection append the new section to the list.
//
func (in *Ini) addSection(sec *Section) {
	if sec == nil {
		return
	}
	if sec.mode != lineModeEmpty && len(sec.name) == 0 {
		return
	}

	sec.nameLower = strings.ToLower(sec.name)

	in.secs = append(in.secs, sec)
}

//
// AsMap return the INI contents as mapping of
// (section-name ":" subsection-name ":" variable-name) as key
// and the variable's values as slice of string.
//
// If section name is not empty, only the keys will be listed in the map.
//
func (in *Ini) AsMap(sectionName, subName string) (out map[string][]string) {
	sep := ":"
	out = make(map[string][]string)

	for x := 0; x < len(in.secs); x++ {
		sec := in.secs[x]

		if len(sectionName) > 0 && sectionName != sec.nameLower {
			continue
		}
		if len(sectionName) > 0 && subName != sec.sub {
			continue
		}

		for y := 0; y < len(sec.vars); y++ {
			v := sec.vars[y]

			if v.mode == lineModeEmpty {
				continue
			}
			if v.mode&lineModeSection > 0 || v.mode&lineModeSubsection > 0 {
				continue
			}

			var key string

			if len(sectionName) > 0 && len(subName) > 0 {
				key += v.key
			} else {
				key += sec.nameLower + sep
				key += sec.sub + sep
				key += v.key
			}

			vals, ok := out[key]
			if !ok {
				out[key] = []string{v.value}
			} else {
				out[key] = libstrings.AppendUniq(vals, v.value)
			}
		}
	}

	return out
}

//
// Get the last key on section and/or subsection.
//
// If key found it will return its value and true; otherwise it will return
// default value in def and false.
//
func (in *Ini) Get(secName, subName, key, def string) (val string, ok bool) {
	if len(in.secs) == 0 || len(secName) == 0 || len(key) == 0 {
		return def, false
	}

	x := len(in.secs) - 1
	sec := strings.ToLower(secName)

	for ; x >= 0; x-- {
		if in.secs[x].nameLower != sec {
			continue
		}

		if in.secs[x].sub != subName {
			continue
		}

		val, ok = in.secs[x].get(key, def)
		if ok {
			return
		}
	}

	return def, false
}

//
// GetBool return key's value as boolean.  If no key found it will return
// default value.
//
func (in *Ini) GetBool(secName, subName, key string, def bool) bool {
	out, ok := in.Get(secName, subName, key, "false")
	if !ok {
		return def
	}

	return IsValueBoolTrue(out)
}

//
// Gets key's values as slice of string in the same section and subsection.
//
func (in *Ini) Gets(secName, subName, key string) (out []string) {
	secName = strings.ToLower(secName)

	for _, sec := range in.secs {
		if sec.mode&lineModeSection == 0 {
			continue
		}
		if sec.nameLower != secName {
			continue
		}
		if sec.sub != subName {
			continue
		}
		vals, ok := sec.gets(key, nil)
		if !ok {
			continue
		}

		out = libstrings.AppendUniq(out, vals...)
	}
	return
}

//
// Prune remove all empty lines, comments, and merge all section and
// subsection with the same name into one group.
//
func (in *Ini) Prune() {
	newSecs := make([]*Section, 0, len(in.secs))

	for _, sec := range in.secs {
		if sec.mode == lineModeEmpty {
			continue
		}
		newSec := &Section{
			mode:      lineModeSection,
			name:      sec.name,
			nameLower: sec.nameLower,
		}
		if len(sec.sub) > 0 {
			newSec.mode |= lineModeSubsection
			newSec.sub = sec.sub
		}
		for _, v := range sec.vars {
			if v.mode == lineModeEmpty || v.mode == lineModeComment {
				continue
			}

			newValue := v.value
			if len(v.value) == 0 {
				newValue = "true"
			}

			newSec.addUniqValue(v.key, newValue)
		}
		newSecs = mergeSection(newSecs, newSec)
	}

	in.secs = newSecs
}

//
// mergeSection merge a section (and subsection) into slice.
//
func mergeSection(secs []*Section, newSec *Section) []*Section {
	for x := 0; x < len(secs); x++ {
		if secs[x].nameLower != newSec.nameLower {
			continue
		}
		if secs[x].sub != newSec.sub {
			continue
		}
		for _, v := range newSec.vars {
			if v.mode == lineModeEmpty || v.mode == lineModeComment {
				continue
			}
			secs[x].addUniqValue(v.keyLower, v.value)
		}
		return secs
	}
	secs = append(secs, newSec)
	return secs
}

//
// Rebase merge the other INI sections into this INI sections.
//
func (in *Ini) Rebase(other *Ini) {
	for _, otherSec := range other.secs {
		in.secs = mergeSection(in.secs, otherSec)
	}
}

//
// Save the current parsed Ini into file `filename`. It will overwrite the
// destination file if it's exist.
//
func (in *Ini) Save(filename string) (err error) {
	f, err := os.Create(filename)
	if err != nil {
		return
	}

	err = in.Write(f)
	if err != nil {
		return
	}

	return f.Close()
}

//
// Subs return all non empty subsections (and its variable) that have the same
// section name.
//
// This function is shortcut to be used in templating.
//
func (in *Ini) Subs(secName string) (subs []*Section) {
	if len(secName) == 0 {
		return
	}

	secName = strings.ToLower(secName)

	for x := 0; x < len(in.secs); x++ {
		if in.secs[x].mode == lineModeEmpty || in.secs[x].mode == lineModeComment {
			continue
		}
		if len(in.secs[x].sub) == 0 {
			continue
		}
		if in.secs[x].nameLower != secName {
			continue
		}

		subs = mergeSection(subs, in.secs[x])
	}

	return subs
}

//
// Val return the last variable value using a string as combination of
// section, subsection, and key with ":" as separator.  If key not found, it
// will return empty string.
//
// For example, to get the value of key "k" in section "s" and subsection
// "sub", call
//
//	V("s:sub:k")
//
// This function is shortcut to be used in templating.
//
func (in *Ini) Val(keyPath string) (val string) {
	keys := strings.Split(keyPath, ":")
	if len(keys) != 3 {
		return
	}

	val, _ = in.Get(keys[0], keys[1], keys[2], "")

	return
}

//
// Vals return all values as slice of string.
// The keyPath is combination of section, subsection, and key using colon ":"
// as separator.
// If key not found, it will return an empty slice.
//
// For example, to get all values of key "k" in section "s" and subsection
// "sub", call
//
//	Vals("s:sub:k")
//
// This function is shortcut to be used in templating.
//
func (in *Ini) Vals(keyPath string) (vals []string) {
	keys := strings.Split(keyPath, ":")
	if len(keys) != 3 {
		return
	}

	vals = in.Gets(keys[0], keys[1], keys[2])

	return
}

//
// Vars return all variables in section and/or subsection as map of string.
// If there is a duplicate in key's name, only the last key value that will be
// store on map value.
//
// This method is a shortcut that can be used in templating.
//
func (in *Ini) Vars(sectionPath string) (vars map[string]string) {
	names := strings.Split(sectionPath, ":")
	switch len(names) {
	case 0:
		return
	case 1:
		names = append(names, "")
	}

	asmap := in.AsMap(names[0], names[1])
	if len(asmap) > 0 {
		vars = make(map[string]string, len(asmap))
	}
	for k, v := range asmap {
		vars[k] = v[len(v)-1]
	}

	return
}

//
// Write the current parsed Ini into writer `w`.
//
func (in *Ini) Write(w io.Writer) (err error) {
	for x := 0; x < len(in.secs); x++ {
		fmt.Fprint(w, in.secs[x])

		for _, v := range in.secs[x].vars {
			fmt.Fprint(w, v)
		}
	}

	return
}

//
// getSection return the last section that have the same name and/or with
// subsection's name.
// Section's name MUST have in lowercase.
//
func (in *Ini) getSection(secName, subName string) *Section {
	x := len(in.secs) - 1
	for ; x >= 0; x-- {
		if in.secs[x].mode == lineModeEmpty || in.secs[x].mode == lineModeComment {
			continue
		}
		if in.secs[x].nameLower != secName {
			continue
		}
		if in.secs[x].sub != subName {
			continue
		}
		return in.secs[x]
	}
	return nil
}
