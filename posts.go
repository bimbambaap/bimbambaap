package handlers

import (
	"context"
	"net/http"
	"os"
	"strconv"

	"github.com/cloudinary/cloudinary-go/v2"
	"github.com/cloudinary/cloudinary-go/v2/api/uploader"
	"github.com/gin-gonic/gin"

	"github.com/yourusername/api/database"
	"github.com/yourusername/api/models"
)

// GetFeed — publieke feed, nieuwste posts eerst
func GetFeed(c *gin.Context) {
	// Pagination via ?page=1&limit=20
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "20"))
	if page < 1 { page = 1 }
	if limit < 1 || limit > 50 { limit = 20 }
	offset := (page - 1) * limit

	rows, err := database.DB.Query(
		`SELECT p.id, p.user_id, p.caption, p.image_url, p.created_at,
		        u.id, u.username, u.avatar_url
		 FROM posts p
		 JOIN users u ON u.id = p.user_id
		 ORDER BY p.created_at DESC
		 LIMIT $1 OFFSET $2`,
		limit, offset,
	)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Kon feed niet laden"})
		return
	}
	defer rows.Close()

	posts := []models.Post{}
	for rows.Next() {
		var p models.Post
		var author models.User
		err := rows.Scan(
			&p.ID, &p.UserID, &p.Caption, &p.ImageURL, &p.CreatedAt,
			&author.ID, &author.Username, &author.AvatarURL,
		)
		if err != nil {
			continue
		}
		p.Author = &author
		posts = append(posts, p)
	}

	c.JSON(http.StatusOK, gin.H{
		"posts": posts,
		"page":  page,
		"limit": limit,
	})
}

// CreatePost — upload foto naar Cloudinary, sla URL op in database
func CreatePost(c *gin.Context) {
	userID := c.GetInt("user_id")

	// Haal foto op uit multipart form
	file, fileHeader, err := c.Request.FormFile("image")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Geen foto meegestuurd"})
		return
	}
	defer file.Close()

	// Max 10MB
	if fileHeader.Size > 10*1024*1024 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Foto mag maximaal 10MB zijn"})
		return
	}

	// Upload naar Cloudinary
	imageURL, err := uploadToCloudinary(file)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Foto upload mislukt"})
		return
	}

	caption := c.Request.FormValue("caption")

	// Sla post op
	var post models.Post
	err = database.DB.QueryRow(
		`INSERT INTO posts (user_id, caption, image_url)
		 VALUES ($1, $2, $3)
		 RETURNING id, user_id, caption, image_url, created_at`,
		userID, caption, imageURL,
	).Scan(&post.ID, &post.UserID, &post.Caption, &post.ImageURL, &post.CreatedAt)

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Kon post niet opslaan"})
		return
	}

	c.JSON(http.StatusCreated, post)
}

// GetPost — haal één post op
func GetPost(c *gin.Context) {
	postID, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Ongeldig post ID"})
		return
	}

	var p models.Post
	var author models.User
	err = database.DB.QueryRow(
		`SELECT p.id, p.user_id, p.caption, p.image_url, p.created_at,
		        u.id, u.username, u.avatar_url
		 FROM posts p
		 JOIN users u ON u.id = p.user_id
		 WHERE p.id = $1`,
		postID,
	).Scan(
		&p.ID, &p.UserID, &p.Caption, &p.ImageURL, &p.CreatedAt,
		&author.ID, &author.Username, &author.AvatarURL,
	)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Post niet gevonden"})
		return
	}
	p.Author = &author

	c.JSON(http.StatusOK, p)
}

// DeletePost — alleen de eigenaar mag verwijderen
func DeletePost(c *gin.Context) {
	userID := c.GetInt("user_id")
	postID, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Ongeldig post ID"})
		return
	}

	result, err := database.DB.Exec(
		`DELETE FROM posts WHERE id = $1 AND user_id = $2`,
		postID, userID,
	)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Kon post niet verwijderen"})
		return
	}

	rows, _ := result.RowsAffected()
	if rows == 0 {
		c.JSON(http.StatusForbidden, gin.H{"error": "Post niet gevonden of geen rechten"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Post verwijderd"})
}

// GetUserPosts — alle posts van een specifieke gebruiker
func GetUserPosts(c *gin.Context) {
	username := c.Param("username")

	rows, err := database.DB.Query(
		`SELECT p.id, p.user_id, p.caption, p.image_url, p.created_at
		 FROM posts p
		 JOIN users u ON u.id = p.user_id
		 WHERE u.username = $1
		 ORDER BY p.created_at DESC`,
		username,
	)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Kon posts niet laden"})
		return
	}
	defer rows.Close()

	posts := []models.Post{}
	for rows.Next() {
		var p models.Post
		rows.Scan(&p.ID, &p.UserID, &p.Caption, &p.ImageURL, &p.CreatedAt)
		posts = append(posts, p)
	}

	c.JSON(http.StatusOK, gin.H{"posts": posts})
}

// Helper: upload bestand naar Cloudinary en geef URL terug
func uploadToCloudinary(file interface{}) (string, error) {
	cld, err := cloudinary.NewFromParams(
		os.Getenv("CLOUDINARY_CLOUD_NAME"),
		os.Getenv("CLOUDINARY_API_KEY"),
		os.Getenv("CLOUDINARY_API_SECRET"),
	)
	if err != nil {
		return "", err
	}

	ctx := context.Background()
	result, err := cld.Upload.Upload(ctx, file, uploader.UploadParams{
		Folder: "posts",
	})
	if err != nil {
		return "", err
	}

	return result.SecureURL, nil
}
