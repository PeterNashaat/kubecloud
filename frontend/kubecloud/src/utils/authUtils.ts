import { authService } from './authService'

/**
 * Check if user has a valid access token
 */
export function hasValidToken(): boolean {
  const { accessToken } = authService.getTokens()
  return !!accessToken
}

/**
 * Check if user is authenticated by validating token
 */
export async function isAuthenticated(): Promise<boolean> {
  const { accessToken } = authService.getTokens()
  
  if (!accessToken) {
    return false
  }

  try {
    // Try to refresh the token to validate it
    const { refreshToken } = authService.getTokens()
    if (refreshToken) {
      await authService.refreshToken({ refresh_token: refreshToken })
      return true
    }
    return false
  } catch (error) {
    // Token is invalid, clear it
    authService.clearTokens()
    return false
  }
}

/**
 * Get user authentication status
 */
export function getAuthStatus(): {
  isAuthenticated: boolean
  hasToken: boolean
  tokenExpired: boolean
} {
  const { accessToken, refreshToken } = authService.getTokens()
  
  return {
    isAuthenticated: !!accessToken,
    hasToken: !!accessToken,
    tokenExpired: !accessToken && !!refreshToken // If no access token but has refresh token, access token expired
  }
}

/**
 * Validate and refresh token if needed
 */
export async function validateAndRefreshToken(): Promise<boolean> {
  const { accessToken, refreshToken } = authService.getTokens()
  
  if (!accessToken && !refreshToken) {
    return false
  }

  if (!accessToken && refreshToken) {
    try {
      await authService.refreshToken({ refresh_token: refreshToken })
      return true
    } catch (error) {
      authService.clearTokens()
      return false
    }
  }

  return true
}

/**
 * Redirect user based on authentication status
 */
export function getRedirectPath(isAuthenticated: boolean, isAdmin: boolean = false): string {
  if (isAuthenticated) {
    return isAdmin ? '/admin' : '/dashboard'
  }
  return '/sign-in'
} 