package auth

import (
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"strings"
	"time"

	"github.com/besart951/go_infra_link/backend/internal/domain"
	domainAuth "github.com/besart951/go_infra_link/backend/internal/domain/auth"
	domainUser "github.com/besart951/go_infra_link/backend/internal/domain/user"
	"github.com/google/uuid"
)

type Service struct {
	jwtService        domainAuth.TokenService
	userRepo          domainUser.UserRepository
	userEmailRepo     domainUser.UserEmailRepository
	refreshTokenRepo  domainAuth.RefreshTokenRepository
	loginAttemptRepo  domainAuth.LoginAttemptRepository
	passwordResetRepo domainAuth.PasswordResetTokenRepository
	passwordHasher    domainUser.PasswordHasher
	accessTokenTTL    time.Duration
	refreshTokenTTL   time.Duration
	issuer            string
}

func NewService(
	jwtService domainAuth.TokenService,
	userRepo domainUser.UserRepository,
	userEmailRepo domainUser.UserEmailRepository,
	refreshTokenRepo domainAuth.RefreshTokenRepository,
	loginAttemptRepo domainAuth.LoginAttemptRepository,
	passwordResetRepo domainAuth.PasswordResetTokenRepository,
	passwordHasher domainUser.PasswordHasher,
	accessTokenTTL time.Duration,
	refreshTokenTTL time.Duration,
	issuer string,
) *Service {
	return &Service{
		jwtService:        jwtService,
		userRepo:          userRepo,
		userEmailRepo:     userEmailRepo,
		refreshTokenRepo:  refreshTokenRepo,
		loginAttemptRepo:  loginAttemptRepo,
		passwordResetRepo: passwordResetRepo,
		passwordHasher:    passwordHasher,
		accessTokenTTL:    accessTokenTTL,
		refreshTokenTTL:   refreshTokenTTL,
		issuer:            issuer,
	}
}

const defaultPasswordResetTTL = time.Hour

func (s *Service) Login(email, password string, userAgent, ip *string) (*domainAuth.LoginResult, error) {
	email = strings.TrimSpace(strings.ToLower(email))

	usr, err := s.userEmailRepo.GetByEmail(email)
	if err != nil {
		if errors.Is(err, domain.ErrNotFound) {
			s.auditLoginAttempt(nil, &email, ip, userAgent, false, strPtr("invalid_credentials"))
			return nil, domainAuth.ErrInvalidCredentials
		}
		return nil, err
	}

	if usr.DisabledAt != nil || !usr.IsActive {
		s.auditLoginAttempt(&usr.ID, &email, ip, userAgent, false, strPtr("account_disabled"))
		return nil, domainAuth.ErrAccountDisabled
	}
	if usr.LockedUntil != nil && usr.LockedUntil.After(time.Now().UTC()) {
		s.auditLoginAttempt(&usr.ID, &email, ip, userAgent, false, strPtr("account_locked"))
		return nil, domainAuth.ErrAccountLocked
	}

	if err := s.passwordHasher.Compare(usr.Password, password); err != nil {
		usr.FailedLoginAttempts++
		if err := s.userRepo.Update(usr); err != nil {
			consumeBestEffortError(err)
		}
		s.auditLoginAttempt(&usr.ID, &email, ip, userAgent, false, strPtr("invalid_credentials"))
		return nil, domainAuth.ErrInvalidCredentials
	}

	// Best-effort: reset counters and record last login.
	now := time.Now().UTC()
	usr.FailedLoginAttempts = 0
	usr.LockedUntil = nil
	usr.LastLoginAt = &now
	if err := s.userRepo.Update(usr); err != nil {
		consumeBestEffortError(err)
	}
	s.auditLoginAttempt(&usr.ID, &email, ip, userAgent, true, nil)

	return s.issueTokens(usr, userAgent, ip)
}

func (s *Service) CreatePasswordResetToken(adminID, userID uuid.UUID) (string, time.Time, error) {
	users, err := s.userRepo.GetByIds([]uuid.UUID{userID})
	if err != nil {
		return "", time.Time{}, err
	}
	if len(users) == 0 {
		return "", time.Time{}, domain.ErrNotFound
	}

	salt, err := generateRandomToken()
	if err != nil {
		return "", time.Time{}, err
	}
	raw, err := generateRandomToken()
	if err != nil {
		return "", time.Time{}, err
	}

	h := sha256.Sum256([]byte(salt + raw))
	hash := hex.EncodeToString(h[:])

	expiresAt := time.Now().UTC().Add(defaultPasswordResetTTL)

	rec := &domainAuth.PasswordResetToken{
		UserID:           userID,
		TokenHash:        hash,
		TokenSalt:        salt,
		ExpiresAt:        expiresAt,
		CreatedByAdminID: &adminID,
	}
	if err := s.passwordResetRepo.Create(rec); err != nil {
		return "", time.Time{}, err
	}

	// Token format: <salt>.<raw>
	return salt + "." + raw, expiresAt, nil
}

