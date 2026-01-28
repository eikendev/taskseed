// Package identity derives deterministic identifiers for task instances.
package identity

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
)

// InstanceID returns a deterministic identifier for a calendar, rule, and occurrence.
func InstanceID(calendarURL, ruleID, occurrence string) string {
	canonical := fmt.Sprintf("%s|%s|%s", calendarURL, ruleID, occurrence)
	hash := sha256.Sum256([]byte(canonical))
	return hex.EncodeToString(hash[:])
}
