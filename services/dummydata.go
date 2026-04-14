package services

import (
	"crypto/sha256"
	"encoding/binary"
	"encoding/hex"
	"fmt"
	"math/rand"

	"github.com/google/uuid"
)

// Dummy data for duress mode.
//
// The threat model (https://nnix.com/projects/cherubgyre/) calls for
// "randomized, fictitious content" that prevents a coercer from pattern-
// matching duress-mode responses across users. Each user-visible value is
// derived from a seed = sha256(real_username), so:
//
//   - Two requests from the same duress-mode session return the same
//     fake content (a coercer refreshing the screen sees stable data).
//   - Two different real users produce completely different fake content,
//     so scanning the install base does not yield a recognizable pattern.
//
// The seed intentionally does NOT include time or request state — stability
// is the point.

// DummyProfile represents the fake profile returned from /profile in duress
// mode. Matches the real profile shape so clients cannot distinguish.
type DummyProfile struct {
	Username string `json:"username"`
	Avatar   string `json:"avatar"`
}

// DummyFollower is one entry in a dummy follower/following list.
type DummyFollower struct {
	Username string `json:"username"`
	Avatar   string `json:"avatar"`
}

func seedRand(seed string) *rand.Rand {
	h := sha256.Sum256([]byte("cherubgyre/dummydata:" + seed))
	s := int64(binary.BigEndian.Uint64(h[:8]))
	return rand.New(rand.NewSource(s))
}

func fakeUsername(r *rand.Rand) string {
	// Three lowercase letters + underscore + three digits — same shape as
	// the legacy username format, so it looks like a real cherubgyre user.
	// Once Phase 3h lands, replace with the wordlist-based slug format.
	const letters = "abcdefghijklmnopqrstuvwxyz"
	b := make([]byte, 7)
	for i := 0; i < 3; i++ {
		b[i] = letters[r.Intn(len(letters))]
	}
	b[3] = '_'
	for i := 4; i < 7; i++ {
		b[i] = byte('0' + r.Intn(10))
	}
	return string(b)
}

func fakeAvatar(username string) string {
	return "https://api.dicebear.com/7.x/identicon/svg?seed=" + username + "&rowColor=000000"
}

// GetDummyProfile returns the fake profile for a duress-mode /profile call.
// seed should be the authenticated user's real username so the output is
// stable across requests.
func GetDummyProfile(seed string) DummyProfile {
	r := seedRand(seed + ":profile")
	username := fakeUsername(r)
	return DummyProfile{
		Username: username,
		Avatar:   fakeAvatar(username),
	}
}

// GetDummyFollowers returns a seeded-random list of 1-5 fake followers.
func GetDummyFollowers(seed string) []DummyFollower {
	r := seedRand(seed + ":followers")
	n := 1 + r.Intn(5)
	out := make([]DummyFollower, n)
	for i := 0; i < n; i++ {
		u := fakeUsername(r)
		out[i] = DummyFollower{Username: u, Avatar: fakeAvatar(u)}
	}
	return out
}

// GetDummyFollowing returns a seeded-random list of 1-5 fake followed users.
// Returns a []string for parity with the real /following response.
func GetDummyFollowing(seed string) []string {
	r := seedRand(seed + ":following")
	n := 1 + r.Intn(5)
	out := make([]string, n)
	for i := 0; i < n; i++ {
		out[i] = fakeUsername(r)
	}
	return out
}

// GetDummyInviteCode returns a fake invite code that looks like a real
// UUID. Unlike the other dummy helpers this intentionally does not use a
// seeded source — the real /invite endpoint generates a fresh UUID every
// call, so the dummy version matches that behavior.
func GetDummyInviteCode() string {
	return uuid.New().String()
}

// DummyUsernameFromHash derives a deterministic fake username from an
// arbitrary input, useful for places where we need to render a "looks
// real" fake identifier without a rand source.
func DummyUsernameFromHash(seed string) string {
	sum := sha256.Sum256([]byte(seed))
	return fmt.Sprintf("%s_%s",
		hex.EncodeToString(sum[:2])[:3],
		hex.EncodeToString(sum[2:4])[:3],
	)
}
