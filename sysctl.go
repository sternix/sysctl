// +build freebsd

// Copyright 2009 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the
// https://github.com/golang/go/blob/master/LICENSE file.

package sysctl

import (
	"C"
	"bytes"
	"syscall"
	"unsafe"
)

// based on https://github.com/golang/sys/blob/master/unix/syscall_bsd.go

type _C_int C.int

var _zero uintptr

//sys	sysctl(mib []_C_int, old *byte, oldlen *uintptr, new *byte, newlen uintptr) (err error) = SYS___SYSCTL

func ByName(name string) (string, error) {
	return Args(name)
}

func Args(name string, args ...int) (string, error) {
	buf, err := Raw(name, args...)
	if err != nil {
		return "", err
	}
	n := len(buf)

	// Throw away terminating NUL.
	if n > 0 && buf[n-1] == '\x00' {
		n--
	}
	return string(buf[0:n]), nil
}

func Uint32(name string) (uint32, error) {
	return Uint32Args(name)
}

func Uint32Args(name string, args ...int) (uint32, error) {
	mib, err := sysctlmib(name, args...)
	if err != nil {
		return 0, err
	}

	n := uintptr(4)
	buf := make([]byte, 4)
	if err := sysctl(mib, &buf[0], &n, nil, 0); err != nil {
		return 0, err
	}
	if n != 4 {
		return 0, syscall.EIO
	}
	return *(*uint32)(unsafe.Pointer(&buf[0])), nil
}

func Uint64(name string, args ...int) (uint64, error) {
	mib, err := sysctlmib(name, args...)
	if err != nil {
		return 0, err
	}

	n := uintptr(8)
	buf := make([]byte, 8)
	if err := sysctl(mib, &buf[0], &n, nil, 0); err != nil {
		return 0, err
	}
	if n != 8 {
		return 0, syscall.EIO
	}
	return *(*uint64)(unsafe.Pointer(&buf[0])), nil
}

func Raw(name string, args ...int) ([]byte, error) {
	mib, err := sysctlmib(name, args...)
	if err != nil {
		return nil, err
	}

	// Find size.
	n := uintptr(0)
	if err := sysctl(mib, nil, &n, nil, 0); err != nil {
		return nil, err
	}
	if n == 0 {
		return nil, nil
	}

	// Read into buffer of that size.
	buf := make([]byte, n)
	if err := sysctl(mib, &buf[0], &n, nil, 0); err != nil {
		return nil, err
	}

	// The actual call may return less than the original reported required
	// size so ensure we deal with that.
	return buf[:n], nil
}

// https://github.com/golang/go/blob/master/src/syscall/syscall_freebsd.go
// Translate "kern.hostname" to []_C_int{0,1,2,3}.
func nametomib(name string) (mib []_C_int, err error) {
	const siz = unsafe.Sizeof(mib[0])

	// NOTE(rsc): It seems strange to set the buffer to have
	// size CTL_MAXNAME+2 but use only CTL_MAXNAME
	// as the size. I don't know why the +2 is here, but the
	// kernel uses +2 for its own implementation of this function.
	// I am scared that if we don't include the +2 here, the kernel
	// will silently write 2 words farther than we specify
	// and we'll get memory corruption.
	var buf [syscall.CTL_MAXNAME + 2]_C_int
	n := uintptr(syscall.CTL_MAXNAME) * siz

	p := (*byte)(unsafe.Pointer(&buf[0]))
	bytes, err := ByteSliceFromString(name)
	if err != nil {
		return nil, err
	}

	// Magic sysctl: "setting" 0.3 to a string name
	// lets you read back the array of integers form.
	if err = sysctl([]_C_int{0, 3}, p, &n, &bytes[0], uintptr(len(name))); err != nil {
		return nil, err
	}
	return buf[0 : n/siz], nil
}

// sysctlmib translates name to mib number and appends any additional args.
func sysctlmib(name string, args ...int) ([]_C_int, error) {
	// Translate name to mib number.
	mib, err := nametomib(name)
	if err != nil {
		return nil, err
	}

	for _, a := range args {
		mib = append(mib, _C_int(a))
	}

	return mib, nil
}

// https://github.com/golang/go/blob/master/src/syscall/zsyscall_freebsd_amd64.go
func sysctl(mib []_C_int, old *byte, oldlen *uintptr, newb *byte, newlen uintptr) (err error) {
	var _p0 unsafe.Pointer
	if len(mib) > 0 {
		_p0 = unsafe.Pointer(&mib[0])
	} else {
		_p0 = unsafe.Pointer(&_zero)
	}
	_, _, e1 := syscall.Syscall6(syscall.SYS___SYSCTL, uintptr(_p0), uintptr(len(mib)), uintptr(unsafe.Pointer(old)), uintptr(unsafe.Pointer(oldlen)), uintptr(unsafe.Pointer(newb)), uintptr(newlen))
	if e1 != 0 {
		err = syscall.Errno(e1)
	}
	return
}

func SetString(name string, value string) error {
	mib, err := nametomib(name)
	if err != nil {
		return err
	}

	valueslc, err := ByteSliceFromString(value)
	if err != nil {
		return err
	}

	if err = sysctl(mib, nil, nil, &valueslc[0], uintptr(len(valueslc))); err != nil {
		return err
	}

	return nil
}

func SetUint32(name string, value uint32) error {
	mib, err := nametomib(name)
	if err != nil {
		return err
	}

	if err = sysctl(mib, nil, nil, (*byte)(unsafe.Pointer(&value)), 4); err != nil {
		return err
	}

	return nil
}

func SetUint64(name string, value uint64) error {
	mib, err := nametomib(name)
	if err != nil {
		return err
	}

	if err = sysctl(mib, nil, nil, (*byte)(unsafe.Pointer(&value)), 8); err != nil {
		return err
	}

	return nil
}

func ByteSliceFromString(s string) ([]byte, error) {
	if i := bytes.IndexByte([]byte(s), 0); i != -1 {
		return nil, syscall.EINVAL
	}
	a := make([]byte, len(s)+1)
	copy(a, s)
	return a, nil
}
