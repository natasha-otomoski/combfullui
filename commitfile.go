package main

import (
	"crypto/aes"
	"fmt"
	"log"
	"os"
	"sync"
	"time"
)

var database_aes [32]byte
var database_file *os.File
var file_mutex sync.Mutex

var initial_writeback_over = false

const fff = "FFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFF"

func CommitDbCtrlC(err error) {
	for initial_writeback_over {

		commit_cache_mutex.Lock()
		commits_mutex.RLock()

		if (len(commit_cache) > 0) || (len(commit_rollback) > 0) {

			commit_cache_mutex.Unlock()
			commits_mutex.RUnlock()

			time.Sleep(10 * time.Millisecond)

			continue
		}

		//CommitDbValidate()

		CommitSumWrite()
		if err != nil {
			log.Fatal(err)
		} else {
			os.Exit(0)
		}

	}
}

func CommitDbOpen(length int64) {
	if !initial_writeback_over {
		return
	}
	file_mutex.Lock()

	f, err := os.OpenFile("commits.db", os.O_APPEND|os.O_WRONLY|os.O_CREATE, 0600)
	if err != nil {
		log.Fatal(err)
	}
	fi, err := f.Stat()
	if err != nil {
		log.Fatal(err)
	}

	var size = fi.Size()

	if (32+8)*length != size {
		log.Fatal(fmt.Errorf("Corrupted commits file size"))
	}

	database_file = f
}
func CommitSumWrite() {
	if !initial_writeback_over {
		return
	}
	file_mutex.Lock()

	f, err := os.OpenFile("commits.db", os.O_APPEND|os.O_WRONLY|os.O_CREATE, 0600)
	if err != nil {
		log.Fatal(err)
	}
	if _, err := f.Write(database_aes[0:]); err != nil {
		f.Close()
		log.Fatal(err)
	}
	if err := f.Sync(); err != nil {
		f.Close()
		log.Fatal(err)
	}

	if err := f.Close(); err != nil {
		log.Fatal(err)
	}
}
func CommitDbUnWrite(key [32]byte) {
	if !initial_writeback_over {
		return
	}

	// begin aes

	var aestmp [32]byte

	var aes, err5 = aes.NewCipher(key[0:])
	if err5 != nil {
		println("ERROR: CANNOT USE CIPHER")
		return
	}

	aes.Decrypt(aestmp[0:16], database_aes[0:16])
	aes.Encrypt(aestmp[16:32], database_aes[16:32])

	for i := 8; i < 16; i++ {
		aestmp[i], aestmp[8+i] = aestmp[8+i], aestmp[i]
	}

	aes.Decrypt(database_aes[0:16], aestmp[0:16])
	aes.Encrypt(database_aes[16:32], aestmp[16:32])

	// end aes

}
func CommitDbWrite(key [32]byte, val [8]byte) {
	if !initial_writeback_over {
		return
	}

	var f *os.File = database_file

	// begin aes

	var aestmp [32]byte

	var aes, err5 = aes.NewCipher(key[0:])
	if err5 != nil {
		f.Close()
		log.Fatal(err5)
	}

	aes.Encrypt(aestmp[0:16], database_aes[0:16])
	aes.Decrypt(aestmp[16:32], database_aes[16:32])

	for i := 8; i < 16; i++ {
		aestmp[i], aestmp[8+i] = aestmp[8+i], aestmp[i]
	}

	aes.Encrypt(database_aes[0:16], aestmp[0:16])
	aes.Decrypt(database_aes[16:32], aestmp[16:32])

	// end aes

	if _, err := f.Write(key[0:]); err != nil {
		f.Close()
		log.Fatal(err)
	}
	if _, err := f.Write(val[0:]); err != nil {
		f.Close()
		log.Fatal(err)
	}
}

func CommitDbClose() {
	if !initial_writeback_over {
		return
	}
	var f *os.File = database_file
	if err := f.Sync(); err != nil {
		log.Fatal(err)
	}

	if err := f.Close(); err != nil {
		log.Fatal(err)
	}
	database_file = nil
	file_mutex.Unlock()
}

func CommitDbTruncate(entries int) {
	if !initial_writeback_over {
		return
	}
	file_mutex.Lock()
	f, err := os.OpenFile("commits.db", os.O_APPEND|os.O_WRONLY|os.O_CREATE, 0600)
	if err != nil {
		log.Fatal(err)
	}
	err2 := f.Truncate(int64(entries) * (32 + 8))
	if err2 != nil {
		log.Fatal(err2)
	}
	if err := f.Sync(); err != nil {
		log.Fatal(err)
	}
	if err := f.Close(); err != nil {
		log.Fatal(err)
	}
	file_mutex.Unlock()
}

