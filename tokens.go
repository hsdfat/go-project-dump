package main

// estimateTokens returns an approximate LLM token count for s.
//
// It uses the widely-used heuristic of ~4 characters per token, which tracks
// closely with byte-pair encoders such as OpenAI's cl100k_base and Anthropic's
// tokenizer for typical source code and prose. It is deliberately dependency-
// free and approximate: the goal is to let users gauge whether a dump fits in a
// model's context window, not to bill against it.
func estimateTokens(s string) int {
	if s == "" {
		return 0
	}
	// Round up so non-empty input never reports zero tokens.
	return (len(s) + 3) / 4
}

// humanCount formats an integer with thousands separators (e.g. 12345 -> "12,345").
func humanCount(n int) string {
	if n < 0 {
		return "-" + humanCount(-n)
	}
	digits := []byte{}
	for n >= 1000 {
		digits = append([]byte{byte('0' + n%10)}, digits...)
		digits = append([]byte{byte('0' + (n/10)%10)}, digits...)
		digits = append([]byte{byte('0' + (n/100)%10)}, digits...)
		digits = append([]byte{','}, digits...)
		n /= 1000
	}
	head := []byte{}
	if n == 0 {
		head = []byte{'0'}
	}
	for n > 0 {
		head = append([]byte{byte('0' + n%10)}, head...)
		n /= 10
	}
	return string(head) + string(digits)
}
