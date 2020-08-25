package trie

// Number represents a formal real number in [0, 1], presented as a bit string.
// The bit string B0, ..., BN represents the number \sum_{i=0}^N Bi * 2^{-i-1}.
// Numbers are intrinsically ordered in the sense of their real number value.
// Numbers are thus real numbers with a finite representation (comprising N bits).
type Number struct {
	// Bytes holds the represented bitstring.
	// Bytes is public to facilitate serialization. It should not be modified directly.
	Bytes []byte
	// Len is the size of the represented bit string (in bits).
	// Len is public to facilitate serialization. It should not be modified directly.
	Len int
}

// Sub represents a formal real number in [0, 1], presented as an infinite bit string.
// The bit string comprises a finite prefix, followed by an infinite number of zeros:
// 	B0, ..., BN, 0, 0, 0, ...
// It is interpreted as a real number (for the sake of ordering), using the formula
//	\sum_{i=0}^\infty Bi * 2^{-i-1}
type Sub Number

// Sup represents a formal real number in [0, 1], presented as an infinite bit string.
// The bit string comprises a finite prefix, followed by an infinite number of ones:
// 	B0, ..., BN, 1, 1, 1, ...
// It is interpreted as a real number (for the sake of ordering), using the formula
//	\sum_{i=0}^\infty Bi * 2^{-i-1}
type Sup Number

func NumbersEqual(x, y Number) bool {
	if x.Len != y.Len {
		return false
	} else {
		return commonPrefixLen(x.Bytes, y.Bytes) >= x.Len
	}
}

// SetBit is a mutable method that sets the i-th bit of the number's underlying bitstring.
// The index i cannot exceed the last index in the number's bitstring.
func (n Number) SetBit(i int, v byte) {
	if v == 0 {
		n.Bytes[i/8] &= ^(1 << (7 - i%8))
	} else {
		n.Bytes[i/8] |= (1 << (7 - i%8))
	}
}

func (n Number) IsZero() bool {
	return n.Len == 0
}

func (n Number) Copy() Number {
	d := make([]byte, len(n.Bytes))
	copy(d, n.Bytes)
	return Number{Bytes: d, Len: n.Len}
}

func (n Number) BitAt(i int) byte {
	d := n.Bytes[i/8] & (byte(1) << (7 - (i % 8)))
	if d == 0 {
		return 0
	} else {
		return 1
	}
}

func commonPrefixLen(a, b []byte) (cpl int) {
	if len(a) > len(b) {
		a = a[:len(b)]
	}
	if len(b) > len(a) {
		b = b[:len(a)]
	}
	for len(a) > 0 {
		if a[0] == b[0] {
			cpl += 8
			a = a[1:]
			b = b[1:]
			continue
		}
		bits := 8
		ab, bb := a[0], b[0]
		for {
			ab >>= 1
			bb >>= 1
			bits--
			if ab == bb {
				cpl += bits
				return
			}
		}
	}
	return
}

// Prev returns the largest number smaller than l, with respect to the formal real number order.
func Prev(l Sub) Sup {
	if i := Number(l).leastSignificantOne(); i < 0 {
		return Sup(Number{})
	} else {
		r := Number(l).Copy()
		r.SetBit(i, 0)
		r.Len = i + 1
		return Sup(r)
	}
}

// Next returns the smallest number larger than l, with respect to the formal real number order.
func Next(u Sup) Sub {
	if i := Number(u).leastSignificantZero(); i < 0 {
		return Sub(Number{})
	} else {
		r := Number(u).Copy()
		r.SetBit(i, 1)
		r.Len = i + 1
		return Sub(r)
	}
}

func (n Number) leastSignificantOne() int {
	for i := 0; i < n.Len; i++ {
		if n.BitAt(n.Len-1-i) == 1 {
			return n.Len - 1 - i
		}
	}
	return -1
}

func (n Number) leastSignificantZero() int {
	for i := 0; i < n.Len; i++ {
		if n.BitAt(n.Len-1-i) == 0 {
			return n.Len - 1 - i
		}
	}
	return -1
}

func numberCommonPrefix(x, y Number) int {
	cpl := commonPrefixLen(x.Bytes, y.Bytes)
	return min(cpl, min(x.Len, y.Len))
}

func min(x, y int) int {
	if x < y {
		return x
	}
	return y
}

// SubLess returns true if and only if xsub < ysub with respect to the order of the underlying formal real numbers.
func SubLess(xsub, ysub Sub) bool {
	x, y := Number(xsub), Number(ysub)
	cpl := numberCommonPrefix(x, y)
	switch {
	case cpl < min(x.Len, y.Len):
		return x.BitAt(cpl) == 0
	case cpl < x.Len: // cpl == y.Len
		return false
	case cpl < y.Len: // cpl == x.Len
		for i := cpl; i < y.Len; i++ {
			if y.BitAt(i) == 1 {
				return true
			}
		}
		return true
	default: // cpl == x.Len, cpl == y.Len
		return false
	}
}

// SupLess returns true if and only if xsup < ysup with respect to the order of the underlying formal real numbers.
func SupLess(xsup, ysup Sup) bool {
	x, y := Number(xsup), Number(ysup)
	cpl := numberCommonPrefix(x, y)
	switch {
	case cpl < min(x.Len, y.Len):
		return x.BitAt(cpl) == 0
	case cpl < x.Len: // cpl == y.Len
		for i := cpl; i < x.Len; i++ {
			if x.BitAt(i) == 0 {
				return true
			}
		}
		return true
	case cpl < y.Len: // cpl == x.Len
		return false
	default: // cpl == x.Len, cpl == y.Len
		return false
	}
}

// SupSubLess returns true if and only if xsup < ysub with respect to the order of the underlying formal real numbers.
func SupSubLess(xsup Sup, ysub Sub) bool {
	x, y := Number(xsup), Number(ysub)
	cpl := numberCommonPrefix(Number(x), Number(y))
	switch {
	case cpl < min(x.Len, y.Len):
		return x.BitAt(cpl) == 0
	case cpl < x.Len: // cpl == y.Len
		return false
	case cpl < y.Len: // cpl == x.Len
		return false
	default: // cpl == x.Len, cpl == y.Len
		return false
	}
}
