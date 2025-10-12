package process

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

// TestRestartPolicy_DefaultValues tests that RestartPolicy has sensible zero values.
func TestRestartPolicy_DefaultValues(t *testing.T) {
	policy := RestartPolicy{}

	assert.False(t, policy.Enabled)
	assert.Equal(t, 0, policy.MaxAttempts)
	assert.Equal(t, RestartStrategy(""), policy.Strategy)
	assert.Equal(t, time.Duration(0), policy.InitialDelay)
	assert.Equal(t, time.Duration(0), policy.MaxDelay)
	assert.Equal(t, time.Duration(0), policy.ResetWindow)
}

// TestRestartPolicy_ExponentialStrategy tests exponential backoff configuration.
func TestRestartPolicy_ExponentialStrategy(t *testing.T) {
	policy := RestartPolicy{
		Enabled:      true,
		MaxAttempts:  5,
		Strategy:     StrategyExponential,
		InitialDelay: 1 * time.Second,
		MaxDelay:     30 * time.Second,
		ResetWindow:  5 * time.Minute,
	}

	assert.True(t, policy.Enabled)
	assert.Equal(t, 5, policy.MaxAttempts)
	assert.Equal(t, StrategyExponential, policy.Strategy)
	assert.Equal(t, 1*time.Second, policy.InitialDelay)
	assert.Equal(t, 30*time.Second, policy.MaxDelay)
	assert.Equal(t, 5*time.Minute, policy.ResetWindow)
}

// TestRestartPolicy_LinearStrategy tests linear backoff configuration.
func TestRestartPolicy_LinearStrategy(t *testing.T) {
	policy := RestartPolicy{
		Enabled:      true,
		MaxAttempts:  10,
		Strategy:     StrategyLinear,
		InitialDelay: 2 * time.Second,
		MaxDelay:     20 * time.Second,
		ResetWindow:  10 * time.Minute,
	}

	assert.True(t, policy.Enabled)
	assert.Equal(t, 10, policy.MaxAttempts)
	assert.Equal(t, StrategyLinear, policy.Strategy)
	assert.Equal(t, 2*time.Second, policy.InitialDelay)
	assert.Equal(t, 20*time.Second, policy.MaxDelay)
	assert.Equal(t, 10*time.Minute, policy.ResetWindow)
}

// TestRestartPolicy_ConstantStrategy tests constant backoff configuration.
func TestRestartPolicy_ConstantStrategy(t *testing.T) {
	policy := RestartPolicy{
		Enabled:      true,
		MaxAttempts:  0, // unlimited
		Strategy:     StrategyConstant,
		InitialDelay: 5 * time.Second,
		MaxDelay:     5 * time.Second, // same as initial for constant
		ResetWindow:  1 * time.Minute,
	}

	assert.True(t, policy.Enabled)
	assert.Equal(t, 0, policy.MaxAttempts)
	assert.Equal(t, StrategyConstant, policy.Strategy)
	assert.Equal(t, 5*time.Second, policy.InitialDelay)
	assert.Equal(t, 5*time.Second, policy.MaxDelay)
	assert.Equal(t, 1*time.Minute, policy.ResetWindow)
}

// TestRestartPolicy_DisabledPolicy tests disabled restart policy.
func TestRestartPolicy_DisabledPolicy(t *testing.T) {
	policy := RestartPolicy{
		Enabled: false,
		// Other fields don't matter when disabled
		MaxAttempts:  5,
		Strategy:     StrategyExponential,
		InitialDelay: 1 * time.Second,
		MaxDelay:     30 * time.Second,
		ResetWindow:  5 * time.Minute,
	}

	assert.False(t, policy.Enabled)
}

// TestRestartStrategy_EnumValues tests that RestartStrategy enum values are correct.
func TestRestartStrategy_EnumValues(t *testing.T) {
	assert.Equal(t, RestartStrategy("unknown"), StrategyUnknown)
	assert.Equal(t, RestartStrategy("exponential"), StrategyExponential)
	assert.Equal(t, RestartStrategy("linear"), StrategyLinear)
	assert.Equal(t, RestartStrategy("constant"), StrategyConstant)
}

