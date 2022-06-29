// Package cherryBase58 file from https://github.com/akamensky/base58/blob/master/base58.go
package cherryBase58

import (
	"fmt"
	"math/big"
)

var (
	bigIntermediateRadix = big.NewInt(430804206899405824) // 58**10
	alphabet             = "123456789ABCDEFGHJKLMNPQRSTUVWXYZabcdefghijkmnopqrstuvwxyz"
	b58table             = [256]byte{
		255, 255, 255, 255, 255, 255, 255, 255,
		255, 255, 255, 255, 255, 255, 255, 255,
		255, 255, 255, 255, 255, 255, 255, 255,
		255, 255, 255, 255, 255, 255, 255, 255,
		255, 255, 255, 255, 255, 255, 255, 255,
		255, 255, 255, 255, 255, 255, 255, 255,
		255, 0, 1, 2, 3, 4, 5, 6,
		7, 8, 255, 255, 255, 255, 255, 255,
		255, 9, 10, 11, 12, 13, 14, 15,
		16, 255, 17, 18, 19, 20, 21, 255,
		22, 23, 24, 25, 26, 27, 28, 29,
		30, 31, 32, 255, 255, 255, 255, 255,
		255, 33, 34, 35, 36, 37, 38, 39,
		40, 41, 42, 43, 255, 44, 45, 46,
		47, 48, 49, 50, 51, 52, 53, 54,
		55, 56, 57, 255, 255, 255, 255, 255,
		255, 255, 255, 255, 255, 255, 255, 255,
		255, 255, 255, 255, 255, 255, 255, 255,
		255, 255, 255, 255, 255, 255, 255, 255,
		255, 255, 255, 255, 255, 255, 255, 255,
		255, 255, 255, 255, 255, 255, 255, 255,
		255, 255, 255, 255, 255, 255, 255, 255,
		255, 255, 255, 255, 255, 255, 255, 255,
		255, 255, 255, 255, 255, 255, 255, 255,
		255, 255, 255, 255, 255, 255, 255, 255,
		255, 255, 255, 255, 255, 255, 255, 255,
		255, 255, 255, 255, 255, 255, 255, 255,
		255, 255, 255, 255, 255, 255, 255, 255,
		255, 255, 255, 255, 255, 255, 255, 255,
		255, 255, 255, 255, 255, 255, 255, 255,
		255, 255, 255, 255, 255, 255, 255, 255,
		255, 255, 255, 255, 255, 255, 255, 255,
	}
)

// Encode takes a slice of bytes and encodes it to base58 string.
// Leading zero bytes are kept in place for precise decoding.
func Encode(input []byte) string {
	output := make([]byte, 0)
	num := new(big.Int).SetBytes(input)
	mod := new(big.Int)

	var primitiveNum int64
	for num.Sign() > 0 {
		num.DivMod(num, bigIntermediateRadix, mod)
		primitiveNum = mod.Int64()

		// This inner loop reduces the amount of calculations with
		// *big.Int by doing them with int64. This improves performance.
		for i := 0; (num.Sign() > 0 || primitiveNum > 0) && i < 10; i++ {
			output = append(output, alphabet[primitiveNum%58])
			primitiveNum /= 58
		}
	}

	for i := 0; i < len(input) && input[i] == 0; i++ {
		output = append(output, alphabet[0])
	}

	// Revert byte order:
	for i := 0; i < len(output)/2; i++ {
		output[i], output[len(output)-1-i] = output[len(output)-1-i], output[i]
	}
	return string(output)
}

// Decode takes string as an input and returns decoded string and error
// If provided string contains characters illegal for base58 the returned error will be <notnil>
func Decode(input string) (output []byte, err error) {
	result := big.NewInt(0)
	tmpBig := new(big.Int)

	for i := 0; i < len(input); {
		var a, m int64 = 0, 58

		// This inner loop reduces the amount of calculations with
		// *big.Int by doing them with int64. This improves performance.
		for f := true; i < len(input) && (f || i%10 != 0); i++ {
			tmp := b58table[input[i]]
			if tmp == 255 {
				msg := "invalid Base58 input string at character \"%c\", position %d"
				return output, fmt.Errorf(msg, input[i], i)
			}
			a = a*58 + int64(tmp)
			if !f {
				m *= 58
			}
			f = false
		}

		result.Mul(result, tmpBig.SetInt64(m))
		result.Add(result, tmpBig.SetInt64(a))
	}

	tmpBytes := result.Bytes()

	var numZeros int
	for numZeros = 0; numZeros < len(input); numZeros++ {
		if input[numZeros] != '1' {
			break
		}
	}
	length := numZeros + len(tmpBytes)
	output = make([]byte, length)
	copy(output[numZeros:], tmpBytes)

	return
}
