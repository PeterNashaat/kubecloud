package app

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"kubecloud/models"
)

// Helper to get admin token
func getAdminToken(t *testing.T, app *App, id int, email, username string) string {
	return GetAuthToken(t, app, id, email, username, true)
}

func TestListUsersHandler(t *testing.T) {
	app, err := SetUp(t)
	require.NoError(t, err)
	router := app.router

	// Register users
	adminUser := &models.User{
		ID:       1,
		Username: "Admin User",
		Email:    "admin@example.com",
		Password: []byte("securepassword"),
		Verified: true,
		Admin:    true,
	}
	normalUser := &models.User{
		ID:       2,
		Username: "Normal User",
		Email:    "user@example.com",
		Password: []byte("securepassword"),
		Verified: true,
		Admin:    false,
	}
	err = app.handlers.db.RegisterUser(adminUser)
	require.NoError(t, err)
	err = app.handlers.db.RegisterUser(normalUser)
	require.NoError(t, err)

	t.Run("Test List all users successfully", func(t *testing.T) {
		token := getAdminToken(t, app, adminUser.ID, adminUser.Email, adminUser.Username)
		req, _ := http.NewRequest("GET", "/api/v1/users", nil)
		req.Header.Set("Authorization", "Bearer "+token)
		resp := httptest.NewRecorder()
		router.ServeHTTP(resp, req)
		assert.Equal(t, http.StatusOK, resp.Code)

		type UsersResponse struct {
			Message string `json:"message"`
			Data    struct {
				Users []models.User `json:"users"`
			} `json:"data"`
		}

		var usersResp UsersResponse
		err := json.Unmarshal(resp.Body.Bytes(), &usersResp)
		assert.NoError(t, err)
		assert.Equal(t, "Users are retrieved successfully", usersResp.Message)
		users := usersResp.Data.Users
		var foundAdmin, foundUser bool
		for _, user := range users {
			if user.Email == adminUser.Email {
				foundAdmin = true
			}
			if user.Email == normalUser.Email {
				foundUser = true
			}
		}
		assert.True(t, foundAdmin, "Admin user should be in the list")
		assert.True(t, foundUser, "Normal user should be in the list")
	})

	t.Run("Test List users with no token", func(t *testing.T) {
		req, _ := http.NewRequest("GET", "/api/v1/users", nil)
		resp := httptest.NewRecorder()
		router.ServeHTTP(resp, req)
		assert.Equal(t, http.StatusUnauthorized, resp.Code)
	})

	t.Run("Test List users with non-admin credentials", func(t *testing.T) {
		token := GetAuthToken(t, app, normalUser.ID, normalUser.Email, normalUser.Username, false)
		req, _ := http.NewRequest("GET", "/api/v1/users", nil)
		req.Header.Set("Authorization", "Bearer "+token)
		resp := httptest.NewRecorder()
		router.ServeHTTP(resp, req)
		assert.Equal(t, http.StatusForbidden, resp.Code)
	})
}

