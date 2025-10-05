package auth

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/sxwebdev/tokenmanager"
)

type DeviceType string

const (
	DeviceTypeUnknown DeviceType = "unknown"
	DeviceTypeWeb     DeviceType = "web"
)

type DeviceInfo struct {
	DeviceID    string     `json:"device_id"`
	DeviceType  DeviceType `json:"device_type"`
	DeviceName  string     `json:"device_name"`
	Fingerprint string     `json:"fingerprint"`
	OSVersion   string     `json:"os_version"`
	AppVersion  string     `json:"app_version"`
}

// SessionData holds additional session data.
type SessionData struct {
	DeviceInfo DeviceInfo `json:"device_info"`
	IP         string     `json:"ip"`
	Country    string     `json:"country"`
	CreatedAt  time.Time  `json:"created_at"`
}

// SessionPair holds both access and refresh tokens for a session.
type SessionPair struct {
	AccessToken    string      `json:"access_token"`
	RefreshToken   string      `json:"refresh_token"`
	AdditionalData SessionData `json:"additional_data"`
}

// AuthorizationResponse is returned after a successful login.
type AuthorizationResponse struct {
	SessionPair      SessionPair                    `json:"session_pair"`
	AccessTokenData  tokenmanager.Data[SessionData] `json:"access_token_data"`
	RefreshTokenData tokenmanager.Data[SessionData] `json:"refresh_token_data"`
}

// RefreshTokenResponse is returned after a successful token refresh.
type RefreshTokenResponse struct {
	SessionPair      SessionPair                    `json:"session_pair"`
	AccessTokenData  tokenmanager.Data[SessionData] `json:"access_token_data"`
	RefreshTokenData tokenmanager.Data[SessionData] `json:"refresh_token_data"`
}

// Manager provides authentication services using two token managers and a session store.
type Manager struct {
	accessManager   *tokenmanager.Manager[SessionData]
	refreshManager  *tokenmanager.Manager[SessionData]
	refreshDuration time.Duration
	sessionStore    tokenmanager.ITokenStore
}

// NewManager creates a new Auth service.
func NewManager(
	accessSecret, refreshSecret string,
	accessDur, refreshDur time.Duration,
	sessionStore tokenmanager.ITokenStore,
) *Manager {
	return &Manager{
		accessManager:   tokenmanager.New[SessionData](sessionStore, accessSecret, accessDur),
		refreshManager:  tokenmanager.New[SessionData](sessionStore, refreshSecret, refreshDur),
		refreshDuration: refreshDur,
		sessionStore:    sessionStore,
	}
}

const sessionPrefix = "session:"

// getSessionKey returns the composite session key as "session:<userID>:<deviceID>".
func getSessionKey(userID, deviceID string) []byte {
	return fmt.Appendf(nil, "%s%s:%s", sessionPrefix, userID, deviceID)
}

// removeSessionKeyPrefix removes the session prefix from a key.
func removeSessionKeyPrefix(key string) string {
	if after, ok := strings.CutPrefix(key, sessionPrefix); ok {
		return after
	}

	return key
}

// GetNewDeviceID generates a unique device ID for the given user.
func (s *Manager) GetNewDeviceID(ctx context.Context, userID string) (string, error) {
	for {
		idBytes := make([]byte, 16)
		if _, err := rand.Read(idBytes); err != nil {
			return "", err
		}

		deviceID := hex.EncodeToString(idBytes)
		key := getSessionKey(userID, deviceID)

		exists, err := s.sessionStore.Exists(ctx, key)
		if err != nil {
			return "", err
		}

		if !exists {
			return deviceID, nil
		}
	}
}

func (s *Manager) Authorization(ctx context.Context, userID string, sessionData SessionData) (*AuthorizationResponse, error) {
	if sessionData.CreatedAt.IsZero() {
		sessionData.CreatedAt = time.Now()
	}

	if sessionData.DeviceInfo.DeviceID != "" { //nolint:nestif
		existsSession, err := s.sessionStore.Exists(ctx, getSessionKey(userID, sessionData.DeviceInfo.DeviceID))
		if err != nil {
			return nil, fmt.Errorf("failed to check session existence: %w", err)
		}

		if !existsSession {
			newDeviceID, err := s.GetNewDeviceID(ctx, userID)
			if err != nil {
				return nil, fmt.Errorf("failed to generate device ID: %w", err)
			}

			sessionData.DeviceInfo.DeviceID = newDeviceID
		}
	} else {
		newDeviceID, err := s.GetNewDeviceID(ctx, userID)
		if err != nil {
			return nil, fmt.Errorf("failed to generate device ID: %w", err)
		}

		sessionData.DeviceInfo.DeviceID = newDeviceID
	}

	if err := s.revokeSessionTokens(ctx, userID, sessionData.DeviceInfo.DeviceID); err != nil {
		return nil, fmt.Errorf("failed to revoke existing session: %w", err)
	}

	accessToken, accessTokenData, err := s.accessManager.CreateToken(ctx, userID, sessionData, tokenmanager.AccessTokenType)
	if err != nil {
		return nil, fmt.Errorf("failed to create access token: %w", err)
	}

	refreshToken, refreshTokenData, err := s.refreshManager.CreateToken(ctx, userID, sessionData, tokenmanager.RefreshTokenType)
	if err != nil {
		err := s.accessManager.RevokeToken(ctx, accessToken)
		if err != nil {
			return nil, fmt.Errorf("failed to revoke access token: %w", err)
		}

		return nil, fmt.Errorf("failed to create refresh token: %w", err)
	}

	pair := SessionPair{
		AccessToken:    accessToken,
		RefreshToken:   refreshToken,
		AdditionalData: sessionData,
	}

	if err := s.setSessionPair(ctx, userID, sessionData.DeviceInfo.DeviceID, pair); err != nil {
		return nil, fmt.Errorf("failed to store session: %w", err)
	}

	return &AuthorizationResponse{
		SessionPair:      pair,
		AccessTokenData:  accessTokenData,
		RefreshTokenData: refreshTokenData,
	}, nil
}

