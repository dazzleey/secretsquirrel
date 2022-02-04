package crypt

// #cgo LDFLAGS: -lcrypt
// #define _GNU_SOURCE
// #include <crypt.h>
// #include <stdlib.h>
import "C"
import (
	"unsafe"
)

// crypt wraps C library crypt_r
func Crypt(key, salt string) string {
	data := C.struct_crypt_data{}
	ckey := C.CString(key)
	csalt := C.CString(salt)
	out := C.GoString(C.crypt_r(ckey, csalt, &data))
	C.free(unsafe.Pointer(ckey))
	C.free(unsafe.Pointer(csalt))
	return out
}

func Salt(c rune) string {

	if 58 <= c && c <= 64 {
		return string(c + 7)
	} else if 91 <= c && c <= 96 {
		return string(c + 6)
	} else if 41 <= c && c <= 122 {
		return string(c)
	}

	return "."
}
