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
			log.Fatalf("FATAL: Configuration non chargée.")
		}

		// Initialiser la connexion à la base de données SQLite avec GORM
		db, err := gorm.Open(sqlite.Open(cfg.Database.Name), &gorm.Config{})
		if err != nil {
			log.Fatalf("FATAL: Impossible de se connecter à la base de données: %v", err)
		}

		sqlDB, err := db.DB()
		if err != nil {
			log.Fatalf("FATAL: Échec de l'obtention de la base de données SQL sous-jacente: %v", err)
		}
		defer sqlDB.Close()

		// Initialiser les repositories
		linkRepo := repository.NewLinkRepository(db)
		clickRepo := repository.NewClickRepository(db)

		// Laissez le log
		log.Println("Repositories initialisés.")

		// Initialiser les services métiers
		linkService := services.NewLinkService(linkRepo)
		_ = services.NewClickService(clickRepo) // Service créé mais pas utilisé directement dans server

		// Laissez le log
		log.Println("Services métiers initialisés.")

		// Initialiser le channel des événements de clic et lancer les workers
		clickEventsChan := make(chan models.ClickEvent, cfg.Analytics.BufferSize)
		workers.StartClickWorkers(cfg.Analytics.Workers, clickEventsChan, clickRepo)

		log.Printf("Channel d'événements de clic initialisé avec un buffer de %d. %d worker(s) de clics démarré(s).",
			cfg.Analytics.BufferSize, cfg.Analytics.Workers)

		// Initialiser et lancer le moniteur d'URLs
		monitorInterval := time.Duration(cfg.Monitor.IntervalMinutes) * time.Minute
		urlMonitor := monitor.NewUrlMonitor(linkRepo, monitorInterval)
		go urlMonitor.Start()
		log.Printf("Moniteur d'URLs démarré avec un intervalle de %v.", monitorInterval)

		// Configurer le routeur Gin et les handlers API
		router := gin.Default()
		api.SetupRoutes(router, linkService)

		// Pas toucher au log
		log.Println("Routes API configurées.")

		// Créer le serveur HTTP Gin
		serverAddr := fmt.Sprintf(":%d", cfg.Server.Port)
		srv := &http.Server{
			Addr:    serverAddr,
			Handler: router,
		}

		// Démarrer le serveur Gin dans une goroutine anonyme pour ne pas bloquer
		go func() {
			log.Printf("Serveur démarré sur le port %d", cfg.Server.Port)
			if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
				log.Fatalf("FATAL: Échec du démarrage du serveur: %v", err)
			}
		}()

		// Gére l'arrêt propre du serveur (graceful shutdown).
		// Créez un channel pour les signaux OS (SIGINT, SIGTERM).
		quit := make(chan os.Signal, 1)
		signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM) // Attendre Ctrl+C ou signal d'arrêt

		// Bloquer jusqu'à ce qu'un signal d'arrêt soit reçu.
		<-quit
		log.Println("Signal d'arrêt reçu. Arrêt du serveur...")

		// Arrêt propre du serveur HTTP avec un timeout.
		log.Println("Arrêt en cours... Donnez un peu de temps aux workers pour finir.")
		time.Sleep(5 * time.Second)

		log.Println("Serveur arrêté proprement.")
	},
}

func init() {
	// Ajouter la commande à RootCmd
	cmd2.RootCmd.AddCommand(RunServerCmd)
}