// TestRestartCause_EnumValues tests that RestartCause enum values are correct.
func TestRestartCause_EnumValues(t *testing.T) {
	assert.Equal(t, RestartCause("manual"), CauseManual)
	assert.Equal(t, RestartCause("crash"), CauseCrash)
	assert.Equal(t, RestartCause("policy"), CausePolicy)
}

// TestCalculateBackoffDelay_ExponentialStrategy tests exponential backoff calculations.
func TestCalculateBackoffDelay_ExponentialStrategy(t *testing.T) {
	policy := RestartPolicy{
		Strategy:     StrategyExponential,
		InitialDelay: 1 * time.Second,
		MaxDelay:     30 * time.Second,
	}

	// attempt 0: 1s * 2^0 = 1s
	assert.Equal(t, 1*time.Second, calculateBackoffDelay(0, policy))

	// attempt 1: 1s * 2^1 = 2s
	assert.Equal(t, 2*time.Second, calculateBackoffDelay(1, policy))

	// attempt 2: 1s * 2^2 = 4s
	assert.Equal(t, 4*time.Second, calculateBackoffDelay(2, policy))

	// attempt 3: 1s * 2^3 = 8s
	assert.Equal(t, 8*time.Second, calculateBackoffDelay(3, policy))

	// attempt 4: 1s * 2^4 = 16s
	assert.Equal(t, 16*time.Second, calculateBackoffDelay(4, policy))

	// attempt 5: 1s * 2^5 = 32s, capped at 30s
	assert.Equal(t, 30*time.Second, calculateBackoffDelay(5, policy))

	// attempt 10: capped at 30s
	assert.Equal(t, 30*time.Second, calculateBackoffDelay(10, policy))
}

// TestCalculateBackoffDelay_LinearStrategy tests linear backoff calculations.
func TestCalculateBackoffDelay_LinearStrategy(t *testing.T) {
	policy := RestartPolicy{
		Strategy:     StrategyLinear,
		InitialDelay: 2 * time.Second,
		MaxDelay:     20 * time.Second,
	}

	// attempt 0: 2s * 1 = 2s
	assert.Equal(t, 2*time.Second, calculateBackoffDelay(0, policy))

	// attempt 1: 2s * 2 = 4s
	assert.Equal(t, 4*time.Second, calculateBackoffDelay(1, policy))

	// attempt 2: 2s * 3 = 6s
	assert.Equal(t, 6*time.Second, calculateBackoffDelay(2, policy))

	// attempt 3: 2s * 4 = 8s
	assert.Equal(t, 8*time.Second, calculateBackoffDelay(3, policy))

	// attempt 9: 2s * 10 = 20s
	assert.Equal(t, 20*time.Second, calculateBackoffDelay(9, policy))

	// attempt 10: 2s * 11 = 22s, capped at 20s
	assert.Equal(t, 20*time.Second, calculateBackoffDelay(10, policy))

	// attempt 100: capped at 20s
	assert.Equal(t, 20*time.Second, calculateBackoffDelay(100, policy))
}

// TestCalculateBackoffDelay_ConstantStrategy tests constant backoff calculations.
func TestCalculateBackoffDelay_ConstantStrategy(t *testing.T) {
	policy := RestartPolicy{
		Strategy:     StrategyConstant,
		InitialDelay: 5 * time.Second,
		MaxDelay:     10 * time.Second,
	}

	// All attempts return constant delay
	assert.Equal(t, 5*time.Second, calculateBackoffDelay(0, policy))
	assert.Equal(t, 5*time.Second, calculateBackoffDelay(1, policy))
	assert.Equal(t, 5*time.Second, calculateBackoffDelay(2, policy))
	assert.Equal(t, 5*time.Second, calculateBackoffDelay(10, policy))
	assert.Equal(t, 5*time.Second, calculateBackoffDelay(100, policy))
}

// TestCalculateBackoffDelay_NoMaxDelay tests behavior when MaxDelay is not set.
func TestCalculateBackoffDelay_NoMaxDelay(t *testing.T) {
	policy := RestartPolicy{
		Strategy:     StrategyExponential,
		InitialDelay: 1 * time.Second,
		MaxDelay:     0, // No cap
	}

	// Large attempts should not be capped
	// attempt 10: 1s * 2^10 = 1024s
	assert.Equal(t, 1024*time.Second, calculateBackoffDelay(10, policy))
}

