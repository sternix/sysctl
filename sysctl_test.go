package sysctl

import (
	"testing"
	//	"syscall"
)

func TestByName(t *testing.T) {
	ot, err := ByName("kern.ostype")
	if err != nil {
		t.Error(err)
	}

	if ot != "FreeBSD" {
		t.Errorf("get %s expected FreeBSD", ot)
	}
}

func TestUint32(t *testing.T) {
	somax, err := Uint32("kern.ipc.soacceptqueue")
	if err != nil {
		t.Error(err)
	}

	// defaults to 128
	if somax < 128 {
		t.Errorf("kern.ipc.soacceptqueue value expected bigger than %d", somax)
	}

	err = SetUint32("kern.ipc.soacceptqueue", 65535)
	if err != nil {
		t.Error(err)
	}

	somax, err = Uint32("kern.ipc.soacceptqueue")
	if err != nil {
		t.Error(err)
	}

	if somax != 65535 {
		t.Errorf("get %d expected 65535", somax)
	}
}

func TestString(t *testing.T) {
	hostname, err := ByName("kern.hostname")
	if err != nil {
		t.Error(err)
	}

	err = SetString("kern.hostname", "sysctl.golang.org")
	if err != nil {
		t.Error(err)
	}

	hn, err := ByName("kern.hostname")
	if err != nil {
		t.Error(err)
	}

	if hn != "sysctl.golang.org" {
		t.Errorf("get %s expected sysctl.golang.org", hn)
	}

	// restore orj hostname
	err = SetString("kern.hostname", hostname)
	if err != nil {
		t.Error(err)
	}
}

func TestBigString(t *testing.T) {
	_, err := ByName("kern.conftxt")
	if err != nil {
		t.Error(err)
	}

	_, err = ByName("kern.geom.conftxt")
	if err != nil {
		t.Error(err)
	}
}

func asciiToGoString(arr []byte) string {
	n := len(arr)

	// Throw away terminating NUL.
	if n > 0 && arr[n-1] == '\x00' {
		n--
	}

	return string(arr[:n])
}

func TestRaw(t *testing.T) {
	hostname, err := Raw("kern.hostname")
	if err != nil {
		t.Error(err)
	}

	err = SetRaw("kern.hostname", []byte("sysctl.golang.org"))
	if err != nil {
		t.Error(err)
	}

	hn, err := Raw("kern.hostname")
	if err != nil {
		t.Error(err)
	}

	strhn := asciiToGoString(hn)

	if strhn != "sysctl.golang.org" {
		t.Errorf("get '%s' expected sysctl.golang.org", string(strhn))
	}

	// restore orj hostname
	err = SetRaw("kern.hostname", hostname)
	if err != nil {
		t.Error(err)
	}
}
