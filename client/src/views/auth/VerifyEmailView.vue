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
        <template v-if="status === 'form'">
          <h1 class="auth-title">{{ t("auth.verifyEmail.title") }}</h1>
          <form v-if="!confirmingSkip" class="status-content" @submit.prevent="submit({ password })">
            <p class="status-text form-item">{{ t("auth.verifyEmail.passwordPrompt") }}</p>
            <p v-if="wrongPassword" class="status-text status-text--error" role="alert">
              {{ t("auth.verifyEmail.wrongPassword") }}
            </p>
            <AuthField
              v-model="password"
              :label="t('auth.verifyEmail.passwordLabel')"
              type="password"
              :icon="IconLock"
              autocomplete="current-password"
            />
            <AppButton type="submit" size="lg" block class="form-item" :disabled="!password">
              {{ t("auth.verifyEmail.confirm") }}
            </AppButton>
            <button type="button" class="link-muted form-item" @click="confirmingSkip = true">
              {{ t("auth.verifyEmail.skipPassword") }}
            </button>
          </form>
          <div v-else class="status-content" role="alertdialog">
            <p class="status-text status-text--error">{{ t("auth.verifyEmail.skipWarning") }}</p>
            <AppButton size="lg" block @click="submit({ skipPassword: true })">
              {{ t("auth.verifyEmail.skipConfirm") }}
            </AppButton>
            <button type="button" class="link-muted" @click="confirmingSkip = false">
              {{ t("auth.verifyEmail.skipCancel") }}
            </button>
          </div>
        </template>

        <template v-else-if="status === 'loading'">
          <h1 class="auth-title">VERIFYING</h1>
          <div class="status-content form-item">
            <div class="status-icon status-icon--loading">
              <AppSpinner :size="48" />
            </div>
            <p class="status-text">กำลังยืนยันอีเมลของคุณในระบบ...</p>
          </div>
        </template>

        <template v-else-if="status === 'success'">
          <h1 class="auth-title">SUCCESS</h1>
          <div class="status-content form-item">
            <div class="status-icon status-icon--success">
              <IconCircleCheck :size="48" />
            </div>
            <h2 class="status-subtitle">ยืนยันอีเมลสำเร็จ!</h2>
            <p v-if="passwordResetRequired" class="status-text">
              {{ t("auth.verifyEmail.resetRequired") }}
            </p>
            <p v-else class="status-text">
              บัญชีของคุณพร้อมใช้งานแล้ว สามารถเข้าสู่ระบบได้ทันที
            </p>

            <RouterLink v-if="passwordResetRequired" to="/forgot-password" class="btn-primary w-full mt-4">
              {{ t("auth.verifyEmail.resetCta") }}
            </RouterLink>
            <RouterLink v-else to="/login" class="btn-primary w-full mt-4">
              LOGIN NOW
            </RouterLink>
          </div>
        </template>

        <template v-else-if="status === 'error'">
          <h1 class="auth-title">EXPIRED</h1>
          <div class="status-content form-item">
            <div class="status-icon status-icon--error">
              <IconAlertCircle :size="48" />
            </div>
            <h2 class="status-subtitle">ลิงก์ไม่ถูกต้องหรือหมดอายุ</h2>
            <p v-if="attemptsExceeded" class="status-text">
              {{ t("auth.verifyEmail.attemptsExceeded") }}
            </p>
            <p v-else class="status-text">
              กรุณาลองกดส่งลิงก์ยืนยันอีเมลใหม่อีกครั้งด้านล่าง
            </p>

            <button
              class="btn-primary w-full mt-4"
              :disabled="resendStatus === 'sending' || resendStatus === 'sent'"
              @click="resend"
            >
              <span v-if="resendStatus === 'idle'">RESEND LINK</span>
              <AppSpinner v-else-if="resendStatus === 'sending'" />
              <span v-else-if="resendStatus === 'sent'">SENT! CHECK EMAIL</span>
              <span v-else>ERROR, TRY AGAIN</span>
            </button>

            <p class="form-footer mt-4">
              <RouterLink to="/login" class="link-muted"
                >Back to Login</RouterLink
              >
            </p>
          </div>
        </template>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import AppSpinner from "@/components/ui/AppSpinner.vue"
  import AppButton from "@/components/ui/AppButton.vue"
  import AuthField from "@/components/auth/AuthField.vue"
  import { ref, onMounted } from "vue"
  import { useRoute, RouterLink } from "vue-router"
  import { useI18n } from "vue-i18n"
  import { authApi } from "@/api/api"
  import { gsap } from "gsap"
  import {
    CircleCheck as IconCircleCheck,
    AlertCircle as IconAlertCircle,
    Lock as IconLock,
  } from "lucide-vue-next"

  const route = useRoute()
  const { t } = useI18n()

  type Status = "form" | "loading" | "success" | "error"
  type ResendStatus = "idle" | "sending" | "sent" | "failed"

  const status = ref<Status>("form")
  const password = ref("")
  const token = ref("")
  const passwordResetRequired = ref(false)
  const wrongPassword = ref(false)
  const confirmingSkip = ref(false)
  const attemptsExceeded = ref(false)
  const resendStatus = ref<ResendStatus>("idle")
  const email = ref("")

  const containerRef = ref<HTMLElement | null>(null)
  const brandRef = ref<HTMLElement | null>(null)
  const dividerRef = ref<HTMLElement | null>(null)
  const formPanelRef = ref<HTMLElement | null>(null)

  function runAnimation() {
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
  }

  onMounted(() => {
    runAnimation()
    token.value = (route.query.token as string) || ""
    // Opening the link must not verify anything by itself; the person confirms below.
    if (!token.value) status.value = "error"
  })

  // A wrong password keeps the form (and the token); only an explicit skip wipes credentials.
  async function submit(choice: { password: string } | { skipPassword: true }) {
    status.value = "loading"
    wrongPassword.value = false
    try {
      const { data } = await authApi.verifyEmail(token.value, choice)
      passwordResetRequired.value = data.password_reset_required
      status.value = "success"
    } catch (err) {
      const code = (err as { response?: { data?: { code?: string } } }).response?.data?.code
      if (code === "WRONG_PASSWORD") {
        wrongPassword.value = true
        confirmingSkip.value = false
        status.value = "form"
        return
      }
      attemptsExceeded.value = code === "VERIFY_ATTEMPTS_EXCEEDED"
      status.value = "error"
    }
  }

  async function resend() {
    if (!email.value) {
      const input = window.prompt("กรุณากรอกอีเมลของคุณ")
      if (!input) return
      email.value = input
    }

    resendStatus.value = "sending"
    try {
      await authApi.resendVerification(email.value)
      resendStatus.value = "sent"
    } catch {
      resendStatus.value = "failed"
    }
  }
