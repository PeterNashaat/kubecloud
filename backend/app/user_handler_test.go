package app

import (
	"bytes"
	"encoding/json"
	"kubecloud/internal"
	"kubecloud/models"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestRegisterHandler(t *testing.T) {
	t.Run("Register User Successfully", func(t *testing.T) {
		app := SetUp(t)
		router := app.router

		payload := map[string]interface{}{
			"name":             "Test User",
			"email":            "testuser@example.com",
			"password":         "securepassword",
			"confirm_password": "securepassword",
		}
		body, _ := json.Marshal(payload)

		req, _ := http.NewRequest("POST", "/api/v1/user/register", bytes.NewReader(body))
		req.Header.Set("Content-Type", "application/json")
		resp := httptest.NewRecorder()
		router.ServeHTTP(resp, req)
		if resp.Code != http.StatusCreated {
			t.Logf("Expected status %d, got %d", http.StatusCreated, resp.Code)
			t.Logf("Response body: %s", resp.Body.String())
		}
		assert.Equal(t, http.StatusCreated, resp.Code)

	})

	t.Run("Register User with Invalid Request Format", func(t *testing.T) {
		app := SetUp(t)
		router := app.router

		body, _ := json.Marshal(map[string]interface{}{})

		req, _ := http.NewRequest("POST", "/api/v1/user/register", bytes.NewReader(body))
		req.Header.Set("Content-Type", "application/json")
		resp := httptest.NewRecorder()
		router.ServeHTTP(resp, req)

		assert.Equal(t, resp.Code, http.StatusBadRequest)

	})

	t.Run("Register Existing Verified User", func(t *testing.T) {
		app := SetUp(t)
		router := app.router

		err := app.handlers.db.RegisterUser(&models.User{
			ID:       1,
			Username: "Test User",
			Email:    "dupe@example.com",
			Password: []byte("securepassword"),
			Verified: true,
		})

		assert.NoError(t, err)

		payload := map[string]interface{}{
			"name":             "Test User",
			"email":            "dupe@example.com",
			"password":         "securepassword",
			"confirm_password": "securepassword",
		}
		body, _ := json.Marshal(payload)

		req, _ := http.NewRequest("POST", "/api/v1/user/register", bytes.NewReader(body))
		req.Header.Set("Content-Type", "application/json")
		resp := httptest.NewRecorder()
		router.ServeHTTP(resp, req)
		assert.Equal(t, http.StatusConflict, resp.Code)

	})
}

func TestVerifyRegisterCode(t *testing.T) {
	t.Run("Test Verify Register Code", func(t *testing.T) {
		app := SetUp(t)
		router := app.router

		err := app.handlers.db.RegisterUser(&models.User{
			ID:        1,
			Username:  "Test User",
			Email:     "dupe@example.com",
			Password:  []byte("securepassword"),
			Code:      123,
			Verified:  false,
			UpdatedAt: time.Now(),
		})

		assert.NoError(t, err)
		payload := map[string]interface{}{
			"email": "dupe@example.com",
			"code":  123,
		}
		body, _ := json.Marshal(payload)

		req, _ := http.NewRequest("POST", "/api/v1/user/register/verify", bytes.NewReader(body))
		req.Header.Set("Content-Type", "application/json")
		resp := httptest.NewRecorder()
		router.ServeHTTP(resp, req)
		assert.Equal(t, http.StatusCreated, resp.Code)

	})

	t.Run("Test Verify Register Code with Invalid request format", func(t *testing.T) {
		app := SetUp(t)
		router := app.router

		payload := map[string]interface{}{
			"email": "dupe@example.com",
		}
		body, _ := json.Marshal(payload)

		req, _ := http.NewRequest("POST", "/api/v1/user/register/verify", bytes.NewReader(body))
		req.Header.Set("Content-Type", "application/json")
		resp := httptest.NewRecorder()
		router.ServeHTTP(resp, req)
		assert.Equal(t, http.StatusBadRequest, resp.Code)
		var result map[string]interface{}
		json.Unmarshal(resp.Body.Bytes(), &result)
		assert.Contains(t, result["message"], "Invalid request format")

	})
	t.Run("Test Verify Register Code with registered user", func(t *testing.T) {
		app := SetUp(t)
		router := app.router

		err := app.handlers.db.RegisterUser(&models.User{
			ID:        1,
			Username:  "Test User",
			Email:     "dupe@example.com",
			Password:  []byte("securepassword"),
			Code:      123,
			Verified:  true,
			UpdatedAt: time.Now(),
		})

		assert.NoError(t, err)
		payload := map[string]interface{}{
			"email": "dupe@example.com",
			"code":  123,
		}
		body, _ := json.Marshal(payload)

		req, _ := http.NewRequest("POST", "/api/v1/user/register/verify", bytes.NewReader(body))
		req.Header.Set("Content-Type", "application/json")
		resp := httptest.NewRecorder()
		router.ServeHTTP(resp, req)
		assert.Equal(t, http.StatusBadRequest, resp.Code)

		var result map[string]interface{}
		json.Unmarshal(resp.Body.Bytes(), &result)
		assert.Contains(t, result["error"], "user already registered")
	})

	t.Run("Test Verify Register Code with wrong code", func(t *testing.T) {
		app := SetUp(t)
		router := app.router

		err := app.handlers.db.RegisterUser(&models.User{
			ID:        1,
			Username:  "Test User",
			Email:     "dupe@example.com",
			Password:  []byte("securepassword"),
			Code:      123,
			Verified:  false,
			UpdatedAt: time.Now(),
		})

		assert.NoError(t, err)
		payload := map[string]interface{}{
			"email": "dupe@example.com",
			"code":  333,
		}
		body, _ := json.Marshal(payload)

		req, _ := http.NewRequest("POST", "/api/v1/user/register/verify", bytes.NewReader(body))
		req.Header.Set("Content-Type", "application/json")
		resp := httptest.NewRecorder()
		router.ServeHTTP(resp, req)
		assert.Equal(t, http.StatusBadRequest, resp.Code)

		var result map[string]interface{}
		json.Unmarshal(resp.Body.Bytes(), &result)
		assert.Contains(t, result["error"], "wrong code")

	})

	t.Run("Test Verify Register Code with expired code", func(t *testing.T) {
		app := SetUp(t)
		router := app.router

		err := app.handlers.db.RegisterUser(&models.User{
			ID:        1,
			Username:  "Test User",
			Email:     "dupe@example.com",
			Password:  []byte("securepassword"),
			Code:      123,
			Verified:  false,
			UpdatedAt: time.Now().Add(-1 * time.Hour),
		})

		assert.NoError(t, err)
		payload := map[string]interface{}{
			"email": "dupe@example.com",
			"code":  123,
		}
		body, _ := json.Marshal(payload)

		req, _ := http.NewRequest("POST", "/api/v1/user/register/verify", bytes.NewReader(body))
		req.Header.Set("Content-Type", "application/json")
		resp := httptest.NewRecorder()
		router.ServeHTTP(resp, req)
		assert.Equal(t, http.StatusBadRequest, resp.Code)

		var result map[string]interface{}
		json.Unmarshal(resp.Body.Bytes(), &result)
		assert.Contains(t, result["error"], "code has expired")

	})

}

func TestLoginUserHandler(t *testing.T) {
	t.Run("Test LoginUserHandler", func(t *testing.T) {
		app := SetUp(t)
		router := app.router

		// Register user
		email := "loginuser@example.com"
		password := "securepassword"
		hashed, _ := internal.HashAndSaltPassword([]byte(password))
		user := &models.User{
			Username: "Login User",
			Email:    email,
			Password: hashed,
			Verified: true,
		}
		err := app.handlers.db.RegisterUser(user)
		assert.NoError(t, err)

		payload := map[string]interface{}{
			"email":    email,
			"password": password,
		}
		body, _ := json.Marshal(payload)
		req, _ := http.NewRequest("POST", "/api/v1/user/login", bytes.NewReader(body))
		req.Header.Set("Content-Type", "application/json")
		resp := httptest.NewRecorder()
		router.ServeHTTP(resp, req)
		assert.Equal(t, http.StatusCreated, resp.Code)

		var result map[string]interface{}
		json.Unmarshal(resp.Body.Bytes(), &result)
		assert.Equal(t, "token pair generated", result["message"])
		assert.NotNil(t, result["data"])
	})

	t.Run("Test LoginUserHandler with Invalid Request Format", func(t *testing.T) {
		app := SetUp(t)
		router := app.router

		body, _ := json.Marshal(map[string]interface{}{"email": "abc"})
		req, _ := http.NewRequest("POST", "/api/v1/user/login", bytes.NewReader(body))
		req.Header.Set("Content-Type", "application/json")
		resp := httptest.NewRecorder()
		router.ServeHTTP(resp, req)
		assert.Equal(t, http.StatusBadRequest, resp.Code)
	})

	t.Run("Test LoginUserHandler with non-existing user", func(t *testing.T) {
		app := SetUp(t)
		router := app.router

		payload := map[string]interface{}{
			"email":    "notfound@example.com",
			"password": "irrelevant",
		}
		body, _ := json.Marshal(payload)
		req, _ := http.NewRequest("POST", "/api/v1/user/login", bytes.NewReader(body))
		req.Header.Set("Content-Type", "application/json")
		resp := httptest.NewRecorder()
		router.ServeHTTP(resp, req)
		assert.Equal(t, http.StatusBadRequest, resp.Code)
		var result map[string]interface{}
		json.Unmarshal(resp.Body.Bytes(), &result)
		assert.Contains(t, result["error"], "email or password is incorrect")
	})

	t.Run("Test LoginUserHandler with wrong password", func(t *testing.T) {
		app := SetUp(t)
		router := app.router

		email := "loginuser2@example.com"
		password := "securepassword"
		hashed, _ := internal.HashAndSaltPassword([]byte(password))
		user := &models.User{
			Username: "Login User2",
			Email:    email,
			Password: hashed,
			Verified: true,
		}
		err := app.handlers.db.RegisterUser(user)
		assert.NoError(t, err)

		payload := map[string]interface{}{
			"email":    email,
			"password": "wrongpassword",
		}
		body, _ := json.Marshal(payload)
		req, _ := http.NewRequest("POST", "/api/v1/user/login", bytes.NewReader(body))
		req.Header.Set("Content-Type", "application/json")
		resp := httptest.NewRecorder()
		router.ServeHTTP(resp, req)
		assert.Equal(t, http.StatusUnauthorized, resp.Code)
		var result map[string]interface{}
		json.Unmarshal(resp.Body.Bytes(), &result)
		assert.Contains(t, result["error"], "email or password is incorrect")
	})
}

func TestRefreshTokenHandler(t *testing.T) {
	t.Run("Test RefreshTokenHandler", func(t *testing.T) {
		app := SetUp(t)
		router := app.router

		email := "refreshtoken@example.com"
		password := "securepassword"
		hashed, _ := internal.HashAndSaltPassword([]byte(password))
		user := &models.User{
			Username: "Refresh User",
			Email:    email,
			Password: hashed,
			Verified: true,
		}
		err := app.handlers.db.RegisterUser(user)
		assert.NoError(t, err)

		tokenPair, _ := app.handlers.tokenManager.CreateTokenPair(1, "Refresh User", false)

		payload := map[string]interface{}{
			"refresh_token": tokenPair.RefreshToken,
		}
		body, _ := json.Marshal(payload)
		req, _ := http.NewRequest("POST", "/api/v1/user/refresh", bytes.NewReader(body))
		req.Header.Set("Content-Type", "application/json")
		resp := httptest.NewRecorder()
		router.ServeHTTP(resp, req)
		assert.Equal(t, http.StatusCreated, resp.Code)

		var result map[string]interface{}
		json.Unmarshal(resp.Body.Bytes(), &result)
		assert.Equal(t, "access token refreshed successfully", result["message"])
		assert.NotNil(t, result["data"])
	})

	t.Run("Test RefreshTokenHandler with Invalid Request Format", func(t *testing.T) {
		app := SetUp(t)
		router := app.router

		body, _ := json.Marshal(map[string]interface{}{})
		req, _ := http.NewRequest("POST", "/api/v1/user/refresh", bytes.NewReader(body))
		req.Header.Set("Content-Type", "application/json")
		resp := httptest.NewRecorder()
		router.ServeHTTP(resp, req)
		assert.Equal(t, http.StatusBadRequest, resp.Code)
	})

	t.Run("Test RefreshTokenHandler with Invalid or Expired Token", func(t *testing.T) {
		app := SetUp(t)
		router := app.router

		payload := map[string]interface{}{
			"refresh_token": "invalidtoken",
		}
		body, _ := json.Marshal(payload)
		req, _ := http.NewRequest("POST", "/api/v1/user/refresh", bytes.NewReader(body))
		req.Header.Set("Content-Type", "application/json")
		resp := httptest.NewRecorder()
		router.ServeHTTP(resp, req)
		assert.Equal(t, http.StatusUnauthorized, resp.Code)
		var result map[string]interface{}
		json.Unmarshal(resp.Body.Bytes(), &result)
		assert.Contains(t, result["error"], "Invalid or expired refresh token")
	})
}

func TestForgotPasswordHandler(t *testing.T) {
	t.Run("Test ForgotPasswordHandler", func(t *testing.T) {
		app := SetUp(t)
		router := app.router

		email := "forgotuser@example.com"
		user := &models.User{
			Username: "Forgot User",
			Email:    email,
			Password: []byte("securepassword"),
			Verified: true,
		}
		err := app.handlers.db.RegisterUser(user)
		assert.NoError(t, err)

		payload := map[string]interface{}{
			"email": email,
		}
		body, _ := json.Marshal(payload)
		req, _ := http.NewRequest("POST", "/api/v1/user/forgot_password", bytes.NewReader(body))
		req.Header.Set("Content-Type", "application/json")
		resp := httptest.NewRecorder()
		router.ServeHTTP(resp, req)
		assert.Equal(t, http.StatusOK, resp.Code)
		var result map[string]interface{}
		json.Unmarshal(resp.Body.Bytes(), &result)
		assert.Equal(t, "Verification code sent", result["message"])
		assert.NotNil(t, result["data"])
	})

	t.Run("Test ForgotPasswordHandler with Invalid Request format", func(t *testing.T) {
		app := SetUp(t)
		router := app.router
		body, _ := json.Marshal(map[string]interface{}{})
		req, _ := http.NewRequest("POST", "/api/v1/user/forgot_password", bytes.NewReader(body))
		req.Header.Set("Content-Type", "application/json")
		resp := httptest.NewRecorder()
		router.ServeHTTP(resp, req)
		assert.Equal(t, http.StatusBadRequest, resp.Code)
	})

	t.Run("Test ForgotPasswordHandler with non-existing user", func(t *testing.T) {
		app := SetUp(t)
		router := app.router
		payload := map[string]interface{}{
			"email": "notfound@example.com",
		}
		body, _ := json.Marshal(payload)
		req, _ := http.NewRequest("POST", "/api/v1/user/forgot_password", bytes.NewReader(body))
		req.Header.Set("Content-Type", "application/json")
		resp := httptest.NewRecorder()
		router.ServeHTTP(resp, req)
		assert.Equal(t, http.StatusNotFound, resp.Code)
		var result map[string]interface{}
		json.Unmarshal(resp.Body.Bytes(), &result)
		assert.Contains(t, result["error"], "failed to get user")
	})

}

func TestVerifyForgetPasswordCodeHandler(t *testing.T) {
	t.Run("Test VerifyForgetPasswordCodeHandler", func(t *testing.T) {
		app := SetUp(t)
		router := app.router

		email := "resetuser@example.com"
		user := &models.User{
			ID:        1,
			Username:  "Reset User",
			Email:     email,
			Password:  []byte("securepassword"),
			Code:      4321,
			Verified:  false,
			UpdatedAt: time.Now(),
		}
		err := app.handlers.db.RegisterUser(user)
		assert.NoError(t, err)
		payload := map[string]interface{}{
			"email": email,
			"code":  4321,
		}
		body, _ := json.Marshal(payload)
		req, _ := http.NewRequest("POST", "/api/v1/user/forgot_password/verify", bytes.NewReader(body))
		req.Header.Set("Content-Type", "application/json")
		resp := httptest.NewRecorder()
		router.ServeHTTP(resp, req)
		assert.Equal(t, http.StatusCreated, resp.Code)
		var result map[string]interface{}
		json.Unmarshal(resp.Body.Bytes(), &result)
		assert.Equal(t, "Verification successful", result["message"])
		assert.NotNil(t, result["data"])
	})

	t.Run("Test VerifyForgetPasswordCodeHandler with Invalid request format", func(t *testing.T) {
		app := SetUp(t)
		router := app.router
		body, _ := json.Marshal(map[string]interface{}{})
		req, _ := http.NewRequest("POST", "/api/v1/user/forgot_password/verify", bytes.NewReader(body))
		req.Header.Set("Content-Type", "application/json")
		resp := httptest.NewRecorder()
		router.ServeHTTP(resp, req)
		assert.Equal(t, http.StatusBadRequest, resp.Code)
	})

	t.Run("Test VerifyForgetPasswordCodeHandler with wrong code", func(t *testing.T) {
		app := SetUp(t)
		router := app.router
		email := "wrongreset@example.com"
		user := &models.User{
			ID:        2,
			Username:  "Wrong Reset",
			Email:     email,
			Password:  []byte("securepassword"),
			Code:      4321,
			Verified:  false,
			UpdatedAt: time.Now(),
		}
		err := app.handlers.db.RegisterUser(user)
		assert.NoError(t, err)
		payload := map[string]interface{}{
			"email": email,
			"code":  9999,
		}
		body, _ := json.Marshal(payload)
		req, _ := http.NewRequest("POST", "/api/v1/user/forgot_password/verify", bytes.NewReader(body))
		req.Header.Set("Content-Type", "application/json")
		resp := httptest.NewRecorder()
		router.ServeHTTP(resp, req)
		assert.Equal(t, http.StatusBadRequest, resp.Code)
		var result map[string]interface{}
		json.Unmarshal(resp.Body.Bytes(), &result)
		assert.Contains(t, result["message"], "Invalid code")
	})

	t.Run("Test VerifyForgetPasswordCodeHandler with expired code", func(t *testing.T) {
		app := SetUp(t)
		router := app.router
		email := "expiredreset@example.com"
		user := &models.User{
			ID:        3,
			Username:  "Expired Reset",
			Email:     email,
			Password:  []byte("securepassword"),
			Code:      4321,
			Verified:  false,
			UpdatedAt: time.Now().Add(-2 * time.Hour),
		}
		err := app.handlers.db.RegisterUser(user)
		assert.NoError(t, err)

		payload := map[string]interface{}{
			"email": email,
			"code":  4321,
		}
		body, _ := json.Marshal(payload)
		req, _ := http.NewRequest("POST", "/api/v1/user/forgot_password/verify", bytes.NewReader(body))
		req.Header.Set("Content-Type", "application/json")
		resp := httptest.NewRecorder()
		router.ServeHTTP(resp, req)
		assert.Equal(t, http.StatusBadRequest, resp.Code)
		var result map[string]interface{}
		json.Unmarshal(resp.Body.Bytes(), &result)
		assert.Contains(t, result["error"], "code has expired")
	})

	t.Run("Test VerifyForgetPasswordCodeHandler with non-existing user", func(t *testing.T) {
		app := SetUp(t)
		router := app.router
		payload := map[string]interface{}{
			"email": "notfoundreset@example.com",
			"code":  4321,
		}
		body, _ := json.Marshal(payload)
		req, _ := http.NewRequest("POST", "/api/v1/user/forgot_password/verify", bytes.NewReader(body))
		req.Header.Set("Content-Type", "application/json")
		resp := httptest.NewRecorder()
		router.ServeHTTP(resp, req)
		assert.Equal(t, http.StatusBadRequest, resp.Code)
		var result map[string]interface{}
		json.Unmarshal(resp.Body.Bytes(), &result)
		assert.Contains(t, result["error"], "record not found")
	})
}

func TestChangePasswordHandler(t *testing.T) {
	t.Run("Test ChangePasswordHandler", func(t *testing.T) {
		app := SetUp(t)
		router := app.router

		email := "changepass@example.com"
		username := "Change Pass"
		user := &models.User{
			ID:        1,
			Username:  username,
			Email:     email,
			Password:  []byte("oldpassword"),
			Code:      5555,
			Verified:  true,
			UpdatedAt: time.Now(),
		}
		err := app.handlers.db.RegisterUser(user)
		assert.NoError(t, err)

		token := GetAuthToken(t, app, 1, email, username, false)

		payload := map[string]interface{}{
			"email":            email,
			"password":         "newsecurepassword",
			"confirm_password": "newsecurepassword",
		}
		body, _ := json.Marshal(payload)
		req, _ := http.NewRequest("PUT", "/api/v1/user/change_password", bytes.NewReader(body))
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("Authorization", "Bearer "+token)

		resp := httptest.NewRecorder()
		router.ServeHTTP(resp, req)
		assert.Equal(t, http.StatusAccepted, resp.Code)
		var result map[string]interface{}
		json.Unmarshal(resp.Body.Bytes(), &result)
		assert.Equal(t, "password is updated successfully", result["message"])
	})

	t.Run("Test ChangePasswordHandler with Invalid Request format", func(t *testing.T) {
		app := SetUp(t)
		router := app.router
		email := "changepass@example.com"
		username := "Change Pass"
		token := GetAuthToken(t, app, 1, email, username, false)
		body, _ := json.Marshal(map[string]interface{}{})
		req, _ := http.NewRequest("PUT", "/api/v1/user/change_password", bytes.NewReader(body))
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("Authorization", "Bearer "+token)
		resp := httptest.NewRecorder()
		router.ServeHTTP(resp, req)
		assert.Equal(t, http.StatusBadRequest, resp.Code)
	})

	t.Run("Test ChangePasswordHandler with passwords mismatch", func(t *testing.T) {
		app := SetUp(t)
		router := app.router
		email := "changepassmismatch@example.com"
		username := "Mismatch"
		user := &models.User{
			ID:        4,
			Username:  username,
			Email:     email,
			Password:  []byte("oldpassword"),
			Code:      5555,
			Verified:  true,
			UpdatedAt: time.Now(),
		}
		err := app.handlers.db.RegisterUser(user)
		assert.NoError(t, err)

		token := GetAuthToken(t, app, 4, email, username, false)

		payload := map[string]interface{}{
			"email":            email,
			"password":         "newsecurepassword",
			"confirm_password": "differentpassword",
			"code":             5555,
		}
		body, _ := json.Marshal(payload)
		req, _ := http.NewRequest("PUT", "/api/v1/user/change_password", bytes.NewReader(body))
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("Authorization", "Bearer "+token)
		resp := httptest.NewRecorder()
		router.ServeHTTP(resp, req)
		assert.Equal(t, http.StatusBadRequest, resp.Code)
	})

}

func TestChargeBalanceHandler(t *testing.T) {
	t.Run("Test ChargeBalance", func(t *testing.T) {
		app := SetUp(t)
		router := app.router

		email := "chargeuser@example.com"
		username := "Charge User"
		stripeCustomerID := "cus_test123"
		user := &models.User{
			ID:               1,
			Username:         username,
			Email:            email,
			Password:         []byte("securepassword"),
			Verified:         true,
			StripeCustomerID: stripeCustomerID,
			Mnemonic:         "winner giant reward damage expose pulse recipe manual brand volcano dry avoid",
		}
		err := app.handlers.db.RegisterUser(user)
		assert.NoError(t, err)
		token := GetAuthToken(t, app, 1, email, username, false)

		payload := map[string]interface{}{
			"card_type":         "visa",
			"payment_method_id": "tok_test",
			"amount":            10,
		}
		body, _ := json.Marshal(payload)
		req, _ := http.NewRequest("POST", "/api/v1/user/balance/charge", bytes.NewReader(body))
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("Authorization", "Bearer "+token)
		resp := httptest.NewRecorder()
		router.ServeHTTP(resp, req)
		assert.Equal(t, http.StatusCreated, resp.Code)
	})

	t.Run("Test ChargeBalance with Invalid Request format", func(t *testing.T) {
		app := SetUp(t)
		router := app.router
		email := "chargeuser@example.com"
		username := "Charge User"
		user := &models.User{
			ID:       1,
			Username: username,
			Email:    email,
			Password: []byte("securepassword"),
			Verified: true,
			Mnemonic: "winner giant reward damage expose pulse recipe manual brand volcano dry avoid",
		}
		err := app.handlers.db.RegisterUser(user)
		assert.NoError(t, err)

		token := GetAuthToken(t, app, 1, email, username, false)
		payload := map[string]interface{}{
			"card_type": "visa",
			"amount":    10,
		}
		body, _ := json.Marshal(payload)
		req, _ := http.NewRequest("POST", "/api/v1/user/balance/charge", bytes.NewReader(body))
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("Authorization", "Bearer "+token)
		resp := httptest.NewRecorder()
		router.ServeHTTP(resp, req)
		assert.Equal(t, http.StatusBadRequest, resp.Code)
	})

	t.Run("Test ChargeBalance with invalid amount", func(t *testing.T) {
		app := SetUp(t)
		router := app.router
		email := "chargeuser3@example.com"
		username := "Charge User3"
		token := GetAuthToken(t, app, 1, email, username, false)
		payload := map[string]interface{}{
			"card_type":     "visa",
			"payment_token": "tok_test",
			"amount":        0,
		}
		body, _ := json.Marshal(payload)
		req, _ := http.NewRequest("POST", "/api/v1/user/balance/charge", bytes.NewReader(body))
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("Authorization", "Bearer "+token)
		resp := httptest.NewRecorder()
		router.ServeHTTP(resp, req)
		assert.Equal(t, http.StatusBadRequest, resp.Code)
	})

	t.Run("Test ChargeBalance with non-existing user", func(t *testing.T) {
		app := SetUp(t)
		router := app.router
		token := GetAuthToken(t, app, 1, "notfound@example.com", "Not Found", false)
		payload := map[string]interface{}{
			"card_type":     "visa",
			"payment_token": "tok_test",
			"amount":        100,
		}
		body, _ := json.Marshal(payload)
		req, _ := http.NewRequest("POST", "/api/v1/user/charge_balance", bytes.NewReader(body))
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("Authorization", "Bearer "+token)
		resp := httptest.NewRecorder()
		router.ServeHTTP(resp, req)
		assert.Equal(t, http.StatusNotFound, resp.Code)
	})
}

func TestGetUserHandler(t *testing.T) {
	t.Run("Test Get user successfully", func(t *testing.T) {
		app := SetUp(t)
		router := app.router

		email := "getuser@example.com"
		username := "Get User"
		user := &models.User{
			ID:       1,
			Username: username,
			Email:    email,
			Password: []byte("securepassword"),
			Verified: true,
		}
		err := app.handlers.db.RegisterUser(user)
		assert.NoError(t, err)

		token := GetAuthToken(t, app, 1, email, username, false)

		req, _ := http.NewRequest("GET", "/api/v1/user/", nil)
		req.Header.Set("Authorization", "Bearer "+token)
		resp := httptest.NewRecorder()
		router.ServeHTTP(resp, req)
		assert.Equal(t, http.StatusOK, resp.Code)

		var result map[string]interface{}
		err = json.Unmarshal(resp.Body.Bytes(), &result)
		assert.NoError(t, err)
		assert.Equal(t, "User is retrieved successfully", result["message"])
		assert.NotNil(t, result["data"])
		userData := result["data"].(map[string]interface{})["user"].(map[string]interface{})
		assert.Equal(t, email, userData["email"])
		assert.Equal(t, username, userData["username"])
	})

	t.Run("Test Get non-existing user", func(t *testing.T) {
		app := SetUp(t)
		router := app.router
		token := GetAuthToken(t, app, 999, "notfound@example.com", "Not Found", false)
		req, _ := http.NewRequest("GET", "/api/v1/user/", nil)
		req.Header.Set("Authorization", "Bearer "+token)
		resp := httptest.NewRecorder()
		router.ServeHTTP(resp, req)
		assert.Equal(t, http.StatusNotFound, resp.Code)
		var result map[string]interface{}
		_ = json.Unmarshal(resp.Body.Bytes(), &result)
		assert.Contains(t, result["message"], "User is not found")
	})

}

func TestGetUserBalanceHandler(t *testing.T) {
	t.Run("Test Get balance successfully", func(t *testing.T) {
		app := SetUp(t)
		router := app.router

		email := "balanceuser@example.com"
		username := "Balance User"
		mnemonic := "winner giant reward damage expose pulse recipe manual brand volcano dry avoid"
		user := &models.User{
			ID:       1,
			Username: username,
			Email:    email,
			Password: []byte("securepassword"),
			Verified: true,
			Mnemonic: mnemonic,
			Debt:     42.5,
		}
		err := app.handlers.db.RegisterUser(user)
		assert.NoError(t, err)

		token := GetAuthToken(t, app, 1, email, username, false)

		req, _ := http.NewRequest("GET", "/api/v1/user/balance", nil)
		req.Header.Set("Authorization", "Bearer "+token)
		resp := httptest.NewRecorder()
		router.ServeHTTP(resp, req)
		assert.Equal(t, http.StatusOK, resp.Code)

		var result map[string]interface{}
		err = json.Unmarshal(resp.Body.Bytes(), &result)
		assert.NoError(t, err)
		assert.Equal(t, "Balance is fetched", result["message"])
		assert.NotNil(t, result["data"])
		data := result["data"].(map[string]interface{})
		assert.Contains(t, data, "balance_usd")
		assert.Contains(t, data, "debt_usd")
	})

	t.Run("Test Get balance for non-existing user", func(t *testing.T) {
		app := SetUp(t)
		router := app.router
		token := GetAuthToken(t, app, 999, "notfound@example.com", "Not Found", false)
		req, _ := http.NewRequest("GET", "/api/v1/user/balance", nil)
		req.Header.Set("Authorization", "Bearer "+token)
		resp := httptest.NewRecorder()
		router.ServeHTTP(resp, req)
		assert.Equal(t, http.StatusNotFound, resp.Code)
		var result map[string]interface{}
		_ = json.Unmarshal(resp.Body.Bytes(), &result)
		assert.Contains(t, result["message"], "User is not found")
	})

}

func TestRedeemVoucherHandler(t *testing.T) {
	t.Run("Test redeem voucher successfully", func(t *testing.T) {
		app := SetUp(t)
		router := app.router

		email := "voucheruser@example.com"
		username := "Voucher User"
		mnemonic := "winner giant reward damage expose pulse recipe manual brand volcano dry avoid"
		user := &models.User{
			ID:       1,
			Username: username,
			Email:    email,
			Password: []byte("securepassword"),
			Verified: true,
			Mnemonic: mnemonic,
		}
		err := app.handlers.db.RegisterUser(user)
		assert.NoError(t, err)

		voucher := &models.Voucher{
			ID:        1,
			Code:      "VOUCHER123",
			Value:     50.0,
			Redeemed:  false,
			CreatedAt: time.Now(),
			ExpiresAt: time.Now().Add(1 * time.Hour),
		}
		err = app.handlers.db.CreateVoucher(voucher)
		assert.NoError(t, err)

		token := GetAuthToken(t, app, 1, email, username, false)
		req, _ := http.NewRequest("PUT", "/api/v1/user/redeem/VOUCHER123", nil)
		req.Header.Set("Authorization", "Bearer "+token)
		resp := httptest.NewRecorder()
		router.ServeHTTP(resp, req)
		assert.Equal(t, http.StatusOK, resp.Code)
		var result map[string]interface{}
		err = json.Unmarshal(resp.Body.Bytes(), &result)
		assert.NoError(t, err)
		assert.Equal(t, "Voucher is redeemed successfully. TFT transfer in progress.", result["message"])
		assert.NotNil(t, result["data"])
		assert.NotEmpty(t, result["data"].(map[string]interface{})["workflow_id"])
	})

	t.Run("Test redeem non-existing voucher", func(t *testing.T) {
		app := SetUp(t)
		router := app.router
		email := "voucheruser2@example.com"
		username := "Voucher User2"
		user := &models.User{
			ID:       2,
			Username: username,
			Email:    email,
			Password: []byte("securepassword"),
			Verified: true,
			Mnemonic: "winner giant reward damage expose pulse recipe manual brand volcano dry avoid",
		}
		err := app.handlers.db.RegisterUser(user)
		assert.NoError(t, err)
		token := GetAuthToken(t, app, 2, email, username, false)
		req, _ := http.NewRequest("PUT", "/api/v1/user/redeem/Voucher123", nil)
		req.Header.Set("Authorization", "Bearer "+token)
		resp := httptest.NewRecorder()
		router.ServeHTTP(resp, req)
		assert.Equal(t, http.StatusNotFound, resp.Code)
	})

	t.Run("Test redeem already redeemed voucher", func(t *testing.T) {
		app := SetUp(t)
		router := app.router
		email := "voucheruser3@example.com"
		username := "Voucher User3"
		user := &models.User{
			ID:       3,
			Username: username,
			Email:    email,
			Password: []byte("securepassword"),
			Verified: true,
			Mnemonic: "winner giant reward damage expose pulse recipe manual brand volcano dry avoid",
		}
		err := app.handlers.db.RegisterUser(user)
		assert.NoError(t, err)
		voucher := &models.Voucher{
			ID:        2,
			Code:      "REDEEMEDVOUCHER",
			Value:     30.0,
			Redeemed:  true,
			CreatedAt: time.Now(),
			ExpiresAt: time.Now().Add(1 * time.Hour),
		}
		err = app.handlers.db.CreateVoucher(voucher)
		assert.NoError(t, err)
		token := GetAuthToken(t, app, 3, email, username, false)
		req, _ := http.NewRequest("PUT", "/api/v1/user/redeem/REDEEMEDVOUCHER", nil)
		req.Header.Set("Authorization", "Bearer "+token)
		resp := httptest.NewRecorder()
		router.ServeHTTP(resp, req)
		assert.Equal(t, http.StatusBadRequest, resp.Code)
	})

	t.Run("Test redeem expired voucher", func(t *testing.T) {
		app := SetUp(t)
		router := app.router
		email := "voucheruser4@example.com"
		username := "Voucher User4"
		user := &models.User{
			ID:       4,
			Username: username,
			Email:    email,
			Password: []byte("securepassword"),
			Verified: true,
			Mnemonic: "winner giant reward damage expose pulse recipe manual brand volcano dry avoid",
		}
		err := app.handlers.db.RegisterUser(user)
		assert.NoError(t, err)
		voucher := &models.Voucher{
			ID:        3,
			Code:      "EXPIREDVOUCHER",
			Value:     20.0,
			Redeemed:  false,
			CreatedAt: time.Now().Add(-2 * time.Hour),
			ExpiresAt: time.Now().Add(-1 * time.Hour),
		}
		err = app.handlers.db.CreateVoucher(voucher)
		assert.NoError(t, err)
		token := GetAuthToken(t, app, 4, email, username, false)
		req, _ := http.NewRequest("PUT", "/api/v1/user/redeem/EXPIREDVOUCHER", nil)
		req.Header.Set("Authorization", "Bearer "+token)
		resp := httptest.NewRecorder()
		router.ServeHTTP(resp, req)
		assert.Equal(t, http.StatusBadRequest, resp.Code)
	})

	t.Run("Test redeem voucher for non-existing user", func(t *testing.T) {
		app := SetUp(t)
		router := app.router
		token := GetAuthToken(t, app, 999, "notfound@example.com", "Not Found", false)
		req, _ := http.NewRequest("PUT", "/api/v1/user/redeem/VOUCHER123", nil)
		req.Header.Set("Authorization", "Bearer "+token)
		resp := httptest.NewRecorder()
		router.ServeHTTP(resp, req)
		assert.Equal(t, http.StatusNotFound, resp.Code)
	})

	t.Run("Test redeem voucher with missing code param", func(t *testing.T) {
		app := SetUp(t)
		router := app.router
		email := "voucheruser5@example.com"
		username := "Voucher User5"
		user := &models.User{
			ID:       5,
			Username: username,
			Email:    email,
			Password: []byte("securepassword"),
			Verified: true,
			Mnemonic: "winner giant reward damage expose pulse recipe manual brand volcano dry avoid",
		}
		err := app.handlers.db.RegisterUser(user)
		assert.NoError(t, err)
		token := GetAuthToken(t, app, 5, email, username, false)
		req, _ := http.NewRequest("PUT", "/api/v1/user/redeem/", nil)
		req.Header.Set("Authorization", "Bearer "+token)
		resp := httptest.NewRecorder()
		router.ServeHTTP(resp, req)
		assert.True(t, resp.Code == http.StatusNotFound)
	})

}

func TestListSSHKeysHandler(t *testing.T) {
	t.Run("Test list SSH keys empty", func(t *testing.T) {
		app := SetUp(t)
		router := app.router
		email := "sshuser@example.com"
		username := "SSH User"
		user := &models.User{
			ID:       1,
			Username: username,
			Email:    email,
			Password: []byte("securepassword"),
			Verified: true,
		}
		err := app.handlers.db.RegisterUser(user)
		assert.NoError(t, err)
		token := GetAuthToken(t, app, 1, email, username, false)
		req, _ := http.NewRequest("GET", "/api/v1/user/ssh-keys", nil)
		req.Header.Set("Authorization", "Bearer "+token)
		resp := httptest.NewRecorder()
		router.ServeHTTP(resp, req)
		assert.Equal(t, http.StatusOK, resp.Code)
		var result map[string]interface{}
		err = json.Unmarshal(resp.Body.Bytes(), &result)
		assert.NoError(t, err)
		assert.Equal(t, "SSH keys retrieved successfully", result["message"])
		assert.NotNil(t, result["data"])
		sshKeys := result["data"].([]interface{})
		assert.Len(t, sshKeys, 0)
	})

	t.Run("Test list SSH keys with multiple keys", func(t *testing.T) {
		app := SetUp(t)
		router := app.router
		email := "sshuser2@example.com"
		username := "SSH User2"
		user := &models.User{
			ID:       2,
			Username: username,
			Email:    email,
			Password: []byte("securepassword"),
			Verified: true,
		}
		err := app.handlers.db.RegisterUser(user)
		assert.NoError(t, err)
		sshKey1 := &models.SSHKey{
			UserID:    2,
			Name:      "key1",
			PublicKey: "ssh-rsa AAAAB3NzaC1yc2EAAAADAQABAAABAQC1",
		}
		sshKey2 := &models.SSHKey{
			UserID:    2,
			Name:      "key2",
			PublicKey: "ssh-rsa AAAAB3NzaC1yc2EAAAADAQABAAABAQC2",
		}
		err = app.handlers.db.CreateSSHKey(sshKey1)
		assert.NoError(t, err)
		err = app.handlers.db.CreateSSHKey(sshKey2)
		assert.NoError(t, err)
		token := GetAuthToken(t, app, 2, email, username, false)
		req, _ := http.NewRequest("GET", "/api/v1/user/ssh-keys", nil)
		req.Header.Set("Authorization", "Bearer "+token)
		resp := httptest.NewRecorder()
		router.ServeHTTP(resp, req)
		assert.Equal(t, http.StatusOK, resp.Code)
		var result map[string]interface{}
		err = json.Unmarshal(resp.Body.Bytes(), &result)
		assert.NoError(t, err)
		assert.Equal(t, "SSH keys retrieved successfully", result["message"])
		assert.NotNil(t, result["data"])
		sshKeys := result["data"].([]interface{})
		assert.Len(t, sshKeys, 2)
	})
}

func TestAddSSHKeyHandler(t *testing.T) {
	t.Run("Add SSH key successfully", func(t *testing.T) {
		app := SetUp(t)
		router := app.router
		email := "addsshuser@example.com"
		username := "Add SSH User"
		user := &models.User{
			ID:       10,
			Username: username,
			Email:    email,
			Password: []byte("securepassword"),
			Verified: true,
		}
		err := app.handlers.db.RegisterUser(user)
		assert.NoError(t, err)
		token := GetAuthToken(t, app, 10, email, username, false)
		payload := map[string]interface{}{
			"name":       "mykey",
			"public_key": "ssh-rsa AAAAB3NzaC1yc2EAAAADAQABAAABgQDzy9yGz+CsKhjYB3FLr27SaoPQVi/tOZDZ06LnO7NuVUj0yR3e7IJO26cxs6j7tRAGTrA7choRMlQJdCFQfkDCaAL+31fPSihHhB3kxUTnZymaWgZ6s/JxjI/2/kKcLjxMWpMYTs18ZdRJf1DgoiyTV6yhlxAhWJvMxTtC5++h5+Ir7mHoN5QdrRt5AjKEcTEJjoKC3it4itHz7w45hi4y07kFYIk4HcMGrInh1IC/BriU7xKlwYcP2tp0W4GIraDJoD8OR3cgcYd/AFXSnVDtomCq5MaKBUli6FWLCK7E3+0AtYxxLkQ/zFkPsYSFAGGqVp8uq2hI46d0TxhgcG2EsWiF/2yOjtMdX1ab3Ns23p8Q0l/8JxXn6WT9xhme9eb2v8UjukN0AR8j+hp5xoQuSEgXAxkg4PFEa2seYEcE8xZPOSavuQl4wEAjXH/1BHnRHxrBBWixN2xdclHRAKQRwR+EHg8wDQ0EAAxtoCCAVHOepBrmV0JDxJGHQ8euvbs= test@gmail.com",
		}
		body, _ := json.Marshal(payload)
		req, _ := http.NewRequest("POST", "/api/v1/user/ssh-keys", bytes.NewReader(body))
		req.Header.Set("Authorization", "Bearer "+token)
		req.Header.Set("Content-Type", "application/json")
		resp := httptest.NewRecorder()
		router.ServeHTTP(resp, req)
		assert.Equal(t, http.StatusCreated, resp.Code)
		var result map[string]interface{}
		err = json.Unmarshal(resp.Body.Bytes(), &result)
		assert.NoError(t, err)
		assert.Equal(t, "SSH key added successfully", result["message"])
		assert.NotNil(t, result["data"])
	})

	t.Run("Add SSH key with invalid request format", func(t *testing.T) {
		app := SetUp(t)
		router := app.router
		email := "addsshuser2@example.com"
		username := "Add SSH User2"
		user := &models.User{
			ID:       11,
			Username: username,
			Email:    email,
			Password: []byte("securepassword"),
			Verified: true,
		}
		err := app.handlers.db.RegisterUser(user)
		assert.NoError(t, err)
		token := GetAuthToken(t, app, 11, email, username, false)
		// Missing public_key
		payload := map[string]interface{}{
			"name": "mykey2",
		}
		body, _ := json.Marshal(payload)
		req, _ := http.NewRequest("POST", "/api/v1/user/ssh-keys", bytes.NewReader(body))
		req.Header.Set("Authorization", "Bearer "+token)
		req.Header.Set("Content-Type", "application/json")
		resp := httptest.NewRecorder()
		router.ServeHTTP(resp, req)
		assert.Equal(t, http.StatusBadRequest, resp.Code)
	})

	t.Run("Add SSH key with invalid SSH key format", func(t *testing.T) {
		app := SetUp(t)
		router := app.router
		email := "addsshuser3@example.com"
		username := "Add SSH User3"
		user := &models.User{
			ID:       12,
			Username: username,
			Email:    email,
			Password: []byte("securepassword"),
			Verified: true,
		}
		err := app.handlers.db.RegisterUser(user)
		assert.NoError(t, err)
		token := GetAuthToken(t, app, 12, email, username, false)
		payload := map[string]interface{}{
			"name":       "badkey",
			"public_key": "not-a-valid-ssh-key",
		}
		body, _ := json.Marshal(payload)
		req, _ := http.NewRequest("POST", "/api/v1/user/ssh-keys", bytes.NewReader(body))
		req.Header.Set("Authorization", "Bearer "+token)
		req.Header.Set("Content-Type", "application/json")
		resp := httptest.NewRecorder()
		router.ServeHTTP(resp, req)
		assert.Equal(t, http.StatusBadRequest, resp.Code)
		var result map[string]interface{}
		err = json.Unmarshal(resp.Body.Bytes(), &result)
		assert.NoError(t, err)
		assert.Contains(t, result["message"], "Validation Error")
	})

	t.Run("Add SSH key with duplicate public key", func(t *testing.T) {
		app := SetUp(t)
		router := app.router
		email := "addsshuser5@example.com"
		username := "Add SSH User5"
		user := &models.User{
			ID:       14,
			Username: username,
			Email:    email,
			Password: []byte("securepassword"),
			Verified: true,
		}
		err := app.handlers.db.RegisterUser(user)
		assert.NoError(t, err)
		token := GetAuthToken(t, app, 14, email, username, false)
		payload1 := map[string]interface{}{
			"name":       "keyA",
			"public_key": "ssh-rsa AAAAB3NzaC1yc2EAAAADAQABAAABgQDzy9yGz+CsKhjYB3FLr27SaoPQVi/tOZDZ06LnO7NuVUj0yR3e7IJO26cxs6j7tRAGTrA7choRMlQJdCFQfkDCaAL+31fPSihHhB3kxUTnZymaWgZ6s/JxjI/2/kKcLjxMWpMYTs18ZdRJf1DgoiyTV6yhlxAhWJvMxTtC5++h5+Ir7mHoN5QdrRt5AjKEcTEJjoKC3it4itHz7w45hi4y07kFYIk4HcMGrInh1IC/BriU7xKlwYcP2tp0W4GIraDJoD8OR3cgcYd/AFXSnVDtomCq5MaKBUli6FWLCK7E3+0AtYxxLkQ/zFkPsYSFAGGqVp8uq2hI46d0TxhgcG2EsWiF/2yOjtMdX1ab3Ns23p8Q0l/8JxXn6WT9xhme9eb2v8UjukN0AR8j+hp5xoQuSEgXAxkg4PFEa2seYEcE8xZPOSavuQl4wEAjXH/1BHnRHxrBBWixN2xdclHRAKQRwR+EHg8wDQ0EAAxtoCCAVHOepBrmV0JDxJGHQ8euvbs= test@gmail.com",
		}
		payload2 := map[string]interface{}{
			"name":       "keyA",
			"public_key": "ssh-rsa AAAAB3NzaC1yc2EAAAADAQABAAABgQDzy9yGz+CsKhjYB3FLr27SaoPQVi/tOZDZ06LnO7NuVUj0yR3e7IJO26cxs6j7tRAGTrA7choRMlQJdCFQfkDCaAL+31fPSihHhB3kxUTnZymaWgZ6s/JxjI/2/kKcLjxMWpMYTs18ZdRJf1DgoiyTV6yhlxAhWJvMxTtC5++h5+Ir7mHoN5QdrRt5AjKEcTEJjoKC3it4itHz7w45hi4y07kFYIk4HcMGrInh1IC/BriU7xKlwYcP2tp0W4GIraDJoD8OR3cgcYd/AFXSnVDtomCq5MaKBUli6FWLCK7E3+0AtYxxLkQ/zFkPsYSFAGGqVp8uq2hI46d0TxhgcG2EsWiF/2yOjtMdX1ab3Ns23p8Q0l/8JxXn6WT9xhme9eb2v8UjukN0AR8j+hp5xoQuSEgXAxkg4PFEa2seYEcE8xZPOSavuQl4wEAjXH/1BHnRHxrBBWixN2xdclHRAKQRwR+EHg8wDQ0EAAxtoCCAVHOepBrmV0JDxJGHQ8euvbs= test@gmail.com",
		}
		body1, _ := json.Marshal(payload1)
		req1, _ := http.NewRequest("POST", "/api/v1/user/ssh-keys", bytes.NewReader(body1))
		req1.Header.Set("Authorization", "Bearer "+token)
		req1.Header.Set("Content-Type", "application/json")
		resp1 := httptest.NewRecorder()
		router.ServeHTTP(resp1, req1)
		assert.Equal(t, http.StatusCreated, resp1.Code)
		body2, _ := json.Marshal(payload2)
		req2, _ := http.NewRequest("POST", "/api/v1/user/ssh-keys", bytes.NewReader(body2))
		req2.Header.Set("Authorization", "Bearer "+token)
		req2.Header.Set("Content-Type", "application/json")
		resp2 := httptest.NewRecorder()
		router.ServeHTTP(resp2, req2)
		assert.Equal(t, http.StatusInternalServerError, resp2.Code)
	})

}
