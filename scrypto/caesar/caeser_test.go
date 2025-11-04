package caesar_test

import (
	"testing"

	"github.com/ductran999/shared-pkg/scrypto/caesar"
	"github.com/stretchr/testify/assert"
)

type testTable struct {
	name  string
	input string
	nonce int
}

func Test_CaesarCryptoGraphy(t *testing.T) {
	testcases := []testTable{
		{
			name:  "message with non chars",
			input: "daniel!",
			nonce: 8,
		},
		{
			name:  "message encrypt with nonce negative",
			input: "daniel",
			nonce: -2,
		},
		{
			name:  "message encrypt with nonce not between 0 and 25",
			input: "daniel",
			nonce: 88,
		},
	}

	for _, tc := range testcases {
		t.Run(tc.name, func(t *testing.T) {
			cipher := caesar.CaesarEncrypt(tc.input, tc.nonce)
			plaintext := caesar.CaesarDecrypt(cipher, tc.nonce)

			assert.Equal(t, tc.input, plaintext)
		})
	}
}
