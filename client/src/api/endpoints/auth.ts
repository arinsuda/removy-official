import api from "../index"
import { REFRESH_POST_TIMEOUT_MS } from "../refresh"

import type {
  AuthResponse,
  ForgotPasswordRequest,
  LoginRequest,
  RegisterRequest,
  ResetPasswordRequest,
} from "@/types"

export interface GoogleStatusResponse {
  enabled: boolean
  google_connected: boolean
  can_disconnect: boolean
  google_email?: string
}

export interface VerifyEmailResponse {
  message: string
  /** true when the password was not proven: the account has no password until a reset */
  password_reset_required: boolean
}

export const authApi = {
  login: (data: LoginRequest) => api.post<AuthResponse>("/auth/login", data),

  register: (data: RegisterRequest) =>
    api.post<AuthResponse>("/auth/register", data),

  refreshToken: (token?: string) => api.post<AuthResponse>("/auth/refresh", { refresh_token: token }, { timeout: REFRESH_POST_TIMEOUT_MS }),

  getSession: () => api.get<{ has_session: boolean }>("/auth/session"),

  // The stored refresh token goes in the body so the server can revoke it when the cookie is missing.
  logout: () => api.post("/auth/logout", { refresh_token: localStorage.getItem("refresh_token") ?? undefined }),

  // POST, never GET: link scanners must not be able to verify an address.
  // Send the current password, or skipPassword: true for an explicit "I don't know it"
  // (the only choice that wipes the account's credentials).
  verifyEmail: (token: string, choice: { password: string } | { skipPassword: true }) =>
    api.post<VerifyEmailResponse>("/auth/verify-email", {
      token,
      ...("password" in choice
        ? { password: choice.password }
        : { skip_password: choice.skipPassword }),
    }),

  resendVerification: (email: string) =>
    api.post("/auth/resend-verification", { email }),

  logoutAll: () => api.post("auth/logout-all"),

  // ส่ง email เพื่อรับ reset link
  forgotPassword: (data: ForgotPasswordRequest) =>
    api.post("auth/forgot-password", data),

  // ใช้ token จาก email ตั้งรหัสผ่านใหม่
  resetPassword: (data: ResetPasswordRequest) =>
    api.post("auth/reset-password", data),

  getGoogleLoginUrl: (returnUrl: string = "/") => {
    const envUrl = import.meta.env.VITE_API_BASE_URL
    const apiBase = envUrl
      ? envUrl.replace(/\/+$/, "").endsWith("/api")
        ? envUrl.replace(/\/+$/, "")
        : `${envUrl.replace(/\/+$/, "")}/api`
      : "/api"
    return `${apiBase}/auth/google/login?return_url=${encodeURIComponent(returnUrl)}`
  },

  startGoogleLink: (returnUrl: string = "/") =>
    api.post<{ authorization_url: string }>("/auth/google/link", { return_url: returnUrl }),

  disconnectGoogle: () => api.delete("/auth/google/link"),

  getGoogleStatus: () => api.get<GoogleStatusResponse>("/auth/google/status"),

  getGoogleConfig: () => api.get<{ enabled: boolean }>("/auth/google/config"),
}
