import api from "../index"

import type { AuthResponse } from "@/types"

export type LinkPlatform = "planzi" | "popcorn"
export type LinkConfirmMethod = "local_password" | "email_code"

export interface LinkHandoffResponse {
  redirect_url: string
  expires_in: number
}

export interface LinkSessionResponse extends AuthResponse {
  linked: boolean
  provisioned: boolean
}

export interface LinkConfirmRequired {
  code: "LINK_CONFIRM_REQUIRED"
  pending_id: string
  email: string
  methods: LinkConfirmMethod[]
}

export interface LinkStatusResponse {
  links: { platform: LinkPlatform; remote_user_id: string; linked_at: string }[]
}

export const linkApi = {
  createHandoff: (target: LinkPlatform) =>
    api.post<LinkHandoffResponse>("/link/handoff", { target }),

  redeem: (token: string) =>
    api.post<LinkSessionResponse>("/auth/handoff/redeem", { token }),

  confirm: (data: {
    pending_id: string
    method: LinkConfirmMethod
    password?: string
    code?: string
  }) => api.post<LinkSessionResponse>("/auth/handoff/confirm", data),

  sendCode: (pendingId: string) =>
    api.post("/auth/handoff/send-code", { pending_id: pendingId }),

  remote: () => api.get<{ web_url: string }>("/link/remote"),

  status: () => api.get<LinkStatusResponse>("/link/status"),

  unlink: (platform: LinkPlatform) => api.delete(`/link/${platform}`),
}
