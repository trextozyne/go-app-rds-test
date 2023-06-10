package main

import (
	"database/sql"
	"fmt"
	"log"
	"net/http"
	"os"
	"time"
	"path/filepath"

	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
	_ "github.com/go-sql-driver/mysql"
	"golang.org/x/sync/errgroup"
)

var (
	g            errgroup.Group
	RDSEndpoint  string
	DatabaseName string // Replace with your actual database name
)

type ErrorResponse struct {
	Error string `json:"error"`
}

type App struct {
	DB *sql.DB
}

func main() {
	// Get the absolute path to the .env file..
	envPath := filepath.Join(os.Getenv("HOME"), ".env")

	// Load the environment variables from the .env file.
	err := godotenv.Load(envPath)
	if err != nil {
		fmt.Println(envPath)
		log.Fatalf("Error loading environment variables file")
	}

	// Retrieve the RDS_ENDPOINT environment variable
	RDSEndpoint := os.Getenv("RDSEndpoint")

	DatabaseName = os.Getenv("DatabaseName")

	if RDSEndpoint == "" {
		log.Fatal("RDS_ENDPOINT environment variable is not set")
	}

	if DatabaseName == "" {
		log.Fatal("DatabaseName environment variable is not set")
	}
	// Construct the database connection string
	dbURL := fmt.Sprintf("tosyne:Salvat1on@tcp(%s)/%s", RDSEndpoint, DatabaseName)

	// Open a connection to the database
	db, err := sql.Open("mysql", dbURL)
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	// Create the 'items' table in the database
	err = createItemsTable(db)
	if err != nil {
		log.Fatal(err)
	}

	app := &App{
		DB: db,
	}

	router := gin.Default()
	router.GET("/hostname", app.getHostname)
	router.GET("/ping", ping)
	router.GET("/health", getHealthStatus)

	server := &http.Server{
		Addr:         ":8080",
		Handler:      router,
		ReadTimeout:  5 * time.Second,
		WriteTimeout: 10 * time.Second,
	}

	if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		log.Fatal(err)
	}
}

func createItemsTable(db *sql.DB) error {
	_, err := db.Exec(`CREATE TABLE IF NOT EXISTS items (
		id INT AUTO_INCREMENT PRIMARY KEY,
		name VARCHAR(255) NOT NULL
	)`)
	if err != nil {
		return err
	}
	log.Println("Table 'items' created successfully")
	return nil
}

func (app *App) getHostname(c *gin.Context) {
	// Get the hostname
	name, err := os.Hostname()
	if err != nil {
		log.Println(err)
		c.JSON(http.StatusInternalServerError, ErrorResponse{Error: "Internal Server Error"})
		return
	}

	// Insert the hostname into the database
	_, err = app.DB.Exec("INSERT INTO items (name) VALUES (?)", name)
	if err != nil {
		log.Println(err)
		c.JSON(http.StatusInternalServerError, ErrorResponse{Error: "Internal Server Error"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"hostname": name})
}

func getHealthStatus(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"status": "ready"})
}

func ping(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"message": "pong"})
}