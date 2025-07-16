package internal

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/vedhavyas/go-subkey/sr25519"
)

func TestIsUserVerified(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/api/v1/status" && r.URL.Query().Get("client_id") == "dummyaddress" {
			w.WriteHeader(http.StatusOK)
			w.Write([]byte(`{"result":{"status":"VERIFIED"}}`))
			return
		}
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"result":{"status":"NOT_VERIFIED"}}`))
	}))
	defer ts.Close()

	client, err := NewKYCClient(ts.URL, "", "", "testdomain")
	if err != nil {
		t.Fatalf("failed to create client: %v", err)
	}

	verified, err := client.IsUserVerified("dummyaddress", nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !verified {
		t.Errorf("expected verified to be true for status VERIFIED")
	}

	verified, err = client.IsUserVerified("otheraddress", nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if verified {
		t.Errorf("expected verified to be false for status NOT_VERIFIED")
	}
}

func TestIsUserVerified_Non200Status(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(`{"error":"server error"}`))
	}))
	defer ts.Close()

	client, _ := NewKYCClient(ts.URL, "", "", "testdomain")
	verified, err := client.IsUserVerified("dummyaddress", nil)
	if err == nil {
		t.Errorf("expected error for non-200 status")
	}
	if verified {
		t.Errorf("expected verified to be false on error")
	}
}

func TestIsUserVerified_MalformedJSON(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`not a json`))
	}))
	defer ts.Close()

	client, _ := NewKYCClient(ts.URL, "", "", "testdomain")
	verified, err := client.IsUserVerified("dummyaddress", nil)
	if err == nil {
		t.Errorf("expected error for malformed JSON")
	}
	if verified {
		t.Errorf("expected verified to be false on error")
	}
}

func TestIsUserVerified_MissingStatusField(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"result":{}}`))
	}))
	defer ts.Close()

	client, _ := NewKYCClient(ts.URL, "", "", "testdomain")
	verified, err := client.IsUserVerified("dummyaddress", nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if verified {
		t.Errorf("expected verified to be false when status field is missing")
	}
}

func TestCreateSponsorship_Success(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost || r.URL.Path != "/api/v1/sponsorships" {
			w.WriteHeader(http.StatusNotFound)
			return
		}
		w.WriteHeader(http.StatusCreated)
		w.Write([]byte(`{"result":"ok"}`))
	}))
	defer ts.Close()

	// Use dummy keypairs
	phrase := "bottom drive obey lake curtain smoke basket hold race lonely fit walk"
	kp, _ := sr25519.Scheme{}.FromPhrase(phrase, "")
	client, _ := NewKYCClient(ts.URL, "sponsor", phrase, "testdomain")
	err := client.CreateSponsorship("sponsee", kp)
	if err != nil {
		t.Errorf("expected no error, got %v", err)
	}
}

func TestCreateSponsorship_Non201(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte(`{"error":"bad request"}`))
	}))
	defer ts.Close()

	phrase := "bottom drive obey lake curtain smoke basket hold race lonely fit walk"
	kp, _ := sr25519.Scheme{}.FromPhrase(phrase, "")
	client, _ := NewKYCClient(ts.URL, "sponsor", phrase, "testdomain")
	err := client.CreateSponsorship("sponsee", kp)
	if err == nil {
		t.Errorf("expected error for non-201 response")
	}
	if !strings.Contains(err.Error(), "bad request") {
		t.Errorf("expected error message to contain 'bad request', got %v", err)
	}
}
