package internal

import (
	"bytes"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/rs/zerolog/log"
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
	var kr subkey.KeyPair
	var err error
	if sponsorPhrase != "" {
		kr, err = sr25519.Scheme{}.FromPhrase(sponsorPhrase, "")
		if err != nil {
			return nil, fmt.Errorf("failed to create sponsor keypair: %w", err)
		}
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
	// Create challenge messages
	sponsorChallenge := c.createChallengeMessage()
	sponseeChallenge := c.createChallengeMessage()

	// Sign challenges
	sponsorSignature, err := signMessage(c.SponsorKeyPair, sponsorChallenge)
	if err != nil {
		return fmt.Errorf("failed to sign sponsor challenge: %w", err)
	}
	sponseeSignature, err := signMessage(sponseeKeyPair, sponseeChallenge)
	if err != nil {
		return fmt.Errorf("failed to sign sponsee challenge: %w", err)
	}

	// Debug logs for troubleshooting
	log.Debug().Msgf("KYC Sponsorship Debug: sponsorAddress=%s, sponseeAddress=%s", c.SponsorAddress, sponseeAddress)
	log.Debug().Msgf("KYC Sponsorship Debug: sponsorChallenge=%s, sponseeChallenge=%s", sponsorChallenge, sponseeChallenge)
	log.Debug().Msgf("KYC Sponsorship Debug: sponsorSignature=%s, sponseeSignature=%s", sponsorSignature, sponseeSignature)

	// Prepare HTTP request
	url := fmt.Sprintf("%s/api/v1/sponsorships", c.APIURL)
	req, err := http.NewRequest("POST", url, bytes.NewBuffer([]byte{}))
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	// Set required headers
	req.Header.Set("X-Client-ID", c.SponsorAddress)
	req.Header.Set("X-Sponsee-ID", sponseeAddress)
	req.Header.Set("X-Challenge", hex.EncodeToString([]byte(sponsorChallenge)))
	req.Header.Set("X-Sponsee-Challenge", hex.EncodeToString([]byte(sponseeChallenge)))
	req.Header.Set("X-Signature", sponsorSignature)
	req.Header.Set("X-Sponsee-Signature", sponseeSignature)

	// Send request
	resp, err := c.HTTPClient.Do(req)
	if err != nil {
		return fmt.Errorf("failed to send sponsorship request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusCreated {
		// Read response body for error details
		bodyBytes := new(bytes.Buffer)
		_, err := bodyBytes.ReadFrom(resp.Body)
		if err != nil {
			return fmt.Errorf("sponsorship creation failed with status: %s and failed to read response body: %w", resp.Status, err)
		}
		return fmt.Errorf("sponsorship creation failed with status: %s and response: %s", resp.Status, bodyBytes.String())
	}

	log.Info().Msgf("Sponsorship created successfully between sponsor %s and sponsee %s", c.SponsorAddress, sponseeAddress)
	return nil
}

// IsUserVerified checks if a user is verified (directly or via sponsorship) by calling the tf-kyc-verifier API
func (c *KYCClient) IsUserVerified(address string, keyPair subkey.KeyPair) (bool, error) {
	url := fmt.Sprintf("%s/api/v1/status?client_id=%s", c.APIURL, address)
	resp, err := c.HTTPClient.Get(url)
	if err != nil {
		return false, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return false, fmt.Errorf("KYC verifier returned status: %s", resp.Status)
	}

	var result struct {
		Result struct {
			Status string `json:"status"`
		} `json:"result"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return false, err
	}

	return result.Result.Status == "VERIFIED", nil
}
