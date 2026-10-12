package main

func rotateRunes(s string, shift int) string {
	resultingSlice := []rune(s)
	sRuneLen := len(resultingSlice)
	if sRuneLen == 0 {
		return ""
	}

	reducedShift := shift % sRuneLen // to prevent possible overflow
	for i, r := range []rune(s) {
		newIdx := (i - reducedShift) % sRuneLen
		if newIdx < 0 {
			newIdx += sRuneLen
		}
		resultingSlice[newIdx] = r
	}

	return string(resultingSlice)
}