func TestDeleteUsersHandler(t *testing.T) {
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
	userToDelete := &models.User{
		ID:       2,
		Username: "Delete Me",
		Email:    "deleteme@example.com",
		Password: []byte("securepassword"),
		Verified: true,
		Admin:    false,
	}
	nonAdminUser := &models.User{
		ID:       3,
		Username: "Normal User",
		Email:    "user@example.com",
		Password: []byte("securepassword"),
		Verified: true,
		Admin:    false,
	}
	err = app.handlers.db.RegisterUser(adminUser)
	require.NoError(t, err)
	err = app.handlers.db.RegisterUser(userToDelete)
	require.NoError(t, err)
	err = app.handlers.db.RegisterUser(nonAdminUser)
	require.NoError(t, err)

	t.Run("Test Delete user successfully", func(t *testing.T) {
		token := getAdminToken(t, app, adminUser.ID, adminUser.Email, adminUser.Username)
		req, _ := http.NewRequest("DELETE", "/api/v1/users/2", nil)
		req.Header.Set("Authorization", "Bearer "+token)
		resp := httptest.NewRecorder()
		router.ServeHTTP(resp, req)
		assert.Equal(t, http.StatusOK, resp.Code)
		var result map[string]interface{}
		err := json.Unmarshal(resp.Body.Bytes(), &result)
		assert.NoError(t, err)
		assert.Equal(t, "User is deleted successfully", result["message"])
	})

	t.Run("Test Admin deletes its account", func(t *testing.T) {
		token := getAdminToken(t, app, adminUser.ID, adminUser.Email, adminUser.Username)
		req, _ := http.NewRequest("DELETE", "/api/v1/users/1", nil)
		req.Header.Set("Authorization", "Bearer "+token)
		resp := httptest.NewRecorder()
		router.ServeHTTP(resp, req)
		assert.Equal(t, http.StatusForbidden, resp.Code)
	})

	t.Run("Test Delete with invalid user id", func(t *testing.T) {
		token := getAdminToken(t, app, adminUser.ID, adminUser.Email, adminUser.Username)
		req, _ := http.NewRequest("DELETE", "/api/v1/users/aaa", nil)
		req.Header.Set("Authorization", "Bearer "+token)
		resp := httptest.NewRecorder()
		router.ServeHTTP(resp, req)
		assert.Equal(t, http.StatusBadRequest, resp.Code)
	})

	t.Run("Test Delete with no user id", func(t *testing.T) {
		token := getAdminToken(t, app, adminUser.ID, adminUser.Email, adminUser.Username)
		req, _ := http.NewRequest("DELETE", "/api/v1/users/", nil)
		req.Header.Set("Authorization", "Bearer "+token)
		resp := httptest.NewRecorder()
		router.ServeHTTP(resp, req)
		assert.Equal(t, http.StatusNotFound, resp.Code)
	})

	t.Run("Test Delete non-existing user", func(t *testing.T) {
		token := getAdminToken(t, app, adminUser.ID, adminUser.Email, adminUser.Username)
		req, _ := http.NewRequest("DELETE", "/api/v1/users/9999", nil)
		req.Header.Set("Authorization", "Bearer "+token)
		resp := httptest.NewRecorder()
		router.ServeHTTP(resp, req)
		assert.Equal(t, http.StatusNotFound, resp.Code)
	})

	t.Run("Test Delete user with no token", func(t *testing.T) {
		req, _ := http.NewRequest("DELETE", "/api/v1/users/2", nil)
		resp := httptest.NewRecorder()
		router.ServeHTTP(resp, req)
		assert.Equal(t, http.StatusUnauthorized, resp.Code)
	})

	t.Run("Test Delete with non-admin user", func(t *testing.T) {
		token := GetAuthToken(t, app, nonAdminUser.ID, nonAdminUser.Email, nonAdminUser.Username, false)
		req, _ := http.NewRequest("DELETE", "/api/v1/users/2", nil)
		req.Header.Set("Authorization", "Bearer "+token)
		resp := httptest.NewRecorder()
		router.ServeHTTP(resp, req)
		assert.Equal(t, http.StatusForbidden, resp.Code)
	})
}

func TestGenerateVouchersHandler(t *testing.T) {
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

	t.Run("Test GenerateVouchers successfully", func(t *testing.T) {
		token := getAdminToken(t, app, adminUser.ID, adminUser.Email, adminUser.Username)
		payload := map[string]interface{}{
			"count":             2,
			"value":             10.0,
			"expire_after_days": 7,
		}
		body, _ := json.Marshal(payload)
		req, _ := http.NewRequest("POST", "/api/v1/vouchers/generate", bytes.NewReader(body))
		req.Header.Set("Authorization", "Bearer "+token)
		req.Header.Set("Content-Type", "application/json")
		resp := httptest.NewRecorder()
		router.ServeHTTP(resp, req)
		assert.Equal(t, http.StatusCreated, resp.Code)
		var result map[string]interface{}
		err := json.Unmarshal(resp.Body.Bytes(), &result)
		assert.NoError(t, err)
		assert.Equal(t, "Vouchers are generated successfully", result["message"])
		assert.NotNil(t, result["data"])
		data, ok := result["data"].(map[string]interface{})
		assert.True(t, ok)
		vouchers, ok := data["vouchers"].([]interface{})
		assert.True(t, ok)
		assert.Len(t, vouchers, 2)
	})

	t.Run("Test GenerateVouchers with invalid request format", func(t *testing.T) {
		token := getAdminToken(t, app, adminUser.ID, adminUser.Email, adminUser.Username)
		body, _ := json.Marshal(map[string]interface{}{}) // missing required fields
		req, _ := http.NewRequest("POST", "/api/v1/vouchers/generate", bytes.NewReader(body))
		req.Header.Set("Authorization", "Bearer "+token)
		req.Header.Set("Content-Type", "application/json")
		resp := httptest.NewRecorder()
		router.ServeHTTP(resp, req)
		assert.Equal(t, http.StatusBadRequest, resp.Code)
	})

	t.Run("Test GenerateVouchers with no token", func(t *testing.T) {
		payload := map[string]interface{}{
			"count":             1,
			"value":             5.0,
			"expire_after_days": 3,
		}
		body, _ := json.Marshal(payload)
		req, _ := http.NewRequest("POST", "/api/v1/vouchers/generate", bytes.NewReader(body))
		req.Header.Set("Content-Type", "application/json")
		resp := httptest.NewRecorder()
		router.ServeHTTP(resp, req)
		assert.Equal(t, http.StatusUnauthorized, resp.Code)
	})

	t.Run("Test GenerateVouchers with non-admin user", func(t *testing.T) {
		token := GetAuthToken(t, app, nonAdminUser.ID, nonAdminUser.Email, nonAdminUser.Username, false)
		payload := map[string]interface{}{
			"count":             1,
			"value":             5.0,
			"expire_after_days": 3,
		}
		body, _ := json.Marshal(payload)
		req, _ := http.NewRequest("POST", "/api/v1/vouchers/generate", bytes.NewReader(body))
		req.Header.Set("Authorization", "Bearer "+token)
		req.Header.Set("Content-Type", "application/json")
		resp := httptest.NewRecorder()
		router.ServeHTTP(resp, req)
		assert.Equal(t, http.StatusForbidden, resp.Code)
	})
}

