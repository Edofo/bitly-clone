package cli

import (
	"fmt"
	"log"
	"os"

	cmd2 "github.com/Edofo/bitly-clone/cmd"
	"github.com/Edofo/bitly-clone/internal/repository"
	"github.com/Edofo/bitly-clone/internal/services"
	"github.com/spf13/cobra"

	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

var shortCodeFlag string

var StatsCmd = &cobra.Command{
	Use:   "stats",
	Short: "Affiche les statistiques (nombre de clics) pour un lien court.",
	Long: `Cette commande permet de récupérer et d'afficher le nombre total de clics
pour une URL courte spécifique en utilisant son code.

Exemple:
  url-shortener stats --code="xyz123"`,
	Run: func(cmd *cobra.Command, args []string) {
		if shortCodeFlag == "" {
			fmt.Println("Erreur: Le flag --code est requis.")
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

		defer sqlDB.Close()

		linkRepo := repository.NewLinkRepository(db)
		linkService := services.NewLinkService(linkRepo)

		link, totalClicks, err := linkService.GetLinkStats(shortCodeFlag)
		if err != nil {
			if err == gorm.ErrRecordNotFound {
				fmt.Printf("Erreur: Aucun lien trouvé avec le code '%s'\n", shortCodeFlag)
			} else {
				fmt.Printf("Erreur lors de la récupération des statistiques: %v\n", err)
			}
			os.Exit(1)
		}

		fmt.Printf("Statistiques pour le code court: %s\n", link.ShortCode)
		fmt.Printf("URL longue: %s\n", link.LongURL)
		fmt.Printf("Total de clics: %d\n", totalClicks)
	},
}

func init() {
	StatsCmd.Flags().StringVar(&shortCodeFlag, "code", "", "Code court pour lequel afficher les statistiques")

	StatsCmd.MarkFlagRequired("code")

	cmd2.RootCmd.AddCommand(StatsCmd)
}
