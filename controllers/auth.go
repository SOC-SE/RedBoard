package controllers

import (
	"errors"
	"net/http"
	"strings"

	"github.com/brian-l-johnson/Redteam-Dashboard-go/v2/models"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"

	"github.com/gin-contrib/sessions"
	"github.com/gin-gonic/gin"
)

type AuthController struct{}

// Login godoc
// @Summary Login
// @Description Login a user
// @Tags auth
// @Accept json
// @Produce json
// @Param login body models.LoginReq true "Login Data"
// @Success 200 {object} map[string]interface{}
// @Router /auth/login [post]
func (a AuthController) Login(c *gin.Context) {
	var lr models.LoginReq
	if err := c.ShouldBindJSON(&lr); err != nil {
		c.IndentedJSON(http.StatusBadRequest, gin.H{"status": "error", "message": "invalid request"})
		return
	}

	db := models.GetDB()
	var user models.User
	result := db.First(&user, "name = ?", lr.User)

	if result.Error != nil {
		// Don't reveal whether user exists
		c.IndentedJSON(http.StatusUnauthorized, gin.H{"status": "error", "message": "invalid credentials"})
		return
	}

	if !user.CheckPassword(lr.Password) {
		c.IndentedJSON(http.StatusUnauthorized, gin.H{"status": "error", "message": "invalid credentials"})
		return
	}

	if !user.Active {
		c.IndentedJSON(http.StatusForbidden, gin.H{"status": "error", "message": "account not activated"})
		return
	}

	session := sessions.Default(c)
	session.Set("user", lr.User)
	session.Set("uid", user.UID)
	session.Set("roles", strings.Join(user.Roles, ","))
	session.Save()

	c.IndentedJSON(http.StatusOK, gin.H{
		"status":  "success",
		"message": "login successful",
		"user":    user.Name,
		"roles":   user.Roles,
	})
}

// Logout godoc
// @Summary Logout
// @Description Logout user
// @Tags auth
// @Accept json
// @Produce json
// @Success 200 {object} map[string]string
// @Router /auth/logout [get]
func (a AuthController) Logout(c *gin.Context) {
	session := sessions.Default(c)
	session.Clear()
	session.Save()
	c.IndentedJSON(http.StatusOK, gin.H{"status": "success", "message": "logged out"})
}

// Status godoc
// @Summary Auth Status
// @Description Check login status
// @Tags auth
// @Accept json
// @Produce json
// @Success 200 {object} map[string]interface{}
// @Router /auth/status [get]
func (a AuthController) Status(c *gin.Context) {
	session := sessions.Default(c)
	username := session.Get("user")
	roles := session.Get("roles")

	if username == nil {
		c.IndentedJSON(http.StatusOK, gin.H{
			"authenticated": false,
			"message":       "not logged in",
		})
		return
	}

	c.IndentedJSON(http.StatusOK, gin.H{
		"authenticated": true,
		"user":          username,
		"roles":         roles,
	})
}

// Register godoc
// @Summary Register User
// @Description Register a new user
// @Tags auth
// @Accept json
// @Produce json
// @Param register body models.RegisterReq true "Registration Data"
// @Success 201 {object} map[string]string
// @Router /auth/register [post]
func (a AuthController) Register(c *gin.Context) {
	var regreq models.RegisterReq
	if err := c.ShouldBindJSON(&regreq); err != nil {
		c.IndentedJSON(http.StatusBadRequest, gin.H{"status": "error", "message": err.Error()})
		return
	}

	db := models.GetDB()

	// Check if user exists
	var existing models.User
	result := db.First(&existing, "name = ?", regreq.Name)
	if result.Error == nil {
		c.IndentedJSON(http.StatusConflict, gin.H{"status": "error", "message": "username already exists"})
		return
	}

	newUser := models.MakeUser(regreq.Name)
	bytes, err := bcrypt.GenerateFromPassword([]byte(regreq.Password), 14)
	if err != nil {
		c.IndentedJSON(http.StatusInternalServerError, gin.H{"status": "error", "message": "error creating account"})
		return
	}
	newUser.PasswordHash = string(bytes)
	newUser.Active = false

	result = db.Create(&newUser)
	if result.Error != nil {
		c.IndentedJSON(http.StatusInternalServerError, gin.H{"status": "error", "message": result.Error.Error()})
		return
	}

	c.IndentedJSON(http.StatusCreated, gin.H{"status": "success", "message": "user created, awaiting activation"})
}

// ListUsers godoc
// @Summary List users
// @Description List all users (admin only)
// @Tags auth
// @Accept json
// @Produce json
// @Success 200 {array} models.User
// @Router /auth/users [get]
func (a AuthController) ListUsers(c *gin.Context) {
	db := models.GetDB()
	var users []models.User
	db.Order("name ASC").Find(&users)
	c.IndentedJSON(http.StatusOK, users)
}

// UpdateUser godoc
// @Summary Update user
// @Description Update user attributes (admin only)
// @Tags auth
// @Accept json
// @Produce json
// @Param uid path string true "User ID"
// @Param user body models.UserReq true "User Data"
// @Success 200 {object} map[string]string
// @Router /auth/users/{uid} [put]
func (a AuthController) UpdateUser(c *gin.Context) {
	db := models.GetDB()
	var user models.User
	result := db.First(&user, "uid = ?", c.Param("uid"))
	if result.Error != nil {
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			c.IndentedJSON(http.StatusNotFound, gin.H{"status": "error", "message": "user not found"})
			return
		}
		c.IndentedJSON(http.StatusInternalServerError, gin.H{"status": "error", "message": result.Error.Error()})
		return
	}

	var userReq models.UserReq
	if err := c.ShouldBindJSON(&userReq); err != nil {
		c.IndentedJSON(http.StatusBadRequest, gin.H{"status": "error", "message": err.Error()})
		return
	}

	user.Active = userReq.Active
	user.Roles = userReq.Roles

	result = db.Save(&user)
	if result.Error != nil {
		c.IndentedJSON(http.StatusInternalServerError, gin.H{"status": "error", "message": result.Error.Error()})
		return
	}

	c.IndentedJSON(http.StatusOK, gin.H{"status": "success", "message": "user updated"})
}

// DeleteUser godoc
// @Summary Delete user
// @Description Delete a user (admin only)
// @Tags auth
// @Accept json
// @Produce json
// @Param uid path string true "User ID"
// @Success 200 {object} map[string]string
// @Router /auth/user/{uid} [delete]
func (a AuthController) DeleteUser(c *gin.Context) {
	db := models.GetDB()
	var user models.User
	result := db.First(&user, "uid = ?", c.Param("uid"))
	if result.Error != nil {
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			c.IndentedJSON(http.StatusNotFound, gin.H{"status": "error", "message": "user not found"})
			return
		}
		c.IndentedJSON(http.StatusInternalServerError, gin.H{"status": "error", "message": result.Error.Error()})
		return
	}

	if user.Name == "admin" {
		c.IndentedJSON(http.StatusForbidden, gin.H{"status": "error", "message": "cannot delete admin user"})
		return
	}

	result = db.Delete(&user)
	if result.Error != nil {
		c.IndentedJSON(http.StatusInternalServerError, gin.H{"status": "error", "message": result.Error.Error()})
		return
	}

	c.IndentedJSON(http.StatusOK, gin.H{"status": "success", "message": "user deleted"})
}