func TestListVouchersHandler(t *testing.T) {
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

	voucher1 := &models.Voucher{
		Code:      "VOUCHER1",
		Value:     10.0,
		CreatedAt: time.Now(),
		ExpiresAt: time.Now().Add(24 * time.Hour),
	}
	voucher2 := &models.Voucher{
		Code:      "VOUCHER2",
		Value:     20.0,
		CreatedAt: time.Now(),
		ExpiresAt: time.Now().Add(48 * time.Hour),
	}
	err = app.handlers.db.CreateVoucher(voucher1)
	require.NoError(t, err)
	err = app.handlers.db.CreateVoucher(voucher2)
	require.NoError(t, err)

	t.Run("Test List Vouchers successfully", func(t *testing.T) {
		token := getAdminToken(t, app, adminUser.ID, adminUser.Email, adminUser.Username)
		req, _ := http.NewRequest("GET", "/api/v1/vouchers", nil)
		req.Header.Set("Authorization", "Bearer "+token)
		resp := httptest.NewRecorder()
		router.ServeHTTP(resp, req)
		assert.Equal(t, http.StatusOK, resp.Code)

		type VouchersResponse struct {
			Message string `json:"message"`
			Data    struct {
				Vouchers []models.Voucher `json:"vouchers"`
			} `json:"data"`
		}

		var vouchersResp VouchersResponse
		err := json.Unmarshal(resp.Body.Bytes(), &vouchersResp)
		assert.NoError(t, err)
		assert.Equal(t, "Vouchers are Retrieved successfully", vouchersResp.Message)
		vouchers := vouchersResp.Data.Vouchers
		var found1, found2 bool
		for _, v := range vouchers {
			if v.Code == voucher1.Code {
				found1 = true
			}
			if v.Code == voucher2.Code {
				found2 = true
			}
		}
		assert.True(t, found1, "Voucher1 should be in the list")
		assert.True(t, found2, "Voucher2 should be in the list")
	})

	t.Run("Test ListVouchersHandler with no token", func(t *testing.T) {
		req, _ := http.NewRequest("GET", "/api/v1/vouchers", nil)
		resp := httptest.NewRecorder()
		router.ServeHTTP(resp, req)
		assert.Equal(t, http.StatusUnauthorized, resp.Code)
	})

	t.Run("Test ListVouchersHandler with non-admin user", func(t *testing.T) {
		token := GetAuthToken(t, app, nonAdminUser.ID, nonAdminUser.Email, nonAdminUser.Username, false)
		req, _ := http.NewRequest("GET", "/api/v1/vouchers", nil)
		req.Header.Set("Authorization", "Bearer "+token)
		resp := httptest.NewRecorder()
		router.ServeHTTP(resp, req)
		assert.Equal(t, http.StatusForbidden, resp.Code)
	})
}

