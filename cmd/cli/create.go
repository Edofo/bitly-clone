package cli

import (
	"fmt"
	"log"
	"net/url"
	"os"

	cmd2 "github.com/Edofo/bitly-clone/cmd"
	"github.com/Edofo/bitly-clone/internal/repository"
	"github.com/Edofo/bitly-clone/internal/services"
	"github.com/spf13/cobra"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

var longURLFlag string

var CreateCmd = &cobra.Command{
	Use:   "create",
	Short: "Crée une URL courte à partir d'une URL longue.",
	Long: `Cette commande raccourcit une URL longue fournie et affiche le code court généré.

Exemple:
  url-shortener create --url="https://www.google.com/search?q=go+lang"`,
	Run: func(cmd *cobra.Command, args []string) {
		if longURLFlag == "" {
			fmt.Println("Erreur: Le flag --url est requis.")
			os.Exit(1)
		}

		_, err := url.ParseRequestURI(longURLFlag)
		if err != nil {
			fmt.Printf("Erreur: URL invalide '%s': %v\n", longURLFlag, err)
			os.Exit(1)
		}

		cfg := cmd2.Cfg
		if cfg == nil {
			fmt.Println("Erreur: Configuration non chargée.")
			os.Exit(1)
		}

		db, err := gorm.Open(sqlite.Open(cfg.Database.Name), &gorm.Config{})
		if err != nil {
			log.Fatalf("FATAL: Impossible de se connecter à la base de données: %v", err)
		}

		sqlDB, err := db.DB()
		if err != nil {
			log.Fatalf("FATAL: Échec de l'obtention de la base de données SQL sous-jacente: %v", err)
		}

		defer func() {
			if err := sqlDB.Close(); err != nil {
				log.Printf("Warning: Failed to close database connection: %v", err)
			}
		}()

		linkRepo := repository.NewLinkRepository(db)
		linkService := services.NewLinkService(linkRepo)

		link, err := linkService.CreateLink(longURLFlag)
		if err != nil {
			fmt.Printf("Erreur lors de la création du lien court: %v\n", err)
			os.Exit(1)
		}

		fullShortURL := fmt.Sprintf("%s/%s", cfg.Server.BaseURL, link.ShortCode)
		fmt.Printf("URL courte créée avec succès:\n")
		fmt.Printf("Code: %s\n", link.ShortCode)
		fmt.Printf("URL complète: %s\n", fullShortURL)
	},
}

func init() {
	CreateCmd.Flags().StringVar(&longURLFlag, "url", "", "URL longue à raccourcir")

	if err := CreateCmd.MarkFlagRequired("url"); err != nil {
		log.Fatalf("Failed to mark url flag as required: %v", err)
	}

	cmd2.RootCmd.AddCommand(CreateCmd)
}
