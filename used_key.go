package main

//import "fmt"
import "sync"

const enable_used_key_feature = true

var used_key_mutex sync.RWMutex

var used_height_commits map[uint64][][32]byte
var used_key_min map[[32]byte]used_key_minimum
var used_commit_keys map[[32]byte][][32]byte

type used_key_minimum struct {
	height uint64
	commit [32]byte
}

func init() {
	used_height_commits = make(map[uint64][][32]byte)
	used_key_min = make(map[[32]byte]used_key_minimum)
	used_commit_keys = make(map[[32]byte][][32]byte)
}

func used_key_fetch(key [32]byte) (min used_key_minimum, has bool) {
	used_key_mutex.RLock()
	min, has = used_key_min[key]
	used_key_mutex.RUnlock()
	return min, has
}

func used_key_try_add(commitment [32]byte, best *[32]byte, height uint64) uint64 {
	commits_mutex.RLock()

	commitment = commit(commitment[0:])

	var tag, ok = commits[commitment]

	if ok {
		//fmt.Println("detected used key", commitment)
		var ctag = tag

		if uint64(ctag.height) < height {
			height = uint64(ctag.height)
			//fmt.Println("earlier used key", commitment, height)

			*best = commitment
		}
	}

	commits_mutex.RUnlock()
	return height
}

func used_key_add_new_minimal_commit_height(key, commitment [32]byte, height uint64) {

	var empty used_key_minimum

	used_key_mutex.Lock()

	if used_key_min[key] != empty {

		var min_height = used_key_min[key].height
		var min_commit = used_key_min[key].commit

		for i, v := range used_height_commits[min_height] {
			if v == commitment {
				l := len(v) - 1
				if l == 0 {
					used_height_commits[min_height] = nil
				} else {
					used_height_commits[min_height][i] = used_height_commits[min_height][l]
					used_height_commits[min_height] = used_height_commits[min_height][0:l]
				}
				break
			}
		}

		for i, v := range used_commit_keys[min_commit] {
			if v == key {
				l := len(v) - 1
				if l == 0 {
					used_commit_keys[min_commit] = nil
				} else {
					used_commit_keys[min_commit][i] = used_commit_keys[min_commit][l]
					used_commit_keys[min_commit] = used_commit_keys[min_commit][0:l]
				}
				break
			}
		}

	}

	used_key_min[key] = used_key_minimum{height, commitment}

	var found1, found2 bool

	for _, v := range used_height_commits[height] {
		if v == commitment {
			found1 = true
			break
		}
	}
	for _, v := range used_commit_keys[commitment] {
		if v == key {
			found2 = true
			break
		}
	}

	if !found1 {
		used_height_commits[height] = append(used_height_commits[height], commitment)
	}
	if !found2 {
		used_commit_keys[commitment] = append(used_commit_keys[commitment], key)
	}

	used_key_mutex.Unlock()
}

func used_key_commit_reorg(commitment [32]byte, height uint64) {

	var empty, min used_key_minimum

	min.commit = commitment
	min.height = height

	used_key_mutex.Lock()

	for _, key := range used_commit_keys[min.commit] {
		if used_key_min[key] == min {
			used_key_min[key] = empty
		}
	}
	used_key_mutex.Unlock()
}

func used_key_height_reorg(height uint64) {
	used_key_mutex.Lock()

	delete(used_height_commits, height)

	used_key_mutex.Unlock()
}