// RefreshToken validates a refresh token, revokes the session, and issues new tokens.
// If the session is not found, a new device ID is generated.
func (s *Manager) RefreshToken(ctx context.Context, refreshToken string) (*RefreshTokenResponse, error) {
	if refreshToken == "" {
		return nil, ErrEmptyRefreshToken
	}

	td, valid := s.refreshManager.ValidateToken(ctx, refreshToken, tokenmanager.RefreshTokenType)
	if !valid {
		return nil, errors.New("invalid or expired refresh token")
	}

	userID := td.UserID
	deviceID := td.AdditionalData.DeviceInfo.DeviceID

	existsSession, err := s.sessionStore.Exists(ctx, getSessionKey(userID, deviceID))
	if err != nil {
		return nil, fmt.Errorf("failed to check session existence: %w", err)
	}

	if !existsSession {
		return nil, errors.New("session not found")
	}

	if err := s.revokeSessionTokens(ctx, userID, deviceID); err != nil {
		return nil, fmt.Errorf("failed to revoke session tokens: %w", err)
	}

	accessToken, accessTokenData, err := s.accessManager.CreateToken(ctx, userID, td.AdditionalData, tokenmanager.AccessTokenType)
	if err != nil {
		return nil, fmt.Errorf("failed to create new access token: %w", err)
	}

	newRefreshToken, refreshTokenData, err := s.refreshManager.CreateToken(ctx, userID, td.AdditionalData, tokenmanager.RefreshTokenType)
	if err != nil {
		err := s.accessManager.RevokeToken(ctx, accessToken)
		if err != nil {
			return nil, fmt.Errorf("failed to revoke new access token: %w", err)
		}

		return nil, fmt.Errorf("failed to create new refresh token: %w", err)
	}

	pair := SessionPair{
		AccessToken:    accessToken,
		RefreshToken:   newRefreshToken,
		AdditionalData: td.AdditionalData,
	}

	if err := s.setSessionPair(ctx, userID, deviceID, pair); err != nil {
		return nil, fmt.Errorf("failed to update session: %w", err)
	}

	return &RefreshTokenResponse{
		SessionPair:      pair,
		AccessTokenData:  accessTokenData,
		RefreshTokenData: refreshTokenData,
	}, nil
}

// Authenticate validates an access token and returns its TokenData.
func (s *Manager) Authenticate(ctx context.Context, accessToken string) (*tokenmanager.Data[SessionData], error) {
	if accessToken == "" {
		return nil, ErrEmptyAccessToken
	}

	td, valid := s.accessManager.ValidateToken(ctx, accessToken, tokenmanager.AccessTokenType)
	if !valid {
		return nil, errors.New("invalid or expired access token")
	}

	return td, nil
}

// Logout revokes the session associated with the provided access token.
func (s *Manager) Logout(ctx context.Context, accessToken string) error {
	if accessToken == "" {
		return ErrEmptyAccessToken
	}

	td, valid := s.accessManager.ValidateToken(ctx, accessToken, tokenmanager.AccessTokenType)
	if !valid {
		return errors.New("invalid or expired access token")
	}

	userID := td.UserID
	deviceID := td.AdditionalData.DeviceInfo.DeviceID

	return s.revokeSessionTokens(ctx, userID, deviceID)
}

// DeleteSession revokes the session for the given deviceID if it belongs to the user.
func (s *Manager) DeleteSession(ctx context.Context, userID, deviceID string) error {
	return s.revokeSessionTokens(ctx, userID, deviceID)
}

// GetAllSessions returns all active sessions for the given user.
// It uses the store's Keys method to list all keys starting with "session:<userID>:".
func (s *Manager) GetAllSessions(ctx context.Context, userID string) (map[string]SessionPair, error) {
	prefix := []byte(sessionPrefix + userID + ":")

	keys, err := s.sessionStore.Keys(ctx, prefix)
	if err != nil {
		return nil, err
	}

	sessions := make(map[string]SessionPair)
	for _, key := range keys {
		var pair SessionPair

		err := s.sessionStore.GetFromJSON(ctx, []byte(key), &pair)
		if err != nil {
			continue
		}

		extractedKey := removeSessionKeyPrefix(key)

		parts := strings.Split(extractedKey, ":")
		if len(parts) == 2 {
			sessions[parts[1]] = pair
		}
	}

	return sessions, nil
}

