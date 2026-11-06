<template>
  <div class="auth-page">
    <div class="auth-bg"><div class="auth-bg__overlay" /></div>
    <div class="auth-container" ref="containerRef">
      <div class="auth-brand" ref="brandRef">
        <div class="brand-logo">
          <span class="brand-movie">POP</span><span class="brand-hub">CORN</span>
        </div>
        <p class="brand-tagline">
          Forgot your password?<br />
          No worries — we'll send you a reset link.
        </p>
      </div>
      <div class="auth-divider" ref="dividerRef" />
      <div class="auth-form-panel" ref="formPanelRef">
        <!-- ── Sent state ── -->
        <template v-if="sent">
          <div class="state-icon">
            <IconMailCheck :size="30" />
          </div>
          <h1 class="auth-title">CHECK YOUR EMAIL</h1>
          <p class="auth-subtitle">
            If an account exists for <strong>{{ submittedEmail }}</strong
            >,<br />
            a reset link has been sent. Check your inbox.
          </p>
          <div class="sent-actions form-item">
            <button
              type="button"
              class="btn-ghost"
              :disabled="resendCooldown > 0"
              @click="handleResend"
            >
              <span v-if="resendCooldown > 0" class="cooldown-text">
                Resend in {{ resendCooldown }}s
              </span>
              <span v-else>Resend email</span>
            </button>
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
              Back to Login
            </RouterLink>
          </div>
        </template>

        <!-- ── Form state ── -->
        <template v-else>
          <h1 class="auth-title">FORGOT PASSWORD</h1>
          <form @submit.prevent="handleSubmit" class="auth-form" novalidate>
            <div class="form-group form-item">
              <label class="form-label">Email address</label>
              <div
                class="input-wrapper"
                :class="{
                  'input-wrapper--error': touched && error,
                  'input-wrapper--valid': touched && !error && email,
                }"
              >
                <span class="input-icon"><IconMail :size="16" /></span>
                <input
                  v-model="email"
                  type="email"
                  class="form-input"
                  autocomplete="email"
                  placeholder="your@email.com"
                  :disabled="isLoading"
                  @input="validate"
                  @blur="handleBlur"
                />
                <span
                  v-if="touched && !error && email"
                  class="input-valid-icon"
                >
                  <IconCheck :size="14" />
                </span>
              </div>
              <Transition name="field-error">
                <p v-if="touched && error" class="field-error-msg">
                  <IconAlertCircle :size="12" /> {{ error }}
                </p>
              </Transition>
            </div>

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
              :disabled="isLoading || !!error || !email"
            >
              <AppSpinner v-if="isLoading" /><span v-else
                >SEND RESET LINK</span
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
import AppSpinner from "@/components/ui/AppSpinner.vue"
  import { ref, onMounted, onUnmounted } from "vue"
  import { gsap } from "gsap"
  import {
    Mail as IconMail,
    MailCheck as IconMailCheck,
    AlertCircle as IconAlertCircle,
    Check as IconCheck,
    X as IconX,
  } from "lucide-vue-next"
  import { authApi } from "@/api/api"

  const email = ref("")
  const touched = ref(false)
  const error = ref("")
  const isLoading = ref(false)
  const serverError = ref("")
  const sent = ref(false)
  const submittedEmail = ref("")
  const resendCooldown = ref(0)
  let cooldownTimer: ReturnType<typeof setInterval> | null = null

  const EMAIL_RE = /^[^\s@]+@[^\s@]+\.[^\s@]+$/

  function validate() {
    const v = email.value.trim()
    if (!v) {
      error.value = "Email is required"
      return
    }
    if (!EMAIL_RE.test(v)) {
      error.value = "Enter a valid email address"
      return
    }
    error.value = ""
  }

  function startCooldown() {
    resendCooldown.value = 60
    cooldownTimer = setInterval(() => {
      resendCooldown.value--
      if (resendCooldown.value <= 0 && cooldownTimer) {
        clearInterval(cooldownTimer)
        cooldownTimer = null
      }
    }, 1000)
  }

  async function handleSubmit() {
    touched.value = true
    validate()
    if (error.value || !email.value) {
      gsap.fromTo(
        formPanelRef.value,
        { x: -8 },
        { x: 0, duration: 0.45, ease: "elastic.out(1, 0.3)" },
      )
      return
    }
    serverError.value = ""
    isLoading.value = true
    try {
      await authApi.forgotPassword({ email: email.value.trim() })
      submittedEmail.value = email.value.trim()
      // fade ออกก่อน transition เป็น sent state
      await gsap.to(formPanelRef.value, {
        opacity: 0,
        y: -12,
        duration: 0.25,
        ease: "power2.in",
      })
      sent.value = true
      await gsap.fromTo(
        formPanelRef.value,
        { opacity: 0, y: 12 },
        { opacity: 1, y: 0, duration: 0.4, ease: "power3.out" },
      )
      startCooldown()
    } catch {
      serverError.value = "Something went wrong. Please try again."
      gsap.fromTo(
        formPanelRef.value,
        { x: -8 },
        { x: 0, duration: 0.45, ease: "elastic.out(1, 0.3)" },
      )
    } finally {
      isLoading.value = false
    }
  }

  async function handleResend() {
    if (resendCooldown.value > 0) return
    try {
      await authApi.forgotPassword({ email: submittedEmail.value })
      startCooldown()
    } catch {
      /* silent */
    }
  }

  function handleBlur() {
    touched.value = true
    validate()
  }

  // ── GSAP entrance ──────────────────────────────────────────────────────
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

  onUnmounted(() => {
    if (cooldownTimer) clearInterval(cooldownTimer)
  })
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

  /* ── sent state ── */
  .state-icon {
    width: 56px;
    height: 56px;
    border-radius: 50%;
    background: rgba(245, 67, 44, 0.12);
    border: 1px solid rgba(245, 67, 44, 0.25);
    color: var(--color-brand);
    display: flex;
    align-items: center;
    justify-content: center;
    margin: 0 auto 1.25rem;
  }
  .auth-subtitle {
    font-size: 0.85rem;
    color: #a3a3a3;
    text-align: center;
    line-height: 1.75;
    margin-bottom: 1.75rem;
  }
  .auth-subtitle strong {
    color: #e0e0e0;
    font-weight: 500;
  }
  .sent-actions {
    display: flex;
    flex-direction: column;
    gap: 10px;
  }
  .btn-ghost {
    width: 100%;
    padding: 0.72rem;
    background: transparent;
    color: #a3a3a3;
    border: 1px solid rgba(255, 255, 255, 0.1);
    border-radius: 9999px;
    font-family: "Noto Sans Thai", "Arial Black", sans-serif;
    font-size: 0.9rem;
    font-weight: 700;
    letter-spacing: 1px;
    cursor: pointer;
    transition:
      color 0.2s,
      border-color 0.2s;
    display: flex;
    align-items: center;
    justify-content: center;
    min-height: 46px;
  }
  .btn-ghost:hover:not(:disabled) {
    color: #fff;
    border-color: rgba(255, 255, 255, 0.25);
  }
  .btn-ghost:disabled {
    opacity: 0.5;
    cursor: not-allowed;
  }
  .cooldown-text {
    font-size: 0.85rem;
    color: #777;
  }

  /* ── form ── */
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

  /* ── per-field error ── */
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

  /* ── server error alert ── */
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

  /* ── button ── */
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
