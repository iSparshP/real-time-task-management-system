package auth

import (
	"errors"
	"fmt"
	"time"
	"unicode"

	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

var (
	ErrInvalidCredentials = errors.New("invalid credentials")
	ErrUserExists         = errors.New("user already exists")
	ErrTokenExpired       = errors.New("token has expired")
	ErrInvalidToken       = errors.New("invalid token")
	ErrUserNotFound       = errors.New("user not found")
)

type Service struct {
	db        *gorm.DB
	jwtSecret []byte
	config    Config
}

func NewService(db *gorm.DB, config Config) *Service {
	return &Service{
		db:        db,
		jwtSecret: []byte(config.JWTSecret),
		config:    config,
	}
}

func (s *Service) Register(req RegisterRequest) (*AuthResponse, error) {
	// Validate password strength
	if err := validatePassword(req.Password); err != nil {
		return nil, err
	}

	// Check if user exists
	var existingUser User
	if result := s.db.Where("email = ?", req.Email).First(&existingUser); result.Error == nil {
		return nil, ErrUserExists
	}

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		return nil, err
	}

	user := &User{
		Email:     req.Email,
		Password:  string(hashedPassword),
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	// Save user to DB
	if err := s.db.Create(user).Error; err != nil {
		return nil, err
	}

	token, err := s.generateToken(user)
	if err != nil {
		return nil, err
	}

	return &AuthResponse{
		Token: token,
		User:  *user,
	}, nil
}

func (s *Service) Login(req LoginRequest) (*AuthResponse, error) {
	var user User
	if err := s.db.Where("email = ?", req.Email).First(&user).Error; err != nil {
		return nil, ErrInvalidCredentials
	}

	// Compare passwords
	if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(req.Password)); err != nil {
		return nil, ErrInvalidCredentials
	}

	token, err := s.generateToken(&user)
	if err != nil {
		return nil, err
	}

	return &AuthResponse{
		Token: token,
		User:  user,
	}, nil
}

func (s *Service) generateToken(user *User) (string, error) {
	claims := jwt.MapClaims{
		"user_id": user.ID,
		"email":   user.Email,
		"exp":     time.Now().Add(time.Hour * 24).Unix(), // 24 hour expiry
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(s.jwtSecret)
}

func (s *Service) ValidateToken(tokenString string) (string, error) {
	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		// Verify signing method
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return s.jwtSecret, nil
	})

	if err != nil {
		return "", ErrInvalidCredentials
	}

	if !token.Valid {
		return "", ErrInvalidCredentials
	}

	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok {
		return "", ErrInvalidCredentials
	}

	// Check token expiration
	if exp, ok := claims["exp"].(float64); ok {
		if time.Now().Unix() > int64(exp) {
			return "", ErrInvalidCredentials
		}
	}

	userID, ok := claims["user_id"].(string)
	if !ok {
		return "", ErrInvalidCredentials
	}

	return userID, nil
}

func (s *Service) RefreshToken(refreshToken string) (*AuthResponse, error) {
	userID, err := s.ValidateToken(refreshToken)
	if err != nil {
		return nil, err
	}

	var user User
	if err := s.db.First(&user, "id = ?", userID).Error; err != nil {
		return nil, ErrInvalidCredentials
	}

	token, err := s.generateToken(&user)
	if err != nil {
		return nil, err
	}

	return &AuthResponse{
		Token: token,
		User:  user,
	}, nil
}

func validatePassword(password string) error {
	// Minimum length
	if len(password) < 8 {
		return errors.New("password must be at least 8 characters")
	}

	// Check for number
	hasNumber := false
	for _, char := range password {
		if unicode.IsNumber(char) {
			hasNumber = true
			break
		}
	}
	if !hasNumber {
		return errors.New("password must contain at least one number")
	}

	// Add more validation as needed
	return nil
}
