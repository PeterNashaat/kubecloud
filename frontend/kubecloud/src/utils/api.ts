import { WorkflowStatus } from '@/types/ewf'
import { useNotificationStore } from '../stores/notifications'
import { useUserStore } from '../stores/user'
import { useRouter } from 'vue-router'

export interface ApiResponse<T = any> {
  data: T
  status: number
  message: string
}

export interface ApiError {
  message: string
  status?: number
  code?: string
}

export interface ApiOptions {
  method?: 'GET' | 'POST' | 'PUT' | 'DELETE' | 'PATCH'
  headers?: Record<string, string>
  body?: any
  timeout?: number
  showNotifications?: boolean
  loadingMessage?: string
  successMessage?: string
  errorMessage?: string
  requiresAuth?: boolean
}

class ApiClient {
  private baseURL: string
  private defaultTimeout: number

  constructor(baseURL?: string, timeout: number = 10000) {
    this.baseURL = baseURL || (typeof window !== 'undefined' && (window as any).__ENV__?.VITE_API_BASE_URL) || import.meta.env.VITE_API_BASE_URL || 'http://localhost:8080/api'
    this.defaultTimeout = timeout
  }

  private getAuthHeaders(): Record<string, string> {
    const token = localStorage.getItem('token')
    if (token) {
      return {
        'Authorization': `Bearer ${token}`
      }
    }
    return {}
  }

  private async request<T>(
    endpoint: string,
    options: ApiOptions = {}
  ): Promise<ApiResponse<T>> {
    const {
      method = 'GET',
      headers = {},
      body,
      timeout = this.defaultTimeout,
      showNotifications = true,
      loadingMessage,
      successMessage,
      errorMessage,
      requiresAuth = false
    } = options

    const notificationStore = useNotificationStore()
    const userStore = useUserStore()
    const router = useRouter()
    const controller = new AbortController()
    const timeoutId = setTimeout(() => controller.abort(), timeout)

    // Track loading notification state
    let loadingNotificationId: string | null = null
    let loadingTimeoutId: NodeJS.Timeout | null = null

    try {
      // Show loading notification after a minimum delay (500ms) to avoid flashing for quick requests
      if (showNotifications && loadingMessage) {
        loadingTimeoutId = setTimeout(() => {
          loadingNotificationId = notificationStore.info('Loading', loadingMessage, { duration: 0 })
        }, 500)
      }

      // Add auth headers if required
      const requestHeaders: Record<string, string> = {
        'Content-Type': 'application/json',
        ...(requiresAuth ? this.getAuthHeaders() : {}),
        ...headers
      }

      let response = await fetch(`${this.baseURL}${endpoint}`, {
        method,
        headers: requestHeaders,
        body: body ? JSON.stringify(body) : undefined,
        signal: controller.signal
      })

      clearTimeout(timeoutId)

      // Handle 401/403 for token refresh
      if ((response.status === 401 || response.status === 403) && requiresAuth) {
        try {
          await userStore.refreshToken()
          // Retry the original request with the new token
          requestHeaders['Authorization'] = `Bearer ${localStorage.getItem('token')}`
          response = await fetch(`${this.baseURL}${endpoint}`, {
            method,
            headers: requestHeaders,
            body: body ? JSON.stringify(body) : undefined,
            signal: controller.signal
          })
          if (!response.ok) throw new Error('Retry after refresh failed')
        } catch (refreshError) {
          userStore.logout()
          router.push('/sign-in')
          throw new Error('Session expired. Please log in again.')
        }
      }

      if (!response.ok) {
        const errorData = await response.json().catch(() => ({}))
        throw new Error(errorData.error || errorData.message || `HTTP ${response.status}: ${response.statusText}`)
      }

      // Handle 204 No Content
      if (response.status === 204) {
        if (loadingNotificationId) {
          notificationStore.removeNotification(loadingNotificationId)
        }
        if (showNotifications && successMessage) {
          notificationStore.success('Success', successMessage)
        }
        return {
          data: {} as T,
          status: response.status,
          message: 'No Content'
        }
      }

      const data = await response.json()

      // Clear loading notification if it was shown
      if (loadingNotificationId) {
        notificationStore.removeNotification(loadingNotificationId)
      }

      // Show success notification if requested
      if (showNotifications && successMessage) {
        notificationStore.success('Success', successMessage)
      }

      return {
        data,
        status: response.status,
        message: 'Success'
      }
    } catch (error) {
      clearTimeout(timeoutId)
      
      // Clear loading notification if it was shown
      if (loadingNotificationId) {
        notificationStore.removeNotification(loadingNotificationId)
      }
      
      let errorMessage = 'An unexpected error occurred'
      
      if (error instanceof Error) {
        if (error.name === 'AbortError') {
          errorMessage = 'Request timed out'
        } else {
          errorMessage = error.message
        }
      }

      // Show error notification if requested
      if (showNotifications) {
        notificationStore.error(
          'Error',
          errorMessage || errorMessage,
          { duration: 8000 }
        )
      }

      throw {
        message: errorMessage,
        status: 500,
        code: 'UNKNOWN_ERROR'
      } as ApiError
    } finally {
      // Clean up loading timeout if request completed before it fired
      if (loadingTimeoutId) {
        clearTimeout(loadingTimeoutId)
      }
    }
  }

