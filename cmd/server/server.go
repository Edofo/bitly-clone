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

// RunServerCmd représente la commande 'run-server' de Cobra.
// C'est le point d'entrée pour lancer le serveur de l'application.
var RunServerCmd = &cobra.Command{
	Use:   "run-server",
	Short: "Lance le serveur API de raccourcissement d'URLs et les processus de fond.",
	Long: `Cette commande initialise la base de données, configure les APIs,
démarre les workers asynchrones pour les clics et le moniteur d'URLs,
puis lance le serveur HTTP.`,
	Run: func(cmd *cobra.Command, args []string) {
		// Charger la configuration chargée globalement via cmd.Cfg
		cfg := cmd2.Cfg
		if cfg == nil {
			log.Fatalf("FATAL: Configuration not loaded.")
		}

		// Initialiser la connexion à la base de données SQLite avec GORM
		db, err := gorm.Open(sqlite.Open(cfg.Database.Name), &gorm.Config{})
		if err != nil {
			log.Fatalf("FATAL: Unable to connect to database: %v", err)
		}

		sqlDB, err := db.DB()
		if err != nil {
			log.Fatalf("FATAL: Failed to get underlying SQL database: %v", err)
		}
		defer sqlDB.Close()

		// Initialiser les repositories
		linkRepo := repository.NewLinkRepository(db)
		clickRepo := repository.NewClickRepository(db)

		// Laissez le log
		log.Println("Repositories initialized.")

		// Initialiser les services métiers
		linkService := services.NewLinkService(linkRepo)
		_ = services.NewClickService(clickRepo) // Service créé mais pas utilisé directement dans server

		// Laissez le log
		log.Println("Business services initialized.")

		// Initialiser le channel des événements de clic et lancer les workers
		clickEventsChan := make(chan models.ClickEvent, cfg.Analytics.BufferSize)
		workers.StartClickWorkers(cfg.Analytics.Workers, clickEventsChan, clickRepo)

		log.Printf("Click events channel initialized with buffer size %d. %d click worker(s) started.",
			cfg.Analytics.BufferSize, cfg.Analytics.Workers)

		// Initialiser et lancer le moniteur d'URLs
		monitorInterval := time.Duration(cfg.Monitor.IntervalMinutes) * time.Minute
		urlMonitor := monitor.NewUrlMonitor(linkRepo, monitorInterval)
		go urlMonitor.Start()
		log.Printf("URL monitor started with interval %v.", monitorInterval)

		// Configurer le routeur Gin et les handlers API
		router := gin.Default()
		api.SetupRoutes(router, linkService)

		// Pas toucher au log
		log.Println("API routes configured.")

		// Créer le serveur HTTP Gin
		serverAddr := fmt.Sprintf(":%d", cfg.Server.Port)
		srv := &http.Server{
			Addr:    serverAddr,
			Handler: router,
		}

		// Démarrer le serveur Gin dans une goroutine anonyme pour ne pas bloquer
		go func() {
			log.Printf("Server started on port %d", cfg.Server.Port)
			if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
				log.Fatalf("FATAL: Failed to start server: %v", err)
			}
		}()

		// Gére l'arrêt propre du serveur (graceful shutdown).
		// Créez un channel pour les signaux OS (SIGINT, SIGTERM).
		quit := make(chan os.Signal, 1)
		signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM) // Attendre Ctrl+C ou signal d'arrêt

		// Bloquer jusqu'à ce qu'un signal d'arrêt soit reçu.
		<-quit
		log.Println("Shutdown signal received. Stopping server...")

		// Arrêt propre du serveur HTTP avec un timeout.
		log.Println("Shutting down... Giving workers time to finish.")
		time.Sleep(5 * time.Second)

		log.Println("Server stopped gracefully.")
	},
}

func init() {
	// Ajouter la commande à RootCmd
	cmd2.RootCmd.AddCommand(RunServerCmd)
}
