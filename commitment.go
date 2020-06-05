package main

import (
	"crypto/sha256"
)

func commit(hash []byte) [32]byte {
	var buf [64]byte
	var sli []byte
	sli = buf[0:0]

	var whitepaper = hex2byte32([]byte("6AFBAC595C1D07A3D4C5179758F5BCE4462A6C263F6E6DFCD942011433ADAAE7"))

	sli = append(sli, whitepaper[0:]...)
	sli = append(sli, hash[0:]...)

	return sha256.Sum256(sli)
}

func merkle(a []byte, b []byte) [32]byte {
	var buf [64]byte
	var sli []byte
	sli = buf[0:0]

	sli = append(sli, a[0:]...)
	sli = append(sli, b[0:]...)

	return sha256.Sum256(sli)
}