// TerminateAllSessions revokes all sessions for the given user except the current one.
func (s *Manager) TerminateAllSessions(ctx context.Context, userID, deviceID string) error {
	sessionKeys, err := s.sessionStore.Keys(ctx, []byte(sessionPrefix+userID+":"))
	if err != nil {
		return err
	}

	sessionKey := string(getSessionKey(userID, deviceID))

	for _, key := range sessionKeys {
		if key == sessionKey {
			continue
		}

		err := s.revokeSessionTokensByKey(ctx, []byte(key))
		if err != nil {
			return fmt.Errorf("failed to delete session: %w", err)
		}
	}

	return nil
}

// UpdateAdditionalData updates the additional data in the session for the given access token.
func (s *Manager) UpdateAdditionalData(ctx context.Context, accessToken string, newAdditionalData SessionData) error {
	if accessToken == "" {
		return ErrEmptyAccessToken
	}

	tokenData, ok := s.accessManager.ValidateToken(ctx, accessToken, tokenmanager.AccessTokenType)
	if !ok {
		return errors.New("failed to validate access token")
	}

	if newAdditionalData.DeviceInfo.DeviceID != tokenData.AdditionalData.DeviceInfo.DeviceID {
		return fmt.Errorf("device ID mismatch: %s != %s", newAdditionalData.DeviceInfo.DeviceID, tokenData.AdditionalData.DeviceInfo.DeviceID)
	}

	// Get the existing session pair from the session store.
	sessionPair, err := s.getSessionPair(ctx, tokenData.UserID, tokenData.AdditionalData.DeviceInfo.DeviceID)
	if err != nil {
		return fmt.Errorf("failed to get session pair: %w", err)
	}

	tokenData.AdditionalData = newAdditionalData
	sessionPair.AdditionalData = newAdditionalData

	// Update the session store with the new additional data.
	if err := s.setSessionPair(ctx, tokenData.UserID, tokenData.AdditionalData.DeviceInfo.DeviceID, sessionPair); err != nil {
		return fmt.Errorf("failed to update additional data: %w", err)
	}

	// Update the additional data in both access and refresh token managers.
	if err := s.accessManager.UpdateAdditionalData(ctx, sessionPair.AccessToken, newAdditionalData); err != nil {
		return fmt.Errorf("failed to update access token additional data: %w", err)
	}

	// Update the additional data in the refresh token manager.
	if err := s.refreshManager.UpdateAdditionalData(ctx, sessionPair.RefreshToken, newAdditionalData); err != nil {
		return fmt.Errorf("failed to update refresh token additional data: %w", err)
	}

	return nil
}

// setSessionPair stores the session pair in the session store.
func (s *Manager) setSessionPair(ctx context.Context, userID, deviceID string, pair SessionPair) error {
	key := getSessionKey(userID, deviceID)
	return s.sessionStore.SetJSON(ctx, key, pair, s.refreshDuration)
}

// getSessionPair retrieves the session pair for the given user and device ID.
func (s *Manager) getSessionPair(ctx context.Context, userID, deviceID string) (SessionPair, error) {
	key := getSessionKey(userID, deviceID)

	var sessionPair SessionPair

	err := s.sessionStore.GetFromJSON(ctx, key, &sessionPair)
	if err != nil {
		return SessionPair{}, fmt.Errorf("failed to get session pair: %w", err)
	}

	return sessionPair, nil
}

// revokeSessionTokens revokes (deletes) the session for the given user and device.
func (s *Manager) revokeSessionTokens(ctx context.Context, userID, deviceID string) error {
	return s.revokeSessionTokensByKey(ctx, getSessionKey(userID, deviceID))
}

func (s *Manager) revokeSessionTokensByKey(ctx context.Context, key []byte) error {
	ok, err := s.sessionStore.Exists(ctx, key)
	if err != nil {
		return fmt.Errorf("failed to check session existence: %w", err)
	}

	if !ok {
		return nil
	}

	var session SessionPair
	if err := s.sessionStore.GetFromJSON(ctx, key, &session); err != nil {
		return fmt.Errorf("failed to get session: %w", err)
	}

	if session.AccessToken != "" {
		err := s.accessManager.RevokeToken(ctx, session.AccessToken)
		if err != nil {
			return fmt.Errorf("failed to revoke access token: %w", err)
		}
	}

	if session.RefreshToken != "" {
		err := s.refreshManager.RevokeToken(ctx, session.RefreshToken)
		if err != nil {
			return fmt.Errorf("failed to revoke refresh token: %w", err)
		}
	}

	return s.sessionStore.Delete(ctx, key)
}