  // Convenience methods
  async get<T>(endpoint: string, options?: Omit<ApiOptions, 'method'>): Promise<ApiResponse<T>> {
    return this.request<T>(endpoint, { ...options, method: 'GET' })
  }

  async post<T>(endpoint: string, body?: any, options?: Omit<ApiOptions, 'method' | 'body'>): Promise<ApiResponse<T>> {
    return this.request<T>(endpoint, { ...options, method: 'POST', body })
  }

  async put<T>(endpoint: string, body?: any, options?: Omit<ApiOptions, 'method' | 'body'>): Promise<ApiResponse<T>> {
    return this.request<T>(endpoint, { ...options, method: 'PUT', body })
  }

  async patch<T>(endpoint: string, body?: any, options?: Omit<ApiOptions, 'method' | 'body'>): Promise<ApiResponse<T>> {
    return this.request<T>(endpoint, { ...options, method: 'PATCH', body })
  }

  async delete<T>(endpoint: string, options?: Omit<ApiOptions, 'method'>): Promise<ApiResponse<T>> {
    return this.request<T>(endpoint, { ...options, method: 'DELETE' })
  }
}

// Create default instance
export const api = new ApiClient()

// Export the class for custom instances
export { ApiClient }

// Utility functions for common API patterns
export const withRetry = async <T>(
  fn: () => Promise<T>,
  maxRetries: number = 3,
  delay: number = 1000
): Promise<T> => {
  let lastError: Error

  for (let i = 0; i < maxRetries; i++) {
    try {
      return await fn()
    } catch (error) {
      lastError = error as Error
      
      if (i < maxRetries - 1) {
        await new Promise(resolve => setTimeout(resolve, delay * Math.pow(2, i)))
      }
    }
  }

  throw lastError!
}

export const debounce = <T extends (...args: any[]) => any>(
  func: T,
  wait: number
): ((...args: Parameters<T>) => void) => {
  let timeout: ReturnType<typeof setTimeout>
  
  return (...args: Parameters<T>) => {
    clearTimeout(timeout)
    timeout = setTimeout(() => func(...args), wait)
  }
}

export async function deployCluster(payload: any) {
  const res = await fetch('/api/v1/deploy', {
    method: 'POST',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify(payload),
    credentials: 'include', // if using cookies/auth
  });
  if (!res.ok) throw new Error(await res.text());
  return res.json();
}

export function listenToEvents(taskId: string, onMessage: (data: any) => void) {
  const eventSource = new EventSource(`/api/v1/events?task_id=${taskId}`);
  eventSource.onmessage = (event) => {
    try {
      const data = JSON.parse(event.data);
      onMessage(data);
    } catch {
      onMessage(event.data);
    }
  };
  eventSource.onerror = (err) => {
    eventSource.close();
  };
  return eventSource;
} 
export async function getWorkflowStatus(workflowID: string): Promise<ApiResponse<{ data: WorkflowStatus }>> {
  return api.get(`/v1/workflow/${workflowID}`, { requiresAuth: false, showNotifications: false })
}



export function createWorkflowStatusChecker(workflowID: string, options?: {
  interval?: number;
}): {
  status: Promise<WorkflowStatus>;
  cancel: () => void;
} {
  const interval = options?.interval ?? 3000;

  let intervalId: NodeJS.Timeout;
  let rejectFn: (reason?: any) => void;

  const statusPromise = new Promise<WorkflowStatus>((resolve, reject) => {
    rejectFn = reject;

    const check = async () => {
      try {
        const result = await getWorkflowStatus(workflowID);
        const status= result.data.data;

        if (status === WorkflowStatus.StatusCompleted || status === WorkflowStatus.StatusFailed) {
          clearInterval(intervalId);
          resolve(status);
        }
      } catch (error) {
        clearInterval(intervalId);
        reject(error);
      }
    };

    intervalId = setInterval(check, interval);
  });

  const cancel = () => {
    clearInterval(intervalId);
    rejectFn?.('Polling canceled.');
  };

  return { status: statusPromise, cancel };
}



