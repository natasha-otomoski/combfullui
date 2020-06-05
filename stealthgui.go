package main

import (
	"crypto/sha256"
	"fmt"
	"github.com/gorilla/mux"
	"net/http"
)

func paginator(w http.ResponseWriter, key [32]byte, off uint64) {
	if off == 0 {
		fmt.Fprintf(w, "<center>You are on page %d of 10000000000000000 <a href=\"/wallet/stealth/%X/%016d\">Page %d &gt;</a></center>\n", off+1, key, off+1, off+2)
	} else {
		fmt.Fprintf(w, "<center><a href=\"/wallet/stealth/%X/%016d\">&lt; Page %d </a> You are on page %d of 10000000000000000 <a href=\"/wallet/stealth/%X/%016d\">Page %d &gt;</a></center>\n", key, off-1, off, off+1, key, off+1, off+2)
	}
}

func stack_load_stealth_data(w http.ResponseWriter, r *http.Request) {

	vars := mux.Vars(r)
	var data = vars["data"]

	fmt.Fprintf(w, `<html><head></head><body>`)
	defer fmt.Fprintf(w, `</body></html>`)

	stack_load_data_internal(w, data)

	var rawdata = hex2byte72([]byte(data))

	var key [32]byte
	copy(key[0:32], rawdata[0:32])

	var off uint64
	off += uint64(rawdata[40])
	off <<= 8
	off += uint64(rawdata[39])
	off <<= 8
	off += uint64(rawdata[38])
	off <<= 8
	off += uint64(rawdata[37])
	off <<= 8
	off += uint64(rawdata[36])
	off <<= 8
	off += uint64(rawdata[35])
	off <<= 8
	off += uint64(rawdata[34])
	off <<= 8
	off += uint64(rawdata[33])

	fmt.Fprintf(w, `<a href="/wallet/stealth/%X/%016d">&larr; Back to your stealth addresses</a><br />`, key, off)

}

func wallet_stealth_view(w http.ResponseWriter, r *http.Request) {

	vars := mux.Vars(r)
	var backingkey, offset = vars["backingkey"], vars["offset"]

	err1 := checkHEX32(backingkey)
	if err1 != nil {
		fmt.Fprintf(w, "error: stealth backing key invalid: %s", err1)
		return
	}

	err2 := checkDEC8(offset)
	if err2 != nil {
		fmt.Fprintf(w, "error: stealth offset invalid: %s", err2)
		return
	}

	var key = hex2byte32([]byte(backingkey))

	var off uint64
	off += uint64(offset[0] - '0')
	off *= 10
	off += uint64(offset[1] - '0')
	off *= 10
	off += uint64(offset[2] - '0')
	off *= 10
	off += uint64(offset[3] - '0')
	off *= 10
	off += uint64(offset[4] - '0')
	off *= 10
	off += uint64(offset[5] - '0')
	off *= 10
	off += uint64(offset[6] - '0')
	off *= 10
	off += uint64(offset[7] - '0')
	off *= 10
	off += uint64(offset[8] - '0')
	off *= 10
	off += uint64(offset[9] - '0')
	off *= 10
	off += uint64(offset[10] - '0')
	off *= 10
	off += uint64(offset[11] - '0')
	off *= 10
	off += uint64(offset[12] - '0')
	off *= 10
	off += uint64(offset[13] - '0')
	off *= 10
	off += uint64(offset[14] - '0')
	off *= 10
	off += uint64(offset[15] - '0')

	fmt.Fprintf(w, `<html><head></head><body><a href="/wallet/">&larr; Back to wallet</a><br />`)
	defer fmt.Fprintf(w, `</body></html>`)

	fmt.Fprintf(w, "<h1>Stealth addresses redirecting to: %X</h1>\n", key)

	fmt.Fprintf(w, `<h2>Any Bitcoin send to mine address will be LOST FOREVER, however haircomb will be generated.</h2> `+"\n")
	fmt.Fprintf(w, `<p>Only pay 546 satoshi to mine address. The more times you pay to different Bitcoin mine addresses the bigger your chance to win a free haircomb.</p> `+"\n")

	paginator(w, key, off)

	var stack [72]byte
	copy(stack[0:32], key[0:32])
	stack[40] = byte(off >> 56)
	stack[39] = byte(off >> 48)
	stack[38] = byte(off >> 40)
	stack[37] = byte(off >> 32)
	stack[36] = byte(off >> 24)
	stack[35] = byte(off >> 16)
	stack[34] = byte(off >> 8)
	stack[33] = byte(off >> 0)

	for i := uint16(0); i < 256; i++ {

		stack[32] = byte(i)

		var addr = sha256.Sum256(stack[0:])

		segments_stack_mutex.Lock()

		_, ok3 := segments_stack[addr]

		segments_stack_mutex.Unlock()

		balance_mutex.RLock()

		var bal = balance_node[addr]

		balance_mutex.RUnlock()

		var ckey = commit(addr[:])

		commits_mutex.RLock()
		basetag, is_commited := commits[ckey]

		if is_commited && !ok3 {
			if _, ok6 := combbases[ckey]; ok6 {

				var btag = basetag
				var bheight = uint64(btag.height)
				bal += Coinbase(bheight)
			}

		}

		commits_mutex.RUnlock()

		fmt.Fprintf(w, `<ul><li> <tt>%X</tt> %d.%08d COMB `, addr, combs(bal), nats(bal))

		if !is_commited {
			if wallet_selfmining_links {

				fmt.Fprintf(w, `<a href="/mining/mine/%X/%s">mine</a> `, ckey, serializeutxotag(forcecoinbasefirst(makefaketag())))

			} else {
				fmt.Fprintf(w, `(%s mine address) `, bech32get(ckey[0:]))
			}
		}

		if !ok3 {
			fmt.Fprintf(w, "<a href=\"/stack/stealthdata/%X\">sweep</a>\n", stack[0:])
		}

		fmt.Fprint(w, "</li></ul>\n")
	}

	paginator(w, key, off)

}
