<template>
  <div class="auth-page">
    <div class="auth-bg"><div class="auth-bg__overlay" /></div>
    <div class="auth-container" ref="containerRef">
      <div class="auth-brand" ref="brandRef">
        <div class="brand-logo">
          <span class="brand-movie">POP</span><span class="brand-hub">CORN</span>
        </div>
        <p class="brand-tagline">
          Welcome to Popcorn – Your Ultimate Destination<br />
          for Movie Reviews and Recommendations!
        </p>
      </div>
      <div class="auth-divider" ref="dividerRef" />
      <div class="auth-form-panel" ref="formPanelRef">
        <h1 class="auth-title">LOGIN</h1>
        <form @submit.prevent="handleLogin" class="auth-form" novalidate>
          <div class="form-group form-item">
            <label class="form-label">Username or email</label>
            <div class="input-wrapper">
              <span class="input-icon"><IconUser :size="16" /></span>
              <input
                v-model="form.identifier"
                type="text"
                class="form-input"
                autocomplete="username"
                :disabled="isLoading"
              />
            </div>
          </div>
          <div class="form-group form-item">
            <label class="form-label">Password</label>
            <div class="input-wrapper">
              <span class="input-icon"><IconLock :size="16" /></span>
              <input
                v-model="form.password"
                :type="showPassword ? 'text' : 'password'"
                class="form-input"
                autocomplete="current-password"
                :disabled="isLoading"
              />
              <button
                type="button"
                class="input-toggle"
                @click="showPassword = !showPassword"
                tabindex="-1"
              >
                <IconEyeOff v-if="showPassword" :size="16" /><IconEye
                  v-else
                  :size="16"
                />
              </button>
            </div>
          </div>
          <Transition name="alert">
            <div v-if="errorMsg" class="form-alert" role="alert">
              <span class="form-alert__icon">
                <IconAlertCircle :size="15" />
              </span>
              <span class="form-alert__text">{{ errorMsg }}</span>
              <button
                class="form-alert__close"
                type="button"
                @click="errorMsg = ''"
              >
                <IconX :size="13" />
              </button>
            </div>
          </Transition>
          <div class="form-forgot form-item">
            <RouterLink to="/forgot-password" class="link-muted"
              >forget password</RouterLink
            >
          </div>
          <button
            type="submit"
            class="btn-primary form-item"
            :disabled="isLoading"
          >
            <span v-if="isLoading" class="spinner" /><span v-else>LOGIN</span>
          </button>

          <div class="oauth-divider form-item">
            <span class="oauth-divider__line" />
            <span class="oauth-divider__text">OR</span>
            <span class="oauth-divider__line" />
          </div>

          <button
            type="button"
            class="btn-google form-item"
            :disabled="isLoading"
            @click="handleGoogleLogin"
            aria-label="Continue with Google"
          >
            <svg class="google-logo" viewBox="0 0 24 24" width="18" height="18">
              <path fill="#4285F4" d="M22.56 12.25c0-.78-.07-1.53-.2-2.25H12v4.26h5.92c-.26 1.37-1.04 2.53-2.21 3.31v2.77h3.57c2.08-1.92 3.28-4.74 3.28-8.09z"/>
              <path fill="#34A853" d="M12 23c2.97 0 5.46-.98 7.28-2.66l-3.57-2.77c-.98.66-2.23 1.06-3.71 1.06-2.86 0-5.29-1.93-6.16-4.53H2.18v2.84C3.99 20.53 7.7 23 12 23z"/>
              <path fill="#FBBC05" d="M5.84 14.1c-.22-.66-.35-1.36-.35-2.1s.13-1.44.35-2.1V7.06H2.18C1.43 8.55 1 10.22 1 12s.43 3.45 1.18 4.94l2.85-2.22.81-.62z"/>
              <path fill="#EA4335" d="M12 5.38c1.62 0 3.06.56 4.21 1.64l3.15-3.15C17.45 2.09 14.97 1 12 1 7.7 1 3.99 3.47 2.18 7.06l3.66 2.84c.87-2.6 3.3-4.52 6.16-4.52z"/>
            </svg>
            <span>Continue with Google</span>
          </button>

          <p class="form-footer form-item">
            <RouterLink to="/register" class="link-underline"
              >I don't have an account.</RouterLink
            >
          </p>
        </form>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
  import { ref, onMounted } from "vue"
  import { useRouter, useRoute } from "vue-router"
  import { useAuthStore } from "@/stores/auth"
  import { authApi } from "@/api/api"
  import { gsap } from "gsap"
  import {
    User as IconUser,
    Lock as IconLock,
    Eye as IconEye,
    EyeOff as IconEyeOff,
    AlertCircle as IconAlertCircle,
    X as IconX,
  } from "lucide-vue-next"

  const router = useRouter()
  const route = useRoute()
  const authStore = useAuthStore()

  const form = ref({ identifier: "", password: "" })
  const showPassword = ref(false)
  const isLoading = ref(false)
  const errorMsg = ref("")

  // --- Refs ---
  const containerRef = ref<HTMLElement | null>(null)
  const brandRef = ref<HTMLElement | null>(null)
  const dividerRef = ref<HTMLElement | null>(null)
  const formPanelRef = ref<HTMLElement | null>(null)

  onMounted(() => {
    const errParam = route.query.error as string | undefined
    if (errParam) {
      if (errParam === "user_cancelled") {
        errorMsg.value = "คุณได้ยกเลิกการเข้าสู่ระบบด้วย Google"
      } else if (errParam === "google_account_link_required") {
        errorMsg.value = "พบบัญชีอีเมลนี้ในระบบ กรุณาเข้าสู่ระบบด้วยรหัสผ่านแล้วเชื่อมต่อ Google ในการตั้งค่า"
      } else if (errParam === "google_identity_already_connected") {
        errorMsg.value = "บัญชี Google นี้ถูกเชื่อมต่อกับผู้ใช้อื่นในระบบแล้ว"
      } else if (errParam === "account_disabled") {
        errorMsg.value = "บัญชีผู้ใช้นี้ถูกระงับการใช้งาน"
      } else {
        errorMsg.value = "เกิดข้อผิดพลาดในการเข้าสู่ระบบด้วย Google กรุณาลองใหม่อีกครั้ง"
      }
      router.replace({ query: {} })
    }

    const tl = gsap.timeline({ defaults: { ease: "power3.out" } })

    gsap.set(containerRef.value, { opacity: 0, y: 32 })
    gsap.set(brandRef.value, { opacity: 0, x: -28 })
    gsap.set(dividerRef.value, { scaleY: 0, transformOrigin: "top center" })
    gsap.set(formPanelRef.value, { opacity: 0, x: 28 })
    gsap.set(".form-item", { opacity: 0, y: 14 })

    tl.to(containerRef.value, { opacity: 1, y: 0, duration: 0.55 })
      .to(brandRef.value, { opacity: 1, x: 0, duration: 0.5 }, "-=0.3")
      .to(dividerRef.value, { scaleY: 1, duration: 0.45 }, "-=0.25")
      .to(formPanelRef.value, { opacity: 1, x: 0, duration: 0.5 }, "-=0.35")
      .to(
        ".form-item",
        {
          opacity: 1,
          y: 0,
          duration: 0.4,
          stagger: 0.08,
        },
        "-=0.25",
      )
  })

  async function handleGoogleLogin() {
    try {
      const { data } = await authApi.getGoogleConfig()
      if (!data.enabled) {
        errorMsg.value = "ระบบยังไม่ได้เปิดใช้งาน Google OAuth ในไฟล์ .env ของเซิร์ฟเวอร์ (GOOGLE_OAUTH_ENABLED=true)"
        return
      }
      const redirect = (route.query.redirect as string) || "/"
      window.location.assign(authApi.getGoogleLoginUrl(redirect))
    } catch {
      const redirect = (route.query.redirect as string) || "/"
      window.location.assign(authApi.getGoogleLoginUrl(redirect))
    }
  }

  async function handleLogin() {
    errorMsg.value = ""
    if (!form.value.identifier || !form.value.password) {
      errorMsg.value = "Please fill in all fields"
      // shake animation เมื่อ error
      gsap.fromTo(
        formPanelRef.value,
        { x: -8 },
        { x: 0, duration: 0.4, ease: "elastic.out(1, 0.3)" },
      )
      return
    }
    isLoading.value = true
    try {
      await authStore.login(form.value)
      const redirect = route.query.redirect as string | undefined
      // fade out ก่อน navigate
      await gsap.to(containerRef.value, {
        opacity: 0,
        y: -20,
        duration: 0.35,
        ease: "power2.in",
      })
      router.push(redirect ?? { name: "home" })
    } catch (e: any) {
      const status = e?.response?.status

      if (status === 401) {
        errorMsg.value = "ชื่อผู้ใช้หรือรหัสผ่านไม่ถูกต้อง"
      } else if (status === 429) {
        errorMsg.value = "พยายามเข้าสู่ระบบบ่อยเกินไป กรุณารอสักครู่"
      } else if (status >= 500) {
        errorMsg.value = "เกิดข้อผิดพลาดจากเซิร์ฟเวอร์ กรุณาลองใหม่ภายหลัง"
      } else {
        errorMsg.value = "เกิดข้อผิดพลาด กรุณาลองใหม่"
      }

      gsap.fromTo(
        formPanelRef.value,
        { x: -8 },
        { x: 0, duration: 0.45, ease: "elastic.out(1, 0.3)" },
      )
    } finally {
      isLoading.value = false
    }
  }
