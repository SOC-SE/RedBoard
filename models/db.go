package models

import (
	"crypto/rand"
	"errors"
	"fmt"
	"math/big"
	"os"

	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

var db *gorm.DB

func Init() {
	fmt.Println("Initializing database...")
	var err error

	// Configure logger based on environment
	logLevel := logger.Warn
	if os.Getenv("GIN_MODE") != "release" {
		logLevel = logger.Info
	}

	db, err = gorm.Open(sqlite.Open("dashboard.db"), &gorm.Config{
		Logger: logger.Default.LogMode(logLevel),
	})
	if err != nil {
		panic("failed to open database file")
	}

	// Run migrations
	db.AutoMigrate(&User{})
	db.AutoMigrate(&Host{})
	db.AutoMigrate(&Port{})
	db.AutoMigrate(&Team{})
	db.AutoMigrate(&Job{})
	db.AutoMigrate(&JobStatus{})
	db.AutoMigrate(&PortBaseline{})
	db.AutoMigrate(&ScanHistory{})

	// Create indexes for better query performance
	db.Exec("CREATE INDEX IF NOT EXISTS idx_hosts_team_id ON hosts(team_id)")
	db.Exec("CREATE INDEX IF NOT EXISTS idx_hosts_ip ON hosts(ip)")
	db.Exec("CREATE INDEX IF NOT EXISTS idx_ports_host_id ON ports(host_id)")
	db.Exec("CREATE INDEX IF NOT EXISTS idx_teams_tid ON teams(t_id)")
	db.Exec("CREATE INDEX IF NOT EXISTS idx_jobs_status ON jobs(status)")
	db.Exec("CREATE INDEX IF NOT EXISTS idx_jobs_tid ON jobs(t_id)")

	// Create admin user if not exists
	var user User
	result := db.First(&user, "name=?", "admin")
	if errors.Is(result.Error, gorm.ErrRecordNotFound) {
		fmt.Println("Admin user does not exist, creating...")

		adminUser := MakeUser("admin")
		genpw, err := GenerateRandomString(32)
		if err != nil {
			panic("unable to generate random password")
		}
		adminUser.SetPassword(genpw)
		adminUser.Active = true
		adminUser.Roles = []string{"admin", "viewer", "scanner"}

		result = db.Create(&adminUser)
		if result.Error != nil {
			panic("unable to save admin user")
		}

		fmt.Println("========================================")
		fmt.Printf("  Admin user created\n")
		fmt.Printf("  Username: admin\n")
		fmt.Printf("  Password: %s\n", genpw)
		fmt.Println("========================================")
	}

	// Initialize job status for nmap if not exists
	var scans []JobStatus
	result = db.Find(&scans)
	if len(scans) < 1 {
		sc := JobStatus{Name: "nmap", JobIndex: 0}
		db.Save(&sc)
	}

	fmt.Println("Database initialization complete")
}

func GetDB() *gorm.DB {
	return db
}

func GenerateRandomString(n int) (string, error) {
	const letters = "0123456789ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz"
	ret := make([]byte, n)
	for i := 0; i < n; i++ {
		num, err := rand.Int(rand.Reader, big.NewInt(int64(len(letters))))
		if err != nil {
			return "", err
		}
		ret[i] = letters[num.Int64()]
	}
	return string(ret), nil
}
