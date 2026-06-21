import { create } from "zustand"
import { persist } from "zustand/middleware"
import {
  authApi,
  setAuthTokens,
  clearAuthTokens,
  ApiError,
  type LoginScope,
} from "@/api"
import type { LoginResponse } from "@/api"

interface User {
  code: string
  id: number
  role: string
  /** scope=platform 时固定 0；scope=tenant 时为真实 tenant_id */
  tenant_id: number
  nickname?: string
  real_name?: string
  avatar?: string
  email?: string
  platform_roles?: string[]
}

interface AuthState {
  token: string | null
  refreshToken: string | null
  user: User | null
  /** 登录作用域：tenant / platform / null（未登录） */
  scope: LoginScope | null
  isAuthenticated: boolean
  isLoading: boolean
  error: string | null
  lastApiError: string | null

  /** 租户域登录（业务用户登录）。跳转到 /app/dashboard。 */
  tenantLogin: (account: string, password: string, tenantId: number) => Promise<boolean>
  /** 平台域登录（super_admin 登录）。跳转到 /platform/dashboard。 */
  platformLogin: (account: string, password: string) => Promise<boolean>
  logout: () => void
  clearError: () => void
  clearApiError: () => void
}

export const useAuthStore = create<AuthState>()(
  persist(
    (set) => ({
      token: null,
      refreshToken: null,
      user: null,
      scope: null,
      isAuthenticated: false,
      isLoading: false,
      error: null,
      lastApiError: null,

      tenantLogin: async (account, password, tenantId) => {
        set({ isLoading: true, error: null })
        try {
          const data: LoginResponse = await authApi.tenantLogin({
            account,
            password,
            tenant_id: tenantId,
          })

          setAuthTokens(data.token, data.refresh_token)

          set({
            token: data.token,
            refreshToken: data.refresh_token,
            user: data.user,
            scope: data.scope,
            isAuthenticated: true,
            isLoading: false,
            error: null,
          })

          return true
        } catch (err) {
          const errorMessage =
            err instanceof ApiError
              ? err.message
              : err instanceof Error
                ? err.message
                : "登录失败"
          set({
            isLoading: false,
            error: errorMessage,
            isAuthenticated: false,
            token: null,
            refreshToken: null,
            user: null,
            scope: null,
          })
          return false
        }
      },

      platformLogin: async (account, password) => {
        set({ isLoading: true, error: null })
        try {
          const data: LoginResponse = await authApi.platformLogin({
            account,
            password,
          })

          setAuthTokens(data.token, data.refresh_token)

          set({
            token: data.token,
            refreshToken: data.refresh_token,
            user: data.user,
            scope: data.scope,
            isAuthenticated: true,
            isLoading: false,
            error: null,
          })

          return true
        } catch (err) {
          const errorMessage =
            err instanceof ApiError
              ? err.message
              : err instanceof Error
                ? err.message
                : "登录失败"
          set({
            isLoading: false,
            error: errorMessage,
            isAuthenticated: false,
            token: null,
            refreshToken: null,
            user: null,
            scope: null,
          })
          return false
        }
      },

      logout: () => {
        clearAuthTokens()
        set({
          token: null,
          refreshToken: null,
          user: null,
          scope: null,
          isAuthenticated: false,
          error: null,
          lastApiError: null,
        })
      },

      clearError: () => {
        set({ error: null })
      },

      clearApiError: () => {
        set({ lastApiError: null })
      },
    }),
    {
      name: "auth-storage",
      partialize: (state) => ({
        token: state.token,
        refreshToken: state.refreshToken,
        user: state.user,
        scope: state.scope,
        isAuthenticated: state.isAuthenticated,
      }),
      onRehydrateStorage: () => (state) => {
        const token = localStorage.getItem("token")
        if (!token && state) {
          state.isAuthenticated = false
          state.token = null
          state.refreshToken = null
          state.user = null
          state.scope = null
        }
      },
    },
  ),
)