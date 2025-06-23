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

// longURLFlag stockera la valeur du flag --url
var longURLFlag string

// CreateCmd représente la commande 'create'
var CreateCmd = &cobra.Command{
	Use:   "create",
	Short: "Crée une URL courte à partir d'une URL longue.",
	Long: `Cette commande raccourcit une URL longue fournie et affiche le code court généré.

Exemple:
  url-shortener create --url="https://www.google.com/search?q=go+lang"`,
	Run: func(cmd *cobra.Command, args []string) {
		// Valider que le flag --url a été fourni
		if longURLFlag == "" {
			fmt.Println("Erreur: Le flag --url est requis.")
			os.Exit(1)
		}

		// Validation basique du format de l'URL
		_, err := url.ParseRequestURI(longURLFlag)
		if err != nil {
			fmt.Printf("Erreur: URL invalide '%s': %v\n", longURLFlag, err)
			os.Exit(1)
		}

		// Charger la configuration chargée globalement via cmd.Cfg
		cfg := cmd2.Cfg
		if cfg == nil {
			fmt.Println("Erreur: Configuration non chargée.")
			os.Exit(1)
		}

		// Initialiser la connexion à la base de données SQLite
		db, err := gorm.Open(sqlite.Open(cfg.Database.Name), &gorm.Config{})
		if err != nil {
			log.Fatalf("FATAL: Impossible de se connecter à la base de données: %v", err)
		}

		sqlDB, err := db.DB()
		if err != nil {
			log.Fatalf("FATAL: Échec de l'obtention de la base de données SQL sous-jacente: %v", err)
		}

		// S'assurer que la connexion est fermée à la fin de l'exécution de la commande
		defer sqlDB.Close()

		// Initialiser les repositories et services nécessaires
		linkRepo := repository.NewLinkRepository(db)
		linkService := services.NewLinkService(linkRepo)

		// Appeler le LinkService et la fonction CreateLink pour créer le lien court
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

// init() s'exécute automatiquement lors de l'importation du package.
// Il est utilisé pour définir les flags que cette commande accepte.
func init() {
	// Définir le flag --url pour la commande create
	CreateCmd.Flags().StringVar(&longURLFlag, "url", "", "URL longue à raccourcir")

	// Marquer le flag comme requis
	CreateCmd.MarkFlagRequired("url")

	// Ajouter la commande à RootCmd
	cmd2.RootCmd.AddCommand(CreateCmd)
}
