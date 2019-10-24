package platform_test

import (
	"testing"

	"github.com/2beens/ispend/internal/platform"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestGenerateRandomString(t *testing.T) {
	randStringLen1 := 10
	randString1 := platform.GenerateRandomString(randStringLen1)
	assert.Len(t, randString1, randStringLen1)
	randStringLen2 := 45
	randString2 := platform.GenerateRandomString(randStringLen2)
	assert.Len(t, randString2, randStringLen2)
}

func TestHashPassword(t *testing.T) {
	raw := "myinspirationsucks"
	hashedPass, err := platform.HashPassword(raw)
	require.NoError(t, err)
	assert.Len(t, hashedPass, 60)

	passOk := platform.CheckPasswordHash(raw, hashedPass)
	assert.True(t, passOk)

	rawFraudPass := "myinspirationstillsucks"
	passOk = platform.CheckPasswordHash(rawFraudPass, hashedPass)
	assert.False(t, passOk)
}
