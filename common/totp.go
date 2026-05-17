package common

import (
	"crypto/hmac"
	"crypto/sha1"
	"encoding/base32"
	"encoding/binary"
	"fmt"
	"strings"
	"time"
)

func GenerateTOTP(secret string) string {
	key, err := base32.StdEncoding.WithPadding(base32.NoPadding).DecodeString(strings.ToUpper(secret))
	if err != nil {
		return ""
	}
	counter := make([]byte, 8)
	binary.BigEndian.PutUint64(counter, uint64(time.Now().Unix()/30))
	hash := hmac.New(sha1.New, key)
	hash.Write(counter)
	h := hash.Sum(nil)
	offset := h[len(h)-1] & 0x0f
	code := ((int(h[offset])&0x7f)<<24 | int(h[offset+1]&0xff)<<16 | int(h[offset+2]&0xff)<<8 | int(h[offset+3]&0xff)) % 1000000
	return fmt.Sprintf("%06d", code)
}
