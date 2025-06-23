package server

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	cmd2 "github.com/Edofo/bitly-clone/cmd"
	"github.com/Edofo/bitly-clone/internal/api"
	"github.com/Edofo/bitly-clone/internal/models"
	"github.com/Edofo/bitly-clone/internal/monitor"
	"github.com/Edofo/bitly-clone/internal/repository"
	"github.com/Edofo/bitly-clone/internal/services"
	"github.com/Edofo/bitly-clone/internal/workers"
	"github.com/gin-gonic/gin"
	"github.com/spf13/cobra"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

var RunServerCmd = &cobra.Command{
	Use:   "run-server",
	Short: "Lance le serveur API de raccourcissement d'URLs et les processus de fond.",
	Long: `Cette commande initialise la base de données, configure les APIs,
démarre les workers asynchrones pour les clics et le moniteur d'URLs,
puis lance le serveur HTTP.`,
	Run: func(cmd *cobra.Command, args []string) {
		cfg := cmd2.Cfg
		if cfg == nil {
			log.Fatalf("FATAL: Configuration not loaded.")
		}

		db, err := gorm.Open(sqlite.Open(cfg.Database.Name), &gorm.Config{})
		if err != nil {
			log.Fatalf("FATAL: Unable to connect to database: %v", err)
		}

		sqlDB, err := db.DB()
		if err != nil {
			log.Fatalf("FATAL: Failed to get underlying SQL database: %v", err)
		}
		defer func() {
			if err := sqlDB.Close(); err != nil {
				log.Printf("Warning: Failed to close database connection: %v", err)
			}
		}()

		linkRepo := repository.NewLinkRepository(db)
		clickRepo := repository.NewClickRepository(db)

		log.Println("Repositories initialized.")

		linkService := services.NewLinkService(linkRepo)
		_ = services.NewClickService(clickRepo) 

		log.Println("Business services initialized.")

		clickEventsChan := make(chan models.ClickEvent, cfg.Analytics.BufferSize)
		workers.StartClickWorkers(cfg.Analytics.Workers, clickEventsChan, clickRepo)

		log.Printf("Click events channel initialized with buffer size %d. %d click worker(s) started.",
			cfg.Analytics.BufferSize, cfg.Analytics.Workers)

		monitorInterval := time.Duration(cfg.Monitor.IntervalMinutes) * time.Minute
		urlMonitor := monitor.NewUrlMonitor(linkRepo, monitorInterval)
		go urlMonitor.Start()
		log.Printf("URL monitor started with interval %v.", monitorInterval)

		router := gin.Default()
		api.SetupRoutes(router, linkService, clickEventsChan)

		log.Println("API routes configured.")

		serverAddr := fmt.Sprintf(":%d", cfg.Server.Port)
		srv := &http.Server{
			Addr:    serverAddr,
			Handler: router,
		}

		go func() {
			log.Printf("Server started on port %d", cfg.Server.Port)
			if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
				log.Fatalf("FATAL: Failed to start server: %v", err)
			}
		}()

		quit := make(chan os.Signal, 1)
		signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM) 

		<-quit
		log.Println("Shutdown signal received. Stopping server...")

		log.Println("Shutting down... Giving workers time to finish.")
		time.Sleep(5 * time.Second)

		log.Println("Server stopped gracefully.")
	},
}

func init() {
	cmd2.RootCmd.AddCommand(RunServerCmd)
}
