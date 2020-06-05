package main

import (
	"crypto/sha256"
	"fmt"
	"github.com/gorilla/mux"
	"github.com/sipa/bech32/ref/go/src/bech32"
	"net/http"
)

func bech32get(data []byte) string {
	var buf [32]int
	for i, b := range data {
		buf[i] = int(b)
	}

	encoded, err := bech32.SegwitAddrEncode("bc", 0, buf[0:])
	if err != nil {
		return ""
	}
	return encoded
}

func sign_use_key(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintf(w, `<html><head></head><body><a href="/wallet/">&larr; Back to wallet</a><br />`)
	defer fmt.Fprintf(w, `</body></html>`)

	vars := mux.Vars(r)
	var walletkey, destination = vars["walletkey"], vars["destination"]

	err1 := checkHEX32(walletkey)
	if err1 != nil {
		fmt.Fprintf(w, "error signing by using key: %s", err1)
		return
	}
	err2 := checkHEX32(destination)
	if err2 != nil {
		fmt.Fprintf(w, "error signing by using destination: %s", err2)
		return
	}

	wkey := hex2byte32([]byte(walletkey))
	dest := hex2byte32([]byte(destination))

	var privkey [21][32]byte
	var ok bool

	wallet_mutex.RLock()

	if privkey, ok = wallet[wkey]; !ok {
		wallet_mutex.RUnlock()

		fmt.Fprintf(w, "error signing no such key in wallet")
		return
	}

	wallet_mutex.RUnlock()

	if wkey == dest {
		fmt.Fprintf(w, "error please do not pay to wallet itself, it will form a loop and the funds will be lost")
		return
	}

	{
		var loopdetector = make(map[[32]byte]struct{})
		loopdetector[wkey] = struct{}{}

		if loopdetect(make(map[[32]byte]struct{}), loopdetector, dest) {

			fmt.Fprintf(w, "error token cannot be paid to address where it previously resided, it will form a loop and the funds will be lost")
			return
		}

	}

	var txbuf [736]byte
	var txsli []byte
	txsli = txbuf[0:0]

	txsli = append(txsli, wkey[0:]...)
	txsli = append(txsli, dest[0:]...)

	var txid = sha256.Sum256(txsli)
	depths := CutCombWhere(txid[0:])

	fmt.Fprintf(w, "<h1>txid: %x</h1>\n", txid)
	fmt.Fprintf(w, "<p>To pay to address <tt>%X</tt> follow these steps:</p>\n", dest)
	fmt.Fprintf(w, "<p>1. Press confirm transaction below to store it on your node. Never double-spend this coin again after doing this.</p>\n")
	fmt.Fprintf(w, "<p>2. Pay 546 satoshi each to 21 addresses below:</p>\n")
	fmt.Fprintf(w, "<p>3. Not sooner than these addresses are funded with 6 confirmations, give the history file to the receiver of the token.</p>\n")

	for i := range depths {
		for j := uint16(0); j < uint16(LEVELS)-uint16(depths[i]); j++ {
			privkey[i] = sha256.Sum256(privkey[i][0:])
		}

		var p1 = commit(privkey[i][0:])
		var p2 = serializeutxotag(makefaketag())
		var p3 = privkey[i]

		txsli = append(txsli, privkey[i][0:]...)

		if sign_debug {

			fmt.Fprintf(w, `<ul><li><a href="/mining/mine/%X/%s">%X</a></li></ul>`, p1, p2, p3)
		} else {
			fmt.Fprintf(w, `<ul><li>%s  <b>Any Bitcoin sent here will be lost forever</b></li></ul>`, bech32get(p1[0:]))
		}
	}

	fmt.Fprintf(w, "<a href=\"/tx/recv/%X\">Confirm transaction (remember to save wallet before you shutdown)</a>\n", txsli)

}

func sign_gui(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintf(w, `
		<html><head></head><body><a href="/wallet/">&larr; Back to wallet</a><br />
			From: <input id="from" style="font-family:monospace;width:45em" />
			To: <input id="to" style="font-family:monospace;width:45em" />
			<a href="#" id="pay">pay</a>

			<script type="text/javascript">
			  var foo = function() {
			    document.getElementById("pay").href="/sign/pay/"+
				document.getElementById("from").value+"/"+
				document.getElementById("to").value; 
			    return false;
			  };
			document.getElementById("from").oninput = foo;
			document.getElementById("to").oninput = foo;
			document.getElementById("from").onpropertychange = foo;
			document.getElementById("to").onpropertychange = foo;

			</script>
		</body></html>
	`)
}
