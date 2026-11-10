<template>
  <div class="erase">
    <p class="erase__text">{{ t("settings.erase.description", { days: ERASURE_GRACE_DAYS }) }}</p>
    <div>
      <AppButton variant="danger" @click="open = true">{{ t("settings.erase.open") }}</AppButton>
    </div>
  </div>

  <AppModal v-model="open" role="alertdialog" size="sm" :title="t('settings.erase.modalTitle')" :dismiss-disabled="loading">
    <form id="erase-account-form" class="modal" @submit.prevent="submit">
      <p class="modal__text">{{ t("settings.erase.modalBody", { days: ERASURE_GRACE_DAYS }) }}</p>
      <ul class="modal__list">
        <li>{{ t("settings.erase.pointGrace", { days: ERASURE_GRACE_DAYS }) }}</li>
        <li>{{ t("settings.erase.pointIrreversible") }}</li>
        <li>{{ t("settings.erase.pointReviews") }}</li>
      </ul>
      <AppInput
        v-if="mode === 'password'"
        v-model="password"
        type="password"
        autocomplete="current-password"
        :label="t('settings.danger.password')"
        :error="error"
        :disabled="loading"
      />
      <AppInput
        v-else
        v-model="username"
        autocomplete="off"
        :label="t('settings.danger.confirmUsername', { username: user.username })"
        :error="error"
        :disabled="loading"
      />
      <button type="button" class="modal__switch" :disabled="loading" @click="switchMode">
        {{ mode === "password" ? t("settings.danger.noPassword") : t("settings.danger.usePassword") }}
      </button>
    </form>
    <template #footer>
      <div class="modal__actions">
        <AppButton variant="outline" :disabled="loading" @click="open = false">{{ t("common.cancel") }}</AppButton>
        <AppButton type="submit" form="erase-account-form" variant="danger" :loading="loading" :disabled="!canSubmit">
          {{ t("settings.erase.confirm") }}
        </AppButton>
      </div>
    </template>
  </AppModal>
</template>

<script setup lang="ts">
import { computed, ref, watch } from "vue"
import { useI18n } from "vue-i18n"
import { useRouter } from "vue-router"
import { userApi } from "@/api/api"
import { useAuthStore } from "@/stores/auth"
import { useSessionReset } from "@/composables/useSessionReset"
import { useToast } from "@/composables/useToast"
import AppButton from "@/components/ui/AppButton.vue"
import AppInput from "@/components/ui/AppInput.vue"
import AppModal from "@/components/ui/AppModal.vue"
import type { UserProfile } from "@/types/user"
import { deleteErrorKey } from "./deleteAccount"
import { ERASURE_GRACE_DAYS } from "./eraseAccount"

defineProps<{ user: UserProfile }>()
const { t, locale } = useI18n()
const router = useRouter()
const auth = useAuthStore()
const toast = useToast()
const resetSession = useSessionReset()

// Same re-auth as the deactivate flow: the client cannot tell whether a password exists, so the
// server's CONFIRM_MISMATCH / WRONG_PASSWORD reply flips the mode.
const mode = ref<"password" | "username">("password")
const open = ref(false)
const loading = ref(false)
const password = ref("")
const username = ref("")
const error = ref("")

const canSubmit = computed(() => !loading.value && (mode.value === "password" ? password.value.length > 0 : username.value.trim().length > 0))

function switchMode() {
  mode.value = mode.value === "password" ? "username" : "password"
  error.value = ""
}

watch(open, (v) => {
  if (v) return
  password.value = ""
  username.value = ""
  error.value = ""
})

async function submit() {
  if (!canSubmit.value) return
  loading.value = true
  error.value = ""
  try {
    const { data } = await userApi.requestErasure(mode.value === "password" ? { password: password.value } : { confirm_username: username.value.trim() })
    const date = new Intl.DateTimeFormat(locale.value, { dateStyle: "long" }).format(new Date(data.erasure_due_at))
    open.value = false
    resetSession()
    auth.endSession()
    toast.success(t("settings.erase.done", { date }))
    await router.push({ name: "login" })
  } catch (e: unknown) {
    const res = (e as { response?: { status?: number; data?: { code?: string } } })?.response
    if (res?.data?.code === "CONFIRM_MISMATCH" && mode.value === "password") mode.value = "username"
    else if (res?.data?.code === "WRONG_PASSWORD" && mode.value === "username") mode.value = "password"
    error.value = t(deleteErrorKey(res?.status, res?.data?.code))
  } finally {
    loading.value = false
  }
}
</script>

<style scoped>
.erase { display: flex; flex-direction: column; gap: 0.75rem; margin-top: 1.25rem; padding-top: 1.25rem; border-top: 1px solid var(--color-border); }
.erase__text { font-size: 0.875rem; color: var(--color-text-secondary); }
.modal { display: flex; flex-direction: column; gap: 0.875rem; }
.modal__text { font-size: 0.875rem; color: var(--color-text-secondary); }
.modal__list { display: flex; flex-direction: column; gap: 0.375rem; padding-left: 1.25rem; list-style: disc; font-size: 0.8125rem; color: var(--color-text-secondary); }
.modal__switch { align-self: flex-start; font-size: 0.8125rem; color: var(--color-text-secondary); text-decoration: underline; cursor: pointer; }
.modal__switch:hover { color: var(--color-text-primary); }
.modal__actions { display: flex; justify-content: flex-end; gap: 0.5rem; }
</style>
