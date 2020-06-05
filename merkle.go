package main

import (
	"crypto/sha256"
	"fmt"
	"github.com/gorilla/mux"
	"net/http"
)

func merkle_construct(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	var addr0, addr1, addr2, secret0, secret1, chaining, siggy = vars["addr0"], vars["addr1"], vars["addr2"], vars["secret0"], vars["secret1"], vars["chaining"], vars["siggy"]

	err1 := checkHEX2(siggy)
	if err1 != nil {
		fmt.Fprintf(w, "error decoding merkle branch siggy from hex: %s", err1)
		return
	}

	var sig16 = hex2uint16([]byte(siggy))

	merkle_construct_internal(w, addr0, addr1, addr2, secret0, secret1, chaining, sig16)
}

func merkle_construct_internal(w http.ResponseWriter, addr0, addr1, addr2, secret0, secret1, chaining string, siggy uint16) {

	var tube [3][17][32]byte

	tube[0][0] = hex2byte32([]byte(addr0))
	tube[1][0] = hex2byte32([]byte(addr1))
	tube[2][0] = hex2byte32([]byte(addr2))

	for y := 0; y < 16; y++ {
		for x := 0; x < 3; x++ {
			tube[x][y+1] = merkle(tube[x][y][:], tube[(x+1)%3][y][:])
		}
	}

	var sig = int(siggy)

	var bitsum byte = 0
	for i := sig; i > 0; i >>= 1 {
		bitsum += byte(i & 1)
	}

	var b0 = tube[(bitsum+2)%3][16]

	fmt.Fprintf(w, "merkle=%X\n", b0)

	var chainz [2][32]byte
	var chtipz [2][32]byte
	chainz[0] = hex2byte32([]byte(secret0))
	chainz[1] = hex2byte32([]byte(secret1))

	for i := 0; i < 65535-sig; i++ {
		chainz[0] = sha256.Sum256(chainz[0][0:])
	}
	for i := 0; i < sig; i++ {
		chainz[1] = sha256.Sum256(chainz[1][0:])
	}
	chtipz = chainz

	for i := 0; i < sig; i++ {
		chtipz[0] = sha256.Sum256(chtipz[0][0:])
	}
	for i := 0; i < 65535-sig; i++ {
		chtipz[1] = sha256.Sum256(chtipz[1][0:])
	}

	fmt.Fprintf(w, "commit0=%X\n", commit(chainz[0][0:]))
	fmt.Fprintf(w, "commit1=%X\n", commit(chainz[1][0:]))

	var a0buf [96]byte

	copy(a0buf[32:64], chtipz[0][0:])
	copy(a0buf[64:96], chtipz[1][0:])

	var a0 = sha256.Sum256(a0buf[0:])
	var e0 = merkle(a0[0:], b0[0:])

	fmt.Fprintf(w, "nextchainer=%X\n", a0)
	fmt.Fprintf(w, "pay-to-root=%X\n", e0)

	fmt.Fprintf(w, "fullbranch=")

	fmt.Fprintf(w, "%X", chtipz[0])
	fmt.Fprintf(w, "%X", chtipz[1])
	fmt.Fprintf(w, "%X", chainz[0])
	fmt.Fprintf(w, "%X", chainz[1])

	var x = 0
	for y := uint(0); y < 16; y++ {
		if ((sig >> y) & 1) == 1 {
			x++
			x %= 3
		} else {
			x += 2
			x %= 3

		}
		fmt.Fprintf(w, "%X", tube[x][y])
		if ((sig >> y) & 1) == 1 {
			x += 2
			x %= 3
		}
	}

	fmt.Fprintf(w, "%X", tube[0][0])

	var chainer [32]byte = hex2byte32([]byte(chaining))
	fmt.Fprintf(w, "%X", chainer)
	fmt.Fprintf(w, "\n")
}

