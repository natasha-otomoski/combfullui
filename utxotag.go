package main

import (
	"time"
)

type utxotag struct {
	direction bool
	height    uint32
	txnum     uint16
	outnum    uint16
}

const UTAG_UNMINE int = -1
const UTAG_END_OF_BLOCK int = 0
const UTAG_MINE int = 1

func posttag(t *utxotag) {
	if t.direction {
		t.height--
		t.txnum = 0
		t.outnum = 0
		t.direction = false
	} else {
		t.txnum = 0
		t.outnum = 0
	}
}

func utag_mining_sign(t utxotag) int {
	if t.height >= 99999999 && t.txnum >= 9999 && t.outnum >= 9999 {
		return UTAG_END_OF_BLOCK
	}
	if t.direction {
		return UTAG_UNMINE
	}
	return UTAG_MINE
}

func makefaketag() (tag utxotag) {
	var t = time.Now().UnixNano()

	var fakeheight = (t - 1231006505000000000) / 600000000000
	var tremainder = (t - 1231006505000000000) % 600000000000

	var faketxnnum = tremainder / 60000000
	var ttleftover = tremainder % 60000000

	var outnum = ttleftover / 6000

	tag.height = uint32(fakeheight)
	tag.txnum = uint16(faketxnnum)
	tag.outnum = uint16(outnum)
	return tag
}

func forcecoinbasefirst(t utxotag) utxotag {
	t.txnum = 0
	t.outnum = 0
	return t
}

func serializeutxotag_internal(t utxotag, buf []byte) {
	if t.height > 99999999 {
		panic("up to 99999999 blocks currently possible")
	}
	if t.txnum > 9999 {
		panic("up to 9999 transactions per block allowed")
	}
	if t.outnum > 9999 {
		panic("up to 9999 outputs per transaction allowed")
	}
	var dir byte
	if t.direction {
		dir = 5
	}
	buf[0] = '0' + byte((t.height/10000000)%10) + dir
	buf[1] = '0' + byte((t.height/1000000)%10)
	buf[2] = '0' + byte((t.height/100000)%10)
	buf[3] = '0' + byte((t.height/10000)%10)
	buf[4] = '0' + byte((t.height/1000)%10)
	buf[5] = '0' + byte((t.height/100)%10)
	buf[6] = '0' + byte((t.height/10)%10)
	buf[7] = '0' + byte((t.height/1)%10)
	buf[8] = '0' + byte((t.txnum/1000)%10)
	buf[9] = '0' + byte((t.txnum/100)%10)
	buf[10] = '0' + byte((t.txnum/10)%10)
	buf[11] = '0' + byte((t.txnum/1)%10)
	buf[12] = '0' + byte((t.outnum/1000)%10)
	buf[13] = '0' + byte((t.outnum/100)%10)
	buf[14] = '0' + byte((t.outnum/10)%10)
	buf[15] = '0' + byte((t.outnum/1)%10)
}

func serializeutxotag(t utxotag) []byte {
	var buf [16]byte
	serializeutxotag_internal(t, buf[0:])
	return buf[0:]
}

func deserializeutxotag(buf []byte) (t utxotag) {
	t.height += uint32(buf[0] - '0')
	t.height *= 10
	t.height += uint32(buf[1] - '0')
	t.height *= 10
	t.height += uint32(buf[2] - '0')
	t.height *= 10
	t.height += uint32(buf[3] - '0')
	t.height *= 10
	t.height += uint32(buf[4] - '0')
	t.height *= 10
	t.height += uint32(buf[5] - '0')
	t.height *= 10
	t.height += uint32(buf[6] - '0')
	t.height *= 10
	t.height += uint32(buf[7] - '0')
	t.txnum += uint16(buf[8] - '0')
	t.txnum *= 10
	t.txnum += uint16(buf[9] - '0')
	t.txnum *= 10
	t.txnum += uint16(buf[10] - '0')
	t.txnum *= 10
	t.txnum += uint16(buf[11] - '0')
	t.outnum += uint16(buf[12] - '0')
	t.outnum *= 10
	t.outnum += uint16(buf[13] - '0')
	t.outnum *= 10
	t.outnum += uint16(buf[14] - '0')
	t.outnum *= 10
	t.outnum += uint16(buf[15] - '0')
	if t.height >= 50000000 && t.height < 99999999 {
		t.direction = true
		t.height -= 50000000
	}
	return t
}

func utag_cmp_height(l *utxotag, r *utxotag) int {
	return int(l.height) - int(r.height)
}

func utag_cmp(l *utxotag, r *utxotag) int {
	if l.height != r.height {
		return int(l.height) - int(r.height)
	}
	if l.txnum != r.txnum {
		return int(l.txnum) - int(r.txnum)
	}
	return int(l.outnum) - int(r.outnum)
}
