package domain

import (
	"fmt"
	"time"

	"github.com/google/uuid"
)

type SigningAlgorithm string

const (
	SigningAlgorithmRS256 SigningAlgorithm = "RS256"
	SigningAlgorithmRS384 SigningAlgorithm = "RS384"
	SigningAlgorithmRS512 SigningAlgorithm = "RS512"
)

func (a SigningAlgorithm) Valid() bool {
	switch a {
	case SigningAlgorithmRS256, SigningAlgorithmRS384, SigningAlgorithmRS512:
		return true
	default:
		return false
	}
}

type SigningKeyStatus string

const (
	SigningKeyStatusPending  SigningKeyStatus = "pending"
	SigningKeyStatusActive   SigningKeyStatus = "active"
	SigningKeyStatusInactive SigningKeyStatus = "inactive"
	SigningKeyStatusExpired  SigningKeyStatus = "expired"
)

type SigningKeyPlanReason string

const (
	SigningKeyPlanReasonBootstrap SigningKeyPlanReason = "bootstrap"
	SigningKeyPlanReasonRecovery  SigningKeyPlanReason = "recovery"
	SigningKeyPlanReasonRotation  SigningKeyPlanReason = "rotation"
)

const MaximumTokenTTL = 24 * time.Hour

type PublicKey struct {
	Modulus  []byte
	Exponent []byte
}

type KeyPair struct {
	KeyID      string
	Algorithm  SigningAlgorithm
	PublicKey  PublicKey
	PrivateKey Secret
}

type SigningKeyWindow struct {
	IntroducedAt time.Time
	ActiveFrom   time.Time
	ActiveUntil  time.Time
	DeleteAfter  time.Time
}

type SigningKey struct {
	ID        uuid.UUID
	KeyID     string
	Algorithm SigningAlgorithm
	PublicKey PublicKey
	Window    SigningKeyWindow
	CreatedAt time.Time
}

func (k SigningKey) StatusAt(now time.Time) SigningKeyStatus {
	switch {
	case now.Before(k.Window.ActiveFrom):
		return SigningKeyStatusPending
	case now.Before(k.Window.ActiveUntil):
		return SigningKeyStatusActive
	case now.Before(k.Window.DeleteAfter):
		return SigningKeyStatusInactive
	default:
		return SigningKeyStatusExpired
	}
}

func (k SigningKey) IsPublishedAt(now time.Time) bool {
	return !now.Before(k.Window.IntroducedAt) && now.Before(k.Window.DeleteAfter)
}

func (k SigningKey) CanSignAt(now time.Time) bool {
	return k.StatusAt(now) == SigningKeyStatusActive
}

// SigningKeySchedule is the lifecycle every signing key goes through.
type SigningKeySchedule struct {
	// PendingDuration is how long a new key is published in the JWKS before it signs, so peers can fetch it first.
	PendingDuration time.Duration

	// ActiveDuration is how long a key is the only key that signs.
	ActiveDuration time.Duration

	// InactiveDuration is how long a key stays published after it stops signing, so the tokens it signed still verify;
	// it must not be shorter than the token TTL.
	InactiveDuration time.Duration
}

func (s SigningKeySchedule) Validate(tokenTTL time.Duration) error {
	if s.PendingDuration <= 0 || s.ActiveDuration <= 0 || s.InactiveDuration <= 0 {
		return fmt.Errorf("signing key durations must be positive: pending=%s active=%s inactive=%s",
			s.PendingDuration, s.ActiveDuration, s.InactiveDuration)
	}
	if tokenTTL <= 0 {
		return fmt.Errorf("token TTL must be positive, got %s", tokenTTL)
	}
	if tokenTTL > MaximumTokenTTL {
		return fmt.Errorf("token TTL %s exceeds the maximum of %s", tokenTTL, MaximumTokenTTL)
	}
	if tokenTTL > s.InactiveDuration {
		return fmt.Errorf("token TTL %s must not exceed the inactive duration %s, or tokens signed at the end of the active phase outlive their key",
			tokenTTL, s.InactiveDuration)
	}
	return nil
}

func (s SigningKeySchedule) TotalLifetime() time.Duration {
	return s.PendingDuration + s.ActiveDuration + s.InactiveDuration
}

func (s SigningKeySchedule) PlanNext(latest *SigningKey, now time.Time) (SigningKeyWindow, SigningKeyPlanReason, bool) {
	if latest == nil {
		return s.windowFrom(now, now), SigningKeyPlanReasonBootstrap, true
	}
	if !now.Before(latest.Window.ActiveUntil) {
		return s.windowFrom(now, now), SigningKeyPlanReasonRecovery, true
	}
	if now.Before(latest.Window.ActiveUntil.Add(-s.PendingDuration)) {
		return SigningKeyWindow{}, "", false
	}
	return s.windowFrom(now, latest.Window.ActiveUntil), SigningKeyPlanReasonRotation, true
}

type SigningKeyRotation struct {
	CurrentKeyWindow SigningKeyWindow
	NextKeyWindow    SigningKeyWindow
}

func (s SigningKeySchedule) PlanRotation(latest *SigningKey, now time.Time) (SigningKeyRotation, error) {
	if latest == nil {
		return SigningKeyRotation{}, ErrSigningKeyNotFound
	}
	switch latest.StatusAt(now) {
	case SigningKeyStatusPending:
		return SigningKeyRotation{}, ErrSigningKeyAlreadyPending
	case SigningKeyStatusActive:
	default:
		return SigningKeyRotation{}, ErrSigningKeyNotFound
	}
	activeFrom := now.Add(s.PendingDuration)
	if latest.Window.ActiveUntil.Before(activeFrom) {
		activeFrom = latest.Window.ActiveUntil
	}
	currentKeyWindow := latest.Window
	currentKeyWindow.ActiveUntil = activeFrom
	currentKeyWindow.DeleteAfter = activeFrom.Add(s.InactiveDuration)
	return SigningKeyRotation{CurrentKeyWindow: currentKeyWindow, NextKeyWindow: s.windowFrom(now, activeFrom)}, nil
}

func (s SigningKeySchedule) windowFrom(introducedAt, activeFrom time.Time) SigningKeyWindow {
	activeUntil := activeFrom.Add(s.ActiveDuration)
	return SigningKeyWindow{
		IntroducedAt: introducedAt,
		ActiveFrom:   activeFrom,
		ActiveUntil:  activeUntil,
		DeleteAfter:  activeUntil.Add(s.InactiveDuration),
	}
}