func merkle_mine(c [32]byte) {
	segments_merkle_mutex.Lock()

	merkledata_each_epsilonzeroes(c, func(e0 *[32]byte) bool {
		var e [2][32]byte
		e[0] = *e0

		var e0q = merkle(e[0][0:], c[0:])

		//fmt.Printf("e0q=%X\n", e0q)

		e[1] = segments_merkle_lever[e0q]

		var tx = merkle(e[0][0:], e[1][0:])

		//fmt.Printf("mine tx=%X\n", tx)

		segments_merkle_mutex.Unlock()
		reactivate(tx, e)
		segments_merkle_mutex.Lock()

		return true
	})
	segments_merkle_mutex.Unlock()
}

func merkle_unmine(c [32]byte) {
	segments_merkle_mutex.Lock()
	merkledata_each_epsilonzeroes(c, func(e0 *[32]byte) bool {

		var e [2][32]byte
		e[0] = *e0
		e[1] = e0_to_e1[e[0]]
		var tx = merkle(e[0][0:], e[1][0:])

		//fmt.Printf("unmine tx=%X\n", tx)

		segments_merkle_mutex.Unlock()
		reactivate(tx, e)
		segments_merkle_mutex.Lock()
		return true
	})
	segments_merkle_mutex.Unlock()
}

func merkle_scan_leg_activity(tx [32]byte) (activity uint8) {

	var data [4][32]byte

	segments_merkle_mutex.RLock()

	if data1, ok1 := segments_merkle_blackheart[tx]; ok1 {
		data = data1
	} else if data2, ok2 := segments_merkle_whiteheart[tx]; ok2 {
		data = data2
	} else {
		segments_merkle_mutex.RUnlock()

		//println("no heart")
		return 0
	}

	segments_merkle_mutex.RUnlock()

	var j = 0
outer:
	for i := 0; i < 2; i++ {

		var rawroottag, ok2 = commits[commit(data[i][0:])]

		if !ok2 {
			continue
		}

		var roottag = rawroottag

		var hash = data[i]

		for ; j < sigvariability; j++ {
			hash = sha256.Sum256(hash[0:])
			if hash == data[i+2] {
				j++
				break
			}

			var candidaterawtag, ok3 = commits[commit(hash[0:])]

			if !ok3 {
				continue
			}
			var candidatetag = candidaterawtag

			if utag_cmp(&roottag, &candidatetag) > 0 {

				//fmt.Println("miscompared hash=", hash)

				//panic("")

				continue outer
			}

		}
		//fmt.Println("solved activity", hash)
		activity |= 1 << uint(i)
	}
	//fmt.Println("activity, j", activity, j)
	return activity
}

func notify_transaction(w http.ResponseWriter, a1, a0, u1, u2, q1, q2 [32]byte, z [16][32]byte, b1 [32]byte) (bool, [32]byte) {

	var e [2][32]byte

	var a1_is_zero = a1 == e[0]

	var sig int

	var hash = q1

	for i := 0; i < 65536; i++ {
		if hash == u1 {
			sig = i
			break
		}

		hash = sha256.Sum256(hash[0:])
	}
	if hash != u1 {
		fmt.Fprintf(w, "error merkle solution sig hash 1 does not match")
		return false, [32]byte{}
	}
	hash = q2
	for i := 0; i < 65535-sig; i++ {
		hash = sha256.Sum256(hash[0:])
	}
	if hash != u2 {
		fmt.Fprintf(w, "error merkle solution sig hash 2 does not match")
		return false, [32]byte{}
	}

	var b0 = b1

	for i := byte(0); i < 16; i++ {
		if ((sig >> i) & 1) == 1 {
			b0 = merkle(b0[0:], z[i][0:])
		} else {
			b0 = merkle(z[i][0:], b0[0:])
		}
	}

	e[0] = merkle(a0[0:], b0[0:])

	var cq1 = commit(q1[0:])
	var cq2 = commit(q2[0:])

	segments_merkle_mutex.Lock()

	segments_merkle_uncommit[commit(e[0][0:])] = e[0]

	merkledata_store_epsilonzeroes(cq1, e[0])
	merkledata_store_epsilonzeroes(cq2, e[0])

	//fmt.Printf("e0 = %X\n", e[0])

	if a1_is_zero {
		e[1] = b1

	} else {
		e[1] = merkle(a1[0:], b1[0:])

	}
	var tx = merkle(e[0][0:], e[1][0:])
	if a1_is_zero {
		segments_merkle_whiteheart[tx] = [4][32]byte{q1, q2, u1, u2}
	} else {
		segments_merkle_blackheart[tx] = [4][32]byte{q1, q2, u1, u2}
	}

	var e0q1 = merkle(e[0][0:], cq1[0:])
	var e0q2 = merkle(e[0][0:], cq2[0:])
	segments_merkle_lever[e0q1] = e[1]
	segments_merkle_lever[e0q2] = e[1]

	segments_merkle_mutex.Unlock()

	//var ahash = merkle(a0[0:], a1[0:])

	//segments_merkle_heart[ahash] = [2]commitment_t{q1,q2}
	commits_mutex.Lock()
	//segments_merkle_mutex.Lock()
	reactivate(tx, e)
	//segments_merkle_mutex.Unlock()
	commits_mutex.Unlock()
	return true, e[0]
}

