package config

import (
	"fmt"
	"log" // Pour logger les informations ou erreurs de chargement de config

	"github.com/spf13/viper" // La bibliothèque pour la gestion de configuration
)

type Config struct {
	Server struct {
		Port int `mapstructure:"port"`
		BaseURL string `mapstructure:"base_url"`
	} `mapstructure:"server"`
	Database struct {
		Name string `mapstructure:"name"`
	} `mapstructure:"database"`
	Analytics struct {
		BufferSize int `mapstructure:"buffer_size"`
		Workers int `mapstructure:"workers"`
	} `mapstructure:"analytics"`
	Monitor struct {
		IntervalMinutes int `mapstructure:"interval_minutes"`
	} `mapstructure:"monitor"`
}

// LoadConfig charge la configuration de l'application en utilisant Viper.
// Elle recherche un fichier 'config.yaml' dans le dossier 'configs/'.
// Elle définit également des valeurs par défaut si le fichier de config est absent ou incomplet.
func LoadConfig() (*Config, error) {
	// Spécifie le chemin où Viper doit chercher les fichiers de config.
	// on cherche dans le dossier 'configs' relatif au répertoire d'exécution.
	viper.AddConfigPath("configs")
	viper.SetConfigName("config")
	viper.SetConfigType("yaml")

	viper.SetDefault("server.port", 8080)
	viper.SetDefault("server.base_url", "http://localhost:8080")
	viper.SetDefault("database.name", "url_shortener.db")
	viper.SetDefault("analytics.buffer_size", 1000)
	viper.SetDefault("analytics.workers", 5)
	viper.SetDefault("monitor.interval_minutes", 5)

	// Lire le fichier de configuration.
	if err := viper.ReadInConfig(); err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); !ok {
			return nil, fmt.Errorf("error reading config file: %w", err)
		}
	}

	// Démapper (unmarshal) la configuration lue (ou les valeurs par défaut) dans la structure Config.
	var cfg Config = Config{}
	if err := viper.Unmarshal(&cfg); err != nil {
		return nil, fmt.Errorf("error unmarshalling config: %w", err)
	}

	// Log  pour vérifier la config chargée
	log.Printf("Configuration loaded: Server Port=%d, DB Name=%s, Analytics Buffer=%d, Monitor Interval=%dmin",
		cfg.Server.Port, cfg.Database.Name, cfg.Analytics.BufferSize, cfg.Monitor.IntervalMinutes)

	return &cfg, nil // Retourne la configuration chargée
}
