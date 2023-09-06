package cherryCrypto

import (
	"fmt"
	"testing"
)

func TestBase64Decode(t *testing.T) {
	token := "eyJ0eXAiOiJKV1QiLCJhbGciOiJIUzI1NiJ9"

	val, err := Base64Decode(token)
	fmt.Println(val)
	fmt.Println(err)
}
