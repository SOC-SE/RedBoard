package server

import (
	"crypto/rand"
	"encoding/base64"
	"net/http"
	"os"
	"strings"
	"text/template"

	"github.com/brian-l-johnson/Redteam-Dashboard-go/v2/controllers"
	docs "github.com/brian-l-johnson/Redteam-Dashboard-go/v2/docs"
	"github.com/brian-l-johnson/Redteam-Dashboard-go/v2/middleware"
	"github.com/gin-contrib/sessions"
	"github.com/gin-contrib/sessions/cookie"
	"github.com/gin-gonic/gin"
	swaggerfiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
)

func getAPIBaseURL() string {
	return os.Getenv("API_BASE_URL")
}

func isAdmin(roles string) bool {
	return strings.Contains(roles, "admin")
}

// getSessionSecret generates or retrieves secure session secret
func getSessionSecret() []byte {
	secret := os.Getenv("SESSION_SECRET")
	if secret == "" {
		// Generate random secret if not provided
		randomBytes := make([]byte, 32)
		rand.Read(randomBytes)
		return randomBytes
	}
	decoded, err := base64.StdEncoding.DecodeString(secret)
	if err != nil {
		return []byte(secret)
	}
	return decoded
}

func NewRouter() *gin.Engine {
	router := gin.New()

	// Use secure cookie store instead of memstore
	store := cookie.NewStore(getSessionSecret())
	store.Options(sessions.Options{
		Path:     "/",
		MaxAge:   86400 * 7, // 7 days
		HttpOnly: true,
		Secure:   os.Getenv("GIN_MODE") == "release",
		SameSite: http.SameSiteLaxMode,
	})
	router.Use(sessions.Sessions("session", store))

	router.Use(gin.Logger())
	router.Use(gin.Recovery())
	router.Use(middleware.SecurityHeaders())

	// Health endpoint (no auth required)
	health := new(controllers.HealthController)
	router.GET("/health", health.Status)

	// Auth endpoints
	auth := new(controllers.AuthController)
	router.POST("/auth/login", auth.Login)
	router.GET("/auth/status", auth.Status)
	router.POST("/auth/register", auth.Register)
	router.GET("/auth/logout", auth.Logout)
	router.GET("/auth/users", middleware.Authorize("admin"), auth.ListUsers)
	router.PUT("/auth/users/:uid", middleware.Authorize("admin"), auth.UpdateUser)
	router.DELETE("/auth/user/:uid", middleware.Authorize("admin"), auth.DeleteUser)
	router.POST("/auth/admin/create-user", middleware.Authorize("admin"), auth.AdminCreateUser)
	router.PUT("/auth/admin/reset-password/:uid", middleware.Authorize("admin"), auth.AdminResetPassword)

	// Team endpoints
	team := new(controllers.TeamController)
	router.GET("/teams", middleware.Authorize("viewer"), team.GetTeams)
	router.GET("/teams/:tid", middleware.Authorize("viewer"), team.GetTeam)
	router.POST("/teams", middleware.Authorize("admin"), team.CreateTeam)
	router.PUT("/teams/:tid", middleware.Authorize("admin"), team.UpdateTeam)
	router.DELETE("/teams/:tid", middleware.Authorize("admin"), team.DeleteTeam)

	// Job endpoints
	jobs := new(controllers.JobController)
	router.GET("/jobs/manager", middleware.Authorize("viewer"), jobs.GetJobManagerState)
	router.GET("/jobs/:jobtype/next", middleware.Authorize("scanner"), jobs.NewJob)
	router.GET("/jobs", middleware.Authorize("any"), jobs.GetJobs)
	router.POST("/jobs/nmap/:jid", middleware.Authorize("scanner"), jobs.UploadScan)
	router.POST("/jobs/:jid/cancel", middleware.Authorize("admin"), jobs.CancelJob)

	// Host endpoints
	host := new(controllers.HostController)
	router.GET("/hosts/by-team/:tid", middleware.Authorize("viewer"), host.GetHostsByTeam)
	router.GET("/hosts/by-team/", middleware.Authorize("viewer"), host.GetAllHostsByTeam)
	router.GET("/dashboard/data", middleware.Authorize("viewer"), host.GetDashboardData)

	// Swagger
	docs.SwaggerInfo.BasePath = "/"
	router.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerfiles.Handler))

	// Static files
	router.Static("/static", "./static")

	// Template functions
	router.SetFuncMap(template.FuncMap{
		"getAPIBaseURL": getAPIBaseURL,
		"isAdmin":       isAdmin,
	})
	router.LoadHTMLGlob("templates/*")

	// HTML routes
	router.GET("/login.html", func(c *gin.Context) {
		c.HTML(http.StatusOK, "login.html", gin.H{"title": "Login"})
	})

	router.GET("/register.html", func(c *gin.Context) {
		c.HTML(http.StatusOK, "register.html", gin.H{"title": "Register"})
	})

	router.GET("/main.html", middleware.AuthorizeHTML("any"), func(c *gin.Context) {
		session := sessions.Default(c)
		c.HTML(http.StatusOK, "main.html", gin.H{
			"user":  session.Get("user"),
			"roles": session.Get("roles"),
			"title": "Dashboard",
		})
	})

	router.GET("/teams.html", middleware.AuthorizeHTML("admin"), func(c *gin.Context) {
		session := sessions.Default(c)
		c.HTML(http.StatusOK, "teams.html", gin.H{
			"user":  session.Get("user"),
			"roles": session.Get("roles"),
			"title": "Team Management",
		})
	})

	router.GET("/users.html", middleware.AuthorizeHTML("admin"), func(c *gin.Context) {
		session := sessions.Default(c)
		c.HTML(http.StatusOK, "users.html", gin.H{
			"user":  session.Get("user"),
			"roles": session.Get("roles"),
			"title": "User Management",
		})
	})

	router.GET("/jobs.html", middleware.AuthorizeHTML("admin"), func(c *gin.Context) {
		session := sessions.Default(c)
		c.HTML(http.StatusOK, "jobs.html", gin.H{
			"user":  session.Get("user"),
			"roles": session.Get("roles"),
			"title": "Job Queue",
		})
	})

	router.GET("/logout.html", func(c *gin.Context) {
		session := sessions.Default(c)
		session.Clear()
		session.Save()
		c.Redirect(http.StatusFound, "/login.html")
	})

	router.GET("/", func(c *gin.Context) {
		c.Redirect(http.StatusFound, "/main.html")
	})

	return router
}
