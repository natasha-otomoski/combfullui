package main

import (
	"fmt"
	"github.com/gorilla/mux"
	"log"
	"net"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"
)

func clean_db_signal() {
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM)
	go func() {
		<-c

		go CommitDbCtrlC(fmt.Errorf("STOP SIGNAL"))

		<-c

		os.Exit(-3)
	}()
}

func main() {

	ln, err6 := net.Listen("tcp", "127.0.0.1:2121")
	if err6 != nil {
		log.Fatal(err6)
	}

	CommitDbLoad()
	initial_writeback_over = true

	r := mux.NewRouter()
	s0 := r.PathPrefix("/").Subrouter()
	s0.HandleFunc("/", main_page)
	s0.HandleFunc("/shutdown", shutdown_page)
	s0.HandleFunc("/version.json", version_page)

	s1 := r.PathPrefix("/wallet").Subrouter()
	s1.HandleFunc("/view", wallet_view)
	s1.HandleFunc("/generator", wallet_generate_key)
	s1.HandleFunc("/brain/{numkeys}/{pass}", wallet_generate_brain)
	s1.HandleFunc("/stealth/{backingkey}/{offset}", wallet_stealth_view)
	s1.HandleFunc("/index.html", wallet_view)
	s1.HandleFunc("/", wallet_view)

	s2 := r.PathPrefix("/sign").Subrouter()

	s2.HandleFunc("/decide/{decider}/{number}", sign_use_decider)
	s2.HandleFunc("/pay/{walletkey}/{destination}", sign_use_key)
	s2.HandleFunc("/multipay/{walletkey}/{change}/{stackbottom}", stackbuilder)
	s2.HandleFunc("/from/{walletkey}", wallet_preview_pay)
	s2.HandleFunc("/index.html", sign_gui)
	s2.HandleFunc("/", sign_gui)

	s3 := r.PathPrefix("/mining").Subrouter()

	s3.HandleFunc("/mine/{commit}/{utxotag}", miner_mine_commit)

	s4 := r.PathPrefix("/tx").Subrouter()

	s4.HandleFunc("/recv/{txn}", tx_receive_transaction)

	s5 := r.PathPrefix("/utxo").Subrouter()
	s5.HandleFunc("/bisect/{cut_off_and_mask}", bisect_view)
	s5.HandleFunc("/commit/{hash}", commit_view)
	s5.HandleFunc("/index.html", utxo_view)
	s5.HandleFunc("/", utxo_view)

	s6 := r.PathPrefix("/merkle").Subrouter()

	s6.HandleFunc("/construct/{addr0}/{addr1}/{addr2}/{secret0}/{secret1}/{chaining}/{siggy}", merkle_construct)
	s6.HandleFunc("/data/{data}", merkle_load_data)

	s7 := r.PathPrefix("/stack").Subrouter()

	s7.HandleFunc("/data/{data}", stack_load_data)
	s7.HandleFunc("/multipaydata/{wallet}/{change}/{data}", stack_load_multipay_data)
	s7.HandleFunc("/stealthdata/{data}", stack_load_stealth_data)
	s7.HandleFunc("/index.html", stacks_view)
	s7.HandleFunc("/", stacks_view)

	s8 := r.PathPrefix("/height").Subrouter()
	s8.HandleFunc("/get", height_view)

	s9 := r.PathPrefix("/export").Subrouter()
	s9.HandleFunc("/history/{target}", routes_all_export)
	s9.HandleFunc("/save/{filename}", routes_all_save)
	s9.HandleFunc("/index.html", history_view)

	s10 := r.PathPrefix("/import").Subrouter()
	s10.HandleFunc("/general", general_purpose_import)
	s10.HandleFunc("/", gui_import)

	s11 := r.PathPrefix("/basiccontract").Subrouter()
	s11.HandleFunc("/amtdecidedlater/{decider}", gui_adl_contract)
	s11.HandleFunc("/trade/{decider}", gui_trade_contract)
	s11.HandleFunc("/amtdecidedlatermkl/{decider}/{min}/{max}/{left}/{right}", gui_adl_contract_merkle)
	s11.HandleFunc("/trademkl/{decider}/{trade}/{forward}/{rollback}", gui_trade_contract_merkle)
	s11.HandleFunc("/", gui_contract)

	s12 := r.PathPrefix("/purse").Subrouter()
	s12.HandleFunc("/view", purse_browse)
	s12.HandleFunc("/generator", purse_generate_key)
	s12.HandleFunc("/index.html", purse_browse)
	s12.HandleFunc("/", purse_browse)

	srv := &http.Server{
		Handler:      r,
		WriteTimeout: 24 * time.Hour,
		ReadTimeout:  24 * time.Hour,
	}
	clean_db_signal()

	err := srv.Serve(ln)
	if err != nil {
		return
	}

}
