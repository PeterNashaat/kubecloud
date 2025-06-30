import { defineStore } from 'pinia'
import { ref, computed } from 'vue'
import { authService, type LoginRequest, type RegisterRequest, type VerifyCodeRequest } from '@/utils/authService'

export interface User {
  id: number
  username: string
  email: string
  admin: boolean
  verified: boolean
  updated_at: string
  credit_card_balance: number
  credited_balance: number
}

export interface AuthState {
  user: User | null
  token: string | null
  isLoading: boolean
  error: string | null
}

export const useUserStore = defineStore('user', () => {
  // State
  const user = ref<User | null>(null)
  const token = ref<string | null>(null)
  const isLoading = ref(false)
  const error = ref<string | null>(null)

  // Computed properties
  const isAdmin = computed(() => user.value?.admin)
  const isLoggedIn = computed(() => !!token.value)

  // Actions
  const login = async (email: string, password: string) => {
    isLoading.value = true
    error.value = null

    try {
      const loginData: LoginRequest = { email, password }
      const response = await authService.login(loginData)
      
      // Store tokens
      authService.storeTokens(response.access_token, response.refresh_token)
      
      // Set token in store
      token.value = response.access_token
      
      // Always decode user info from JWT
      try {
        const payload = JSON.parse(atob(response.access_token.split('.')[1]))
        user.value = {
          id: payload.user_id || 0,
          username: payload.username || 'User',
          email: payload.email || '',
          admin: payload.admin || false,
          verified: payload.verified ?? true,
          updated_at: payload.updated_at || new Date().toISOString(),
          credit_card_balance: payload.credit_card_balance || 0,
          credited_balance: payload.credited_balance || 0
        }
        localStorage.setItem('user', JSON.stringify(user.value))
      } catch (e) {
        user.value = {
          id: 0,
          username: 'User',
          email: '',
          admin: false,
          verified: true,
          updated_at: new Date().toISOString(),
          credit_card_balance: 0,
          credited_balance: 0
        }
        localStorage.setItem('user', JSON.stringify(user.value))
      }
    } catch (err) {
      error.value = err instanceof Error ? err.message : 'Login failed'
      throw err
    } finally {
      isLoading.value = false
    }
  }

  const logout = () => {
    user.value = null
    token.value = null
    error.value = null
    // Clear localStorage
    authService.clearTokens()
    localStorage.removeItem('user')
  }

  const register = async (userData: Omit<User, 'id' | 'updated_at' | 'credit_card_balance' | 'credited_balance'>) => {
    isLoading.value = true
    error.value = null

    try {
      const registerData: RegisterRequest = {
        name: userData.username,
        email: userData.email,
        password: userData.username, // This should be the actual password from the form
        confirm_password: userData.username // This should be the actual confirm password from the form
      }
      
      const response = await authService.register(registerData)
      
      // Registration requires email verification, so we don't log in immediately
      return response
    } catch (err) {
      error.value = err instanceof Error ? err.message : 'Registration failed'
      throw err
    } finally {
      isLoading.value = false
    }
  }

  const verifyCode = async (email: string, code: number) => {
    isLoading.value = true
    error.value = null

    try {
      const verifyData: VerifyCodeRequest = { email, code }
      const response = await authService.verifyCode(verifyData)
      
      // Store tokens
      authService.storeTokens(response.access_token, response.refresh_token)
      
      // Set token in store
      token.value = response.access_token
      
      // Always decode user info from JWT
      try {
        const payload = JSON.parse(atob(response.access_token.split('.')[1]))
        user.value = {
          id: payload.user_id || 0,
          username: payload.username || 'User',
          email: payload.email || '',
          admin: payload.admin || false,
          verified: payload.verified ?? true,
          updated_at: payload.updated_at || new Date().toISOString(),
          credit_card_balance: payload.credit_card_balance || 0,
          credited_balance: payload.credited_balance || 0
        }
        localStorage.setItem('user', JSON.stringify(user.value))
      } catch (e) {
        user.value = {
          id: 0,
          username: 'User',
          email: '',
          admin: false,
          verified: true,
          updated_at: new Date().toISOString(),
          credit_card_balance: 0,
          credited_balance: 0
        }
        localStorage.setItem('user', JSON.stringify(user.value))
      }
      
      return response
    } catch (err) {
      error.value = err instanceof Error ? err.message : 'Verification failed'
      throw err
    } finally {
      isLoading.value = false
    }
  }

  const updateProfile = async (updates: Partial<User>) => {
    if (!user.value) throw new Error('User not logged in')

    isLoading.value = true
    error.value = null

    try {
      // TODO: Implement profile update API call
      user.value = { ...user.value, ...updates }
      localStorage.setItem('user', JSON.stringify(user.value))
    } catch (err) {
      error.value = err instanceof Error ? err.message : 'Profile update failed'
      throw err
    } finally {
      isLoading.value = false
    }
  }

  const refreshToken = async () => {
    const { refreshToken } = authService.getTokens()
    if (!refreshToken) return

    try {
      const response = await authService.refreshToken({ refresh_token: refreshToken })
      authService.storeTokens(response.access_token, response.refresh_token)
      token.value = response.access_token
    } catch (err) {
      // If refresh fails, logout user
      logout()
      throw err
    }
  }

  const initializeAuth = () => {
    // Check for stored auth data on app start
    const storedUser = localStorage.getItem('user')
    const { accessToken } = authService.getTokens()

    // Always set token if it exists in localStorage
    if (accessToken) {
      token.value = accessToken
    }

    // Set user data if it exists
    if (storedUser) {
      try {
        user.value = JSON.parse(storedUser)
      } catch (err) {
        // Clear invalid stored data
        localStorage.removeItem('user')
      }
    }
  }

  return {
    // State
    user: computed(() => user.value),
    token: computed(() => token.value),
    isLoading: computed(() => isLoading.value),
    error: computed(() => error.value),

    // Computed
    isAdmin,
    isLoggedIn,

    // Actions
    login,
    logout,
    register,
    verifyCode,
    updateProfile,
    refreshToken,
    initializeAuth,
  }
}) 