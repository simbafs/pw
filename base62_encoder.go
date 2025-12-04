package main

import "math/big"

const base62Alphabet = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"

type Base62Encoder struct{}

func (Base62Encoder) Encode(b []byte) string {
	if len(b) == 0 {
		return ""
	}

	n := new(big.Int).SetBytes(b)
	zero := big.NewInt(0)
	base := big.NewInt(62)

	if n.Cmp(zero) == 0 {
		return string(base62Alphabet[0])
	}

	var out []byte
	for n.Cmp(zero) > 0 {
		q, r := new(big.Int), new(big.Int)
		q.DivMod(n, base, r)
		n = q
		idx := r.Int64()
		out = append(out, base62Alphabet[idx])
	}

	for i, j := 0, len(out)-1; i < j; i, j = i+1, j-1 {
		out[i], out[j] = out[j], out[i]
	}

	return string(out)
}
