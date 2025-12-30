package supabase

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"strings"

	"github.com/supabase-community/supabase-go"
)

var client *supabase.Client
var supabaseURL string
var supabaseKey string

func InitClient() {
	supabaseURL = os.Getenv("SUPABASE_URL")
	supabaseKey = os.Getenv("SUPABASE_KEY")

	schema := os.Getenv("SUPABASE_SCHEMA")
	if schema == "" {
		// After segmented-schema migration, the service tables live under schema `company`.
		schema = "company"
	}

	c, err := supabase.NewClient(supabaseURL, supabaseKey, &supabase.ClientOptions{
		Schema: schema,
		Headers: map[string]string{
			"X-Client-Info": "carteira-contabil-focus-integration-service",
		},
	})
	if err != nil {
		log.Fatalf("Erro ao iniciar Supabase: %v", err)
	}
	client = c
}

func GetClient() *supabase.Client {
	return client
}

func RpcPublic(name string, body any) (string, error) {
	if strings.TrimSpace(supabaseURL) == "" || strings.TrimSpace(supabaseKey) == "" {
		return "", fmt.Errorf("supabase env não configurado (SUPABASE_URL/SUPABASE_KEY)")
	}

	// PostgREST expects RPC under /rest/v1/rpc/<name>
	url := strings.TrimRight(supabaseURL, "/") + "/rest/v1/rpc/" + name

	var payload []byte
	if body == nil {
		payload = []byte(`{}`)
	} else {
		b, err := json.Marshal(body)
		if err != nil {
			return "", fmt.Errorf("erro ao serializar payload RPC: %w", err)
		}
		payload = b
	}

	req, err := http.NewRequest("POST", url, bytes.NewReader(payload))
	if err != nil {
		return "", fmt.Errorf("erro ao criar request RPC: %w", err)
	}

	// Force schema = public for RPC
	req.Header.Set("Accept-Profile", "public")
	req.Header.Set("Content-Profile", "public")
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json")
	req.Header.Set("apikey", supabaseKey)
	req.Header.Set("Authorization", "Bearer "+supabaseKey)
	req.Header.Set("X-Client-Info", "carteira-contabil-focus-integration-service")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return "", fmt.Errorf("erro ao executar RPC: %w", err)
	}
	defer resp.Body.Close()

	respBody, _ := io.ReadAll(resp.Body)

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		// PostgREST error is JSON; return raw for debugging.
		return "", fmt.Errorf("rpc %s failed: HTTP %d: %s", name, resp.StatusCode, strings.TrimSpace(string(respBody)))
	}

	// For void functions PostgREST may return empty body
	return strings.TrimSpace(string(respBody)), nil
}


