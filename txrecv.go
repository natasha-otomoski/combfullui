package main

import (
	"crypto/sha256"
	"fmt"
	"github.com/gorilla/mux"
	"net/http"
)

var tx_legs_activity map[[32]byte]uint32

func init() {
	tx_legs_activity = make(map[[32]byte]uint32)
}

func tx_scan_leg_activity(tx [32]byte) (activity uint32) {

	var data, ok1 = segments_transaction_data[tx]
	if !ok1 {
		return 0
	}

outer:
	for i := 0; i < 21; i++ {

		var rawroottag, ok2 = commits[commit(data[i][0:])]

		if !ok2 {
			continue
		}

		var roottag = rawroottag

		var hash = data[i]

		for j := 0; j < 65536; j++ {
			hash = sha256.Sum256(hash[0:])

			var candidaterawtag, ok3 = commits[commit(hash[0:])]

			if !ok3 {
				continue
			}
			var candidatetag = candidaterawtag

			if utag_cmp(&roottag, &candidatetag) >= 0 {
				continue outer
			}
		}
		activity |= 1 << uint(i)
	}
	return activity
}

func tx_receive_transaction(w http.ResponseWriter, r *http.Request) {

	vars := mux.Vars(r)
	var txn = vars["txn"]

	fmt.Fprintf(w, `<html><head></head><body>`)
	defer fmt.Fprintf(w, `</body></html>`)

	var back = tx_receive_transaction_internal(w, txn)
	fmt.Fprintf(w, `<html><head></head><body><a href="/sign/pay/%X/%X">&larr; Back to payment</a><br />`, back[0], back[1])

}

func tx_receive_transaction_internal(w http.ResponseWriter, txn string) [2][32]byte {

	err1 := checkHEX736(txn)
	if err1 != nil {
		fmt.Fprintf(w, "error decoding transaction from hex: %s", err1)
		return [2][32]byte{}
	}

	transaction := hex2byte736([]byte(txn))

	var txcommitsandfrom [22][32]byte
	var txidandto [2][32]byte

	copy(txcommitsandfrom[21][0:], transaction[0:32])

	for j := 2; j < 23; j++ {

		copy(txcommitsandfrom[j-2][0:], transaction[j*32:j*32+32])
	}

	copy(txidandto[1][0:], transaction[32:64])

	if txcommitsandfrom[21] == txidandto[1] {
		fmt.Fprintf(w, "warning: transaction forms a trivial loop from to: %X", txidandto[1])
	}

	txidandto[0] = sha256.Sum256(transaction[0:64])

	var teeth_lengths = CutCombWhere(txidandto[0][0:])

	for i := 0; i < 21; i++ {
		var hashchain = txcommitsandfrom[i]

		for j := uint16(0); j < teeth_lengths[i]; j++ {
			hashchain = sha256.Sum256(hashchain[0:])
		}

		copy(transaction[64+i*32:i*32+32+64], hashchain[0:])
	}

	var actuallyfrom = sha256.Sum256(transaction[64:])

	if actuallyfrom != txcommitsandfrom[21] {
		fmt.Fprintf(w, "error: transaction is from user: %X", actuallyfrom)
		return [2][32]byte{txcommitsandfrom[21], txidandto[1]}
	}
	commits_mutex.RLock()
	txleg_mutex.Lock()
	segments_transaction_mutex.Lock()

	segments_transaction_uncommit[commit(actuallyfrom[0:])] = actuallyfrom

	for i := 0; i < 21; i++ {
		txlegs_store_leg(commit(txcommitsandfrom[i][0:]), txidandto[0])
	}

	if _, ok := segments_transaction_data[txidandto[0]]; !ok {
		txdoublespends_store_doublespend(actuallyfrom, txidandto)
	}

	segments_transaction_data[txidandto[0]] = txcommitsandfrom

	fmt.Fprintf(w, "<pre>%X</pre>\n", actuallyfrom)

	var oldactivity = tx_legs_activity[txidandto[0]]
	var newactivity = tx_scan_leg_activity(txidandto[0])
	tx_legs_activity[txidandto[0]] = newactivity
	if oldactivity != newactivity {
		if oldactivity == 2097151 {
			//var maybecoinbase = commit(actuallyfrom[0:])

			segments_transaction_untrickle(nil, actuallyfrom, 0xffffffffffffffff)

			//segments_coinbase_untrickle_auto(maybecoinbase, actuallyfrom)

			delete(segments_transaction_next, actuallyfrom)
		}
		if newactivity == 2097151 {

			segments_transaction_next[actuallyfrom] = txidandto

			var maybecoinbase = commit(actuallyfrom[0:])
			if _, ok1 := combbases[maybecoinbase]; ok1 {
				segments_coinbase_trickle_auto(maybecoinbase, actuallyfrom)
			}

			segments_transaction_trickle(make(map[[32]byte]struct{}), actuallyfrom)
		}
	}

	segments_transaction_mutex.Unlock()
	txleg_mutex.Unlock()
	commits_mutex.RUnlock()
	return [2][32]byte{txcommitsandfrom[21], txidandto[1]}
}
