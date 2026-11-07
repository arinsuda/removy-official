import { io, type Socket } from "socket.io-client"
import { ref } from "vue"

function getSocketUrl(): string {
  const raw =
    import.meta.env.VITE_SOCKET_URL ||
    import.meta.env.VITE_API_BASE_URL ||
    window.location.origin

  try {
    const u = new URL(raw, window.location.origin)
    return u.origin
  } catch {
    return raw.replace(/\/api\/?$/, "")
  }
}

const SOCKET_URL = getSocketUrl()

let socket: Socket | null = null
const isConnected = ref(false)
const isConnecting = ref(false)
let storageListenerBound = false
// Account the socket was opened for; a token for another account means another tab switched.
let socketUserId: number | null = null

// The access token is a JWT whose "uid" claim names the account. No signature check:
// this only tells accounts apart, the server still authenticates the token.
function userIdFromToken(token: string | null): number | null {
  if (!token) return null
  try {
    const payload = (token.split(".")[1] ?? "").replace(/-/g, "+").replace(/_/g, "/")
    const uid = JSON.parse(atob(payload)).uid
    return typeof uid === "number" ? uid : null
  } catch {
    return null
  }
}

/**
 * Reconnect a socket that the server dropped (revoked or expired token) once a newer
 * token is available. No-op when connected, never connected, or signed out.
 */
export function resume() {
  if (!socket || socket.connected) return
  const token = localStorage.getItem("access_token")
  if (!token) return
  socket.auth = { token }
  isConnecting.value = true
  socket.connect()
}

// Another tab refreshed or re-logged in: pick up its token.
function bindStorageListener() {
  if (storageListenerBound || typeof window === "undefined") return
  storageListenerBound = true
  window.addEventListener("storage", e => {
    if (e.key !== "access_token") return
    const incoming = userIdFromToken(localStorage.getItem("access_token"))
    if (incoming !== null && socketUserId !== null && incoming !== socketUserId) {
      // Another tab signed in as a different account: reload so every store and the
      // socket start clean instead of mixing two accounts.
      window.location.reload()
      return
    }
    resume()
  })
}

export function useSocket() {
  /**
   * เปิดการเชื่อมต่อ (no-op ถ้าเชื่อมอยู่แล้ว)
   * userId ใช้ยืนยันตัวตนฝั่ง server เพื่อ join ห้องเฉพาะ user นั้น
   * ถ้า backend ใช้ cookie/JWT อยู่แล้วผ่าน withCredentials ก็ยังส่ง userId
   * เป็น fallback ไว้เผื่อ server อยากใช้ยืนยันห้องอีกชั้น
   */
  // useSocket.ts
  function connect(): Socket {
    if (socket) {
      resume()
      return socket
    }

    isConnecting.value = true
    bindStorageListener()

    const token = localStorage.getItem("access_token")
    socketUserId = userIdFromToken(token)

    socket = io(SOCKET_URL, {
      withCredentials: true,
      auth: { token },
      transports: ["polling", "websocket"],
      reconnection: true,
      reconnectionAttempts: 10,
      reconnectionDelay: 1500,
      reconnectionDelayMax: 8000,
    })

    // Auto-reconnects (network blips) must use the current token, not the one from connect().
    socket.io.on("reconnect_attempt", () => {
      const fresh = localStorage.getItem("access_token")
      if (fresh && socket) {
        socket.auth = { token: fresh }
      }
    })

    socket.on("connect", () => {
      isConnected.value = true
      isConnecting.value = false
      // ไม่ต้อง emit join แล้ว — backend join ห้องให้อัตโนมัติจาก cookie
    })

    socket.on("disconnect", reason => {
      isConnected.value = false
      if (import.meta.env.DEV) console.warn("[socket] disconnected:", reason)
    })

    // The server rejected the token (revoked or expired session). Retrying with the same
    // token only loops, so stop reconnecting but keep the socket and its listeners;
    // resume() reconnects it once a fresh token is written.
    const stopUnauthorized = () => {
      if (import.meta.env.DEV) console.warn("[socket] unauthorized, not reconnecting")
      isConnecting.value = false
      socket?.disconnect()
    }

    socket.on("error", (reason: unknown) => {
      if (reason === "unauthorized") stopUnauthorized()
    })

    socket.on("connect_error", err => {
      isConnecting.value = false
      if (err.message === "unauthorized") stopUnauthorized()
      if (import.meta.env.DEV)
        console.error("[socket] connect_error:", err.message)
    })

    return socket
  }

  function disconnect() {
    socket?.removeAllListeners()
    socket?.disconnect()
    socket = null
    socketUserId = null
    isConnected.value = false
    isConnecting.value = false
  }

  function on<T = unknown>(event: string, handler: (payload: T) => void) {
    socket?.on(event, handler as (...args: unknown[]) => void)
  }

  function off(event: string, handler?: (...args: unknown[]) => void) {
    socket?.off(event, handler)
  }

  function emit(event: string, payload?: unknown) {
    if (!socket?.connected) {
      if (import.meta.env.DEV) {
        console.warn(`[socket] tried to emit "${event}" while disconnected`)
      }
      return
    }
    socket.emit(event, payload)
  }

  return { connect, disconnect, resume, on, off, emit, isConnected, isConnecting }
}
