package main

import (
	"fmt"
	"github.com/gorilla/mux"
	"net/http"
	"os"
)

const tx_final_confirmations = 6

func is_in_set(key [32]byte, set map[[32]byte]struct{}) bool {
	if set == nil {
		return true
	}
	if _, is := set[key]; is {
		return true
	}
	return false
}

func routes_all_export(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	var target = vars["target"]

	var set map[[32]byte]struct{}

	if len(target) > 1 {
		err1 := checkHEX32(target)
		if err1 != nil {
			fmt.Fprintf(w, "error decoding target from hex: %s", err1)
			return
		}

		var rawtarget = hex2byte32([]byte(target))

		set = anonymize(combbases, rawtarget)
	}

	segments_stack_mutex.RLock()

	for key, val := range segments_stack {

		if is_in_set(key, set) {

			fmt.Fprintf(w, "/stack/data/%X\r\n", val)

		}

	}

	segments_stack_mutex.RUnlock()

	segments_transaction_mutex.RLock()

	for key, val := range segments_transaction_data {

		if is_in_set(val[21], set) {

			var nxt [32]byte

			txdoublespends_each_doublespend_target(val[21], func(txidto *[2][32]byte) bool {
				if (*txidto)[0] == key {
					nxt = (*txidto)[1]
					return false
				}
				return true
			})

			var confirm_height uint64
			commits_mutex.Lock()
			for i := 0; i < 21; i++ {
				var candidaterawtag, ok3 = commits[commit(val[i][0:])]
				if !ok3 {
					confirm_height = 0xffffffffffffffff - tx_final_confirmations
					break
				}
				var candidatetag = candidaterawtag
				var bheight = uint64(candidatetag.height)
				if bheight > confirm_height {
					confirm_height = bheight
				}

			}
			var confirmed = confirm_height+tx_final_confirmations < uint64(commit_currently_loaded.height)

			commits_mutex.Unlock()

			if !confirmed {
				continue
			}

			fmt.Fprintf(w, "/tx/recv/%X%X%X%X%X%X%X%X%X%X%X%X%X%X%X%X%X%X%X%X%X%X%X\r\n",
				val[21], nxt, val[0], val[1],
				val[2], val[3], val[4], val[5], val[6], val[7], val[8],
				val[9], val[10], val[11], val[12], val[13], val[14],
				val[15], val[16], val[17], val[18], val[19], val[20])

		}
	}
	segments_transaction_mutex.RUnlock()
	segments_merkle_mutex.RLock()
	for mkl := range segmets_merkle_userinput {

		fmt.Fprintf(w, "/merkle/data/%X%X%X%X%X%X%X%X%X%X%X%X%X%X%X%X%X%X%X%X%X%X\r\n",
			mkl[0], mkl[1], mkl[2], mkl[3], mkl[4], mkl[5], mkl[6],
			mkl[7], mkl[8], mkl[9], mkl[10], mkl[11], mkl[12], mkl[13],
			mkl[14], mkl[15], mkl[16], mkl[17], mkl[18], mkl[19], mkl[20],
			mkl[21])
	}

	segments_merkle_mutex.RUnlock()
}

func fileExists(filename string) bool {
	info, err := os.Stat(filename)
	if os.IsNotExist(err) {
		return false
	}
	return !info.IsDir()
}

func routes_all_save(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	var data = vars["filename"]

	fmt.Fprintf(w, `
		<html><head></head><body><a href="/">Back to wallet home</a><br />`)

	defer fmt.Fprintf(w, `</body></html>`)

	var goodfilename []byte

	for _, c := range data {
		if (c == '.') || (c >= 'a' && c <= 'z') || (c >= 'A' && c <= 'Z') || (c >= '0' && c <= '9') {
			goodfilename = append(goodfilename, byte(c))
		}
	}
	if len(goodfilename) > 3 {
		if goodfilename[len(goodfilename)-1] == 't' {
			if goodfilename[len(goodfilename)-2] == 'a' {
				if goodfilename[len(goodfilename)-3] == 'd' {
					if goodfilename[len(goodfilename)-4] == '.' {
						goodfilename = goodfilename[0 : len(goodfilename)-4]
					}
				}
			}
		}
	}

	if len(goodfilename) == 0 {
		goodfilename = append(goodfilename, 'u', 'n', 't', 'i', 't', 'l', 'e', 'd')
	}
	goodfilename = append(goodfilename, '.', 'd', 'a', 't')

	if fileExists(string(goodfilename)) {
		fmt.Fprintf(w, "File or directory already exists: %s\n", string(goodfilename))
		return
	}

	f, err := os.OpenFile(string(goodfilename), os.O_WRONLY|os.O_CREATE, 0600)
	if err != nil {
		fmt.Fprintf(w, "Open error: %s\n", err)
		return
	}
	wsaved, psaved, err2 := routes_all_do_save(f)
	if err2 != nil {
		fmt.Fprintf(w, "Write error: %s\n", err2)
		return
	}
	err3 := f.Sync()
	if err3 != nil {
		fmt.Fprintf(w, "Sync error: %s\n", err3)
		return
	}
	err4 := f.Close()
	if err4 != nil {
		fmt.Fprintf(w, "Close error: %s\n", err4)
		return
	}
	fmt.Fprintf(w, "Saved as: %s\n", string(goodfilename))
	wallet_mutex.Lock()
	wallet_saved = wsaved
	wallet_mutex.Unlock()
	purse_mutex.Lock()
	purse_saved = psaved
	purse_mutex.Unlock()

}

