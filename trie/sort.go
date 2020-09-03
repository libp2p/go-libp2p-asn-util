package trie

// Sort returns an ordered list of disjoint ranges.
func (t *Trie) Sort() []Range {
	return t.SortAtDepth(0)
}

func (t *Trie) SortAtDepth(depth int) []Range {
	if t.IsLeaf() {
		if t.IsEmpty() {
			return nil
		} else {
			key, asn := t.Key.Number(), t.Key.ASN
			return []Range{{Lower: Sub(key), Upper: Sup(key), ASN: asn}}
		}
	} else {
		left, right := t.Branch[0].SortAtDepth(depth+1), t.Branch[1].SortAtDepth(depth+1)
		if t.IsEmpty() {
			return append(left, right...)
		} else {
			key, asn := t.Key.Number(), t.Key.ASN
			switch {
			case len(left) == 0 && len(right) == 0:
				return []Range{{Lower: Sub(key), Upper: Sup(key), ASN: asn}}
			case len(left) == 0 && len(right) > 0:
				return MergeRanges(
					rangeBelow(Sub(key), right[0].Lower, asn),
					right,
					rangeAbove(right[len(right)-1].Upper, Sup(key), asn),
				)
			case len(left) > 0 && len(right) == 0:
				return MergeRanges(
					rangeBelow(Sub(key), left[0].Lower, asn),
					left,
					rangeAbove(left[len(left)-1].Upper, Sup(key), asn),
				)
			default:
				return MergeRanges(
					rangeBelow(Sub(key), left[0].Lower, asn),
					left,
					rangeBetween(left[len(left)-1].Upper, right[0].Lower, asn),
					right,
					rangeAbove(right[len(right)-1].Upper, Sup(key), asn),
				)
			}
		}
	}
}

func MergeRanges(r ...[]Range) []Range {
	m := []Range{}
	for _, r := range r {
		m = append(m, r...)
	}
	return m
}

func rangeBelow(lower Sub, upperBoundary Sub, asn string) []Range {
	if !SubLess(lower, upperBoundary) {
		return nil
	} else {
		return []Range{{
			Lower: lower,
			Upper: Prev(upperBoundary),
			ASN:   asn,
		}}
	}
}

func rangeAbove(lowerBoundary Sup, upper Sup, asn string) []Range {
	if !SupLess(lowerBoundary, upper) {
		return nil
	} else {
		return []Range{{
			Lower: Next(lowerBoundary),
			Upper: upper,
			ASN:   asn,
		}}
	}
}

func rangeBetween(lowerBoundary Sup, upperBoundary Sub, asn string) []Range {
	if !SupSubLess(lowerBoundary, upperBoundary) {
		return nil
	} else {
		return []Range{{
			Lower: Next(lowerBoundary),
			Upper: Prev(upperBoundary),
			ASN:   asn,
		}}
	}
}