func reactivate(tx [32]byte, e [2][32]byte) {
	var oldactivity = segments_merkle_activity[tx]
	var newactivity = merkle_scan_leg_activity(tx)
	segments_merkle_activity[tx] = newactivity
	if oldactivity != newactivity {
		if oldactivity == 3 {
			//var maybecoinbase = commit(e[0][0:])

			segments_merkle_untrickle(nil, e[0], 0xffffffffffffffff)
			//segments_coinbase_untrickle_auto(maybecoinbase, e[0])

			segments_merkle_mutex.Lock()
			delete(e0_to_e1, e[0])
			segments_merkle_mutex.Unlock()
		}
		if newactivity == 3 {
			segments_merkle_mutex.Lock()
			if _, ok1 := e0_to_e1[e[0]]; ok1 {

				fmt.Println("Panic: e0 to e1 already have live path")
				panic("")
			}

			e0_to_e1[e[0]] = e[1]
			segments_merkle_mutex.Unlock()
			var maybecoinbase = commit(e[0][0:])
			if _, ok1 := combbases[maybecoinbase]; ok1 {
				segments_coinbase_trickle_auto(maybecoinbase, e[0])
			}

			segments_merkle_trickle(make(map[[32]byte]struct{}), e[0])
		}
	}
}
func merkle_load_data_internal(w http.ResponseWriter, data string) {

	err1 := checkHEX704(data)
	if err1 != nil {
		fmt.Fprintf(w, "error decoding transaction from hex: %s", err1)
		return
	}

	var rawdata = hex2byte704([]byte(data))

	var arraydata [22][32]byte

	for i := range arraydata {
		copy(arraydata[i][0:], rawdata[32*i:32+32*i])
	}

	var z [16][32]byte
	for i := range z {
		z[i] = arraydata[MERKLE_DATA_Z0+i]
	}

	var buf3_a0 [96]byte

	copy(buf3_a0[0:32], arraydata[MERKLE_INPUT_A1][0:32])
	copy(buf3_a0[32:64], arraydata[MERKLE_DATA_U1][0:32])
	copy(buf3_a0[64:96], arraydata[MERKLE_DATA_U2][0:32])

	var a0 = sha256.Sum256(buf3_a0[0:])

	//fmt.Fprintf(w, "a0=%X\n", a0)

	var notified, e0 = notify_transaction(w, arraydata[MERKLE_INPUT_A1], a0, arraydata[MERKLE_DATA_U1],
		arraydata[MERKLE_DATA_U2], arraydata[MERKLE_DATA_Q1], arraydata[MERKLE_DATA_Q2], z, arraydata[MERKLE_DATA_B1])

	if notified {

		segments_merkle_mutex.Lock()

		segmets_merkle_userinput[arraydata] = e0

		segments_merkle_mutex.Unlock()
	}
}
func merkle_load_data(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	var data = vars["data"]

	merkle_load_data_internal(w, data)
}
