package app

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"kubecloud/models"
)

func TestListAllInvoicesHandler(t *testing.T) {
	app, err := SetUp(t)
	require.NoError(t, err)
	router := app.router

	adminUser := &models.User{
		ID:       1,
		Username: "Admin User",
		Email:    "admin@example.com",
		Password: []byte("securepassword"),
		Verified: true,
		Admin:    true,
	}
	nonAdminUser := &models.User{
		ID:       2,
		Username: "Normal User",
		Email:    "user@example.com",
		Password: []byte("securepassword"),
		Verified: true,
		Admin:    false,
	}
	err = app.handlers.db.RegisterUser(adminUser)
	require.NoError(t, err)
	err = app.handlers.db.RegisterUser(nonAdminUser)
	require.NoError(t, err)

	invoice1 := &models.Invoice{
		UserID:    adminUser.ID,
		Total:     100.0,
		Tax:       10.0,
		CreatedAt: time.Now(),
	}
	invoice2 := &models.Invoice{
		UserID:    nonAdminUser.ID,
		Total:     200.0,
		Tax:       20.0,
		CreatedAt: time.Now(),
	}
	err = app.handlers.db.CreateInvoice(invoice1)
	require.NoError(t, err)
	err = app.handlers.db.CreateInvoice(invoice2)
	require.NoError(t, err)

	t.Run("Test List all invoices successfully", func(t *testing.T) {
		token := GetAuthToken(t, app, adminUser.ID, adminUser.Email, adminUser.Username, true)
		req, _ := http.NewRequest("GET", "/api/v1/invoices", nil)
		req.Header.Set("Authorization", "Bearer "+token)
		resp := httptest.NewRecorder()
		router.ServeHTTP(resp, req)
		assert.Equal(t, http.StatusOK, resp.Code)
		var result map[string]interface{}
		err := json.Unmarshal(resp.Body.Bytes(), &result)
		assert.NoError(t, err)
		assert.Equal(t, "Invoices are retrieved successfully", result["message"])
		assert.NotNil(t, result["data"])
		data, ok := result["data"].(map[string]interface{})
		assert.True(t, ok)
		invoices, ok := data["invoices"].([]interface{})
		assert.True(t, ok)
		var found1, found2 bool
		for _, inv := range invoices {
			invMap, ok := inv.(map[string]interface{})
			if !ok {
				continue
			}
			if int(invMap["user_id"].(float64)) == adminUser.ID {
				found1 = true
			}
			if int(invMap["user_id"].(float64)) == nonAdminUser.ID {
				found2 = true
			}
		}
		assert.True(t, found1, "Admin's invoice should be in the list")
		assert.True(t, found2, "Normal user's invoice should be in the list")
	})

	t.Run("Test List all invoices with no token", func(t *testing.T) {
		req, _ := http.NewRequest("GET", "/api/v1/invoices", nil)
		resp := httptest.NewRecorder()
		router.ServeHTTP(resp, req)
		assert.Equal(t, http.StatusUnauthorized, resp.Code)
	})

	t.Run("Test List all invoices with non-admin user", func(t *testing.T) {
		token := GetAuthToken(t, app, nonAdminUser.ID, nonAdminUser.Email, nonAdminUser.Username, false)
		req, _ := http.NewRequest("GET", "/api/v1/invoices", nil)
		req.Header.Set("Authorization", "Bearer "+token)
		resp := httptest.NewRecorder()
		router.ServeHTTP(resp, req)
		assert.Equal(t, http.StatusForbidden, resp.Code)
	})

	t.Run("Test List all invoices with empty list", func(t *testing.T) {
		app2, err := SetUp(t)
		require.NoError(t, err)
		router2 := app2.router
		err = app2.handlers.db.RegisterUser(adminUser)
		require.NoError(t, err)
		token := GetAuthToken(t, app2, adminUser.ID, adminUser.Email, adminUser.Username, true)
		req, _ := http.NewRequest("GET", "/api/v1/invoices", nil)
		req.Header.Set("Authorization", "Bearer "+token)
		resp := httptest.NewRecorder()
		router2.ServeHTTP(resp, req)
		assert.Equal(t, http.StatusOK, resp.Code)
		var result map[string]interface{}
		err = json.Unmarshal(resp.Body.Bytes(), &result)
		assert.NoError(t, err)
		assert.NotNil(t, result["data"])
		data, ok := result["data"].(map[string]interface{})
		assert.True(t, ok)
		invoices, ok := data["invoices"].([]interface{})
		assert.True(t, ok)
		assert.Len(t, invoices, 0)
	})
}

