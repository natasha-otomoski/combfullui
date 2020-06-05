package main

const precision = 50

func mult128to128(o1hi uint64, o1lo uint64, o2hi uint64, o2lo uint64, hi *uint64, lo *uint64) {

	mult64to128(o1lo, o2lo, hi, lo)

	*hi += o1hi * o2lo
	*hi += o2hi * o1lo

}

func mult64to128(op1 uint64, op2 uint64, hi *uint64, lo *uint64) {
	var u1 = (op1 & 0xffffffff)
	var v1 = (op2 & 0xffffffff)
	var t = (u1 * v1)
	var w3 = (t & 0xffffffff)
	var k = (t >> 32)

	op1 >>= 32
	t = (op1 * v1) + k
	k = (t & 0xffffffff)
	var w1 = (t >> 32)

	op2 >>= 32
	t = (u1 * op2) + k
	k = (t >> 32)

	*hi = (op1 * op2) + w1 + k
	*lo = (t << 32) + w3
}

func log2(xx uint64) (uint64, uint64) {
	var b uint64 = 1 << (precision - 1)
	var yhi uint64 = 0
	var ylo uint64 = 0
	var zhi uint64 = xx >> (64 - precision)
	var zlo uint64 = xx << precision

	for (zhi > 0) || (zlo >= 2<<precision) {
		zlo = (zhi << (64 - 1)) | (zlo >> 1)
		zhi = zhi >> 1
		if ylo+(1<<precision) < ylo {
			yhi++
		}

		ylo += 1 << precision
	}

	for i := 0; i < precision; i++ {

		mult128to128(zhi, zlo, zhi, zlo, &zhi, &zlo)

		zlo = (zhi << (64 - precision)) | (zlo >> precision)
		zhi = zhi >> precision

		if (zhi > 0) || (zlo >= 2<<precision) {

			zlo = (zhi << (64 - 1)) | (zlo >> 1)
			zhi = zhi >> 1

			if ylo+b < ylo {
				yhi++
			}

			ylo += b
		}
		b >>= 1
	}

	return yhi, ylo
}

const accuracy = precision

func coinsupply(height uint64) (uint64, uint64) {
	var loghi, loglo = log2(height)

	var hi, lo uint64

	mult128to128(loghi, loglo, loghi, loglo, &hi, &lo)

	lo = lo>>(accuracy) | hi<<(64-accuracy)
	hi = hi >> (accuracy)

	var hi2, lo2 uint64

	mult128to128(hi, lo, hi, lo, &hi2, &lo2)

	lo2 = lo2>>(accuracy) | hi2<<(64-accuracy)
	hi2 = hi2 >> (accuracy)

	var hi3, lo3 uint64

	mult128to128(hi, lo, hi2, lo2, &hi3, &lo3)

	lo3 = lo3>>(accuracy) | hi3<<(64-accuracy)
	hi3 = hi3 >> (accuracy)

	lo3 = lo3>>(accuracy) | hi3<<(64-accuracy)
	hi3 = hi3 >> (accuracy)

	return lo3, loglo
}

func ispow2(height uint64) uint64 {
	if (height & (height - 1)) == 0 {
		return 1
	}
	return 0
}

func Coinbase(height uint64) uint64 {
	if height >= 21835313 {
		return 0
	}

	var decrease, _ = coinsupply(height)

	return 210000000 - decrease

}

func beforefast(h uint64) uint64 {
	var off = 0
	for h > 0 {
		h >>= 1
		off++
	}

	return [...]uint64{18446744073499551616,
		0,
		419999999,
		1259999466,
		2939986442,
		6299851762,
		13018918654,
		26453794335,
		53309660584,
		106967881461,
		214093216040,
		427701149758,
		852858119367,
		1696837614699,
		3365959585698,
		6649790891866,
		13064173115642,
		25470501063675,
		49141161513723,
		93447594119613,
		174117323213707,
		314950508520476,
		544328688863285,
		871247361753599,
		1196062033987716,
		1242382339976706}[off]
}

func RemainingAfter(height uint64) uint64 {
	var h = height
	h |= h >> 1
	h |= h >> 2
	h |= h >> 4
	h |= h >> 8
	h |= h >> 16
	h |= h >> 32
	h >>= 1

	var sum uint64 = 1242382339976706 - beforefast(h)

	for i := uint64(h); i <= 21835313; i++ {

		sum -= Coinbase(i)

		if height == i {
			return sum
		}
	}
	return 0
}