func (s *Service) ConfirmPasswordReset(token, newPassword string) error {
	parts := strings.Split(token, ".")
	if len(parts) != 2 {
		return domainAuth.ErrPasswordResetTokenInvalid
	}
	salt := strings.TrimSpace(parts[0])
	raw := strings.TrimSpace(parts[1])
	if salt == "" || raw == "" {
		return domainAuth.ErrPasswordResetTokenInvalid
	}
	if _, err := hex.DecodeString(salt); err != nil {
		return domainAuth.ErrPasswordResetTokenInvalid
	}
	if _, err := hex.DecodeString(raw); err != nil {
		return domainAuth.ErrPasswordResetTokenInvalid
	}

	h := sha256.Sum256([]byte(salt + raw))
	hash := hex.EncodeToString(h[:])

	rec, err := s.passwordResetRepo.GetByTokenHash(hash)
	if err != nil {
		return err
	}
	if rec == nil {
		return domainAuth.ErrPasswordResetTokenInvalid
	}
	if rec.UsedAt != nil {
		return domainAuth.ErrPasswordResetTokenUsed
	}
	if time.Now().UTC().After(rec.ExpiresAt) {
		return domainAuth.ErrPasswordResetTokenExpired
	}

	used, err := s.passwordResetRepo.MarkUsedByTokenHash(hash, time.Now().UTC())
	if err != nil {
		return err
	}
	if !used {
		return domainAuth.ErrPasswordResetTokenUsed
	}

	hashedPassword, err := s.passwordHasher.Hash(newPassword)
	if err != nil {
		return domainUser.ErrPasswordHashingFailed
	}

	users, err := s.userRepo.GetByIds([]uuid.UUID{rec.UserID})
	if err != nil {
		return err
	}
	if len(users) == 0 {
		return domainAuth.ErrPasswordResetTokenInvalid
	}
	usr := users[0]
	usr.Password = hashedPassword
	usr.FailedLoginAttempts = 0
	usr.LockedUntil = nil
	return s.userRepo.Update(usr)
}

func (s *Service) ListLoginAttempts(page, limit int, search string) (*domain.PaginatedList[domainAuth.LoginAttempt], error) {
	page, limit = domain.NormalizePagination(page, limit, 20)
	return s.loginAttemptRepo.GetPaginatedList(domain.PaginationParams{Page: page, Limit: limit, Search: search})
}

func (s *Service) Refresh(refreshToken string, userAgent, ip *string) (*domainAuth.LoginResult, error) {
	if refreshToken == "" {
		return nil, domainAuth.ErrInvalidToken
	}

	hash := hashToken(refreshToken)
	rec, err := s.refreshTokenRepo.GetByTokenHash(hash)
	if err != nil {
		return nil, err
	}
	if rec == nil {
		return nil, domainAuth.ErrInvalidToken
	}
	if rec.RevokedAt != nil {
		return nil, domainAuth.ErrTokenRevoked
	}
	if time.Now().UTC().After(rec.ExpiresAt) {
		return nil, domainAuth.ErrTokenExpired
	}

	users, err := s.userRepo.GetByIds([]uuid.UUID{rec.UserID})
	if err != nil {
		return nil, err
	}
	if len(users) == 0 {
		return nil, domainAuth.ErrInvalidToken
	}
	usr := users[0]

	revokedAt := time.Now().UTC()
	if err := s.refreshTokenRepo.RevokeByTokenHash(hash, revokedAt); err != nil {
		return nil, err
	}

	return s.issueTokens(usr, userAgent, ip)
}

func (s *Service) Logout(refreshToken string) error {
	if refreshToken == "" {
		return nil
	}

	hash := hashToken(refreshToken)
	return s.refreshTokenRepo.RevokeByTokenHash(hash, time.Now().UTC())
}

func (s *Service) issueTokens(usr *domainUser.User, userAgent, ip *string) (*domainAuth.LoginResult, error) {
	accessExpiry := time.Now().UTC().Add(s.accessTokenTTL)
	accessToken, err := s.jwtService.CreateAccessToken(usr.ID, accessExpiry)
	if err != nil {
		return nil, err
	}

	refreshToken, refreshHash, err := generateRefreshToken()
	if err != nil {
		return nil, err
	}
	refreshExpiry := time.Now().UTC().Add(s.refreshTokenTTL)

	record := &domainAuth.RefreshToken{
		UserID:      usr.ID,
		TokenHash:   refreshHash,
		ExpiresAt:   refreshExpiry,
		CreatedByIP: ip,
		UserAgent:   userAgent,
	}

	if err := s.refreshTokenRepo.Create(record); err != nil {
		return nil, err
	}

	csrfToken, err := generateRandomToken()
	if err != nil {
		return nil, err
	}

	return &domainAuth.LoginResult{
		User:               usr,
		AccessToken:        accessToken,
		AccessTokenExpiry:  accessExpiry,
		RefreshToken:       refreshToken,
		RefreshTokenExpiry: refreshExpiry,
		CSRFFriendlyToken:  csrfToken,
	}, nil
}

func generateRefreshToken() (string, string, error) {
	raw, err := generateRandomToken()
	if err != nil {
		return "", "", err
	}
	return raw, hashToken(raw), nil
}

func generateRandomToken() (string, error) {
	b := make([]byte, 32)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return hex.EncodeToString(b), nil
}

func hashToken(token string) string {
	h := sha256.Sum256([]byte(token))
	return hex.EncodeToString(h[:])
}

func (s *Service) auditLoginAttempt(userID *uuid.UUID, email, ip, userAgent *string, success bool, failureReason *string) {
	if s.loginAttemptRepo == nil {
		return
	}
	attempt := &domainAuth.LoginAttempt{
		UserID:        userID,
		Email:         email,
		IP:            ip,
		UserAgent:     userAgent,
		Success:       success,
		FailureReason: failureReason,
	}
	if err := s.loginAttemptRepo.Create(attempt); err != nil {
		consumeBestEffortError(err)
	}
}

func strPtr(s string) *string {
	return &s
}

func consumeBestEffortError(_ error) {}
