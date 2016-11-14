// Tideland Go Library - Version
//
// Copyright (C) 2014-2015 Frank Mueller / Tideland / Oldenburg / Germany
//
// All rights reserved. Use of this source code is governed
// by the new BSD license.

package version

//--------------------
// IMPORTS
//--------------------

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/tideland/golib/errors"
)

//--------------------
// CONST
//--------------------

const (
	Metadata = "+"
)

//--------------------
// VERSION
//--------------------

// Version defines the interface of a version.
type Version interface {
	fmt.Stringer

	// Major returns the major version.
	Major() int

	// Minor returns the minor version.
	Minor() int

	// Patch return the path version.
	Patch() int

	// PreRelease returns a possible pre-release of the version.
	PreRelease() string

	// Metadata returns a possible build metadata of the version.
	Metadata() string

	// Less returns true if this version is less than the passed one.
	Less(cv Version) bool
}

// vsn implements the version interface.
type vsn struct {
	major      int
	minor      int
	patch      int
	preRelease []string
	metadata   []string
}

// New returns a simple version instance. Parts of pre-release
// and metadata are passed as optional strings separated by
// version.Metadata ("+").
func New(major, minor, patch int, prmds ...string) Version {
	if major < 0 {
		major = 0
	}
	if minor < 0 {
		minor = 0
	}
	if patch < 0 {
		patch = 0
	}
	v := &vsn{
		major: major,
		minor: minor,
		patch: patch,
	}
	isPR := true
	for _, prmd := range prmds {
		if isPR {
			if prmd == Metadata {
				isPR = false
				continue
			}
			v.preRelease = append(v.preRelease, validID(prmd, true))
		} else {
			v.metadata = append(v.metadata, validID(prmd, false))
		}
	}
	return v
}

// Parse retrieves a version out of a string.
func Parse(vsnstr string) (Version, error) {
	// Split version, pre-release, and metadata.
	npmstrs, err := splitVersionString(vsnstr)
	if err != nil {
		return nil, err
	}
	// Parse these parts.
	nums, err := parseNumberString(npmstrs[0])
	if err != nil {
		return nil, err
	}
	prmds := []string{}
	if npmstrs[1] != "" {
		prmds = strings.Split(npmstrs[1], ".")
	}
	if npmstrs[2] != "" {
		prmds = append(prmds, Metadata)
		prmds = append(prmds, strings.Split(npmstrs[2], ".")...)
	}
	// Done.
	return New(nums[0], nums[1], nums[2], prmds...), nil
}

// Major returns the major version.
func (v *vsn) Major() int {
	return v.major
}

// Minor returns the minor version.
func (v *vsn) Minor() int {
	return v.minor
}

// Patch returns the patch version.
func (v *vsn) Patch() int {
	return v.patch
}

// PreRelease returns a possible pre-release of the version.
func (v *vsn) PreRelease() string {
	return strings.Join(v.preRelease, ".")
}

// Metadata returns a possible build metadata of the version.
func (v *vsn) Metadata() string {
	return strings.Join(v.metadata, ".")
}

// Less returns true if this version is less than the passed one.
func (v *vsn) Less(cv Version) bool {
	// Major version.
	if v.major < cv.Major() {
		return true
	}
	if v.major > cv.Major() {
		return false
	}
	// Minor version.
	if v.minor < cv.Minor() {
		return true
	}
	if v.minor > cv.Minor() {
		return false
	}
	// Patch version.
	if v.patch < cv.Patch() {
		return true
	}
	if v.patch > cv.Patch() {
		return false
	}
	// Simple comparing done, now the pre-release is interesting.
	cvpr := []string{}
	if cvprs := cv.PreRelease(); len(cvprs) > 0 {
		cvpr = strings.Split(cvprs, ".")
	}
	return less(v.preRelease, cvpr)
}

// String returns the version as string.
func (v *vsn) String() string {
	vs := fmt.Sprintf("%d.%d.%d", v.major, v.minor, v.patch)
	if len(v.preRelease) > 0 {
		vs += "-" + v.PreRelease()
	}
	if len(v.metadata) > 0 {
		vs += Metadata + v.Metadata()
	}
	return vs
}

//--------------------
// TOOLS
//--------------------

// validID reduces the passed identifier to a valid one. If we care
// for numeric identifiers leading zeros will be removed.
func validID(id string, numeric bool) string {
	out := []rune{}
	letter := false
	digit := false
	hyphen := false
	for _, r := range id {
		switch {
		case r >= 'a' && r <= 'z':
			letter = true
			out = append(out, r)
		case r >= 'A' && r <= 'Z':
			letter = true
			out = append(out, r)
		case r >= '0' && r <= '9':
			digit = true
			out = append(out, r)
		case r == '-':
			hyphen = true
			out = append(out, r)
		}
	}
	if numeric && digit && !letter && !hyphen {
		// Digits only, and we care for it.
		// Remove leading zeros.
		for len(out) > 0 && out[0] == '0' {
			out = out[1:]
		}
		if len(out) == 0 {
			out = []rune{'0'}
		}
	}
	return string(out)
}

// less compares two string slices and returns true
// if a  is less than b.
func less(a, b []string) bool {
	for i := 0; i < len(a) && i < len(b); i++ {
		nl, ok := numericLess(a[i], b[i])
		switch {
		case ok:
			return nl
		case a[i] > b[i]:
			return false
		case a[i] < b[i]:
			return true
		}
	}
	if len(a) > len(b) {
		return true
	}
	return false
}

// numericLess tries to convert a and b into ints and
// compares them then if possible.
func numericLess(a, b string) (bool, bool) {
	an, err := strconv.Atoi(a)
	if err != nil {
		return false, false
	}
	bn, err := strconv.Atoi(b)
	if err != nil {
		return false, false
	}
	return an < bn, true
}

// splitVersionString seperates the version string into numbers,
// pre-release, and metadata strings.
func splitVersionString(vsnstr string) ([]string, error) {
	npXm := strings.SplitN(vsnstr, Metadata, 2)
	switch len(npXm) {
	case 1:
		nXp := strings.SplitN(npXm[0], "-", 2)
		switch len(nXp) {
		case 1:
			return []string{nXp[0], "", ""}, nil
		case 2:
			return []string{nXp[0], nXp[1], ""}, nil
		}
	case 2:
		nXp := strings.SplitN(npXm[0], "-", 2)
		switch len(nXp) {
		case 1:
			return []string{nXp[0], "", npXm[1]}, nil
		case 2:
			return []string{nXp[0], nXp[1], npXm[1]}, nil
		}
	}
	return nil, errors.New(ErrIllegalVersionFormat, errorMessages)
}

// parseNumberString retrieves major, minor, and patch number
// of the passed string.
func parseNumberString(nstr string) ([]int, error) {
	nstrs := strings.Split(nstr, ".")
	if len(nstrs) < 1 || len(nstrs) > 3 {
		return nil, errors.New(ErrIllegalVersionFormat, errorMessages)
	}
	vsn := []int{1, 0, 0}
	for i, nstr := range nstrs {
		num, err := strconv.Atoi(nstr)
		if err != nil {
			return nil, errors.Annotate(err, ErrIllegalVersionFormat, errorMessages)
		}
		if num < 0 {
			return nil, errors.New(ErrIllegalVersionFormat, errorMessages)
		}
		vsn[i] = num
	}
	return vsn, nil
}

// EOF
