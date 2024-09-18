package main

func main() {

	c := []byte("daf11")
	a := c[:3]
	b := c[:4]

	println(&a[0]) // 0x17a249
	println(&b[0]) //0x17a249
	println(&a)
	println(&b)

	x := []string{"212", "af"}
	m := make(map[string]int)
	for _, s := range x {
		if c, ok := m[s]; c > -2 {
			print(ok)
			println(m[s])
			m[s] = c - 1
		}
	}
	var b1 interface{}
	b1 = 3
	var b2 = b1.(int)
	println(b2 + 3)
	println(1)
	aaa := []byte{'a', 'b'}
	var bbb []byte
	println(len(aaa))
	print(len(bbb))

	// var buf [20]byte // big enough for 64bit value base 10 2的64次幂足够十进制20位来保存.
	// i := byte('0' + '2' + '1')
	println((-3) & 7)
	println(3)
}
