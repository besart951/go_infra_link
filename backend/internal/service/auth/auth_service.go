package auth

import (
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"time"

	"github.com/besart951/go_infra_link/backend/internal/domain"
	domainAuth "github.com/besart951/go_infra_link/backend/internal/domain/auth"
	domainUser "github.com/besart951/go_infra_link/backend/internal/domain/user"
	passwordsvc "github.com/besart951/go_infra_link/backend/internal/service/password"
	"github.com/google/uuid"
)

type Service struct {
	jwtService       JWTService
	userRepo         domainUser.UserRepository
	userEmailRepo    domainUser.UserEmailRepository
	refreshTokenRepo domainAuth.RefreshTokenRepository
	passwordService  passwordsvc.Service
	accessTokenTTL   time.Duration
	refreshTokenTTL  time.Duration
	issuer           string
}

func NewService(
	jwtService JWTService,
	userRepo domainUser.UserRepository,
	userEmailRepo domainUser.UserEmailRepository,
	refreshTokenRepo domainAuth.RefreshTokenRepository,
	passwordService passwordsvc.Service,
	accessTokenTTL time.Duration,
	refreshTokenTTL time.Duration,
	issuer string,
) *Service {
	return &Service{
		jwtService:       jwtService,
		userRepo:         userRepo,
		userEmailRepo:    userEmailRepo,
		refreshTokenRepo: refreshTokenRepo,
		passwordService:  passwordService,
		accessTokenTTL:   accessTokenTTL,
		refreshTokenTTL:  refreshTokenTTL,
		issuer:           issuer,
	}
}

type LoginResult struct {
	User               *domainUser.User
	AccessToken        string
	AccessTokenExpiry  time.Time
	RefreshToken       string
	RefreshTokenExpiry time.Time
	CSRFFriendlyToken  string
}

func (s *Service) Login(email, password string, userAgent, ip *string) (*LoginResult, error) {
	usr, err := s.userEmailRepo.GetByEmail(email)
	if err != nil {
		if errors.Is(err, domain.ErrNotFound) {
			return nil, domainAuth.ErrInvalidCredentials
		}
		return nil, err
	}
	if !usr.IsActive {
		return nil, domainAuth.ErrInvalidCredentials
	}

	if err := s.passwordService.Compare(usr.Password, password); err != nil {
		return nil, domainAuth.ErrInvalidCredentials
	}

	return s.issueTokens(usr, userAgent, ip)
}

func (s *Service) Refresh(refreshToken string, userAgent, ip *string) (*LoginResult, error) {
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

func (s *Service) issueTokens(usr *domainUser.User, userAgent, ip *string) (*LoginResult, error) {
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

	return &LoginResult{
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
