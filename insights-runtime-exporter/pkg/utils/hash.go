package utils

import (
	"encoding/base64"
	"hash"
)

// HashString returns a base64 URL encoded version of the hash of the
// provided string. The resulting string length of 12 is chosen to have a
// probability of a collision across 1 billion results of 0.0001.
//
// if `hash` is false, simply return the provided string.
//
// similar to from https://github.com/openshift/insights-operator/blob/80246495256b1a4628dd45998aa7162d8e934f78/pkg/gatherers/workloads/gather_workloads_info.go#L415

func HashString(hash bool, h hash.Hash, s string) string {
	if !hash {
		return s
	}
	h.Reset()
	h.Write([]byte(s))
	return base64.RawURLEncoding.EncodeToString(h.Sum(nil))[:12]
}
