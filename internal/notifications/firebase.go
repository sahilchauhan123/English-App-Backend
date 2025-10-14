package notifications

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"strings"
)

type ServiceAccount struct {
	Type                    string `json:"type"`
	ProjectID               string `json:"project_id"`
	PrivateKeyID            string `json:"private_key_id"`
	PrivateKey              string `json:"private_key"`
	ClientEmail             string `json:"client_email"`
	ClientID                string `json:"client_id"`
	AuthURI                 string `json:"auth_uri"`
	TokenURI                string `json:"token_uri"`
	AuthProviderX509CertURL string `json:"auth_provider_x509_cert_url"`
	ClientX509CertURL       string `json:"client_x509_cert_url"`
	UniverseDomain          string `json:"universe_domain"`
}

// Creates strango.json at startup if missing
func EnsureFirebaseKeyFile(filePath string) {
	if _, err := os.Stat(filePath); err == nil {
		return
	}

	key := ServiceAccount{
		Type:                    os.Getenv("TYPE"),
		ProjectID:               os.Getenv("PROJECT_ID"),
		PrivateKeyID:            os.Getenv("PRIVATE_KEY_ID"),
		PrivateKey:              strings.ReplaceAll(os.Getenv("PRIVATE_KEY"), "\\n", "\n"),
		ClientEmail:             os.Getenv("CLIENT_EMAIL"),
		ClientID:                os.Getenv("CLIENT_ID"),
		AuthURI:                 os.Getenv("AUTH_URI"),
		TokenURI:                os.Getenv("TOKEN_URI"),
		AuthProviderX509CertURL: os.Getenv("AUTH_PROVIDER_X509_CERT_URL"),
		ClientX509CertURL:       os.Getenv("CLIENT_X509_CERT_URL"),
		UniverseDomain:          os.Getenv("UNIVERSE_DOMAIN"),
	}

	file, err := os.Create(filePath)
	if err != nil {
		log.Fatalf("❌ Failed to create strango.json: %v", err)
	}
	defer file.Close()

	encoder := json.NewEncoder(file)
	encoder.SetIndent("", "  ")
	if err := encoder.Encode(key); err != nil {
		log.Fatalf("❌ Failed to write JSON: %v", err)
	}

	os.Chmod(filePath, 0600)
	fmt.Println("✅ Created strango.json securely from environment variables.")
}
