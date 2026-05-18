package handlers

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/bimbambaap/bimbambaap/database"
	"github.com/bimbambaap/bimbambaap/models"
)

// Controleer of huidige user admin is
func isAdmin(c *gin.Context) bool {
	userID := c.GetInt("user_id")
	var admin bool
	database.DB.QueryRow("SELECT is_admin FROM users WHERE id = $1", userID).Scan(&admin)
	return admin
}

// GET /admin/users — alle accounts
func AdminGetUsers(c *gin.Context) {
	if !isAdmin(c) {
		c.JSON(http.StatusForbidden, gin.H{"error": "Geen rechten"})
		return
	}

	rows, err := database.DB.Query(
		`SELECT id, username, email, is_admin, created_at FROM users ORDER BY created_at DESC`,
	)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Kon gebruikers niet laden"})
		return
	}
	defer rows.Close()

	users := []models.User{}
	for rows.Next() {
		var u models.User
		rows.Scan(&u.ID, &u.Username, &u.Email, &u.IsAdmin, &u.CreatedAt)
		users = append(users, u)
	}
	c.JSON(http.StatusOK, gin.H{"users": users})
}

// DELETE /admin/users/:id — verwijder account + alle posts
func AdminDeleteUser(c *gin.Context) {
	if !isAdmin(c) {
		c.JSON(http.StatusForbidden, gin.H{"error": "Geen rechten"})
		return
	}

	targetID, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Ongeldig ID"})
		return
	}

	// Voorkom dat admin zichzelf verwijdert
	if targetID == c.GetInt("user_id") {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Je kan jezelf niet verwijderen"})
		return
	}

	_, err = database.DB.Exec("DELETE FROM users WHERE id = $1", targetID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Kon account niet verwijderen"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Account verwijderd"})
}

// GET /admin/posts — alle posts
func AdminGetPosts(c *gin.Context) {
	if !isAdmin(c) {
		c.JSON(http.StatusForbidden, gin.H{"error": "Geen rechten"})
		return
	}

	rows, err := database.DB.Query(
		`SELECT p.id, p.user_id, p.caption, p.image_url, p.created_at,
		        u.id, u.username
		 FROM posts p
		 JOIN users u ON u.id = p.user_id
		 ORDER BY p.created_at DESC
		 LIMIT 100`,
	)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Kon posts niet laden"})
		return
	}
	defer rows.Close()

	posts := []models.Post{}
	for rows.Next() {
		var p models.Post
		var author models.User
		rows.Scan(&p.ID, &p.UserID, &p.Caption, &p.ImageURL, &p.CreatedAt,
			&author.ID, &author.Username)
		p.Author = &author
		posts = append(posts, p)
	}
	c.JSON(http.StatusOK, gin.H{"posts": posts})
}
