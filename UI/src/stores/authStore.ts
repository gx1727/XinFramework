import { create } from "zustand"
import { persist } from "zustand/middleware"
import { authApi, setAuthTokens, clearAuthTokens, ApiError } from "@/api"
import type { LoginResponse } from "@/api"

interface User {
  code: string
  id: number
  role: string
  tenant_id: number
}

interface AuthState {
  token: string | null
  refreshToken: string | null
  user: User | null
  isAuthenticated: boolean
  isLoading: boolean
  error: string | null
  lastApiError: string | null
  
  login: (account: string, password: string) => Promise<boolean>
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
      isAuthenticated: false,
      isLoading: false,
      error: null,
      lastApiError: null,
      
      login: async (account: string, password: string) => {
        set({ isLoading: true, error: null })
        
        try {
          const data: LoginResponse = await authApi.login({ account, password, tenant_id: 1 })
          
          const refreshToken = data.refresh_token
          const accessToken = data.token
          
          setAuthTokens(accessToken, refreshToken)
          
          set({ 
            token: accessToken,
            refreshToken: refreshToken,
            user: data.user,
            isAuthenticated: true, 
            isLoading: false,
            error: null 
          })
          
          return true
        } catch (err) {
          const errorMessage = err instanceof ApiError ? err.message : (err instanceof Error ? err.message : "登录失败")
          set({ 
            isLoading: false, 
            error: errorMessage,
            isAuthenticated: false,
            token: null,
            refreshToken: null,
            user: null
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
          isAuthenticated: false, 
          error: null,
          lastApiError: null
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
        isAuthenticated: state.isAuthenticated 
      }),
      onRehydrateStorage: () => (state) => {
        const token = localStorage.getItem("token")
        if (!token && state) {
          state.isAuthenticated = false
          state.token = null
          state.refreshToken = null
          state.user = null
        }
      },
    }
  )
)