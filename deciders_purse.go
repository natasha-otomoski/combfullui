package main

import "crypto/rand"
import (
	"crypto/sha256"
	"fmt"
	"github.com/gorilla/mux"
	"net/http"
)
import "sync"

var purse_mutex sync.RWMutex

var purse map[[32]byte][64]byte
var purse_saved int

func init() {
	purse = make(map[[32]byte][64]byte)
}

func decider_load_data_internal(w http.ResponseWriter, keykey string) {
	err1 := checkHEX96(keykey)
	if err1 != nil {
		fmt.Fprintf(w, "error decoding key from hex: %s", err1)
		return
	}

	rawkey := hex2byte96([]byte(keykey))

	for i := 0; i < 32; i++ {
		if rawkey[i] != 0 {
			return
		}
	}

	var buf [64]byte

	copy(buf[0:], rawkey[32:])

	var dec = decider_key_to_short_decider(buf)

	purse_mutex.Lock()

	var newkeys = 0

	if _, ok := purse[dec]; !ok {
		newkeys++
	}
	purse[dec] = buf
	purse_saved += newkeys

	purse_mutex.Unlock()
}

func decider_key_to_short_decider(key [64]byte) (o [32]byte) {
	var buf3_a0 [96]byte
	copy(buf3_a0[32:96], key[0:])
	for i := uint32(0); i < 65535; i++ {
		o = sha256.Sum256(buf3_a0[32:64])
		copy(buf3_a0[32:64], o[0:])
		o = sha256.Sum256(buf3_a0[64:96])
		copy(buf3_a0[64:96], o[0:])
	}
	o = sha256.Sum256(buf3_a0[0:])
	return o
}

func decider_key_to_long_decider(key [64]byte, message uint16) (ret [2][32]byte) {
	var o [32]byte
	for i := uint32(0); i < uint32(message); i++ {
		o = sha256.Sum256(key[0:32])
		copy(key[0:32], o[0:])
	}
	for i := uint32(0); i < 65535-uint32(message); i++ {
		o = sha256.Sum256(key[32:])
		copy(key[32:64], o[0:])
	}
	copy(ret[0][0:], key[0:32])
	copy(ret[1][0:], key[32:64])
	return ret
}

func purse_generate_key(w http.ResponseWriter, r *http.Request) {

	var key [64]byte

	fmt.Fprintf(w, `<html><head></head><body><a href="/purse/">&larr; Back to deciders purse</a><br />`)
	defer fmt.Fprintf(w, `</body></html>`)

	_, err := rand.Read(key[0:])
	if err != nil {
		fmt.Fprintf(w, "error generating true random key: %s", err)
		return
	}

	var short = decider_key_to_short_decider(key)

	purse_mutex.Lock()

	purse[short] = key

	purse_mutex.Unlock()
}

func purse_browse(w http.ResponseWriter, r *http.Request) {

	fmt.Fprintf(w, `<html><head></head><body><a href="/">&larr; Back to home</a><br />`)
	defer fmt.Fprintf(w, `</body></html>`)

	fmt.Fprintf(w, `<h1>Always fully save your wallet from the main page after generating a new decider key in the purse</h1>`)
	fmt.Fprintf(w, `<hr /><div><a href="/purse/generator">Generate decider key</a></div>`)
	fmt.Fprintf(w, `<div>To be signed number (0-65535): <input id="to_be_signed" name="to_be_signed" value="" maxlength="5" /></div>`)
	fmt.Fprintf(w, `<div>Decider to sign with: <input style="font-family:monospace;width:45em" id="decider" name="decider" value="" /></div>`)
	fmt.Fprintf(w, `
			<a href="#" id="pay">decide decider</a>

			<script type="text/javascript">
			  var foo = function() {
			    var number = parseInt(document.getElementById("to_be_signed").value);
			    document.getElementById("pay").href="/sign/decide/"+
				document.getElementById("decider").value+"/"+number.toString(16).toUpperCase().padStart(4, "0"); 
			    return false;
			  };
			  var bar = function(x) {
			    document.getElementById("decider").value = x;
			    foo();
			  }
			document.getElementById("to_be_signed").oninput = foo;
			document.getElementById("to_be_signed").onpropertychange = foo;
			document.getElementById("decider").oninput = foo;
			document.getElementById("decider").onpropertychange = foo;
			</script>
	`)

	fmt.Fprintf(w, `<ul>`)
	purse_mutex.RLock()
	for pub, _ := range purse {
		fmt.Fprintf(w, "<li><a href=\"#\" onclick=\"javascript:bar('%s');\">%s</a></li>\n", Hexpand32(pub[0:]), Hexpand32(pub[0:]))
	}
	purse_mutex.RUnlock()
	fmt.Fprintf(w, `</ul>`)

}

func sign_use_decider(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintf(w, `<html><head></head><body><a href="/purse/">&larr; Back to purse</a><br />`)
	defer fmt.Fprintf(w, `</body></html>`)

	vars := mux.Vars(r)
	var pursekey, destination = vars["decider"], vars["number"]

	err1 := checkHEXPAND32(pursekey)
	if err1 != nil {
		fmt.Fprintf(w, "error signing by using key: %s", err1)
		return
	}
	err2 := checkHEX2(destination)
	if err2 != nil {
		fmt.Fprintf(w, "error signing by using destination: %s", err2)
		return
	}

	var buf = []byte(pursekey)
	Unhexpand(buf)
	pkey := hex2byte32(buf)
	dest := hex2byte2([]byte(destination))

	purse_mutex.RLock()
	purse_value, ok := purse[pkey]

	purse_mutex.RUnlock()

	if !ok {
		fmt.Fprintf(w, "<h1>DECIDER KEY NOT FOUND IN THE WALLET</h1>\n")
		return
	}

	num := 256*uint16(dest[0]) + uint16(dest[1])
	decided := decider_key_to_long_decider(purse_value, uint16_reverse(num))

	fmt.Fprintf(w, "<h1>Only reveal the long decider once the addresses are funded with 6 confirmations</h1>\n")
	fmt.Fprintf(w, "<div>Signed number: %d</div>", num)
	fmt.Fprintf(w, "<div>Long decider: <tt>%s%s%s</tt></div>\n", Hexpand2(dest[0:]), Hexpand32(decided[0][0:]), Hexpand32(decided[1][0:]))

	var dec0 = commit(decided[0][0:])
	var dec1 = commit(decided[1][0:])

	fmt.Fprintf(w, "<div>Address 1: <tt>%s</tt></div>\n", bech32get(dec0[0:]))
	fmt.Fprintf(w, "<div>Address 2: <tt>%s</tt></div>\n", bech32get(dec1[0:]))

	if wallet_selfmining_links {
		fmt.Fprintf(w, `<div>Selfmining link 1: <tt><a href="/mining/mine/%X/%s">mine</a></tt></div>\n`, dec0[0:], serializeutxotag(forcecoinbasefirst(makefaketag())))
		fmt.Fprintf(w, `<div>Selfmining link 2: <tt><a href="/mining/mine/%X/%s">mine</a></tt></div>\n`, dec1[0:], serializeutxotag(forcecoinbasefirst(makefaketag())))
	}

}
