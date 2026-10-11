package main

func rotateRunes(s string, shift int) string {
	resultingSlice := []rune(s)
	sRuneLen := len(resultingSlice)

	for i, r := range []rune(s) {
		reducedShift := shift % sRuneLen // to prevent possible overflow
		newIdx := (i - reducedShift) % sRuneLen
		if newIdx < 0 {
			newIdx += sRuneLen
		}
		resultingSlice[newIdx] = r
	}

	return string(resultingSlice)
}