func CommitDbValidate() bool {
	file_mutex.Lock()
	defer file_mutex.Unlock()
	f, err := os.OpenFile("commits.db", os.O_RDONLY, 0600)
	if err != nil {
		println("Could not open file")
		return false
	}
	defer f.Close()
	var count int
	var load_aes [32]byte
	for true {
		var buf [32]byte
		n1, err := f.Read(buf[0:])
		if err != nil || n1 != 32 {
			break
		}

		var buf2 [8]byte
		n2, err2 := f.Read(buf2[0:])
		if err2 != nil || n2 != 8 {
			break
		}

		var buf1 = hex2byte8(serializeutxotag(commits[buf]))

		for i := range buf1 {
			if buf1[i] != buf2[i] {
				println("File commit not in ram")
				return false
			}
		}

		// begin aes

		var aestmp [32]byte

		var aes, err5 = aes.NewCipher(buf[0:])
		if err5 != nil {
			println("AES fault")
			return false
		}

		aes.Encrypt(aestmp[0:16], load_aes[0:16])
		aes.Decrypt(aestmp[16:32], load_aes[16:32])

		for i := 8; i < 16; i++ {
			aestmp[i], aestmp[8+i] = aestmp[8+i], aestmp[i]
		}

		aes.Encrypt(load_aes[0:16], aestmp[0:16])
		aes.Decrypt(load_aes[16:32], aestmp[16:32])

		// end aes

		count++
	}
	if load_aes != database_aes {
		println("Checksum is bad")
		fmt.Printf("Should be %X\n", load_aes)
		return false
	}
	if len(commits) != count {
		println("File count not equal to ram count")
		return false
	}

	err3 := f.Close()
	if err3 != nil {
		println("Could not close")

		return false
	}
	return true
}

func CommitDbLoad() bool {
	file_mutex.Lock()
	defer file_mutex.Unlock()
	f, err := os.OpenFile("commits.db", os.O_RDWR|os.O_CREATE, 0600)
	if err != nil {
		println("Could not open file")
		return false
	}
	defer f.Close()
	fi, err := f.Stat()
	if err != nil {
		println("Could not stat file")
		return false
	}

	var size = fi.Size()

	if size == 0 {
		return true
	}

	if size%(32+8) != 32 {
		f.Truncate(0)
		println("File has incorrect size")
		return false
	}

	var lasttag [8]byte
	var loaded int
	var dbload_commitments [][2]string
	var load_aes [32]byte

	for i := int64(0); i < size/(32+8); i++ {
		var buf [32]byte
		n1, err := f.Read(buf[0:])
		if err != nil || n1 != 32 {
			f.Truncate(0)
			println("Read 32 fault")
			return false
		}
		var buf2 [8]byte
		n2, err2 := f.Read(buf2[0:])
		if err2 != nil || n2 != 8 {
			f.Truncate(0)
			println("Read 8 fault")
			return false
		}

		var tagsame = true
		for i := 0; i < 4; i++ {
			tagsame = tagsame && lasttag[i] == buf2[i]
		}

		if !tagsame {
			dbload_commitments = append(dbload_commitments, [2]string{fff, "9999999999999999"})
			loaded = 0
		}

		dbload_commitments = append(dbload_commitments, [2]string{fmt.Sprintf("%X", buf), fmt.Sprintf("%X", buf2)})

		loaded++

		// begin aes

		var aestmp [32]byte

		var aes, err5 = aes.NewCipher(buf[0:])
		if err5 != nil {
			println("AES fault")
			return false
		}

		aes.Encrypt(aestmp[0:16], load_aes[0:16])
		aes.Decrypt(aestmp[16:32], load_aes[16:32])

		for i := 8; i < 16; i++ {
			aestmp[i], aestmp[8+i] = aestmp[8+i], aestmp[i]
		}

		aes.Encrypt(load_aes[0:16], aestmp[0:16])
		aes.Decrypt(load_aes[16:32], aestmp[16:32])

		// end aes

		lasttag = buf2
	}
	if loaded > 0 {
		dbload_commitments = append(dbload_commitments, [2]string{fff, "9999999999999999"})
		loaded = 0
	}

	{
		var checksum [32]byte
		n1, err := f.Read(checksum[0:])
		if err != nil || n1 != 32 {
			f.Truncate(0)
			println("Read sum fault")
			return false
		}

		for i := range checksum {
			if checksum[i] != load_aes[i] {
				f.Truncate(0)
				fmt.Printf("checksum %X %X\n", checksum, load_aes)
				println("Chcksum wrong")
				return false
			}
		}
	}
	err2 := f.Truncate(int64(size - 32))
	if err2 != nil {
		println("Could not truncate")

		return false
	}
	err4 := f.Sync()
	if err4 != nil {
		println("Could not sync")

		return false
	}
	err3 := f.Close()
	if err3 != nil {
		println("Could not close")

		return false
	}

	for _, val := range dbload_commitments {
		miner_mine_commit_internal(&DummyHttpWriter{}, val[0], val[1])
	}

	database_aes = load_aes
	return true
}
