package main

import (
	"crypto/sha256"
	"fmt"
	"github.com/gorilla/mux"
	"net/http"
)

func stack_decode(b []byte) (changeto [32]byte, sumto [32]byte, sum uint64) {

	copy(changeto[0:], b[0:32])
	copy(sumto[0:], b[32:64])

	sum = uint64(b[64])
	sum <<= 8
	sum += uint64(b[65])
	sum <<= 8
	sum += uint64(b[66])
	sum <<= 8
	sum += uint64(b[67])
	sum <<= 8
	sum += uint64(b[68])
	sum <<= 8
	sum += uint64(b[69])
	sum <<= 8
	sum += uint64(b[70])
	sum <<= 8
	sum += uint64(b[71])

	return changeto, sumto, sum
}

func stack_load_data(w http.ResponseWriter, r *http.Request) {

	vars := mux.Vars(r)
	var data = vars["data"]

	fmt.Fprintf(w, `<html><head></head><body>`)
	defer fmt.Fprintf(w, `</body></html>`)
	fmt.Fprintf(w, `<a href="/stack/">&larr; Back to liquidity stacks</a><br />`)
	stack_load_data_internal(w, data)

}

func stack_load_data_internal(w http.ResponseWriter, data string) [32]byte {
	err1 := checkHEX72(data)
	if err1 != nil {
		if w != nil {
			fmt.Fprintf(w, "error decoding liquidity stack item from hex: %s", err1)
		}
		return [32]byte{}
	}

	var rawdata = hex2byte72([]byte(data))

	var hash = sha256.Sum256(rawdata[0:])
	var maybecoinbase = commit(hash[0:])

	segments_stack_mutex.Lock()

	segments_stack[hash] = rawdata
	segments_stack_uncommit[maybecoinbase] = hash

	segments_stack_mutex.Unlock()

	if w != nil {
		fmt.Fprintf(w, "loaded stack address: %X", hash)
	}
	commits_mutex.RLock()
	if _, ok1 := combbases[maybecoinbase]; ok1 {
		segments_coinbase_trickle_auto(maybecoinbase, hash)
	}
	commits_mutex.RUnlock()
	segments_stack_trickle(make(map[[32]byte]struct{}), hash)
	return hash
}
