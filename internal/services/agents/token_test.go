package agents_test

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/sxwebdev/sentinel/internal/services/agents"
)

func TestNewAgentToken(t *testing.T) {
	agentID := "01K5XS2JAQR651FDX9FSYGYSDE"
	token, secret, tokenHint, err := agents.NewAgentToken(agentID)
	require.NoError(t, err)

	require.NotEmpty(t, token)
	require.NotEmpty(t, secret)
	require.NotEmpty(t, tokenHint)

	fmt.Println("Token:", token)
	fmt.Println("Secret:", secret)
	fmt.Println("Token Hint:", tokenHint)

	hashedSecret, err := agents.HashSecretArgon2id(secret, agents.DefaultArgon2)
	require.NoError(t, err)
	require.NotEmpty(t, hashedSecret)

	fmt.Println("Hashed Secret:", hashedSecret)

	valid, err := agents.VerifySecret(secret, hashedSecret)
	require.NoError(t, err)
	require.True(t, valid)
}
