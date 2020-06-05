package main

import "fmt"

var ErrBadHex128 error = fmt.Errorf("Not uppercase 128byte hex identifier")
var ErrBadHex704 error = fmt.Errorf("Not uppercase 704byte hex identifier")
var ErrBadHex72 error = fmt.Errorf("Not uppercase 72byte hex identifier")
var ErrBadHex96 error = fmt.Errorf("Not uppercase 96byte hex identifier")
var ErrBadHex736 error = fmt.Errorf("Not uppercase 736byte hex identifier")
var ErrBadHex672 error = fmt.Errorf("Not uppercase 672byte hex identifier")
var ErrBadHex32 error = fmt.Errorf("Not uppercase 32byte hex identifier")
var ErrBadHex8 error = fmt.Errorf("Not uppercase 8byte hex identifier")
var ErrBadHex2 error = fmt.Errorf("Not uppercase 2byte hex identifier")
var ErrBadDec8 error = fmt.Errorf("Not 8byte decimal hex identifier")
var ErrBadDec4 error = fmt.Errorf("Not 4byte decimal hex identifier")
var ErrBadHexpand32 error = fmt.Errorf("Not uppercase 32byte hexpand identifier")
var ErrBadHexpand66 error = fmt.Errorf("Not uppercase 66byte hexpand identifier")

func Hex(x byte) byte {
	return 7*(x/10) + x + '0'
}

func checkDEC8(b string) error {
	if len(b) != 16 {
		return ErrBadDec8
	}
	for i := 0; i < 16; i++ {
		if (b[i] >= '0') && (b[i] <= '9') {
		} else {
			return ErrBadDec8
		}
	}
	return nil
}
func checkDEC4(b string) error {
	if len(b) != 8 {
		return ErrBadDec4
	}
	for i := 0; i < 8; i++ {
		if (b[i] >= '0') && (b[i] <= '9') {
		} else {
			return ErrBadDec4
		}
	}
	return nil
}

