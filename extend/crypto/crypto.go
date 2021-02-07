package cherryCrypto

import (
	"crypto/md5"
	"encoding/base64"
	"fmt"
	"hash/crc32"
)

func MD5(value string) string {
	data := []byte(value)
	has := md5.Sum(data)
	return fmt.Sprintf("%x", has)
}

func Base64Encode(value string) string {
	data := []byte(value)
	return base64.StdEncoding.EncodeToString(data)
}

func CRC32(value string) uint {
	return uint(int(crc32.ChecksumIEEE([]byte(value))))
}
