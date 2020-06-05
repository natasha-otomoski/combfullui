package main

import (
	"fmt"
	"github.com/gorilla/mux"
	"net/http"
)

func utxo_view(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintf(w, `<html><head></head><body>`)
	fmt.Fprintf(w, `<a href="/">&larr; Back to wallet home from utxo set</a><br /><hr />`)
	defer fmt.Fprintf(w, `</body></html>`)

	commits_mutex.RLock()
	balance_mutex.RLock()

	for key, val := range balance_node {
		if _, ok := combbases[key]; !ok {
			continue
		}

		if val > 0 {
			fmt.Fprintf(w, `<ul><li> <tt>%s</tt> %d.%08d COMB </li></ul>`+"\n", bech32get(key[0:]), combs(val), nats(val))
		}

	}

	fmt.Fprintf(w, `<hr />`+"\n")

	for key, val := range balance_node {
		if _, ok := combbases[key]; ok {
			continue
		}

		if val > 0 {
			fmt.Fprintf(w, `<ul><li> <tt>%X</tt> %d.%08d COMB </li></ul>`+"\n", key, combs(val), nats(val))
		}

	}
	balance_mutex.RUnlock()
	commits_mutex.RUnlock()
}

func commit_view(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	var hash = vars["hash"]

	err1 := checkHEX32(hash)
	if err1 != nil {
		fmt.Fprintf(w, "Invalid hash: %s", err1)
		return
	}

	var rawhash = hex2byte32([]byte(hash))

	commits_mutex.RLock()
	if _, ok := commits[rawhash]; ok {
		fmt.Fprint(w, "Y")
	} else {
		fmt.Fprint(w, "N")
	}
	commits_mutex.RUnlock()
}
