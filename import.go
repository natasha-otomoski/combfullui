package main

import (
	"fmt"
	//"github.com/gorilla/mux"
	"bufio"
	"bytes"
	"io"
	"net/http"
	"sync"
)

type import_item struct {
	length int
	prefix string
}

type import_stats struct {
	stacks_count  int
	tx_count      int
	merkle_count  int
	key_count     int
	decider_count int
}

func general_purpose_import(w http.ResponseWriter, r *http.Request) {

	go func(done <-chan struct{}) {
		<-done
		r.Body.Close()
	}(r.Context().Done())

	var stats import_stats

	var bufr = bufio.NewReaderSize(r.Body, 2048)

	var err error

	var wg sync.WaitGroup

	for {
		var buffer bytes.Buffer

		var l []byte
		var isPrefix bool
		for {
			l, isPrefix, err = bufr.ReadLine()
			buffer.Write(l)

			if !isPrefix {
				break
			}

			if err != nil {
				break
			}
		}

		if err == io.EOF {
			break
		} else if err == http.ErrBodyReadAfterClose {
			break
		}

		line := buffer.String()

		if len(line) < 3 {
			continue
		}

		if line[0] != '/' {
			continue
		}

		var import_data = []import_item{{156, "/stack/data/"}, {1481, "/tx/recv/"}, {1421, "/merkle/data/"}, {1357, "/wallet/data/"}, {204, "/purse/data/"}}

		for _, v := range import_data {

			if len(line) != v.length {
				continue
			}
			if line[0:len(v.prefix)] != v.prefix {
				continue
			}
			switch v.length {
			case 156:
				stats.stacks_count++
				wg.Add(1)
				go func() {
					stack_load_data_internal(DummyHttpWriter{}, line[12:])
					wg.Done()
				}()
			case 204:
				stats.decider_count++
				wg.Add(1)
				go func() {
					decider_load_data_internal(DummyHttpWriter{}, line[12:])
					wg.Done()
				}()
			case 1481:
				stats.tx_count++
				wg.Add(1)
				go func() {
					tx_receive_transaction_internal(DummyHttpWriter{}, line[9:])
					wg.Done()
				}()
			case 1421:
				stats.merkle_count++
				wg.Add(1)
				go func() {
					merkle_load_data_internal(DummyHttpWriter{}, line[13:])
					wg.Done()
				}()
			case 1357:
				stats.key_count++
				wg.Add(1)
				go func() {
					key_load_data_internal(DummyHttpWriter{}, line[13:])
					wg.Done()
				}()
			}
		}
	}

	wg.Wait()

	fmt.Fprintf(w, `<html><head></head><body><a href="/">&larr; Back to home</a><br />`)
	defer fmt.Fprintf(w, `</body></html>`)

	fmt.Fprintln(w, stats)
}

func gui_import(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintf(w, `
		<html><head></head><body>
			<a href="/">&larr; Back to home</a><br />
			<form action="/import/general" method="post" enctype="multipart/form-data">
			    Select coin history file to import into wallet:
			    <input type="file" name="fileToUpload" id="fileToUpload">
			    <input type="submit" value="Load Coin History" name="submit">
			</form>
		</body></html>
	`)
}
