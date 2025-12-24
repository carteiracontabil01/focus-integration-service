package supabase

import (
	"log"
	"os"

	"github.com/supabase-community/supabase-go"
)

var client *supabase.Client

func InitClient() {
	url := os.Getenv("SUPABASE_URL")
	key := os.Getenv("SUPABASE_KEY")

	c, err := supabase.NewClient(url, key, nil)
	if err != nil {
		log.Fatalf("Erro ao iniciar Supabase: %v", err)
	}
	client = c
}

func GetClient() *supabase.Client {
	return client
}