</script>

<style scoped>
  .auth-page {
    position: relative;
    min-height: 100vh;
    display: flex;
    align-items: center;
    justify-content: center;
    overflow: hidden;
    background: #0a0a0a;
  }
  .auth-bg {
    position: absolute;
    inset: 0;
    background: url("@/assets/bg.png") center/cover no-repeat;
    filter: brightness(0.35) saturate(0.7);
  }
  .auth-bg__overlay {
    position: absolute;
    inset: 0;
    background: linear-gradient(
      to right,
      rgba(10, 10, 10, 0.5) 0%,
      rgba(10, 10, 10, 0.1) 50%,
      rgba(10, 10, 10, 0.5) 100%
    );
  }
  .auth-container {
    position: relative;
    z-index: 1;
    display: flex;
    align-items: center;
    width: 100%;
    max-width: 860px;
    min-height: 460px;
    background: rgba(12, 12, 12, 0.6);
    backdrop-filter: blur(24px);
    -webkit-backdrop-filter: blur(24px);
    border: 1px solid rgba(255, 255, 255, 0.08);
    border-radius: 16px;
    overflow: hidden;
    margin: 2rem;
  }
  .auth-brand {
    flex: 1;
    display: flex;
    flex-direction: column;
    align-items: center;
    justify-content: center;
    padding: 3rem 2.5rem;
    text-align: center;
  }
  .brand-logo {
    font-family: "Noto Sans Thai", "Arial Black", sans-serif;
    font-size: 3.5rem;
    font-weight: 900;
    letter-spacing: -1px;
    line-height: 1;
    margin-bottom: 1.5rem;
  }
  .brand-movie {
    color: #ffffff;
  }
  .brand-hub {
    color: var(--color-brand);
  }
  .brand-tagline {
    font-size: 0.875rem;
    color: #a3a3a3;
    line-height: 1.7;
    max-width: 280px;
  }
  .auth-divider {
    width: 1px;
    align-self: stretch;
    background: rgba(255, 255, 255, 0.08);
    flex-shrink: 0;
  }
  .auth-form-panel {
    flex: 1;
    display: flex;
    flex-direction: column;
    justify-content: center;
    padding: 3rem 2.5rem;
  }
  .auth-title {
    font-family: "Noto Sans Thai", "Arial Black", sans-serif;
    font-size: 2rem;
    font-weight: 900;
    letter-spacing: 3px;
    text-align: center;
    color: #ffffff;
    margin-bottom: 2rem;
  }
  .auth-form {
    display: flex;
    flex-direction: column;
    gap: 1rem;
  }
  .form-group {
    display: flex;
    flex-direction: column;
    gap: 0.35rem;
  }
  .form-label {
    font-size: 0.82rem;
    font-weight: 500;
    color: #d1d1d1;
  }
  .input-wrapper {
    position: relative;
    display: flex;
    align-items: center;
  }
  .input-icon {
    position: absolute;
    left: 12px;
    color: #777;
    display: flex;
    align-items: center;
    justify-content: center;
    pointer-events: none;
    width: 16px;
    height: 16px;
  }
  .form-input {
    width: 100%;
    padding: 0.65rem 2.8rem;
    background: rgba(228, 228, 228, 0.92);
    border: none;
    border-radius: 9999px;
    font-size: 0.9rem;
    color: #111;
    outline: none;
    transition:
      background 0.2s,
      box-shadow 0.2s;
  }
  .form-input:focus {
    background: #f0f0f0;
    box-shadow: 0 0 0 2px rgba(245, 67, 44, 0.45);
  }
  .form-input:disabled {
    opacity: 0.6;
  }
  .input-toggle {
    position: absolute;
    right: 12px;
    background: none;
    border: none;
    cursor: pointer;
    color: #666;
    display: flex;
    align-items: center;
    justify-content: center;
    padding: 0;
    width: 16px;
    height: 16px;
    transition: color 0.2s;
  }
  .input-toggle:hover {
    color: var(--color-brand);
  }
  .form-error {
    font-size: 0.8rem;
    color: var(--color-brand);
    text-align: center;
    margin: 0;
  }
  .form-forgot {
    text-align: right;
    margin-top: -0.25rem;
  }
  .link-muted {
    font-size: 0.8rem;
    color: #a3a3a3;
    text-decoration: none;
    transition: color 0.2s;
  }
  .link-muted:hover {
    color: #fff;
  }
  .btn-primary {
    margin-top: 0.25rem;
    width: 100%;
    padding: 0.75rem;
    background: var(--color-brand);
    color: #fff;
    border: none;
    border-radius: 9999px;
    font-family: "Noto Sans Thai", "Arial Black", sans-serif;
    font-size: 1.1rem;
    font-weight: 900;
    letter-spacing: 2px;
    cursor: pointer;
    transition:
      background 0.2s,
      transform 0.1s;
    display: flex;
    align-items: center;
    justify-content: center;
    min-height: 48px;
  }
  .btn-primary:hover:not(:disabled) {
    background: var(--color-brand-hover);
    transform: translateY(-1px);
  }
  .btn-primary:active:not(:disabled) {
    transform: translateY(0);
  }
  .btn-primary:disabled {
    opacity: 0.6;
    cursor: not-allowed;
  }
  .spinner {
    width: 20px;
    height: 20px;
    border: 2px solid rgba(255, 255, 255, 0.3);
    border-top-color: #fff;
    border-radius: 50%;
    animation: spin 0.7s linear infinite;
  }
  @keyframes spin {
    to {
      transform: rotate(360deg);
    }
  }
  .form-footer {
    text-align: center;
    margin: 0;
  }
  .link-underline {
    font-size: 0.85rem;
    color: #d1d1d1;
    text-underline-offset: 3px;
    transition: color 0.2s;
  }
  .link-underline:hover {
    color: #fff;
  }
  @media (max-width: 640px) {
    .auth-container {
      flex-direction: column;
      margin: 1rem;
    }
    .auth-divider {
      width: auto;
      height: 1px;
      align-self: stretch;
    }
    .auth-brand {
      padding: 2rem 2rem 1.5rem;
    }
    .auth-form-panel {
      padding: 1.5rem 2rem 2rem;
    }
    .brand-logo {
      font-size: 2.5rem;
    }
  }
  /* ลบ .form-error เดิมออก แล้วใส่อันนี้แทน */
  .form-alert {
    display: flex;
    align-items: center;
    gap: 8px;
    padding: 0.6rem 0.9rem;
    background: rgba(245, 67, 44, 0.12);
    border: 1px solid rgba(245, 67, 44, 0.35);
    border-radius: 10px;
    color: #ff6b6b;
    font-size: 0.8rem;
    line-height: 1.4;
  }

  .form-alert__icon {
    display: flex;
    flex-shrink: 0;
    color: var(--color-brand);
  }

  .form-alert__text {
    flex: 1;
  }

  .form-alert__close {
    display: flex;
    flex-shrink: 0;
    background: none;
    border: none;
    cursor: pointer;
    color: #ff6b6b;
    padding: 0;
    opacity: 0.7;
    transition: opacity 0.2s;
  }

  .form-alert__close:hover {
    opacity: 1;
  }

  /* Transition animation */
  .alert-enter-active {
    animation: alertIn 0.35s cubic-bezier(0.21, 1.02, 0.73, 1) forwards;
  }

  .alert-leave-active {
    animation: alertOut 0.25s ease-in forwards;
  }

  @keyframes alertIn {
    from {
      opacity: 0;
      transform: translateY(-8px) scaleY(0.92);
      max-height: 0;
    }
    to {
      opacity: 1;
      transform: translateY(0) scaleY(1);
      max-height: 80px;
    }
  }

  @keyframes alertOut {
    from {
      opacity: 1;
      transform: translateY(0) scaleY(1);
      max-height: 80px;
    }
    to {
      opacity: 0;
      transform: translateY(-6px) scaleY(0.94);
      max-height: 0;
    }
  }

  .oauth-divider {
    display: flex;
    align-items: center;
    gap: 12px;
    margin: 0.5rem 0;
  }
  .oauth-divider__line {
    flex: 1;
    height: 1px;
    background: rgba(255, 255, 255, 0.12);
  }
  .oauth-divider__text {
    font-size: 0.75rem;
    font-weight: 600;
    color: #8a8a8e;
    letter-spacing: 1px;
  }

  .btn-google {
    width: 100%;
    padding: 0.75rem;
    background: rgba(255, 255, 255, 0.07);
    color: #ffffff;
    border: 1px solid rgba(255, 255, 255, 0.15);
    border-radius: 9999px;
    font-family: inherit;
    font-size: 0.95rem;
    font-weight: 600;
    cursor: pointer;
    transition: background 0.2s, border-color 0.2s, transform 0.1s;
    display: flex;
    align-items: center;
    justify-content: center;
    gap: 10px;
    min-height: 46px;
  }
  .btn-google:hover:not(:disabled) {
    background: rgba(255, 255, 255, 0.14);
    border-color: rgba(255, 255, 255, 0.3);
    transform: translateY(-1px);
  }
  .btn-google:active:not(:disabled) {
    transform: translateY(0);
  }
  .btn-google:disabled {
    opacity: 0.5;
    cursor: not-allowed;
  }
</style>
