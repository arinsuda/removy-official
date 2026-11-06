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
        <h1 class="auth-title">REGISTER</h1>
        <form @submit.prevent="handleRegister" class="auth-form" novalidate>
          <!-- Username -->
          <div class="form-group form-item">
            <label class="form-label">Username</label>
            <div
              class="input-wrapper"
              :class="{
                'input-wrapper--error': touched.username && errors.username,
                'input-wrapper--valid':
                  touched.username && !errors.username && form.username,
              }"
            >
              <span class="input-icon"><IconUser :size="16" /></span>
              <input
                v-model="form.username"
                type="text"
                class="form-input"
                autocomplete="username"
                :disabled="isLoading"
                @input="validateField('username')"
                @blur="touchField('username')"
              />
              <span
                v-if="touched.username && !errors.username && form.username"
                class="input-valid-icon"
              >
                <IconCheck :size="14" />
              </span>
            </div>
            <Transition name="field-error">
              <p
                v-if="touched.username && errors.username"
                class="field-error-msg"
              >
                <IconAlertCircle :size="12" />
                {{ errors.username }}
              </p>
            </Transition>
          </div>

          <!-- Email -->
          <div class="form-group form-item">
            <label class="form-label">Email address</label>
            <div
              class="input-wrapper"
              :class="{
                'input-wrapper--error': touched.email && errors.email,
                'input-wrapper--valid':
                  touched.email && !errors.email && form.email,
              }"
            >
              <span class="input-icon"><IconMail :size="16" /></span>
              <input
                v-model="form.email"
                type="email"
                class="form-input"
                autocomplete="email"
                :disabled="isLoading"
                @input="validateField('email')"
                @blur="touchField('email')"
              />
              <span
                v-if="touched.email && !errors.email && form.email"
                class="input-valid-icon"
              >
                <IconCheck :size="14" />
              </span>
            </div>
            <Transition name="field-error">
              <p v-if="touched.email && errors.email" class="field-error-msg">
                <IconAlertCircle :size="12" />
                {{ errors.email }}
              </p>
            </Transition>
          </div>

          <!-- Password -->
          <div class="form-group form-item">
            <label class="form-label">Password</label>
            <div
              class="input-wrapper"
              :class="{
                'input-wrapper--error': touched.password && errors.password,
                'input-wrapper--valid':
                  touched.password && !errors.password && form.password,
              }"
            >
              <span class="input-icon"><IconLock :size="16" /></span>
              <input
                v-model="form.password"
                :type="showPassword ? 'text' : 'password'"
                class="form-input"
                autocomplete="new-password"
                :disabled="isLoading"
                @input="onPasswordInput"
                @blur="touchField('password')"
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
            <!-- Password strength bar -->
            <div v-if="form.password" class="password-strength">
              <div class="strength-bars">
                <span
                  v-for="i in 4"
                  :key="i"
                  class="strength-bar"
                  :class="strengthBarClass(i)"
                />
              </div>
              <span
                class="strength-label"
                :class="`strength-label--${passwordStrength.level}`"
              >
                {{ passwordStrength.label }}
              </span>
            </div>
            <Transition name="field-error">
              <p
                v-if="touched.password && errors.password"
                class="field-error-msg"
              >
                <IconAlertCircle :size="12" />
                {{ errors.password }}
              </p>
            </Transition>
          </div>

          <!-- Confirm Password -->
          <div class="form-group form-item">
            <label class="form-label">Confirm password</label>
            <div
              class="input-wrapper"
              :class="{
                'input-wrapper--error':
                  touched.confirm_password && errors.confirm_password,
                'input-wrapper--valid':
                  touched.confirm_password &&
                  !errors.confirm_password &&
                  form.confirm_password,
              }"
            >
              <span class="input-icon"><IconLock :size="16" /></span>
              <input
                v-model="form.confirm_password"
                :type="showConfirm ? 'text' : 'password'"
                class="form-input"
                autocomplete="new-password"
                :disabled="isLoading"
                @input="validateField('confirm_password')"
                @blur="touchField('confirm_password')"
              />
              <button
                type="button"
                class="input-toggle"
                @click="showConfirm = !showConfirm"
                tabindex="-1"
              >
                <IconEyeOff v-if="showConfirm" :size="16" /><IconEye
                  v-else
                  :size="16"
                />
              </button>
            </div>
            <Transition name="field-error">
              <p
                v-if="touched.confirm_password && errors.confirm_password"
                class="field-error-msg"
              >
                <IconAlertCircle :size="12" />
                {{ errors.confirm_password }}
              </p>
            </Transition>
          </div>

          <!-- Server / global error -->
          <Transition name="alert">
            <div v-if="serverError" class="form-alert" role="alert">
              <span class="form-alert__icon"
                ><IconAlertCircle :size="15"
              /></span>
              <span class="form-alert__text">{{ serverError }}</span>
              <button
                class="form-alert__close"
                type="button"
                @click="serverError = ''"
              >
                <IconX :size="13" />
              </button>
            </div>
          </Transition>

          <button
            type="submit"
            class="btn-primary form-item"
            :disabled="isLoading"
          >
            <span v-if="isLoading" class="spinner" /><span v-else
              >REGISTER</span
            >
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
            <RouterLink to="/login" class="link-underline"
              >I already have an account.</RouterLink
            >
          </p>
        </form>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
  import { ref, computed, onMounted } from "vue"
  import { useRouter } from "vue-router"
  import { useAuthStore } from "@/stores/auth"
  import { authApi } from "@/api/api"
  import { gsap } from "gsap"
  import {
    User as IconUser,
    Mail as IconMail,
    Lock as IconLock,
    Eye as IconEye,
    EyeOff as IconEyeOff,
    AlertCircle as IconAlertCircle,
    Check as IconCheck,
    X as IconX,
  } from "lucide-vue-next"

  const router = useRouter()
  const authStore = useAuthStore()

  async function handleGoogleLogin() {
    try {
      const { data } = await authApi.getGoogleConfig()
      if (!data.enabled) {
        serverError.value = "ระบบยังไม่ได้เปิดใช้งาน Google OAuth ในไฟล์ .env ของเซิร์ฟเวอร์ (GOOGLE_OAUTH_ENABLED=true)"
        return
      }
      window.location.assign(authApi.getGoogleLoginUrl("/"))
    } catch {
      window.location.assign(authApi.getGoogleLoginUrl("/"))
    }
  }

  const form = ref({
    username: "",
    email: "",
    password: "",
    confirm_password: "",
  })
  const showPassword = ref(false)
  const showConfirm = ref(false)
  const isLoading = ref(false)
  const serverError = ref("")

  // ── Touched state (เพื่อไม่แสดง error ก่อน user แตะ field) ──────────────
  const touched = ref({
    username: false,
    email: false,
    password: false,
    confirm_password: false,
  })
  const errors = ref({
    username: "",
    email: "",
    password: "",
    confirm_password: "",
  })

  // ── Validation rules ─────────────────────────────────────────────────────
  const EMAIL_RE = /^[^\s@]+@[^\s@]+\.[^\s@]+$/
  // username: 3-20 ตัว, ตัวอักษร/ตัวเลข/underscore/dash เท่านั้น
  const USERNAME_RE = /^[a-zA-Z0-9_-]{3,20}$/

  function validateField(field: keyof typeof errors.value) {
    switch (field) {
      case "username": {
        const v = form.value.username.trim()
        if (!v) {
          errors.value.username = "Username is required"
        } else if (v.length < 3) {
          errors.value.username = "At least 3 characters"
        } else if (v.length > 20) {
          errors.value.username = "Maximum 20 characters"
        } else if (!USERNAME_RE.test(v)) {
          errors.value.username = "Letters, numbers, _ and - only"
        } else {
          errors.value.username = ""
        }
        break
      }
      case "email": {
        const v = form.value.email.trim()
        if (!v) {
          errors.value.email = "Email is required"
        } else if (!EMAIL_RE.test(v)) {
          errors.value.email = "Enter a valid email address"
        } else {
          errors.value.email = ""
        }
        break
      }
      case "password": {
        const v = form.value.password
        if (!v) {
          errors.value.password = "Password is required"
        } else if (v.length < 8) {
          errors.value.password = "At least 8 characters"
        } else if (!/[A-Z]/.test(v)) {
          errors.value.password = "Include at least one uppercase letter"
        } else if (!/[0-9]/.test(v)) {
          errors.value.password = "Include at least one number"
        } else {
          errors.value.password = ""
        }
        // re-validate confirm ถ้า user เคย touch แล้ว
        if (touched.value.confirm_password) validateField("confirm_password")
        break
      }
      case "confirm_password": {
        const v = form.value.confirm_password
        if (!v) {
          errors.value.confirm_password = "Please confirm your password"
        } else if (v !== form.value.password) {
          errors.value.confirm_password = "Passwords do not match"
        } else {
          errors.value.confirm_password = ""
        }
        break
      }
    }
  }

  function touchField(field: keyof typeof touched.value) {
    touched.value[field] = true
    validateField(field)
  }

  function touchAll() {
    ;(Object.keys(touched.value) as Array<keyof typeof touched.value>).forEach(
      f => {
        touched.value[f] = true
        validateField(f)
      },
    )
  }

  function isFormValid() {
    return (
      Object.values(errors.value).every(e => e === "") &&
      Object.values(form.value).every(v => v.trim() !== "")
    )
  }

  // ── Password input handler ────────────────────────────────────────────────
  function onPasswordInput() {
    validateField("password")
  }

  // ── Password strength ─────────────────────────────────────────────────────
  const passwordStrength = computed(() => {
    const p = form.value.password
    if (!p) return { score: 0, level: "none", label: "" }
    let score = 0
    if (p.length >= 8) score++
    if (p.length >= 12) score++
    if (/[A-Z]/.test(p) && /[a-z]/.test(p)) score++
    if (/[0-9]/.test(p)) score++
    if (/[^A-Za-z0-9]/.test(p)) score++

    if (score <= 1) return { score: 1, level: "weak", label: "Weak" }
    if (score === 2) return { score: 2, level: "fair", label: "Fair" }
    if (score === 3) return { score: 3, level: "good", label: "Good" }
    return { score: 4, level: "strong", label: "Strong" }
  })

  function strengthBarClass(index: number) {
    const s = passwordStrength.value
    if (s.score === 0) return ""
    const active = index <= s.score
    return active ? `strength-bar--${s.level}` : ""
  }

  // ── GSAP ──────────────────────────────────────────────────────────────────
  const containerRef = ref<HTMLElement | null>(null)
  const brandRef = ref<HTMLElement | null>(null)
  const dividerRef = ref<HTMLElement | null>(null)
  const formPanelRef = ref<HTMLElement | null>(null)

  onMounted(() => {
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
        { opacity: 1, y: 0, duration: 0.38, stagger: 0.07 },
        "-=0.25",
      )
  })

  // ── Submit ────────────────────────────────────────────────────────────────
  async function handleRegister() {
    serverError.value = ""
    touchAll()
    if (!isFormValid()) {
      gsap.fromTo(
        formPanelRef.value,
        { x: -8 },
        { x: 0, duration: 0.45, ease: "elastic.out(1, 0.3)" },
      )
      return
    }
    isLoading.value = true
    try {
      await authStore.register(form.value)
      await gsap.to(containerRef.value, {
        opacity: 0,
        y: -20,
        duration: 0.35,
        ease: "power2.in",
      })
      router.push({ name: "check-email" })
    } catch (e: any) {
      const status = e?.response?.status
      if (status === 409) {
        serverError.value =
          e?.response?.data?.error ?? "Email or username already in use"
      } else if (status >= 500) {
        serverError.value = "Server error. Please try again later."
      } else {
        serverError.value =
          e?.response?.data?.error ?? "Registration failed. Please try again."
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
    padding: 2.5rem;
  }
  .auth-title {
    font-family: "Noto Sans Thai", "Arial Black", sans-serif;
    font-size: 2rem;
    font-weight: 900;
    letter-spacing: 3px;
    text-align: center;
    color: #ffffff;
    margin-bottom: 1.5rem;
  }
  .auth-form {
    display: flex;
    flex-direction: column;
    gap: 0.75rem;
  }
  .form-group {
    display: flex;
    flex-direction: column;
    gap: 0.3rem;
  }
  .form-label {
    font-size: 0.82rem;
    font-weight: 500;
    color: #d1d1d1;
  }

  /* ── Input wrapper states ── */
  .input-wrapper {
    position: relative;
    display: flex;
    align-items: center;
    transition: filter 0.2s;
  }
  .input-wrapper--error .form-input {
    box-shadow: 0 0 0 2px rgba(245, 67, 44, 0.5);
    background: rgba(245, 67, 44, 0.06);
  }
  .input-wrapper--valid .form-input {
    box-shadow: 0 0 0 2px rgba(34, 197, 94, 0.45);
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
    z-index: 1;
  }
  .input-valid-icon {
    position: absolute;
    right: 12px;
    color: #22c55e;
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

  /* ── Per-field error ── */
  .field-error-msg {
    display: flex;
    align-items: center;
    gap: 4px;
    font-size: 0.75rem;
    color: #f87171;
    padding-left: 4px;
    margin: 0;
  }

  .field-error-enter-active {
    animation: fieldErrorIn 0.28s cubic-bezier(0.21, 1.02, 0.73, 1) forwards;
  }
  .field-error-leave-active {
    animation: fieldErrorOut 0.18s ease-in forwards;
  }
  @keyframes fieldErrorIn {
    from {
      opacity: 0;
      transform: translateY(-4px);
      max-height: 0;
    }
    to {
      opacity: 1;
      transform: translateY(0);
      max-height: 40px;
    }
  }
  @keyframes fieldErrorOut {
    from {
      opacity: 1;
      transform: translateY(0);
      max-height: 40px;
    }
    to {
      opacity: 0;
      transform: translateY(-4px);
      max-height: 0;
    }
  }

  /* ── Password strength ── */
  .password-strength {
    display: flex;
    align-items: center;
    gap: 8px;
    padding: 0 4px;
    margin-top: 2px;
  }
  .strength-bars {
    display: flex;
    gap: 4px;
    flex: 1;
  }
  .strength-bar {
    flex: 1;
    height: 3px;
    border-radius: 99px;
    background: rgba(255, 255, 255, 0.1);
    transition: background 0.3s;
  }
  .strength-bar--weak {
    background: #ef4444;
  }
  .strength-bar--fair {
    background: #f97316;
  }
  .strength-bar--good {
    background: #eab308;
  }
  .strength-bar--strong {
    background: #22c55e;
  }

  .strength-label {
    font-size: 0.7rem;
    font-weight: 600;
    letter-spacing: 0.5px;
    text-transform: uppercase;
    min-width: 40px;
    text-align: right;
  }
  .strength-label--weak {
    color: #ef4444;
  }
  .strength-label--fair {
    color: #f97316;
  }
  .strength-label--good {
    color: #eab308;
  }
  .strength-label--strong {
    color: #22c55e;
  }

  /* ── Global / server error alert ── */
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

  /* ── Button ── */
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
