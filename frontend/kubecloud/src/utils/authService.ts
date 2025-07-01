import { api } from './api'

// Types for auth requests and responses
export interface RegisterRequest {
  name: string
  email: string
  password: string
  confirm_password: string
}

export interface RegisterResponse {
  message: string
  timeout: number
}

export interface VerifyCodeRequest {
  email: string
  code: number
}

export interface LoginRequest {
  email: string
  password: string
}

export interface LoginResponse {
  access_token: string
  refresh_token: string
}

// New type to match backend response
export interface BackendLoginResponse {
  message: string
  status: number
  data: LoginResponse
}

export interface RefreshTokenRequest {
  refresh_token: string
}

export interface RefreshTokenResponse {
  access_token: string
  refresh_token: string
}

export interface ForgotPasswordRequest {
  email: string
}

export interface ForgotPasswordResponse {
  message: string
  timeout: number
}

export interface ChangePasswordRequest {
  email: string
  password: string
  confirm_password: string
}

export interface ChangePasswordResponse {
  message: string
}

// Generic API response type
export interface ApiResponse<T> {
  status: number;
  message?: string;
  data: T;
  error?: string;
}

// Auth service class
export class AuthService {
  private static instance: AuthService

  private constructor() {}

  static getInstance(): AuthService {
    if (!AuthService.instance) {
      AuthService.instance = new AuthService()
    }
    return AuthService.instance
  }

  // Register a new user
  async register(data: RegisterRequest): Promise<RegisterResponse> {
    const response = await api.post<ApiResponse<RegisterResponse>>('/v1/user/register', data, {
      showNotifications: true,
      loadingMessage: 'Creating your account...',
      successMessage: 'Verification code sent to your email!',
      errorMessage: 'Registration failed'
    })
    return response.data.data
  }

  // Verify registration code
  async verifyCode(data: VerifyCodeRequest): Promise<LoginResponse> {
    const response = await api.post<ApiResponse<LoginResponse>>('/v1/user/register/verify', data, {
      showNotifications: true,
      successMessage: 'Account verified successfully!',
      errorMessage: 'Verification failed'
    })
    return response.data.data
  }

  // Login user
  async login(data: LoginRequest): Promise<LoginResponse> {
    const response = await api.post<ApiResponse<LoginResponse>>('/v1/user/login', data, {
      showNotifications: true,
      successMessage: 'Welcome back!',
      errorMessage: 'Login failed'
    })
    return response.data.data
  }

  // Refresh access token
  async refreshToken(data: RefreshTokenRequest): Promise<RefreshTokenResponse> {
    const response = await api.post<ApiResponse<RefreshTokenResponse>>('/v1/user/refresh', data, {
      showNotifications: false // Don't show notifications for token refresh
    })
    return response.data.data
  }

  // Forgot password
  async forgotPassword(data: ForgotPasswordRequest): Promise<ForgotPasswordResponse> {
    const response = await api.post<ApiResponse<ForgotPasswordResponse>>('/v1/user/forgot_password', data, {
      showNotifications: true,
      loadingMessage: 'Sending reset code...',
      successMessage: 'Reset code sent to your email!',
      errorMessage: 'Failed to send reset code'
    })
    return response.data.data
  }

  // Verify forgot password code
  async verifyForgotPasswordCode(data: VerifyCodeRequest): Promise<LoginResponse> {
    const response = await api.post<ApiResponse<LoginResponse>>('/v1/user/forgot_password/verify', data, {
      showNotifications: false,
      errorMessage: 'Invalid reset code'
    })
    return response.data.data
  }

  // Change password (requires authentication)
  async changePassword(data: ChangePasswordRequest): Promise<ChangePasswordResponse> {
    const response = await api.post<ApiResponse<ChangePasswordResponse>>('/v1/user/change_password', data, {
      requiresAuth: true,
      showNotifications: true,
      loadingMessage: 'Updating password...',
      successMessage: 'Password updated successfully!',
      errorMessage: 'Failed to update password'
    })
    return response.data.data
  }

  // Store tokens in localStorage
  storeTokens(accessToken: string, refreshToken: string): void {
    localStorage.setItem('token', accessToken)
    localStorage.setItem('refreshToken', refreshToken)
  }

  // Get stored tokens
  getTokens(): { accessToken: string | null; refreshToken: string | null } {
    return {
      accessToken: localStorage.getItem('token'),
      refreshToken: localStorage.getItem('refreshToken')
    }
  }

  // Clear stored tokens
  clearTokens(): void {
    localStorage.removeItem('token')
    localStorage.removeItem('refreshToken')
  }

  // Check if user is authenticated
  isAuthenticated(): boolean {
    return !!localStorage.getItem('token')
  }
}

// Export singleton instance
export const authService = AuthService.getInstance() 