// API 客户端的公共层：fetch wrapper + 401 自动 refresh + 通用类型
// 所有域文件 (auth/user/menu/...) 都基于这里的 api() 调用。

const API_BASE_URL = import.meta.env.VITE_API_BASE_URL || "http://localhost:8087/api/v1"

export { API_BASE_URL }

export interface ApiOptions extends RequestInit {
  params?: Record<string, string | number | boolean>
  retry?: number
  retryDelay?: number
}

export interface ApiResponse<T = unknown> {
  code: number
  msg: string
  data: T
}

export interface PageResponse<T> {
  list: T[]
  total: number
  page?: number
  size?: number
}

export class ApiError extends Error {
  status: number
  code: number
  data?: unknown

  constructor(
    status: number,
    code: number,
    message: string,
    data?: unknown
  ) {
    super(message)
    this.name = "ApiError"
    this.status = status
    this.code = code
    this.data = data
  }
}

let isRefreshing = false
let refreshSubscribers: Array<(token: string) => void> = []

function subscribeTokenRefresh(callback: (token: string) => void) {
  refreshSubscribers.push(callback)
}

function onTokenRefreshed(token: string) {
  refreshSubscribers.forEach((callback) => callback(token))
  refreshSubscribers = []
}

async function buildUrl(endpoint: string, params?: Record<string, string | number | boolean>): Promise<string> {
  const url = new URL(`${API_BASE_URL}${endpoint}`, window.location.origin)
  if (params) {
    Object.entries(params).forEach(([key, value]) => {
      if (value === undefined || value === null) return
      url.searchParams.append(key, String(value))
    })
  }
  return url.toString()
}

export function getToken(): string | null {
  if (typeof window !== "undefined") {
    return localStorage.getItem("token")
  }
  return null
}

export function getRefreshToken(): string | null {
  if (typeof window !== "undefined") {
    return localStorage.getItem("refresh_token")
  }
  return null
}

function setTokens(token: string, refreshToken?: string) {
  localStorage.setItem("token", token)
  if (refreshToken) {
    localStorage.setItem("refresh_token", refreshToken)
  }
}

function clearTokens() {
  localStorage.removeItem("token")
  localStorage.removeItem("refresh_token")
}

function redirectToLogin() {
  clearTokens()
  if (typeof window !== "undefined") {
    window.location.href = "/login"
  }
}

async function refreshAccessToken(): Promise<string | null> {
  const refreshToken = getRefreshToken()
  if (!refreshToken) {
    redirectToLogin()
    return null
  }

  try {
    const response = await fetch(`${API_BASE_URL}/auth/refresh`, {
      method: "POST",
      headers: {
        "Content-Type": "application/json",
      },
      body: JSON.stringify({ refresh_token: refreshToken }),
    })

    if (!response.ok) {
      redirectToLogin()
      return null
    }

    const data = await response.json()

    if (data.code === 0 && data.data) {
      setTokens(data.data.token, data.data.refresh_token)
      return data.data.token
    }

    redirectToLogin()
    return null
  } catch {
    redirectToLogin()
    return null
  }
}

async function delay(ms: number): Promise<void> {
  return new Promise((resolve) => setTimeout(resolve, ms))
}

export async function api<T = unknown>(
  endpoint: string,
  options: ApiOptions = {}
): Promise<T> {
  const {
    params,
    retry = 0,
    retryDelay = 1000,
    ...fetchOptions
  } = options

  const url = await buildUrl(endpoint, params)

  const headers: Record<string, string> = {
    ...(options.headers as Record<string, string> | undefined || {}),
  }
  if (!(options.body instanceof FormData) && !headers["Content-Type"]) {
    headers["Content-Type"] = "application/json"
  }

  const token = getToken()
  if (token) {
    headers["Authorization"] = `Bearer ${token}`
  }

  let lastError: Error | null = null

  for (let attempt = 0; attempt <= retry; attempt++) {
    try {
      const response = await fetch(url, {
        ...fetchOptions,
        headers,
      })

      if (response.status === 401) {
        if (!isRefreshing) {
          isRefreshing = true

          const newToken = await refreshAccessToken()

          isRefreshing = false

          if (newToken) {
            onTokenRefreshed(newToken)
            ;(headers as Record<string, string>)["Authorization"] = `Bearer ${newToken}`

            const retryResponse = await fetch(url, {
              ...fetchOptions,
              headers,
            })

            const data = await retryResponse.json().catch(() => null)

            if (!retryResponse.ok) {
              redirectToLogin()
              throw new ApiError(
                retryResponse.status,
                data?.code || retryResponse.status,
                data?.msg || `HTTP error! status: ${retryResponse.status}`,
                data
              )
            }

            const apiResponse = data as ApiResponse<T>

            if (apiResponse.code !== 0) {
              throw new ApiError(
                200,
                apiResponse.code,
                apiResponse.msg,
                apiResponse.data
              )
            }

            return apiResponse.data as T
          }

          redirectToLogin()
          throw new ApiError(401, 401, "Token refresh failed")
        }

        return new Promise((resolve, reject) => {
          subscribeTokenRefresh(async (newToken) => {
            try {
              ;(headers as Record<string, string>)["Authorization"] = `Bearer ${newToken}`
              const retryResponse = await fetch(url, {
                ...fetchOptions,
                headers,
              })
              const data = await retryResponse.json().catch(() => null)

              if (!retryResponse.ok) {
                redirectToLogin()
                reject(new ApiError(
                  retryResponse.status,
                  (data as ApiResponse<unknown>)?.code || retryResponse.status,
                  (data as ApiResponse<unknown>)?.msg || "Request failed"
                ))
                return
              }

              const apiResponse = data as ApiResponse<T>

              if (apiResponse?.code !== 0) {
                reject(new ApiError(200, apiResponse?.code || 0, apiResponse?.msg || "Request failed", apiResponse?.data))
              } else {
                resolve(apiResponse?.data as T)
              }
            } catch (err) {
              reject(err)
            }
          })
        })
      }

      const data = await response.json().catch(() => null)

      if (!response.ok) {
        throw new ApiError(
          response.status,
          data?.code || response.status,
          data?.msg || `HTTP error! status: ${response.status}`,
          data
        )
      }

      const apiResponse = data as ApiResponse<T>

      if (apiResponse.code !== 0) {
        throw new ApiError(
          200,
          apiResponse.code,
          apiResponse.msg,
          apiResponse.data
        )
      }

      return apiResponse.data as T
    } catch (err) {
      lastError = err as Error

      if (attempt < retry && !(err instanceof ApiError && err.status === 401)) {
        await delay(retryDelay * Math.pow(2, attempt))
        continue
      }

      throw lastError
    }
  }

  throw lastError
}

export function setAuthTokens(token: string, refreshToken?: string) {
  setTokens(token, refreshToken)
}

export function clearAuthTokens() {
  clearTokens()
}
