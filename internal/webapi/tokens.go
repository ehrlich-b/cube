package webapi

import "strings"

func ValidMoveToken(token string) bool {
	if len(token) == 0 {
		return false
	}
	if last := token[len(token)-1]; last == '\'' || last == '2' {
		token = token[:len(token)-1]
	}
	if len(token) == 1 && strings.ContainsRune("MESxyz", rune(token[0])) {
		return true
	}
	if strings.HasSuffix(token, "w") {
		token = token[:len(token)-1]
	}
	if len(token) == 2 && token[0] >= '1' && token[0] <= '7' {
		token = token[1:]
	}
	return len(token) == 1 && strings.ContainsRune("URFDLB", rune(token[0]))
}

func FaceStickerCount(face string) (int, bool) {
	count := 0
	for i := 0; i < len(face); {
		if !strings.ContainsRune("WYROGB?", rune(face[i])) {
			return 0, false
		}
		i++
		n := 1
		if i < len(face) && face[i] >= '1' && face[i] <= '9' {
			n = int(face[i] - '0')
			i++
			if i < len(face) && face[i] >= '0' && face[i] <= '9' {
				n = n*10 + int(face[i]-'0')
				i++
			}
		}
		count += n
	}
	return count, len(face) > 0
}