func checkHEX(b string, length int) bool {
	if len(b) != 2*length {
		return false
	}
	for i := 0; i < 2*length; i++ {
		if ((b[i] >= '0') && (b[i] <= '9')) || ((b[i] >= 'A') && (b[i] <= 'F')) {
		} else {
			return false
		}
	}
	return true
}
func checkHEX128(b string) error {
	if checkHEX(b, 128) {
		return nil
	}
	return ErrBadHex128
}
func checkHEX96(b string) error {
	if checkHEX(b, 96) {
		return nil
	}
	return ErrBadHex96
}
func checkHEX72(b string) error {
	if checkHEX(b, 72) {
		return nil
	}
	return ErrBadHex72
}
func checkHEX704(b string) error {
	if checkHEX(b, 704) {
		return nil
	}
	return ErrBadHex704
}
func checkHEX736(b string) error {
	if checkHEX(b, 736) {
		return nil
	}
	return ErrBadHex736
}
func checkHEX672(b string) error {
	if checkHEX(b, 672) {
		return nil
	}
	return ErrBadHex672
}
func checkHEX32(b string) error {
	if checkHEX(b, 32) {
		return nil
	}
	return ErrBadHex32
}
func checkHEX2(b string) error {
	if checkHEX(b, 2) {
		return nil
	}
	return ErrBadHex2
}
func checkHEX8(b string) error {
	if checkHEX(b, 8) {
		return nil
	}
	return ErrBadHex8
}
func hex2byte8(hex []byte) (out [8]byte) {
	for i := range out {
		out[i] = (x2b(hex[i<<1]) << 4) | x2b(hex[i<<1|1])
	}
	return out
}
func hex2byte2(hex []byte) (out [2]byte) {
	for i := range out {
		out[i] = (x2b(hex[i<<1]) << 4) | x2b(hex[i<<1|1])
	}
	return out
}
func hex2byte32(hex []byte) (out [32]byte) {
	for i := range out {
		out[i] = (x2b(hex[i<<1]) << 4) | x2b(hex[i<<1|1])
	}
	return out
}
func hex2uint16(hex []byte) (out uint16) {
	out = uint16(x2b(hex[0]))<<12 | uint16(x2b(hex[1]))<<8 | uint16(x2b(hex[2]))<<4 | uint16(x2b(hex[3]))
	return out
}
func hex2byte128(hex []byte) (out [128]byte) {
	for i := range out {
		out[i] = (x2b(hex[i<<1]) << 4) | x2b(hex[i<<1|1])
	}
	return out
}
func hex2byte96(hex []byte) (out [96]byte) {
	for i := range out {
		out[i] = (x2b(hex[i<<1]) << 4) | x2b(hex[i<<1|1])
	}
	return out
}
func hex2byte72(hex []byte) (out [72]byte) {
	for i := range out {
		out[i] = (x2b(hex[i<<1]) << 4) | x2b(hex[i<<1|1])
	}
	return out
}
func hex2byte704(hex []byte) (out [704]byte) {
	for i := range out {
		out[i] = (x2b(hex[i<<1]) << 4) | x2b(hex[i<<1|1])
	}
	return out
}
func hex2byte736(hex []byte) (out [736]byte) {
	for i := range out {
		out[i] = (x2b(hex[i<<1]) << 4) | x2b(hex[i<<1|1])
	}
	return out
}
func hex2byte672(hex []byte) (out [672]byte) {
	for i := range out {
		out[i] = (x2b(hex[i<<1]) << 4) | x2b(hex[i<<1|1])
	}
	return out
}
func x2b(hex byte) (lo byte) {
	return [32]byte{13, 14, 15, 0, 0, 10, 11, 12, 0, 0, 0, 0, 0, 0, 0, 0, 3, 2, 1, 0, 7, 6, 5, 4, 0, 0, 9, 8, 0, 0, 0, 0}[(hex^(hex>>4))&31]
}
func checkHEXPAND(b string, length int) bool {
	if len(b) != 2*length {
		return false
	}
	for i := 0; i < 2*length; i++ {
		if (b[i] >= 'G') && (b[i] <= 'V') {
		} else {
			return false
		}
	}
	return true
}
func checkHEXPAND32(b string) error {
	if checkHEXPAND(b, 32) {
		return nil
	}
	return ErrBadHexpand32
}
func checkHEXPAND66(b string) error {
	if checkHEXPAND(b, 66) {
		return nil
	}
	return ErrBadHexpand66
}
func Hexpand2(v []byte) (out [4]byte) {
	if len(v)*2 != len(out) {
		panic("Hexpand2 not 16bit")
	}
	for i := range out {
		if i&1 == 1 {
			out[i] = (v[i>>1] & 0xF) + 'G'
		} else {
			out[i] = (v[i>>1] >> 4) + 'G'
		}
	}
	return out
}

func Hexpand32(v []byte) (out [64]byte) {
	if len(v)*2 != len(out) {
		panic("Hexpand32 not 256bit")
	}
	for i := range out {
		if i&1 == 1 {
			out[i] = (v[i>>1] & 0xF) + 'G'
		} else {
			out[i] = (v[i>>1] >> 4) + 'G'
		}
	}
	return out
}

func hexpand2byte32(hexpand []byte) (out [32]byte) {
	for i := range out {
		out[i] = 16*(hexpand[2*i]-'G') + (hexpand[2*i+1] - 'G')
	}
	return out
}

func hexpand2byte2(hexpand []byte) (out [2]byte) {
	for i := range out {
		out[i] = 16*(hexpand[2*i]-'G') + (hexpand[2*i+1] - 'G')
	}
	return out
}

func Unhexpand(v []byte) {
	for i := range v {
		v[i] -= 'G'
		v[i] = Hex(v[i])
	}
}
