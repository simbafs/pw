package main

import (
	"crypto/sha256"
	"strings"
)

type SegmentType int

const (
	SegmentUpper SegmentType = iota
	SegmentLower
	SegmentDigit
)

type SegNormalizer struct {
	Policy *SitePolicy
}

func (n SegNormalizer) Normalize(s string) string {
	if len(s) == 0 {
		return s
	}

	length := 12 // default length
	// 如果有 site policy，先套 max_len
	if n.Policy != nil && n.Policy.MaxLen > 0 {
		length = n.Policy.MaxLen
	}
	if len(s) < length {
		panic("encoded string too short")
	}

	s = s[:length]

	// 先做「原本的三段規則」
	base := normalizeSegmented(s)

	// 沒有 site policy → 用原規則就好
	if n.Policy == nil {
		return base
	}

	// 有 policy → 再強制補 requirement（大小寫、數字、特殊符號）
	return applyPolicy(base, n.Policy)
}

// 原本三段邏輯
func normalizeSegmented(s string) string {
	if len(s) == 0 {
		return s
	}

	L := len(s)
	if L < 3 {
		return normalizeSingleSegment(s)
	}

	// 分段：盡量平均分三段
	segLens := make([]int, 3)
	base := L / 3
	rem := L % 3
	for i := range 3 {
		segLens[i] = base
		if i < rem {
			segLens[i]++
		}
	}

	hash := sha256.Sum256([]byte(s))

	segTypes := make([]SegmentType, 3)
	for i := range 3 {
		segTypes[i] = SegmentType(hash[i] % 3) // 0,1,2
	}

	out := make([]byte, 0, L)
	start := 0
	for i := range 3 {
		end := min(start+segLens[i], L)
		if start >= end {
			continue
		}
		segment := s[start:end]
		out = append(out, normalizeSegmentByType(segment, segTypes[i])...)
		start = end
	}

	return string(out)
}

// 把整串當單一 segment，用 hash 決定類型
func normalizeSingleSegment(s string) string {
	hash := sha256.Sum256([]byte(s))
	segType := SegmentType(hash[0] % 3)
	return string(normalizeSegmentByType(s, segType))
}

// 把 segment 依指定類型轉成大寫 / 小寫 / 數字
func normalizeSegmentByType(seg string, t SegmentType) []byte {
	out := make([]byte, len(seg))
	for i, r := range seg {
		ch := byte(r)
		idx := max(strings.IndexByte(base62Alphabet, ch), 0)

		switch t {
		case SegmentUpper:
			out[i] = byte('A' + (idx % 26))
		case SegmentLower:
			out[i] = byte('a' + (idx % 26))
		case SegmentDigit:
			out[i] = byte('0' + (idx % 10))
		}
	}
	return out
}
