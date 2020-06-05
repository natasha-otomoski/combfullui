package main

import (
	"fmt"
	"github.com/gorilla/mux"
	"net/http"
)

func bisect_view(w http.ResponseWriter, r *http.Request) {

	vars := mux.Vars(r)
	var cut_off_and_mask = vars["cut_off_and_mask"]

	if len(cut_off_and_mask) < 9 {
		fmt.Fprintf(w, "error: cutoff and mask short")
		return
	}

	var cut_off = cut_off_and_mask[1:9]
	var mask = cut_off_and_mask[9:]

	err2 := checkDEC4(cut_off)
	if err2 != nil {
		fmt.Fprintf(w, "error: cutoff invalid: %s", err2)
		return
	}

	var binmask = []byte(mask)
	for i := range binmask {
		if ((binmask[i] >= '0') && (binmask[i] <= '9')) || ((binmask[i] >= 'A') && (binmask[i] <= 'F')) {
		} else {
			fmt.Fprintf(w, "error: binmask invalid")
			return
		}

		binmask[i] = x2b(binmask[i])
	}

	fmt.Fprintf(w, `<html><head></head><body>`)
	defer fmt.Fprintf(w, `</body></html>`)
	fmt.Fprintf(w, `<a href="/">&larr; Back to home</a><br />`)

	var xors [16][32]byte
	var sums [16]uint32
	var depth = len(mask)

	var cutoffheight = int(cut_off[7]-'0') + 10*(int(cut_off[6]-'0')+10*(int(cut_off[5]-'0')+10*(int(cut_off[4]-'0')+10*(int(cut_off[3]-'0')+10*(int(cut_off[2]-'0')+10*(int(cut_off[1]-'0')+10*int(cut_off[0]-'0')))))))

	fmt.Fprintln(w, "<h1>Compare database integrity with peer</h1>")
	fmt.Fprintln(w, "<p>This tool allows to check if there is problem with your commits.db file</p>")
	fmt.Fprintln(w, "<p>This check is performed using two Haircomb Core nodes, by clicking the same button on both nodes</p>")

	if cut_off_and_mask[0] == 'b' || cut_off_and_mask[0] == 'x' || cut_off_and_mask[0] == 'r' {

		fmt.Fprintf(w, "<h2>You are checking Bitcoin blocks 0 to %d</h2>", cutoffheight)
		fmt.Fprintf(w, "<p>Both you and your peer shall now see number <b>%d</b> here. Compare it.</p>", cutoffheight)

		fmt.Fprintln(w, "<p>Next, compare these 16 values with the peers values.</p>")

		commits_mutex.RLock()

		var height = int(commit_currently_loaded.height)

		if height|1023 < 0xFFFFFC00&cutoffheight {
			fmt.Fprintln(w, "<p>Bitcoin block height that you want to compare is in the future. Please choose lower height or sync your node to continue.</p>")
			commits_mutex.RUnlock()
			return
		}

		if cut_off_and_mask[0] == 'b' {

			var initial_height = 0xFFFFFC00 & height

			const towards_genesis = 481280

			initial_height -= towards_genesis
			cutoffheight -= towards_genesis

			var current = initial_height >> 1

			if initial_height != cutoffheight {

				for i := uint(2); i < 32; i++ {
					if current < cutoffheight {
						current += initial_height >> i
					} else if current > cutoffheight {
						current -= initial_height >> i
					} else {
						current = initial_height >> i
						break
					}
				}

			}
			initial_height += towards_genesis
			cutoffheight += towards_genesis

			if current == 0 {

				fmt.Fprintf(w, "<p>Problem appears to be close to block %d. Use options below to isolate the problem. If everything is different, keep clicking the first value difference.</p>\n", cutoffheight)

			} else {

				fmt.Fprintf(w, `<a href="/utxo/bisect/b%08d%s">Everything is different</a>`+"\n", cutoffheight-current, mask)
				if cutoffheight+current <= initial_height {

					fmt.Fprintf(w, ` <a href="/utxo/bisect/b%08d%s">Everything is the same</a><br />`+"\n", cutoffheight+current, mask)

				} else {
					fmt.Fprintf(w, ` <a href="/utxo/bisect/r%08d%s">Everything is the same</a><br />`+"\n", initial_height+512, mask)
				}

			}
		}

		if cut_off_and_mask[0] == 'r' {

			var current = 512

			for i := uint(1); i < 15; i++ {
				if current < cutoffheight&1023 {
					current += 512 >> i
					if i == 14 {
						current = 0
					}
				} else if current > cutoffheight&1023 {
					current -= 512 >> i
					if i == 14 {
						current = 0
					}
				} else {
					current = 512 >> i
					break
				}
			}

			if current == 0 {

				fmt.Fprintf(w, "<p>Problem appears to be close to block %d. Use options below to isolate the problem. If everything is different, keep clicking the first value difference.</p>\n", cutoffheight)

			} else {

				fmt.Fprintf(w, `<a href="/utxo/bisect/r%08d%s">Everything is different</a>`+"\n", cutoffheight-current, mask)

				fmt.Fprintf(w, ` <a href="/utxo/bisect/r%08d%s">Everything is the same</a><br />`+"\n", cutoffheight+current, mask)

			}
		}
	outer:
		for key, tag := range commits {

			if cutoffheight < int(tag.height) {
				continue
			}

			for i, should := range binmask {
				if i&1 == 0 {
					if key[i/2]>>4 != should {
						continue outer
					}
				} else {
					if key[i/2]&0xf != should {
						continue outer
					}
				}
			}

			var bucket = key[depth/2]
			if depth&1 == 0 {
				bucket >>= 4
			} else {
				bucket &= 0xf
			}

			for i, v := range key {
				xors[bucket][i] ^= v
			}

			sums[bucket]++

		}

		commits_mutex.RUnlock()

		fmt.Fprintln(w, "<p>If only some value is different, both users click <b>difference</b> next to that row.</p>")

		for i, xor := range xors {

			if sums[i] == 0 {
				fmt.Fprintf(w, `Row %d. No data. Please click no data in this row on peer's computer if they have data in this exact row<br />`+"\n", i+1)
			} else {

				fmt.Fprintf(w, `Row %d. %x %d <a href="/utxo/bisect/x%s%s">difference</a> <a href="/utxo/bisect/n%s%s">no data</a><br />`+"\n", i+1, xor[depth/2:], sums[i], cut_off, mask+string(Hex(byte(i))), cut_off, mask+string(Hex(byte(i))))

			}
		}

		fmt.Fprintf(w, "<p>Additional information: Locator is %s</p>\n", cut_off_and_mask)

	} else {

		fmt.Fprintf(w, "<h2>Please verify manually that this is true (DO NOT PAY TO ANY ADDRESS):</h2>")

		commits_mutex.RLock()
	outer2:
		for key, tag := range commits {

			if cutoffheight < int(tag.height) {
				continue
			}

			for i, should := range binmask {
				if i&1 == 0 {
					if key[i/2]>>4 != should {
						continue outer2
					}
				} else {
					if key[i/2]&0xf != should {
						continue outer2
					}
				}
			}

			if tag.height >= 620000 {

				fmt.Fprintf(w, `Block %d output %d%04d contains the earliest payment to address %s<br />`+"\n", tag.height, tag.txnum, tag.outnum, bech32get(key[0:]))

			} else {
				fmt.Fprintf(w, `Block %d transaction %d output %d contains the earliest payment to address %s<br />`+"\n", tag.height, tag.txnum, tag.outnum, bech32get(key[0:]))
			}
		}
		commits_mutex.RUnlock()

		fmt.Fprintf(w, "<h2>Your commits.db is corrupt if anything above is not real</h2>")
		fmt.Fprintf(w, "<h4>Your commits.db is corrupt due to missing data if your peer also visited locator %s and additional statement not shown here is real</h4>", cut_off_and_mask)
	}
}