func routes_all_do_save(w *os.File) (int, int, error) {

	var wsaved int
	var psaved int

	wallet_mutex.RLock()

	for _, val := range wallet {
		n, err := fmt.Fprintf(w, "/wallet/data/%X%X%X%X%X%X%X%X%X%X%X%X%X%X%X%X%X%X%X%X%X\r\n",
			val[0], val[1], val[2], val[3], val[4], val[5], val[6],
			val[7], val[8], val[9], val[10], val[11], val[12], val[13],
			val[14], val[15], val[16], val[17], val[18], val[19], val[20])
		if err != nil || n != 1359 {
			if err == nil {
				err = fmt.Errorf("Wrong size")
			}
			return 0, 0, fmt.Errorf("Routes all save: %s", err)
		}
		wsaved++
	}
	wallet_mutex.RUnlock()
	purse_mutex.RLock()
	for _, priv := range purse {
		n, err := fmt.Fprintf(w, "/purse/data/%X%X\r\n", [32]byte{}, priv)
		if err != nil || n != 206 {
			if err == nil {
				err = fmt.Errorf("Wrong size")
			}
			return 0, 0, fmt.Errorf("Routes all save: %s", err)
		}
		psaved++
	}
	purse_mutex.RUnlock()
	segments_stack_mutex.RLock()
	for _, val := range segments_stack {

		n, err := fmt.Fprintf(w, "/stack/data/%X\r\n", val)
		if err != nil || n != 158 {
			if err == nil {
				err = fmt.Errorf("Wrong size")
			}
			return 0, 0, fmt.Errorf("Routes all save: %s", err)
		}
	}
	segments_stack_mutex.RUnlock()
	segments_transaction_mutex.RLock()
	for key, val := range segments_transaction_data {

		var nxt [32]byte

		txdoublespends_each_doublespend_target(val[21], func(txidto *[2][32]byte) bool {
			if (*txidto)[0] == key {
				nxt = (*txidto)[1]
				return false
			}
			return true
		})

		n, err := fmt.Fprintf(w, "/tx/recv/%X%X%X%X%X%X%X%X%X%X%X%X%X%X%X%X%X%X%X%X%X%X%X\r\n",
			val[21], nxt, val[0], val[1],
			val[2], val[3], val[4], val[5], val[6], val[7], val[8],
			val[9], val[10], val[11], val[12], val[13], val[14],
			val[15], val[16], val[17], val[18], val[19], val[20])
		if err != nil || n != 1483 {
			if err == nil {
				err = fmt.Errorf("Wrong size")
			}
			return 0, 0, fmt.Errorf("Routes all save: %s", err)
		}

	}
	segments_transaction_mutex.RUnlock()
	segments_merkle_mutex.RLock()
	for mkl := range segmets_merkle_userinput {

		n, err := fmt.Fprintf(w, "/merkle/data/%X%X%X%X%X%X%X%X%X%X%X%X%X%X%X%X%X%X%X%X%X%X\r\n",
			mkl[0], mkl[1], mkl[2], mkl[3], mkl[4], mkl[5], mkl[6],
			mkl[7], mkl[8], mkl[9], mkl[10], mkl[11], mkl[12], mkl[13],
			mkl[14], mkl[15], mkl[16], mkl[17], mkl[18], mkl[19], mkl[20],
			mkl[21])
		if err != nil || n != 1423 {
			if err == nil {
				err = fmt.Errorf("Wrong size")
			}
			return 0, 0, fmt.Errorf("Routes all save: %s", err)
		}
	}
	segments_merkle_mutex.RUnlock()
	return wsaved, psaved, nil
}

func history_view(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintf(w, `<html><head></head><body>`)
	defer fmt.Fprintf(w, `</body></html>`)
	fmt.Fprintf(w, `<a href="/">&larr; Back to home page</a><br />`)
	fmt.Fprintf(w, `<a href="/export/history/x">Export complete history (not recommended, please send every receiver their own history using the feature below)</a><br />`)
	fmt.Fprintf(w, `
		<hr />
		History of address: <input id="to" style="font-family:monospace;width:45em" />
		<a href="#" id="dohistory">view history</a>

		<script type="text/javascript">
		  var foo = function() {
		    document.getElementById("dohistory").href="/export/history/"+
			document.getElementById("to").value.toUpperCase();
		    return false;
		  };
		document.getElementById("to").oninput = foo;
		document.getElementById("to").onpropertychange = foo;
		</script>
		<hr />

	`)

}
