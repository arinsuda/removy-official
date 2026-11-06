<template>
  <div class="auth-page">
    <div class="auth-bg"><div class="auth-bg__overlay" /></div>
    <div class="auth-container" ref="containerRef">
      <div class="auth-brand" ref="brandRef">
        <div class="brand-logo">
          <span class="brand-movie">POP</span><span class="brand-hub">CORN</span>
        </div>
        <p class="brand-tagline">
          Set a new password for your account.<br />
          Make it strong and memorable.
        </p>
      </div>
      <div class="auth-divider" ref="dividerRef" />
      <div class="auth-form-panel" ref="formPanelRef">
        <!-- ── Invalid / expired token ── -->
        <template v-if="tokenError">
          <div class="state-icon state-icon--error">
            <IconShieldOff :size="32" />
          </div>
          <h1 class="auth-title">Link Expired</h1>
          <p class="auth-subtitle">
            This reset link is invalid or has expired.<br />
            Please request a new one.
          </p>
          <RouterLink
            to="/forgot-password"
            class="btn-primary"
            style="
              text-decoration: none;
              display: flex;
              align-items: center;
              justify-content: center;
            "
          >
            Request New Link
          </RouterLink>
          <p class="form-footer form-item">
            <RouterLink to="/login" class="link-underline"
              >Back to login</RouterLink
            >
          </p>
        </template>

        <!-- ── Success ── -->
        <template v-else-if="success">
          <div class="state-icon state-icon--success">
            <IconShieldCheck :size="32" />
          </div>
          <h1 class="auth-title">Password Reset!</h1>
          <p class="auth-subtitle">
            Your password has been updated.<br />
            You can now log in with your new password.
          </p>
          <RouterLink
            to="/login"
            class="btn-primary"
            style="
              text-decoration: none;
              display: flex;
              align-items: center;
              justify-content: center;
            "
          >
            Go to Login
          </RouterLink>
        </template>

        <!-- ── Form ── -->
        <template v-else>
          <h1 class="auth-title">NEW PASSWORD</h1>
          <form @submit.prevent="handleSubmit" class="auth-form" novalidate>
            <div class="form-group form-item">
              <label class="form-label">New Password</label>
              <div class="input-wrapper">
                <span class="input-icon"><IconLock :size="16" /></span>
                <input
                  v-model="form.new_password"
                  :type="showNew ? 'text' : 'password'"
                  class="form-input"
                  autocomplete="new-password"
                  :disabled="isLoading"
                  placeholder="At least 8 characters"
                />
                <button
                  type="button"
                  class="input-toggle"
                  @click="showNew = !showNew"
                  tabindex="-1"
                >
                  <IconEyeOff v-if="showNew" :size="16" /><IconEye
                    v-else
                    :size="16"
                  />
                </button>
              </div>
              <!-- password strength bar -->
              <div class="strength-bar" v-if="form.new_password">
                <div
                  class="strength-bar__fill"
                  :class="`strength-bar__fill--${strength.level}`"
                  :style="{ width: strength.width }"
                />
              </div>
              <span
                v-if="form.new_password"
                class="strength-label"
                :class="`strength-label--${strength.level}`"
              >
                {{ strength.label }}
              </span>
            </div>

            <div class="form-group form-item">
              <label class="form-label">Confirm New Password</label>
              <div class="input-wrapper">
                <span class="input-icon"><IconLock :size="16" /></span>
                <input
                  v-model="form.confirm_password"
                  :type="showConfirm ? 'text' : 'password'"
                  class="form-input"
                  autocomplete="new-password"
                  :disabled="isLoading"
                  placeholder="Repeat your new password"
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
              <Transition name="inline-err">
                <span
                  v-if="
                    form.confirm_password &&
                    form.new_password !== form.confirm_password
                  "
                  class="inline-error"
                >
                  <IconAlertCircle :size="12" /> Passwords do not match
                </span>
              </Transition>
            </div>

            <Transition name="alert">
              <div v-if="errorMsg" class="form-alert" role="alert">
                <span class="form-alert__icon"
                  ><IconAlertCircle :size="15"
                /></span>
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

            <button
              type="submit"
              class="btn-primary form-item"
              :disabled="isLoading || !canSubmit"
            >
              <span v-if="isLoading" class="spinner" /><span v-else
                >RESET PASSWORD</span
              >
            </button>
            <p class="form-footer form-item">
              <RouterLink to="/login" class="link-underline"
                >Back to login</RouterLink
              >
            </p>
          </form>
        </template>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
  import { ref, computed, onMounted } from "vue"
  import { useRoute, useRouter } from "vue-router"
  import { gsap } from "gsap"
  import {
    Lock as IconLock,
    Eye as IconEye,
    EyeOff as IconEyeOff,
    AlertCircle as IconAlertCircle,
    X as IconX,
    ShieldOff as IconShieldOff,
    ShieldCheck as IconShieldCheck,
  } from "lucide-vue-next"
  import { authApi } from "@/api/api"

  const route = useRoute()
  const router = useRouter()

  const token = route.query.token as string | undefined
  const userId = route.query.uid ? Number(route.query.uid) : undefined

  const tokenError = ref(!token || !userId || isNaN(userId as number))
  const success = ref(false)
  const isLoading = ref(false)
  const errorMsg = ref("")
  const showNew = ref(false)
  const showConfirm = ref(false)

  const form = ref({ new_password: "", confirm_password: "" })

  // ── password strength ─────────────────────────────────────────────────
  const strength = computed(() => {
    const p = form.value.new_password
    if (!p) return { level: "none", label: "", width: "0%" }
    let score = 0
    if (p.length >= 8) score++
    if (p.length >= 12) score++
    if (/[A-Z]/.test(p)) score++
    if (/[0-9]/.test(p)) score++
    if (/[^A-Za-z0-9]/.test(p)) score++
    if (score <= 1) return { level: "weak", label: "Weak", width: "25%" }
    if (score <= 2) return { level: "fair", label: "Fair", width: "50%" }
    if (score <= 3) return { level: "good", label: "Good", width: "75%" }
    return { level: "strong", label: "Strong", width: "100%" }
  })

  const canSubmit = computed(
    () =>
      form.value.new_password.length >= 8 &&
      form.value.new_password === form.value.confirm_password,
  )

  // ── animations ────────────────────────────────────────────────────────
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
        { opacity: 1, y: 0, duration: 0.4, stagger: 0.08 },
        "-=0.25",
      )
  })

  // ── submit ────────────────────────────────────────────────────────────
  async function handleSubmit() {
    if (!canSubmit.value) return
    errorMsg.value = ""
    isLoading.value = true
    try {
      await authApi.resetPassword({
        token: token!,
        user_id: userId!,
        new_password: form.value.new_password,
        confirm_password: form.value.confirm_password,
      })
      success.value = true
    } catch (e: any) {
      const status = e?.response?.status
      if (status === 404 || status === 410 || status === 422) {
        tokenError.value = true
      } else if (status === 400) {
        errorMsg.value = e?.response?.data?.error ?? "Invalid request."
      } else {
        errorMsg.value = "Something went wrong. Please try again."
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
  /* ── base (copied from LoginView for consistency) ── */
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
    margin-bottom: 0.5rem;
  }
  .auth-subtitle {
    font-size: 0.85rem;
    color: #a3a3a3;
    text-align: center;
    line-height: 1.7;
    margin-bottom: 1.75rem;
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
    padding: 0;
    width: 16px;
    height: 16px;
    transition: color 0.2s;
  }
  .input-toggle:hover {
    color: var(--color-brand);
  }

  /* strength bar */
  .strength-bar {
    height: 3px;
    background: rgba(255, 255, 255, 0.1);
    border-radius: 9999px;
    overflow: hidden;
    margin-top: 6px;
  }
  .strength-bar__fill {
    height: 100%;
    border-radius: 9999px;
    transition:
      width 0.3s ease,
      background 0.3s ease;
  }
  .strength-bar__fill--weak {
    background: var(--color-brand);
  }
  .strength-bar__fill--fair {
    background: #f59e0b;
  }
  .strength-bar__fill--good {
    background: #3b82f6;
  }
  .strength-bar__fill--strong {
    background: #22c55e;
  }
  .strength-label {
    font-size: 0.72rem;
    margin-top: 2px;
  }
  .strength-label--weak {
    color: var(--color-brand);
  }
  .strength-label--fair {
    color: #f59e0b;
  }
  .strength-label--good {
    color: #3b82f6;
  }
  .strength-label--strong {
    color: #22c55e;
  }

  /* inline password mismatch */
  .inline-error {
    display: flex;
    align-items: center;
    gap: 4px;
    font-size: 0.75rem;
    color: var(--color-brand);
    margin-top: 2px;
  }
  .inline-err-enter-active {
    transition: all 0.2s ease;
  }
  .inline-err-leave-active {
    transition: all 0.15s ease;
  }
  .inline-err-enter-from,
  .inline-err-leave-to {
    opacity: 0;
    transform: translateY(-4px);
  }

  /* alert box */
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

  /* state icon (success / error) */
  .state-icon {
    width: 64px;
    height: 64px;
    border-radius: 50%;
    display: flex;
    align-items: center;
    justify-content: center;
    margin: 0 auto 1.25rem;
  }
  .state-icon--success {
    background: rgba(34, 197, 94, 0.15);
    color: #22c55e;
  }
  .state-icon--error {
    background: rgba(245, 67, 44, 0.15);
    color: var(--color-brand);
  }

  /* cta button */
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
</style>
