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

export function useSocket() {
  /**
   * เปิดการเชื่อมต่อ (no-op ถ้าเชื่อมอยู่แล้ว)
   * userId ใช้ยืนยันตัวตนฝั่ง server เพื่อ join ห้องเฉพาะ user นั้น
   * ถ้า backend ใช้ cookie/JWT อยู่แล้วผ่าน withCredentials ก็ยังส่ง userId
   * เป็น fallback ไว้เผื่อ server อยากใช้ยืนยันห้องอีกชั้น
   */
  // useSocket.ts
  function connect(): Socket {
    if (socket?.connected || isConnecting.value) return socket as Socket

    isConnecting.value = true

    const token = localStorage.getItem("access_token")

    socket = io(SOCKET_URL, {
      withCredentials: true,
      auth: { token },
      transports: ["polling", "websocket"],
      reconnection: true,
      reconnectionAttempts: 10,
      reconnectionDelay: 1500,
      reconnectionDelayMax: 8000,
    })

    // The server drops sockets when sessions are revoked; reconnect with the current token, not the one from connect().
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

    // The server rejected the token (revoked or expired session) and dropped us. Retrying
    // with the same token only loops, so stop; the next connect() starts a fresh socket.
    socket.on("error", (reason: unknown) => {
      if (reason !== "unauthorized") return
      if (import.meta.env.DEV) console.warn("[socket] unauthorized, not reconnecting")
      disconnect()
    })

    socket.on("connect_error", err => {
      isConnecting.value = false
      if (import.meta.env.DEV)
        console.error("[socket] connect_error:", err.message)
    })

    return socket
  }

  function disconnect() {
    socket?.removeAllListeners()
    socket?.disconnect()
    socket = null
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

  return { connect, disconnect, on, off, emit, isConnected, isConnecting }
}
