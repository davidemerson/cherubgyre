package services

import (
	"regexp"
	"testing"

	"github.com/google/uuid"
)

var fakeUsernameRE = regexp.MustCompile(`^[a-z]{3}_\d{3}$`)

func TestDummyProfileIsStablePerSeed(t *testing.T) {
	// Same seed → same output. This is the whole point of the
	// seeded-random dummy data: a coercer refreshing the screen in
	// duress mode must see identical fake content both times.
	a := GetDummyProfile("cherub-gyre-chicago")
	b := GetDummyProfile("cherub-gyre-chicago")
	if a != b {
		t.Errorf("GetDummyProfile stability violated: %+v vs %+v", a, b)
	}
}

func TestDummyProfileDiffersBetweenSeeds(t *testing.T) {
	// Different real users must produce different fake profiles so
	// the attacker cannot pattern-match across the install base.
	seeds := []string{
		"cherub-gyre-chicago",
		"pegasus-thicket-boston",
		"wolf-mesa-atlanta",
		"raven-fjord-kyoto",
		"phoenix-delta-lisbon",
	}
	seen := make(map[string]struct{}, len(seeds))
	for _, s := range seeds {
		p := GetDummyProfile(s)
		if _, dup := seen[p.Username]; dup {
			t.Errorf("duplicate fake username %q for distinct seeds", p.Username)
		}
		seen[p.Username] = struct{}{}
	}
}

func TestDummyProfileShape(t *testing.T) {
	p := GetDummyProfile("wolf-mesa-atlanta")
	if !fakeUsernameRE.MatchString(p.Username) {
		t.Errorf("dummy username %q does not match fake shape", p.Username)
	}
	if p.Avatar == "" {
		t.Error("dummy profile has empty avatar")
	}
}

func TestDummyFollowersCountInBounds(t *testing.T) {
	for _, seed := range []string{"a", "bb", "ccc", "wolf-mesa-atlanta"} {
		f := GetDummyFollowers(seed)
		if len(f) < 1 || len(f) > 5 {
			t.Errorf("GetDummyFollowers(%q) returned %d entries, want 1..5", seed, len(f))
		}
		for i, entry := range f {
			if !fakeUsernameRE.MatchString(entry.Username) {
				t.Errorf("follower[%d] = %q has wrong shape", i, entry.Username)
			}
		}
	}
}

func TestDummyFollowingCountInBounds(t *testing.T) {
	for _, seed := range []string{"a", "bb", "ccc", "wolf-mesa-atlanta"} {
		f := GetDummyFollowing(seed)
		if len(f) < 1 || len(f) > 5 {
			t.Errorf("GetDummyFollowing(%q) returned %d entries, want 1..5", seed, len(f))
		}
	}
}

func TestDummyFollowersIsStablePerSeed(t *testing.T) {
	a := GetDummyFollowers("pegasus-thicket-boston")
	b := GetDummyFollowers("pegasus-thicket-boston")
	if len(a) != len(b) {
		t.Fatalf("length mismatch on repeat call: %d vs %d", len(a), len(b))
	}
	for i := range a {
		if a[i] != b[i] {
			t.Errorf("follower[%d] differs on repeat call: %+v vs %+v", i, a[i], b[i])
		}
	}
}

func TestDummyInviteCodeIsValidUUID(t *testing.T) {
	for i := 0; i < 10; i++ {
		code := GetDummyInviteCode()
		if _, err := uuid.Parse(code); err != nil {
			t.Errorf("dummy invite code %q is not a valid UUID: %v", code, err)
		}
	}
}

func TestDummyInviteCodeChanges(t *testing.T) {
	// The real /invite endpoint mints a fresh UUID every call; the
	// dummy version should match that behavior so a coercer who taps
	// /invite twice in duress mode sees two different "codes".
	a := GetDummyInviteCode()
	b := GetDummyInviteCode()
	if a == b {
		t.Error("GetDummyInviteCode should not return the same code twice in a row")
	}
}
