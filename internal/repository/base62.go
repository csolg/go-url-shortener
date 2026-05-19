package repository

const alphabet = "0123456789abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ"

func Encode(num uint64) string {
	if num == 0 {
		return string(alphabet[0])
	}

	base := uint64(len(alphabet))
	s := make([]byte, 0, 11)

	for num > 0 {
		s = append(s, alphabet[num%base])
		num /= base
	}

	for i, j := 0, len(s)-1; i < j; i, j = i+1, j-1 {
		s[i], s[j] = s[j], s[i]
	}

	return string(s)
}

func Decode(s string) (uint64, bool) {
	var num uint64
	base := uint64(len(alphabet))
	maxUint := ^uint64(0)

	for i := 0; i < len(s); i++ {
		char := s[i]

		var val uint64
		switch {
		case char >= '0' && char <= '9':
			val = uint64(char - '0')
		case char >= 'a' && char <= 'z':
			val = uint64(char-'a') + 10
		case char >= 'A' && char <= 'Z':
			val = uint64(char-'A') + 36
		default:
			return 0, false
		}

		if num > (maxUint-val)/base {
			return 0, false
		}
		num = num*base + val
	}

	return num, true
}
