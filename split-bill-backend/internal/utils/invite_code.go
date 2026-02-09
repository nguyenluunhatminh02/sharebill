package utils

import (
	"crypto/rand"
	"math/big"
	"strings"
)

const (
	inviteCodeLength = 8
	inviteCodeChars  = "ABCDEFGHJKLMNPQRSTUVWXYZ23456789" // Removed ambiguous: I, O, 0, 1
)

// GenerateInviteCode generates a random invite code for groups
func GenerateInviteCode() string {
	var sb strings.Builder
	for i := 0; i < inviteCodeLength; i++ {
		idx, _ := rand.Int(rand.Reader, big.NewInt(int64(len(inviteCodeChars))))
		sb.WriteByte(inviteCodeChars[idx.Int64()])
	}
	return sb.String()
}
