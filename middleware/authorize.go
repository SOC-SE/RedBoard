package middleware

import (
	"net/http"
	"os"
	"slices"
	"strings"

	"github.com/gin-contrib/sessions"
	"github.com/gin-gonic/gin"
)

// AuthorizeHTML redirects to login page if not authenticated
func AuthorizeHTML(role string) gin.HandlerFunc {
	return func(c *gin.Context) {
		session := sessions.Default(c)
		user := session.Get("user")

		if user == nil {
			c.Redirect(http.StatusFound, os.Getenv("API_BASE_URL")+"/login.html")
			c.Abort()
			return
		}

		c.Set("user", user)
		rolestring, ok := session.Get("roles").(string)
		if !ok {
			c.Redirect(http.StatusFound, os.Getenv("API_BASE_URL")+"/login.html")
			c.Abort()
			return
		}

		c.Set("roles", rolestring)
		roles := strings.Split(rolestring, ",")

		if role == "any" || slices.Contains(roles, role) {
			c.Next()
			return
		}

		c.AbortWithStatus(http.StatusForbidden)
	}
}

// Authorize checks authentication for API endpoints
func Authorize(role string) gin.HandlerFunc {
	return func(c *gin.Context) {
		session := sessions.Default(c)
		user := session.Get("user")

		if user == nil {
			c.IndentedJSON(http.StatusUnauthorized, gin.H{
				"status":  "error",
				"message": "authentication required",
			})
			c.Abort()
			return
		}

		rolesstring, ok := session.Get("roles").(string)
		if !ok {
			c.IndentedJSON(http.StatusUnauthorized, gin.H{
				"status":  "error",
				"message": "invalid session",
			})
			c.Abort()
			return
		}

		roles := strings.Split(rolesstring, ",")
		if role != "any" && !slices.Contains(roles, role) {
			c.IndentedJSON(http.StatusForbidden, gin.H{
				"status":  "error",
				"message": "insufficient permissions",
			})
			c.Abort()
			return
		}

		c.Set("user", user)
		c.Set("roles", rolesstring)
		c.Next()
	}
}

// SecurityHeaders adds security headers to responses
func SecurityHeaders() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Header("X-Content-Type-Options", "nosniff")
		c.Header("X-Frame-Options", "DENY")
		c.Header("X-XSS-Protection", "1; mode=block")
		c.Header("Referrer-Policy", "strict-origin-when-cross-origin")
		c.Next()
	}
}

// CORS adds CORS headers for development
func CORS() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Header("Access-Control-Allow-Origin", os.Getenv("API_BASE_URL"))
		c.Header("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		c.Header("Access-Control-Allow-Headers", "Content-Type, Authorization")
		c.Header("Access-Control-Allow-Credentials", "true")

		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(http.StatusNoContent)
			return
		}

		c.Next()
	}
}
