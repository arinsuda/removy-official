/// <reference types="vite/client" />
export {}

declare global {
  interface Window {
    $toast?: {
      show: (options: unknown) => void
      success: (msg: string, title?: string) => void
      error: (msg: string, title?: string) => void
      info: (msg: string, title?: string) => void
      warning: (msg: string, title?: string) => void
    }
  }
}