// TestCalculateBackoffDelay_NegativeAttempt tests handling of negative attempt numbers.
func TestCalculateBackoffDelay_NegativeAttempt(t *testing.T) {
	policy := RestartPolicy{
		Strategy:     StrategyExponential,
		InitialDelay: 1 * time.Second,
		MaxDelay:     30 * time.Second,
	}

	// Negative attempts should be treated as 0
	assert.Equal(t, 1*time.Second, calculateBackoffDelay(-1, policy))
	assert.Equal(t, 1*time.Second, calculateBackoffDelay(-100, policy))
}

// TestCalculateBackoffDelay_UnknownStrategy tests behavior with unknown strategy.
func TestCalculateBackoffDelay_UnknownStrategy(t *testing.T) {
	policy := RestartPolicy{
		Strategy:     StrategyUnknown,
		InitialDelay: 3 * time.Second,
		MaxDelay:     10 * time.Second,
	}

	// Unknown strategy defaults to constant behavior
	assert.Equal(t, 3*time.Second, calculateBackoffDelay(0, policy))
	assert.Equal(t, 3*time.Second, calculateBackoffDelay(1, policy))
	assert.Equal(t, 3*time.Second, calculateBackoffDelay(10, policy))
}

// TestCalculateBackoffDelay_LargeAttemptExponential tests overflow protection for very large attempts.
func TestCalculateBackoffDelay_LargeAttemptExponential(t *testing.T) {
	policy := RestartPolicy{
		Strategy:     StrategyExponential,
		InitialDelay: 1 * time.Second,
		MaxDelay:     60 * time.Second,
	}

	// Very large attempt numbers should be capped by MaxDelay
	// This also tests that we don't overflow
	delay := calculateBackoffDelay(1000, policy)
	assert.Equal(t, 60*time.Second, delay)
}

// TestShouldResetRestartCounter_ZeroLastCrashTime tests that counter resets when no previous crash.
func TestShouldResetRestartCounter_ZeroLastCrashTime(t *testing.T) {
	var zeroTime time.Time
	resetWindow := 5 * time.Minute

	// Should reset when lastCrashTime is zero
	assert.True(t, shouldResetRestartCounter(zeroTime, resetWindow))
}

// TestShouldResetRestartCounter_RecentCrash tests that counter does not reset for recent crashes.
func TestShouldResetRestartCounter_RecentCrash(t *testing.T) {
	lastCrashTime := time.Now().Add(-1 * time.Minute) // 1 minute ago
	resetWindow := 5 * time.Minute

	// Should not reset - not enough time has passed
	assert.False(t, shouldResetRestartCounter(lastCrashTime, resetWindow))
}

// TestShouldResetRestartCounter_OldCrash tests that counter resets after enough time has passed.
func TestShouldResetRestartCounter_OldCrash(t *testing.T) {
	lastCrashTime := time.Now().Add(-10 * time.Minute) // 10 minutes ago
	resetWindow := 5 * time.Minute

	// Should reset - enough time has passed
	assert.True(t, shouldResetRestartCounter(lastCrashTime, resetWindow))
}

// TestShouldResetRestartCounter_ExactlyAtResetWindow tests boundary condition.
func TestShouldResetRestartCounter_ExactlyAtResetWindow(t *testing.T) {
	resetWindow := 5 * time.Minute
	lastCrashTime := time.Now().Add(-resetWindow)

	// Should reset when exactly at reset window (>=)
	assert.True(t, shouldResetRestartCounter(lastCrashTime, resetWindow))
}

// TestShouldResetRestartCounter_ZeroResetWindow tests behavior when reset window is zero.
func TestShouldResetRestartCounter_ZeroResetWindow(t *testing.T) {
	lastCrashTime := time.Now().Add(-10 * time.Minute)
	resetWindow := time.Duration(0)

	// Should not reset when reset window is 0 or negative
	assert.False(t, shouldResetRestartCounter(lastCrashTime, resetWindow))
}

// TestShouldResetRestartCounter_NegativeResetWindow tests behavior when reset window is negative.
func TestShouldResetRestartCounter_NegativeResetWindow(t *testing.T) {
	lastCrashTime := time.Now().Add(-10 * time.Minute)
	resetWindow := -5 * time.Minute

	// Should not reset when reset window is negative
	assert.False(t, shouldResetRestartCounter(lastCrashTime, resetWindow))
}