func TestListUserInvoicesHandler(t *testing.T) {
	app, err := SetUp(t)
	require.NoError(t, err)
	router := app.router

	user := &models.User{
		ID:       1,
		Username: "Test User",
		Email:    "user@example.com",
		Password: []byte("securepassword"),
		Verified: true,
		Admin:    false,
	}
	err = app.handlers.db.RegisterUser(user)
	require.NoError(t, err)

	invoice1 := &models.Invoice{
		UserID:    user.ID,
		Total:     100.0,
		Tax:       10.0,
		CreatedAt: time.Now(),
	}
	err = app.handlers.db.CreateInvoice(invoice1)
	require.NoError(t, err)

	t.Run("Test List user invoices successfully", func(t *testing.T) {
		token := GetAuthToken(t, app, user.ID, user.Email, user.Username, false)
		req, _ := http.NewRequest("GET", "/api/v1/user/invoice/", nil)
		req.Header.Set("Authorization", "Bearer "+token)
		resp := httptest.NewRecorder()
		router.ServeHTTP(resp, req)
		assert.Equal(t, http.StatusOK, resp.Code)
		var result map[string]interface{}
		err := json.Unmarshal(resp.Body.Bytes(), &result)
		assert.NoError(t, err)
		assert.Equal(t, "Invoices are retrieved successfully", result["message"])
	})

	t.Run("Test List user invoices with no token", func(t *testing.T) {
		req, _ := http.NewRequest("GET", "/api/v1/user/invoice/", nil)
		resp := httptest.NewRecorder()
		router.ServeHTTP(resp, req)
		assert.Equal(t, http.StatusUnauthorized, resp.Code)
	})

	t.Run("Test List user invoices with empty list", func(t *testing.T) {
		app2, err := SetUp(t)
		require.NoError(t, err)
		router2 := app2.router
		err = app2.handlers.db.RegisterUser(user)
		require.NoError(t, err)
		token := GetAuthToken(t, app2, user.ID, user.Email, user.Username, false)
		req, _ := http.NewRequest("GET", "/api/v1/user/invoice/", nil)
		req.Header.Set("Authorization", "Bearer "+token)
		resp := httptest.NewRecorder()
		router2.ServeHTTP(resp, req)
		assert.Equal(t, http.StatusOK, resp.Code)
		var result map[string]interface{}
		err = json.Unmarshal(resp.Body.Bytes(), &result)
		assert.NoError(t, err)
		assert.NotNil(t, result["data"])
		data, ok := result["data"].(map[string]interface{})
		assert.True(t, ok)
		invoices, ok := data["invoices"].([]interface{})
		assert.True(t, ok)
		assert.Len(t, invoices, 0)
	})
}

func TestDownloadInvoiceHandler(t *testing.T) {
	app, err := SetUp(t)
	require.NoError(t, err)
	router := app.router

	user1 := &models.User{
		ID:       1,
		Username: "User One",
		Email:    "user1@example.com",
		Password: []byte("securepassword"),
		Verified: true,
		Admin:    false,
	}
	user2 := &models.User{
		ID:       2,
		Username: "User Two",
		Email:    "user2@example.com",
		Password: []byte("securepassword"),
		Verified: true,
		Admin:    false,
	}
	err = app.handlers.db.RegisterUser(user1)
	require.NoError(t, err)
	err = app.handlers.db.RegisterUser(user2)
	require.NoError(t, err)

	invoice := &models.Invoice{
		ID: 1,
		UserID:    user1.ID,
		Total:     100.0,
		Tax:       10.0,
		CreatedAt: time.Now(),
	}
	err = app.handlers.db.CreateInvoice(invoice)
	require.NoError(t, err)

	t.Run("Download an invoice successfully", func(t *testing.T) {
		token := GetAuthToken(t, app, user1.ID, user1.Email, user1.Username, false)
		req, _ := http.NewRequest("GET", fmt.Sprintf("/api/v1/user/invoice/%d", invoice.ID), nil)
		req.Header.Set("Authorization", "Bearer "+token)
		resp := httptest.NewRecorder()
		router.ServeHTTP(resp, req)
		assert.Equal(t, http.StatusOK, resp.Code)
		// assert.Equal(t, "application/pdf", resp.Header().Get("Content-Type"))
		// assert.True(t, len(resp.Body.Bytes()) > 0)
	})

	t.Run("Download invoice with no token", func(t *testing.T) {
		req, _ := http.NewRequest("GET", fmt.Sprintf("/api/v1/user/invoice/%d", invoice.ID), nil)
		resp := httptest.NewRecorder()
		router.ServeHTTP(resp, req)
		assert.Equal(t, http.StatusUnauthorized, resp.Code)
	})

	t.Run("Download non-existing invoice", func(t *testing.T) {
		token := GetAuthToken(t, app, user1.ID, user1.Email, user1.Username, false)
		req, _ := http.NewRequest("GET", "/api/v1/user/invoice/99999", nil)
		req.Header.Set("Authorization", "Bearer "+token)
		resp := httptest.NewRecorder()
		router.ServeHTTP(resp, req)
		assert.Equal(t, http.StatusNotFound, resp.Code)
	})

	t.Run("Download invoice with invalid invoice id", func(t *testing.T) {
		token := GetAuthToken(t, app, user1.ID, user1.Email, user1.Username, false)
		req, _ := http.NewRequest("GET", "/api/v1/user/invoice/abc", nil)
		req.Header.Set("Authorization", "Bearer "+token)
		resp := httptest.NewRecorder()
		router.ServeHTTP(resp, req)
		assert.Equal(t, http.StatusBadRequest, resp.Code)
	})
}

