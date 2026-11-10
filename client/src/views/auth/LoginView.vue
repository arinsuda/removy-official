<template>
  <AuthShell ref="shellRef" :title="t('auth.login.title')" :subtitle="t('auth.login.subtitle')">
    <form class="auth-form" novalidate @submit.prevent="handleLogin">
      <AuthField
        v-model="form.identifier"
        :label="t('auth.login.identifierLabel')"
        :icon="IconUser"
        autocomplete="username"
        :disabled="isLoading"
      />
      <AuthField
        v-model="form.password"
        :label="t('auth.login.passwordLabel')"
        type="password"
        :icon="IconLock"
        autocomplete="current-password"
        :disabled="isLoading"
      />
      <AuthAlert :message="errorMsg" @dismiss="dismissError" />
      <UnverifiedResend v-if="unverified" :identifier="form.identifier" />
      <div class="auth-forgot form-item">
        <RouterLink to="/forgot-password" class="auth-link">{{ t("auth.login.forgotPassword") }}</RouterLink>
      </div>
      <AppButton type="submit" size="lg" block class="form-item" :loading="isLoading">
        {{ t("auth.login.submit") }}
      </AppButton>
      <GoogleSignInButton :disabled="isLoading" @click="handleGoogleLogin" />
      <p class="auth-footer form-item">
        <span class="auth-footer__text">{{ t("auth.login.noAccount") }}</span>
        <RouterLink to="/register" class="auth-link auth-link--strong">{{ t("auth.login.register") }}</RouterLink>
      </p>
    </form>
  </AuthShell>
</template>

<script setup lang="ts">
  import { ref, onMounted } from "vue"
  import { useRouter, useRoute } from "vue-router"
  import { useI18n } from "vue-i18n"
  import { isRateLimited } from "@/utils/rateLimit"
  import { asApiError } from "@/utils/apiError"
  import { User as IconUser, Lock as IconLock } from "lucide-vue-next"
  import AuthAlert from "@/components/auth/AuthAlert.vue"
  import AuthField from "@/components/auth/AuthField.vue"
  import AuthShell from "@/components/auth/AuthShell.vue"
  import UnverifiedResend from "@/components/auth/UnverifiedResend.vue"
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
  const unverified = ref(false)

  function dismissError() {
    errorMsg.value = ""
    unverified.value = false
  }

  const shellRef = ref<InstanceType<typeof AuthShell> | null>(null)

  onMounted(() => {
    const errParam = route.query.error as string | undefined
    if (errParam) {
      if (errParam === "user_cancelled") {
        errorMsg.value = t("auth.login.errors.cancelled")
      } else if (errParam === "google_account_link_required") {
        errorMsg.value = t("auth.login.errors.linkRequired")
      } else if (errParam === "google_identity_already_connected") {
        errorMsg.value = t("auth.login.errors.identityConnected")
      } else if (errParam === "account_disabled") {
        errorMsg.value = t("auth.login.errors.disabled")
      } else if (errParam === "google_email_verification_required") {
        errorMsg.value = t("auth.login.googleVerificationRequired")
      } else if (errParam === "session_expired") {
        errorMsg.value = t("auth.login.sessionExpired")
      } else {
        errorMsg.value = t("auth.login.errors.google")
      }
      router.replace({ query: {} })
    }
  })

  async function handleGoogleLogin() {
    try {
      const { data } = await authApi.getGoogleConfig()
      if (!data.enabled) {
        errorMsg.value = t("auth.login.errors.googleDisabled")
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
    dismissError()
    if (!form.value.identifier || !form.value.password) {
      errorMsg.value = t("auth.login.errors.fill")
      shellRef.value?.shake()
      return
    }
    isLoading.value = true
    try {
      await authStore.login(form.value)
      const redirect = route.query.redirect as string | undefined
      await shellRef.value?.playExit()
      router.push(redirect ?? { name: "home" })
    } catch (err) {
      const e = asApiError(err)
      const status = e.response?.status

      if (isRateLimited(err)) {
        errorMsg.value = t("errors.rateLimited")
      } else if (status === 403) {
        errorMsg.value = t("auth.login.unverified")
        unverified.value = true
      } else if (status === 401) {
        errorMsg.value = t("auth.login.errors.invalid")
      } else if ((status ?? 0) >= 500) {
        errorMsg.value = t("auth.login.errors.server")
      } else {
        errorMsg.value = t("auth.login.errors.generic")
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
    display: flex;
    flex-wrap: wrap;
    justify-content: center;
    gap: 0.375rem;
    margin: 0;
    font-size: var(--text-small);
    color: var(--color-text-muted);
  }
  .auth-link {
    font-size: var(--text-small);
    color: var(--color-text-secondary);
    text-decoration: underline;
    text-decoration-color: transparent;
    text-underline-offset: 3px;
    transition:
      color var(--motion-fast) var(--ease-out-soft),
      text-decoration-color var(--motion-fast) var(--ease-out-soft);
  }
  .auth-link--strong {
    font-weight: 600;
    color: var(--color-text-primary);
    text-decoration-color: currentcolor;
  }
  @media (hover: hover) {
    .auth-link:hover {
      color: var(--color-text-primary);
      text-decoration-color: currentcolor;
    }
  }
  @media (prefers-reduced-motion: reduce) {
    .auth-link {
      transition: none;
    }
  }
</style>
