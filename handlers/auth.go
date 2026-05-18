package handlers

import (
	"net/http"
	"os"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/crypto/bcrypt"

	"github.com/bimbambaap/bimbambaap/database"
	"github.com/bimbambaap/bimbambaap/middleware"
	"github.com/bimbambaap/bimbambaap/models"
)

func Register(c *gin.Context) {
	var req models.RegisterRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Hash wachtwoord — sla NOOIT plaintext op
	hashed, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Server fout"})
		return
	}

	var user models.User
	err = database.DB.QueryRow(
		`INSERT INTO users (username, email, password)
		 VALUES ($1, $2, $3)
		 RETURNING id, username, email, created_at`,
		req.Username, req.Email, string(hashed),
	).Scan(&user.ID, &user.Username, &user.Email, &user.CreatedAt)

	if err != nil {
		// Controleer op duplicate username/email
		if isDuplicate(err) {
			c.JSON(http.StatusConflict, gin.H{"error": "Username of email al in gebruik"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Kon account niet aanmaken"})
		return
	}

	token, err := generateToken(user.ID, user.Username)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Kon token niet aanmaken"})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"token": token,
		"user":  user,
	})
}

func Login(c *gin.Context) {
	var req models.LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	var user models.User
	err := database.DB.QueryRow(
		`SELECT id, username, email, password, avatar_url, bio, is_admin, created_at
		 FROM users WHERE email = $1`,
		req.Email,
	).Scan(&user.ID, &user.Username, &user.Email, &user.Password,
		&user.AvatarURL, &user.Bio, &user.IsAdmin, &user.CreatedAt)

	if err != nil {
		// Geef altijd dezelfde foutmelding — verklap niet of email bestaat
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Onjuist email of wachtwoord"})
		return
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(req.Password)); err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Onjuist email of wachtwoord"})
		return
	}

	token, err := generateToken(user.ID, user.Username)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Kon token niet aanmaken"})
		return
	}

	user.Password = "" // Verwijder wachtwoord uit response
	c.JSON(http.StatusOK, gin.H{
		"token": token,
		"user":  user,
	})
}

func GetMe(c *gin.Context) {
	userID := c.GetInt("user_id")

	var user models.User
	err := database.DB.QueryRow(
		`SELECT id, username, email, avatar_url, bio, is_admin, created_at
		 FROM users WHERE id = $1`,
		userID,
	).Scan(&user.ID, &user.Username, &user.Email,
		&user.AvatarURL, &user.Bio, &user.IsAdmin, &user.CreatedAt)

	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Gebruiker niet gevonden"})
		return
	}

	c.JSON(http.StatusOK, user)
}

func UpdateProfile(c *gin.Context) {
	userID := c.GetInt("user_id")

	var req models.UpdateProfileRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	_, err := database.DB.Exec(
		`UPDATE users SET bio = $1, avatar_url = $2 WHERE id = $3`,
		req.Bio, req.AvatarURL, userID,
	)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Kon profiel niet updaten"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Profiel bijgewerkt"})
}

func SetAdmin(c *gin.Context) {
	// Controleer of huidige user admin is
	userID := c.GetInt("user_id")
	var selfAdmin bool
	database.DB.QueryRow("SELECT is_admin FROM users WHERE id = $1", userID).Scan(&selfAdmin)
	if !selfAdmin {
		c.JSON(http.StatusForbidden, gin.H{"error": "Geen rechten"})
		return
	}

	targetID := c.Param("id")

	var currentAdmin bool
	err := database.DB.QueryRow("SELECT is_admin FROM users WHERE id = $1", targetID).Scan(&currentAdmin)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Gebruiker niet gevonden"})
		return
	}

	_, err = database.DB.Exec("UPDATE users SET is_admin = $1 WHERE id = $2", !currentAdmin, targetID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Kon admin status niet wijzigen"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"is_admin": !currentAdmin})
}

// Helper: maak JWT token aan
func generateToken(userID int, username string) (string, error) {
	secret := os.Getenv("JWT_SECRET")
	claims := middleware.Claims{
		UserID:   userID,
		Username: username,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(30 * 24 * time.Hour)), // 30 dagen
			IssuedAt:  jwt.NewNumericDate(time.Now()),
		},
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(secret))
}

// Helper: check of postgres duplicate key error
func isDuplicate(err error) bool {
	return err != nil && len(err.Error()) > 0 &&
		(contains(err.Error(), "duplicate key") || contains(err.Error(), "unique constraint"))
}

func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr ||
		len(s) > 0 && containsStr(s, substr))
}

func containsStr(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
