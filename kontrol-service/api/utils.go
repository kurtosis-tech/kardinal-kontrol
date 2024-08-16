package api

import "math/rand"

const letterBytes = "abcdefghijklmnopqrstuvwxyz0123456789"
const (
	letterIdxBits = 6
	letterIdxMask = 1<<letterIdxBits - 1
	letterIdxMax  = 63 / letterIdxBits

	templatePrefix = "template-"
)

func getRandID() string {
	n := 10
	b := make([]byte, n)
	// A rand.Int63() generates 63 random bits, enough for letterIdxMax letters!
	for i, cache, remain := n-1, rand.Int63(), letterIdxMax; i >= 0; {
		if remain == 0 {
			cache, remain = rand.Int63(), letterIdxMax
		}
		if idx := int(cache & letterIdxMask); idx < len(letterBytes) {
			b[i] = letterBytes[idx]
			i--
		}
		cache >>= letterIdxBits
		remain--
	}

	return string(b)
}

func getRandFlowID() string {
	return getRandID()
}

func getRandTemplateID() string {
	return templatePrefix + getRandID()
}
