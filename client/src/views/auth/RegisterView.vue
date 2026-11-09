<template>
  <AuthShell ref="shellRef" :title="t('auth.register.title')" :subtitle="t('auth.register.subtitle')">
    <form class="auth-form" novalidate @submit.prevent="handleRegister">
      <AuthField
        v-model="form.username"
        :label="t('auth.register.usernameLabel')"
        :icon="IconUser"
        autocomplete="username"
        :disabled="isLoading"
        :error="fieldError('username')"
        :valid="isFieldValid('username')"
        @input="validateField('username')"
        @blur="touchField('username')"
      />
      <AuthField
        v-model="form.email"
        :label="t('auth.register.emailLabel')"
        type="email"
        :icon="IconMail"
        autocomplete="email"
        :disabled="isLoading"
        :error="fieldError('email')"
        :valid="isFieldValid('email')"
        @input="validateField('email')"
        @blur="touchField('email')"
      />
      <AuthField
        v-model="form.password"
        :label="t('auth.register.passwordLabel')"
        :hint="t('auth.register.passwordHint')"
        type="password"
        :icon="IconLock"
        autocomplete="new-password"
        :disabled="isLoading"
        :error="fieldError('password')"
        :valid="isFieldValid('password')"
        @input="validateField('password')"
        @blur="touchField('password')"
      >
        <PasswordStrength :password="form.password" />
      </AuthField>
      <AuthField
        v-model="form.confirm_password"
        :label="t('auth.register.confirmPasswordLabel')"
        type="password"
        :icon="IconLock"
        autocomplete="new-password"
        :disabled="isLoading"
        :error="fieldError('confirm_password')"
        :valid="isFieldValid('confirm_password')"
        @input="validateField('confirm_password')"
        @blur="touchField('confirm_password')"
      />
      <AuthAlert :message="serverError" @dismiss="serverError = ''" />
      <AppButton type="submit" size="lg" block class="form-item" :loading="isLoading">
        {{ t("auth.register.submit") }}
      </AppButton>
      <GoogleSignInButton :disabled="isLoading" @click="handleGoogleLogin" />
      <p class="auth-footer form-item">
        <span class="auth-footer__text">{{ t("auth.register.hasAccount") }}</span>
        <RouterLink to="/login" class="auth-link auth-link--strong">{{ t("auth.register.login") }}</RouterLink>
      </p>
    </form>
  </AuthShell>
</template>

<script setup lang="ts">
  import { ref } from "vue"
  import { useRouter } from "vue-router"
  import { User as IconUser, Mail as IconMail, Lock as IconLock } from "lucide-vue-next"
  import AuthAlert from "@/components/auth/AuthAlert.vue"
  import AuthField from "@/components/auth/AuthField.vue"
  import AuthShell from "@/components/auth/AuthShell.vue"
  import GoogleSignInButton from "@/components/auth/GoogleSignInButton.vue"
  import PasswordStrength from "@/components/auth/PasswordStrength.vue"
  import AppButton from "@/components/ui/AppButton.vue"
  import { useAuthStore } from "@/stores/auth"
  import { authApi } from "@/api/api"
  import { useI18n } from "vue-i18n"
  import { isRateLimited } from "@/utils/rateLimit"

  const { t } = useI18n()

  const router = useRouter()
  const authStore = useAuthStore()

  async function handleGoogleLogin() {
    try {
      const { data } = await authApi.getGoogleConfig()
      if (!data.enabled) {
        serverError.value = t("auth.login.errors.googleDisabled")
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
          errors.value.username = t("auth.register.errors.usernameRequired")
        } else if (v.length < 3) {
          errors.value.username = t("auth.register.errors.usernameMin")
        } else if (v.length > 20) {
          errors.value.username = t("auth.register.errors.usernameMax")
        } else if (!USERNAME_RE.test(v)) {
          errors.value.username = t("auth.register.errors.usernameChars")
        } else {
          errors.value.username = ""
        }
        break
      }
      case "email": {
        const v = form.value.email.trim()
        if (!v) {
          errors.value.email = t("auth.register.errors.emailRequired")
        } else if (!EMAIL_RE.test(v)) {
          errors.value.email = t("auth.register.errors.emailInvalid")
        } else {
          errors.value.email = ""
        }
        break
      }
      case "password": {
        const v = form.value.password
        if (!v) {
          errors.value.password = t("auth.register.errors.passwordRequired")
        } else if (v.length < 8) {
          errors.value.password = t("auth.register.errors.passwordMin")
        } else if (!/[A-Z]/.test(v)) {
          errors.value.password = t("auth.register.errors.passwordUpper")
        } else if (!/[0-9]/.test(v)) {
          errors.value.password = t("auth.register.errors.passwordNumber")
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
          errors.value.confirm_password = t("auth.register.errors.confirmRequired")
        } else if (v !== form.value.password) {
          errors.value.confirm_password = t("auth.register.errors.confirmMismatch")
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

  function fieldError(field: keyof typeof errors.value) {
    return touched.value[field] ? errors.value[field] : ""
  }

  function isFieldValid(field: keyof typeof errors.value) {
    return touched.value[field] && !errors.value[field] && !!form.value[field]
  }

  const shellRef = ref<InstanceType<typeof AuthShell> | null>(null)

  // ── Submit ────────────────────────────────────────────────────────────────
  async function handleRegister() {
    serverError.value = ""
    touchAll()
    if (!isFormValid()) {
      shellRef.value?.shake()
      return
    }
    isLoading.value = true
    try {
      await authStore.register(form.value)
      await shellRef.value?.playExit()
      router.push({ name: "check-email" })
    } catch (e: any) {
      const status = e?.response?.status
      if (isRateLimited(e)) {
        serverError.value = t("errors.rateLimited")
      } else if (status === 409) {
        serverError.value =
          e?.response?.data?.error ?? t("auth.register.errors.conflict")
      } else if (status >= 500) {
        serverError.value = t("auth.register.errors.server")
      } else {
        serverError.value =
          e?.response?.data?.error ?? t("auth.register.errors.failed")
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
