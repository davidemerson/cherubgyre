package services

import (
	"strings"
	"testing"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

const testSecret = "unit-test-secret-that-is-at-least-32-bytes-long"

func withTestSecret(t *testing.T) {
	t.Helper()
	previous := jwtKey
	SetJWTSecret([]byte(testSecret))
	t.Cleanup(func() { jwtKey = previous })
}

func makeToken(t *testing.T, claims Claims, method jwt.SigningMethod, secret []byte) string {
	t.Helper()
	tok := jwt.NewWithClaims(method, &claims)
	s, err := tok.SignedString(secret)
	if err != nil {
		t.Fatalf("sign: %v", err)
	}
	return s
}

func TestValidatePinBounds(t *testing.T) {
	cases := []struct {
		name string
		pin  string
		ok   bool
	}{
		{"too short", "123", false},
		{"exactly min", "1234", true},
		{"typical", "987654", true},
		{"exactly max", strings.Repeat("a", MaxPinLength), true},
		{"over max", strings.Repeat("a", MaxPinLength+1), false},
		{"absurd", strings.Repeat("a", 10_000), false},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			err := ValidatePin(c.pin)
			if c.ok && err != nil {
				t.Errorf("expected ok, got %v", err)
			}
			if !c.ok && err == nil {
				t.Errorf("expected error for %q, got nil", c.name)
			}
		})
	}
}

func TestValidateTokenAcceptsValidHS256(t *testing.T) {
	withTestSecret(t)
	claims := Claims{
		UserID: "test-user",
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(time.Hour)),
		},
	}
	tok := makeToken(t, claims, jwt.SigningMethodHS256, []byte(testSecret))
	ok, err := ValidateToken(tok)
	if err != nil || !ok {
		t.Errorf("ValidateToken on valid HS256: ok=%v err=%v", ok, err)
	}
}

func TestValidateTokenRejectsExpired(t *testing.T) {
	withTestSecret(t)
	claims := Claims{
		UserID: "test-user",
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(-time.Minute)),
		},
	}
	tok := makeToken(t, claims, jwt.SigningMethodHS256, []byte(testSecret))
	ok, err := ValidateToken(tok)
	if ok || err == nil {
		t.Errorf("expected expired token to be rejected, got ok=%v err=%v", ok, err)
	}
}

func TestValidateTokenRejectsWrongSignature(t *testing.T) {
	withTestSecret(t)
	claims := Claims{
		UserID: "test-user",
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(time.Hour)),
		},
	}
	// Sign with a DIFFERENT secret.
	tok := makeToken(t, claims, jwt.SigningMethodHS256, []byte("this-is-not-the-right-secret-xxxx"))
	ok, err := ValidateToken(tok)
	if ok || err == nil {
		t.Errorf("expected wrong-signature token to be rejected, got ok=%v err=%v", ok, err)
	}
}

func TestValidateTokenRejectsAlgNone(t *testing.T) {
	withTestSecret(t)
	// Craft an alg=none token by hand. The parser's key func explicitly
	// checks for *jwt.SigningMethodHMAC so this must be refused even
	// though the library technically supports alg=none.
	claims := Claims{
		UserID: "attacker",
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(time.Hour)),
		},
	}
	tok := jwt.NewWithClaims(jwt.SigningMethodNone, &claims)
	signed, err := tok.SignedString(jwt.UnsafeAllowNoneSignatureType)
	if err != nil {
		t.Fatalf("sign none: %v", err)
	}
	ok, err := ValidateToken(signed)
	if ok || err == nil {
		t.Errorf("expected alg=none token to be rejected, got ok=%v err=%v", ok, err)
	}
}

func TestValidateTokenRejectsMalformed(t *testing.T) {
	withTestSecret(t)
	cases := []string{
		"",
		"not-a-jwt",
		"a.b.c",
		"Bearer garbage",
	}
	for _, tok := range cases {
		if ok, _ := ValidateToken(tok); ok {
			t.Errorf("ValidateToken(%q) returned ok=true", tok)
		}
	}
}

func TestGetUsernameFromTokenExtractsClaim(t *testing.T) {
	withTestSecret(t)
	claims := Claims{
		UserID: "pegasus-thicket-boston",
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(time.Hour)),
		},
	}
	tok := makeToken(t, claims, jwt.SigningMethodHS256, []byte(testSecret))
	got, err := GetUsernameFromToken(tok)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got != "pegasus-thicket-boston" {
		t.Errorf("GetUsernameFromToken = %q, want pegasus-thicket-boston", got)
	}
}

func TestIsDuressTokenReportsFlag(t *testing.T) {
	withTestSecret(t)
	normal := makeToken(t, Claims{
		UserID: "u",
		IsDuress: false,
		RegisteredClaims: jwt.RegisteredClaims{ExpiresAt: jwt.NewNumericDate(time.Now().Add(time.Hour))},
	}, jwt.SigningMethodHS256, []byte(testSecret))
	if IsDuressToken(normal) {
		t.Error("normal token reported as duress")
	}

	duress := makeToken(t, Claims{
		UserID: "u",
		IsDuress: true,
		RegisteredClaims: jwt.RegisteredClaims{ExpiresAt: jwt.NewNumericDate(time.Now().Add(time.Hour))},
	}, jwt.SigningMethodHS256, []byte(testSecret))
	if !IsDuressToken(duress) {
		t.Error("duress token not reported as duress")
	}
}

func TestParseTokenStripsBearerPrefix(t *testing.T) {
	withTestSecret(t)
	claims := Claims{
		UserID: "u",
		RegisteredClaims: jwt.RegisteredClaims{ExpiresAt: jwt.NewNumericDate(time.Now().Add(time.Hour))},
	}
	raw := makeToken(t, claims, jwt.SigningMethodHS256, []byte(testSecret))
	if ok, err := ValidateToken("Bearer " + raw); !ok || err != nil {
		t.Errorf("Bearer-prefixed token rejected: ok=%v err=%v", ok, err)
	}
	if ok, err := ValidateToken("bearer " + raw); !ok || err != nil {
		t.Errorf("lowercase bearer-prefixed token rejected: ok=%v err=%v", ok, err)
	}
}
