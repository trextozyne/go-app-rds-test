package main

import (
	"database/sql"
	"fmt"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/gin-gonic/gin"
	_ "github.com/go-sql-driver/mysql"
	"golang.org/x/sync/errgroup"
)

var (
	g errgroup.Group
)

var RDSEndpoint string

func main() {
	// Retrieve the RDS_ENDPOINT environment variable
	rdsEndpoint := RDSEndpoint
	if rdsEndpoint == "" {
		log.Fatal("RDS_ENDPOINT environment variable is not set")
	}

	// Construct the database connection string
	dbURL := fmt.Sprintf("tosyne:Salvat1on@tcp(%s:port)/database", rdsEndpoint)

	// Open a connection to the database
	db, err := sql.Open("mysql", dbURL)
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	// Create the 'items' table in the database
	_, err = db.Exec(`CREATE TABLE IF NOT EXISTS items (
		id INT AUTO_INCREMENT PRIMARY KEY,
		name VARCHAR(255) NOT NULL
	)`)
	if err != nil {
		log.Fatal(err)
	}

	mainServer := &http.Server{
		Addr:         ":8080",
		Handler:      mainRouter(db),
		ReadTimeout:  5 * time.Second,
		WriteTimeout: 10 * time.Second,
	}

	healthServer := &http.Server{
		Addr:         ":8081",
		Handler:      healthRouter(),
		ReadTimeout:  5 * time.Second,
		WriteTimeout: 10 * time.Second,
	}

	g.Go(func() error {
		err := mainServer.ListenAndServe()
		if err != nil && err != http.ErrServerClosed {
			log.Fatal(err)
		}
		return err
	})

	g.Go(func() error {
		err := healthServer.ListenAndServe()
		if err != nil && err != http.ErrServerClosed {
			log.Fatal(err)
		}
		return err
	})

	if err := g.Wait(); err != nil {
		log.Fatal(err)
	}
}

// Modify the database connection details according to your Amazon RDS configuration
func mainRouter(db *sql.DB) http.Handler {
	engine := gin.New()
	engine.Use(gin.Recovery())
	engine.GET("/hostname", getHostname(db))
	engine.GET("/ping", ping)
	return engine
}

func healthRouter() http.Handler {
	engine := gin.New()
	engine.Use(gin.Recovery())
	engine.GET("/health", getHealthStatus)
	return engine
}

type Item struct {
	ID   int    `json:"id"`
	Name string `json:"name"`
}

func getHostname(db *sql.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Get the hostname
		name, err := os.Hostname()
		if err != nil {
			panic(err)
		}

		// Insert the hostname into the database
		_, err = db.Exec("INSERT INTO items (name) VALUES (?)", name)
		if err != nil {
			panic(err)
		}

		c.IndentedJSON(http.StatusOK, gin.H{"hostname": name})
	}
}

func getHealthStatus(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"status": "ready"})
}

func ping(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"message": "pong"})
}
