<template>
  <AuthShell ref="shellRef" title="LOGIN">
    <form class="auth-form" novalidate @submit.prevent="handleLogin">
      <AuthField
        v-model="form.identifier"
        label="Username or email"
        :icon="IconUser"
        autocomplete="username"
        :disabled="isLoading"
      />
      <AuthField
        v-model="form.password"
        label="Password"
        type="password"
        :icon="IconLock"
        autocomplete="current-password"
        :disabled="isLoading"
      />
      <AuthAlert :message="errorMsg" @dismiss="errorMsg = ''" />
      <div class="auth-forgot form-item">
        <RouterLink to="/forgot-password" class="auth-link">forget password</RouterLink>
      </div>
      <AppButton type="submit" size="lg" block class="form-item" :loading="isLoading">
        LOGIN
      </AppButton>
      <GoogleSignInButton :disabled="isLoading" @click="handleGoogleLogin" />
      <p class="auth-footer form-item">
        <RouterLink to="/register" class="auth-link">I don't have an account.</RouterLink>
      </p>
    </form>
  </AuthShell>
</template>

<script setup lang="ts">
  import { ref, onMounted } from "vue"
  import { useRouter, useRoute } from "vue-router"
  import { useI18n } from "vue-i18n"
  import { isRateLimited } from "@/utils/rateLimit"
  import { User as IconUser, Lock as IconLock } from "lucide-vue-next"
  import AuthAlert from "@/components/auth/AuthAlert.vue"
  import AuthField from "@/components/auth/AuthField.vue"
  import AuthShell from "@/components/auth/AuthShell.vue"
  import GoogleSignInButton from "@/components/auth/GoogleSignInButton.vue"
  import AppButton from "@/components/ui/AppButton.vue"
  import { useAuthStore } from "@/stores/auth"
  import { authApi } from "@/api/api"

  const { t } = useI18n()
  const router = useRouter()
  const route = useRoute()
  const authStore = useAuthStore()

  const form = ref({ identifier: "", password: "" })
  const isLoading = ref(false)
  const errorMsg = ref("")

  const shellRef = ref<InstanceType<typeof AuthShell> | null>(null)

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
      } else if (errParam === "google_email_verification_required") {
        errorMsg.value = t("auth.login.googleVerificationRequired")
      } else if (errParam === "session_expired") {
        errorMsg.value = t("auth.login.sessionExpired")
      } else {
        errorMsg.value = "เกิดข้อผิดพลาดในการเข้าสู่ระบบด้วย Google กรุณาลองใหม่อีกครั้ง"
      }
      router.replace({ query: {} })
    }
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
      shellRef.value?.shake()
      return
    }
    isLoading.value = true
    try {
      await authStore.login(form.value)
      const redirect = route.query.redirect as string | undefined
      await shellRef.value?.playExit()
      router.push(redirect ?? { name: "home" })
    } catch (e: any) {
      const status = e?.response?.status

      if (isRateLimited(e)) {
        errorMsg.value = t("errors.rateLimited")
      } else if (status === 401) {
        errorMsg.value = "ชื่อผู้ใช้หรือรหัสผ่านไม่ถูกต้อง"
      } else if (status >= 500) {
        errorMsg.value = "เกิดข้อผิดพลาดจากเซิร์ฟเวอร์ กรุณาลองใหม่ภายหลัง"
      } else {
        errorMsg.value = "เกิดข้อผิดพลาด กรุณาลองใหม่"
      }

      shellRef.value?.shake()
    } finally {
      isLoading.value = false
    }
  }
</script>

<style scoped>
  .auth-form {
    display: flex;
    flex-direction: column;
    gap: 1rem;
  }
  .auth-forgot {
    margin-top: -0.25rem;
    text-align: right;
  }
  .auth-footer {
    margin: 0;
    text-align: center;
  }
  .auth-link {
    font-size: 0.8125rem;
    color: var(--color-text-secondary);
    text-decoration: underline;
    text-decoration-color: transparent;
    text-underline-offset: 3px;
    transition:
      color var(--motion-fast) var(--ease-out-soft),
      text-decoration-color var(--motion-fast) var(--ease-out-soft);
  }
  @media (hover: hover) {
    .auth-link:hover {
      color: #fff;
      text-decoration-color: currentcolor;
    }
  }
</style>
