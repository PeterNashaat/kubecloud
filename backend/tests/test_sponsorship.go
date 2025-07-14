package main

import (
	"bytes"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/vedhavyas/go-subkey"
	"github.com/vedhavyas/go-subkey/sr25519"
)

// KYCClient holds configuration for tf-kyc-verifier API client
type KYCClient struct {
	APIURL          string
	SponsorAddress  string
	SponsorKeyPair  subkey.KeyPair
	ChallengeDomain string
	HTTPClient      *http.Client
}

// NewKYCClient creates a new KYCClient instance
func NewKYCClient(apiURL, sponsorAddress, sponsorPhrase, challengeDomain string) (*KYCClient, error) {
	kr, err := sr25519.Scheme{}.FromPhrase(sponsorPhrase, "")
	if err != nil {
		return nil, fmt.Errorf("failed to create sponsor keypair: %w", err)
	}
	return &KYCClient{
		APIURL:          apiURL,
		SponsorAddress:  sponsorAddress,
		SponsorKeyPair:  kr,
		ChallengeDomain: challengeDomain,
		HTTPClient:      &http.Client{Timeout: 10 * time.Second},
	}, nil
}

// createChallengeMessage creates the challenge message string
func (c *KYCClient) createChallengeMessage() string {
	timestamp := time.Now().Unix()
	return fmt.Sprintf("%s:%d", c.ChallengeDomain, timestamp)
}

// signMessage signs the message with given keypair and returns hex encoded signature
func signMessage(kr subkey.KeyPair, message string) (string, error) {
	sig, err := kr.Sign([]byte(message)) // Sign raw bytes, not hash
	if err != nil {
		return "", err
	}
	return hex.EncodeToString(sig), nil
}

// CreateSponsorship creates a sponsorship between sponsor and sponsee addresses
func (c *KYCClient) CreateSponsorship(sponseeAddress string, sponseeKeyPair subkey.KeyPair) error {
	log.Printf("[INFO] Creating sponsorship: sponsor=%s, sponsee=%s", c.SponsorAddress, sponseeAddress)
	// Create challenge messages
	sponsorChallenge := c.createChallengeMessage()
	sponseeChallenge := c.createChallengeMessage()
	log.Printf("[DEBUG] Sponsor challenge: %s", sponsorChallenge)
	log.Printf("[DEBUG] Sponsee challenge: %s", sponseeChallenge)

	// Sign challenges
	sponsorSignature, err := signMessage(c.SponsorKeyPair, sponsorChallenge)
	if err != nil {
		log.Printf("[ERROR] Failed to sign sponsor challenge: %v", err)
		return fmt.Errorf("failed to sign sponsor challenge: %w", err)
	}
	sponseeSignature, err := signMessage(sponseeKeyPair, sponseeChallenge)
	if err != nil {
		log.Printf("[ERROR] Failed to sign sponsee challenge: %v", err)
		return fmt.Errorf("failed to sign sponsee challenge: %w", err)
	}

	log.Printf("[DEBUG] Sponsor signature: %s", sponsorSignature)
	log.Printf("[DEBUG] Sponsee signature: %s", sponseeSignature)

	// Prepare HTTP request
	url := fmt.Sprintf("%s/api/v1/sponsorships", c.APIURL)
	log.Printf("[INFO] Sending POST request to %s", url)
	req, err := http.NewRequest("POST", url, bytes.NewBuffer([]byte{}))
	if err != nil {
		log.Printf("[ERROR] Failed to create HTTP request: %v", err)
		return fmt.Errorf("failed to create request: %w", err)
	}

	// Set required headers
	req.Header.Set("X-Client-ID", c.SponsorAddress)
	req.Header.Set("X-Sponsee-ID", sponseeAddress)
	req.Header.Set("X-Challenge", hex.EncodeToString([]byte(sponsorChallenge)))
	req.Header.Set("X-Sponsee-Challenge", hex.EncodeToString([]byte(sponseeChallenge)))
	req.Header.Set("X-Signature", sponsorSignature)
	req.Header.Set("X-Sponsee-Signature", sponseeSignature)

	log.Printf("[DEBUG] Request headers:")
	for k, v := range req.Header {
		log.Printf("  %s: %s", k, v)
	}

	// Send request
	resp, err := c.HTTPClient.Do(req)
	if err != nil {
		log.Printf("[ERROR] Failed to send sponsorship request: %v", err)
		return fmt.Errorf("failed to send sponsorship request: %w", err)
	}
	defer resp.Body.Close()

	bodyBytes, _ := io.ReadAll(resp.Body)
	log.Printf("[INFO] HTTP response status: %s", resp.Status)
	log.Printf("[INFO] HTTP response body: %s", string(bodyBytes))

	if resp.StatusCode != http.StatusCreated {
		return fmt.Errorf("sponsorship creation failed with status: %s and response: %s", resp.Status, string(bodyBytes))
	}

	log.Printf("[SUCCESS] Sponsorship created successfully between sponsor %s and sponsee %s", c.SponsorAddress, sponseeAddress)
	return nil
}