func TestCreditUserHandler(t *testing.T) {
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
		Mnemonic: app.config.SystemAccount.Mnemonic,
	}
	normalUser := &models.User{
		ID:       2,
		Username: "Normal User",
		Email:    "user@example.com",
		Password: []byte("securepassword"),
		Verified: true,
		Admin:    false,
		Mnemonic: app.config.SystemAccount.Mnemonic,
	}
	err = app.handlers.db.RegisterUser(adminUser)
	require.NoError(t, err)
	err = app.handlers.db.RegisterUser(normalUser)
	require.NoError(t, err)

	t.Run("Test Credit user successfully", func(t *testing.T) {
		token := getAdminToken(t, app, adminUser.ID, adminUser.Email, adminUser.Username)
		payload := map[string]interface{}{
			"amount": 100,
			"memo":   "Manual credit",
		}
		body, _ := json.Marshal(payload)
		req, _ := http.NewRequest("POST", "/api/v1/users/2/credit", bytes.NewReader(body))
		req.Header.Set("Authorization", "Bearer "+token)
		req.Header.Set("Content-Type", "application/json")
		resp := httptest.NewRecorder()
		router.ServeHTTP(resp, req)
		assert.Equal(t, http.StatusCreated, resp.Code)
		var result map[string]interface{}
		err := json.Unmarshal(resp.Body.Bytes(), &result)
		assert.NoError(t, err)
		assert.Equal(t, "User is credited successfully", result["message"])
		assert.NotNil(t, result["data"])
		data, ok := result["data"].(map[string]interface{})
		assert.True(t, ok)
		assert.Equal(t, normalUser.Email, data["user"])
		assert.EqualValues(t, 100, data["amount"])
		assert.Equal(t, "Manual credit", data["memo"])
	})

	t.Run("Test Credit user with invalid request format", func(t *testing.T) {
		token := getAdminToken(t, app, adminUser.ID, adminUser.Email, adminUser.Username)
		body, _ := json.Marshal(map[string]interface{}{}) // missing required fields
		req, _ := http.NewRequest("POST", "/api/v1/users/2/credit", bytes.NewReader(body))
		req.Header.Set("Authorization", "Bearer "+token)
		req.Header.Set("Content-Type", "application/json")
		resp := httptest.NewRecorder()
		router.ServeHTTP(resp, req)
		assert.Equal(t, http.StatusBadRequest, resp.Code)
	})

	t.Run("Test Credit user with invalid user id", func(t *testing.T) {
		token := getAdminToken(t, app, adminUser.ID, adminUser.Email, adminUser.Username)
		payload := map[string]interface{}{
			"amount": 100,
			"memo":   "Manual credit",
		}
		body, _ := json.Marshal(payload)
		req, _ := http.NewRequest("POST", "/api/v1/users/abc/credit", bytes.NewReader(body))
		req.Header.Set("Authorization", "Bearer "+token)
		req.Header.Set("Content-Type", "application/json")
		resp := httptest.NewRecorder()
		router.ServeHTTP(resp, req)
		assert.Equal(t, http.StatusBadRequest, resp.Code)
	})

	t.Run("Test Credit non-existing user", func(t *testing.T) {
		token := getAdminToken(t, app, adminUser.ID, adminUser.Email, adminUser.Username)
		payload := map[string]interface{}{
			"amount": 100,
			"memo":   "Manual credit",
		}
		body, _ := json.Marshal(payload)
		req, _ := http.NewRequest("POST", "/api/v1/users/9999/credit", bytes.NewReader(body))
		req.Header.Set("Authorization", "Bearer "+token)
		req.Header.Set("Content-Type", "application/json")
		resp := httptest.NewRecorder()
		router.ServeHTTP(resp, req)
		assert.Equal(t, http.StatusInternalServerError, resp.Code)
	})

	t.Run("Test Credit user with no token", func(t *testing.T) {
		payload := map[string]interface{}{
			"amount": 100,
			"memo":   "Manual credit",
		}
		body, _ := json.Marshal(payload)
		req, _ := http.NewRequest("POST", "/api/v1/users/2/credit", bytes.NewReader(body))
		req.Header.Set("Content-Type", "application/json")
		resp := httptest.NewRecorder()
		router.ServeHTTP(resp, req)
		assert.Equal(t, http.StatusUnauthorized, resp.Code)
	})

	t.Run("Test Credit user with non-admin user", func(t *testing.T) {
		token := GetAuthToken(t, app, normalUser.ID, normalUser.Email, normalUser.Username, false)
		payload := map[string]interface{}{
			"amount": 100,
			"memo":   "Manual credit",
		}
		body, _ := json.Marshal(payload)
		req, _ := http.NewRequest("POST", "/api/v1/users/2/credit", bytes.NewReader(body))
		req.Header.Set("Authorization", "Bearer "+token)
		req.Header.Set("Content-Type", "application/json")
		resp := httptest.NewRecorder()
		router.ServeHTTP(resp, req)
		assert.Equal(t, http.StatusForbidden, resp.Code)
	})
}
