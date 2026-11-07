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
        <h1 class="auth-title">CHECK EMAIL</h1>

        <div class="status-content form-item">
          <div class="status-icon status-icon--info">
            <IconMailCheck :size="48" />
          </div>

          <p class="status-text">
            เราได้ส่งลิงก์สำหรับยืนยันตนไปที่อีเมลของคุณเรียบร้อยแล้ว<br />
            <span class="user-email">{{
              email || "your-email@example.com"
            }}</span>
          </p>

          <div class="info-alert">
            ไม่พบอีเมล? ลองตรวจสอบใน
            <strong>Spam / จดหมายขยะ</strong> หรือกดปุ่มด้านล่างเพื่อส่งอีกครั้ง
          </div>

          <button
            class="btn-primary w-full mt-2"
            :disabled="resendStatus === 'sending' || resendStatus === 'sent'"
            @click="resend"
          >
            <span v-if="resendStatus === 'idle'">RESEND EMAIL</span>
            <AppSpinner v-else-if="resendStatus === 'sending'" />
            <span v-else-if="resendStatus === 'sent'">SENT SUCCESSFUL</span>
            <span v-else>ERROR, TRY AGAIN</span>
          </button>
          <p v-if="rateLimited" class="status-text status-text--error" role="alert">
            {{ t("errors.rateLimited") }}
          </p>

          <p class="form-footer mt-4">
            <RouterLink to="/login" class="link-muted">
              กลับสู่หน้าเข้าสู่ระบบ
            </RouterLink>
          </p>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import AppSpinner from "@/components/ui/AppSpinner.vue"
  import { ref, onMounted } from "vue"
  import { RouterLink } from "vue-router"
  import { useAuthStore } from "@/stores/auth"
  import { authApi } from "@/api/api"
  import { useI18n } from "vue-i18n"
  import { isRateLimited } from "@/utils/rateLimit"

  import { gsap } from "gsap"
  import { MailCheck as IconMailCheck } from "lucide-vue-next"

  const { t } = useI18n()
  const authStore = useAuthStore()
  const email = ref(authStore.user?.email ?? "")

  type ResendStatus = "idle" | "sending" | "sent" | "failed"
  const resendStatus = ref<ResendStatus>("idle")
  const rateLimited = ref(false)

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

  async function resend() {
    if (!email.value) {
      const input = window.prompt("กรุณากรอกอีเมลของคุณ")
      if (!input) return
      email.value = input
    }

    rateLimited.value = false
    resendStatus.value = "sending"
    try {
      await authApi.resendVerification(email.value)
      resendStatus.value = "sent"
    } catch (e) {
      rateLimited.value = isRateLimited(e)
      resendStatus.value = "failed"
    }
  }
</script>

<style scoped>
  @import "@/assets/css/auth-shared.css";

  .status-content {
    display: flex;
    flex-direction: column;
    align-items: center;
    text-align: center;
    gap: 1.2rem;
  }
  .status-icon {
    display: flex;
    align-items: center;
    justify-content: center;
    width: 80px;
    height: 80px;
    border-radius: 50%;
  }
  .status-icon--info {
    color: var(--color-brand);
    background: rgba(245, 67, 44, 0.1);
  }
  .status-text {
    font-size: 0.95rem;
    color: #d1d1d1;
    line-height: 1.6;
  }
  .user-email {
    display: inline-block;
    margin-top: 0.25rem;
    color: #ffffff;
    font-weight: 600;
    text-decoration: underline;
    text-underline-offset: 4px;
    color: var(--color-brand);
  }
  .info-alert {
    padding: 0.75rem 1rem;
    background: rgba(255, 255, 255, 0.03);
    border: 1px solid rgba(255, 255, 255, 0.08);
    border-radius: 12px;
    font-size: 0.82rem;
    color: #a3a3a3;
    line-height: 1.5;
  }
</style>