// Config struct for config.json
type Config struct {
	KYCVerifierAPIURL  string `json:"kyc_verifier_api_url"`
	KYCSponsorPhrase   string `json:"kyc_sponsor_phrase"`
	KYCChallengeDomain string `json:"kyc_challenge_domain"`
}

func loadConfig(path string) (*Config, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer f.Close()
	var cfg Config
	if err := json.NewDecoder(f).Decode(&cfg); err != nil {
		return nil, err
	}
	return &cfg, nil
}

func main() {
	// === LOAD CONFIG ===
	cfg, err := loadConfig("../../backend/config.json")
	if err != nil {
		log.Fatalf("[ERROR] Failed to load config: %v", err)
	}
	log.Printf("[INFO] Loaded config: API URL=%s, ChallengeDomain=%s", cfg.KYCVerifierAPIURL, cfg.KYCChallengeDomain)

	// Sponsor
	sponsorPhrase := cfg.KYCSponsorPhrase
	log.Printf("[INFO] Sponsor phrase: %s", sponsorPhrase)
	sponsorKeyPair, err := sr25519.Scheme{}.FromPhrase(sponsorPhrase, "")
	if err != nil {
		log.Fatalf("[ERROR] Failed to create sponsor keypair: %v", err)
	}
	sponsorAddress, err := sponsorKeyPair.SS58Address(42)
	if err != nil {
		log.Fatalf("[ERROR] Failed to get sponsor address: %v", err)
	}
	log.Printf("[DEBUG] Sponsor address from mnemonic: %s", sponsorAddress)

	sponseePhrase := "enter 12 mnemonic phrase here..."
	log.Printf("[INFO] Sponsee phrase: %s", sponseePhrase)
	sponseeKeyPair, err := sr25519.Scheme{}.FromPhrase(sponseePhrase, "")
	if err != nil {
		log.Fatalf("[ERROR] Failed to create sponsee keypair: %v", err)
	}
	sponseeAddress, err := sponseeKeyPair.SS58Address(42)
	if err != nil {
		log.Fatalf("[ERROR] Failed to get sponsee address: %v", err)
	}
	log.Printf("[INFO] Sponsee address: %s", sponseeAddress)

	// Create KYC client
	kycClient, err := NewKYCClient(cfg.KYCVerifierAPIURL, sponsorAddress, sponsorPhrase, cfg.KYCChallengeDomain)
	if err != nil {
		log.Fatalf("[ERROR] Failed to create KYC client: %v", err)
	}

	// Create sponsorship
	err = kycClient.CreateSponsorship(sponseeAddress, sponseeKeyPair)
	if err != nil {
		log.Fatalf("[ERROR] Failed to create sponsorship: %v", err)
	}
	log.Println("[SUCCESS] Sponsorship created successfully!")
}
