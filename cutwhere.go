package main

const BIG = 9
const LEVELS = 59213
const CHAINS = 21
const ROWS = LEVELS + CHAINS + 1

func cmp(x *[BIG]uint32, y *[BIG]uint32) int64 {
	for i := BIG - 1; i >= 0; i-- {
		var diff int64 = int64(x[i]) - int64(y[i])
		if diff != 0 {
			return diff
		}
	}
	return 0
}

func neg(x *[BIG]uint32, y *[BIG]uint32) {
	for i := 0; i < BIG; i++ {
		x[i] = ^y[i]
	}
}

func add(a *[BIG]uint32, b *[BIG]uint32, c *[BIG]uint32, d int64) {
	for i := 0; i < BIG; i++ {
		d += int64(a[i]) + int64(b[i])
		c[i] = uint32(d)
		d >>= 32
	}
}

func combination(m *[BIG]uint32) (result [CHAINS]uint16) {
	const n = LEVELS + CHAINS - 1
	const k = CHAINS - 1

	var a uint16 = n
	var b uint16 = k

	var pascal_row = [CHAINS][BIG]uint32{
		{1, 0, 0, 0, 0, 0, 0, 0, 0},
		{59233, 0, 0, 0, 0, 0, 0, 0, 0},
		{1754244528, 0, 0, 0, 0, 0, 0, 0, 0},
		{602937712, 8064, 0, 0, 0, 0, 0, 0, 0},
		{3058129352, 119409758, 0, 0, 0, 0, 0, 0, 0},
		{3952490376, 1459879366, 329, 0, 0, 0, 0, 0, 0},
		{867081424, 4203890776, 3251023, 0, 0, 0, 0, 0, 0},
		{571786896, 2495679268, 1737110108, 6, 0, 0, 0, 0, 0},
		{2538551252, 3483201710, 3275810776, 47413, 0, 0, 0, 0, 0},
		{2904167444, 2502439117, 2600040610, 312008899, 0, 0, 0, 0, 0},
		{15236784, 1090272407, 3599736332, 1005569742, 430, 0, 0, 0, 0},
		{2771790704, 1830842580, 4103433979, 1846835476, 2316341, 0, 0, 0, 0},
		{2641563128, 3912880309, 359793491, 4099421989, 2841596421, 2, 0, 0, 0},
		{1353542328, 3504533452, 2312492313, 3545219183, 3716363095, 12124, 0, 0, 0},
		{292641872, 2244364797, 2201954449, 2546316031, 635591981, 51288180, 0, 0, 0},
		{557060880, 4093699000, 1014556106, 3607789982, 1013869665, 618853100, 47, 0, 0},
		{167647410, 4015800687, 123812818, 365245652, 804433926, 698240931, 174486, 0, 0},
		{3649223474, 3784973464, 3050243067, 1727132035, 1555960283, 2531337440, 607796887, 0, 0},
		{2609572944, 3474714567, 3009913845, 2024950701, 594202374, 73291005, 2356901554, 465, 0},
		{2114905744, 1835767214, 651601687, 3420175838, 521655267, 3270496971, 1979891750, 1450919, 0},
		{398112152, 1592865778, 756800062, 2447190614, 424008452, 1245103756, 523637628, 769952, 1},
	}

	var x [BIG]uint32
	var prev uint16 = 0
	var tot uint16 = 0
	neg(&x, m)
	add(&x, &pascal_row[b], &x, 0)
	for i := 0; i < k; i++ {
		for {
			a--
			for j := uint16(0); j < b; j++ {
				neg(&pascal_row[j], &pascal_row[j])
				add(&pascal_row[j], &pascal_row[j+1], &pascal_row[j+1], 1)
				neg(&pascal_row[j], &pascal_row[j])
			}

			if cmp(&pascal_row[b], &x) <= 0 {
				break
			}
		}
		result[i] = n - a - 1 - prev
		tot += result[i]
		prev = n - a
		neg(&x, &x)
		add(&x, &pascal_row[b], &x, 0)
		neg(&x, &x)
		b--
	}
	result[k] = LEVELS - tot
	return result
}

func CutCombWhere(m []byte) [CHAINS]uint16 {
	var x [BIG]uint32
	for i := 0; i < 32; i++ {
		x[i/4] |= uint32(m[i]) << uint(8*(i&3))
	}
	return combination(&x)
}

func main_triangle() {
	var pascal_row [CHAINS][BIG]uint32
	var pascal_row2 [CHAINS][BIG]uint32
	pascal_row[0][0] = 1
	pascal_row[1][0] = 2
	pascal_row[2][0] = 1
	pascal_row2[0][0] = 1
	for j := 0; j < 59231; j++ {
		for i := 0; i < 20; i++ {
			add(&pascal_row[i], &pascal_row[i+1], &pascal_row2[i+1], 0)
		}
		pascal_row = pascal_row2
	}

	for i := range pascal_row {
		print("{")
		for j := range pascal_row[i] {
			print(pascal_row[i][j])
			print(",")
		}
		println("},")
	}
}