</script>

<style scoped>
  /* ดึง CSS Shared ของ Auth มาใช้งานร่วมกัน */
  @import "@/assets/css/auth-shared.css";
  /* (แนะนำให้รวมสไตล์ตัวแปรไว้ในไฟล์กลาง หากไม่มี สามารถ copy style จาก LoginView ไปใส่ได้ครับ) */

  .status-content {
    display: flex;
    flex-direction: column;
    align-items: center;
    text-align: center;
    gap: 1rem;
  }
  .status-icon {
    display: flex;
    align-items: center;
    justify-content: center;
    width: 80px;
    height: 80px;
    border-radius: 50%;
    margin-bottom: 0.5rem;
  }
  .status-icon--loading {
    color: var(--color-brand);
    background: rgba(245, 67, 44, 0.1);
  }
  .status-icon--success {
    color: #22c55e;
    background: rgba(34, 197, 94, 0.1);
  }
  .status-icon--error {
    color: #ef4444;
    background: rgba(239, 68, 68, 0.1);
  }
  .status-subtitle {
    font-size: 1.25rem;
    font-weight: 700;
    color: #ffffff;
  }
  .status-text--error {
    color: #ff5a52;
  }
  .status-text {
    font-size: 0.9rem;
    color: #a3a3a3;
    line-height: 1.6;
    max-width: 320px;
  }
</style>
