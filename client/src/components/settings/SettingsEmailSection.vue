<template>
  <SettingsCard id="email" :title="t('settings.email.title')">
    <div class="row">
      <div class="row__input">
      <AppInput
        v-model="emailInput"
        type="email"
        :label="t('settings.email.label')"
        :disabled="step !== 'input'"
        :error="step === 'input' ? error : ''"
        :hint="step === 'input' ? t('settings.email.inputHint', { current: user.email }) : ''"
        :placeholder="t('settings.email.placeholder')"
      />
      </div>
      <div class="row__actions">
        <AppButton v-if="!isVerified && step === 'idle'" variant="outline" size="sm" @click="resendVerification">
          <Mail :size="12" aria-hidden="true" />{{ t("settings.email.verify") }}
        </AppButton>
        <AppButton v-else-if="step === 'idle'" variant="outline" size="sm" @click="startChange">
          <Pencil :size="12" aria-hidden="true" />{{ t("settings.email.change") }}
        </AppButton>
        <template v-else-if="step === 'input'">
          <AppButton variant="primary" size="sm" :loading="sending" :disabled="!canSend" @click="requestOtp">
            <Mail :size="12" aria-hidden="true" />{{ t("settings.email.sendCode") }}
          </AppButton>
          <AppButton variant="ghost" size="sm" @click="cancel">{{ t("common.cancel") }}</AppButton>
        </template>
        <AppButton v-else-if="step === 'otp'" variant="ghost" size="sm" @click="cancel">{{ t("common.cancel") }}</AppButton>
        <span v-else-if="step === 'done'" class="done"><CheckCircle :size="14" aria-hidden="true" />{{ t("settings.email.done") }}</span>
      </div>
    </div>

    <div v-if="step === 'otp'" class="otp">
      <AppInput v-model="otp" inputmode="numeric" :maxlength="EMAIL_OTP_LENGTH" :placeholder="OTP_PLACEHOLDER" :label="t('settings.email.currentCodeLabel')" :hint="t('settings.email.currentCodeHint', { email: user.email })" @input="otp = otp.replace(/\D/g, '')" />
      <AppInput v-model="newOtp" inputmode="numeric" :maxlength="EMAIL_OTP_LENGTH" :placeholder="OTP_PLACEHOLDER" :label="t('settings.email.newCodeLabel')" :hint="t('settings.email.newCodeHint', { email: emailInput })" @input="newOtp = newOtp.replace(/\D/g, '')" />
      <p v-if="otpError" class="otp__error" role="alert">{{ otpError }}</p>
      <AppButton variant="primary" :loading="confirming" :disabled="otp.length !== EMAIL_OTP_LENGTH || newOtp.length !== EMAIL_OTP_LENGTH" @click="verifyOtp">
        {{ t("settings.email.confirm") }}
      </AppButton>
    </div>

    <ConfirmModal
      v-model="showConfirm"
      list-type="email_change"
      :item-name="user.email"
      @confirm="confirmRequestOtp"
      @cancel="showConfirm = false"
    />
  </SettingsCard>
</template>

<script setup lang="ts">
import { computed, ref, watch } from "vue"
import { useI18n } from "vue-i18n"
import { CheckCircle, Mail, Pencil } from "lucide-vue-next"
import { authApi, userApi } from "@/api/api"
import { useAuthStore } from "@/stores/auth"
import { useUIStore } from "@/stores/ui"
import AppButton from "@/components/ui/AppButton.vue"
import AppInput from "@/components/ui/AppInput.vue"
import ConfirmModal from "@/components/profile/components/ConfirmModal.vue"
import type { UserProfile } from "@/types/user"
import SettingsCard from "./SettingsCard.vue"
import { EMAIL_CHANGED_NOTICE_MS, EMAIL_FORMAT_PATTERN, EMAIL_OTP_LENGTH, OTP_PLACEHOLDER, emailErrorKey } from "./emailChange"

const props = defineProps<{ user: UserProfile }>()
const { t } = useI18n()
const authStore = useAuthStore()
const ui = useUIStore()

// idle  -> showing the current email
// input -> typing the new email
// otp   -> codes sent to the CURRENT and the NEW email, waiting for both
// done  -> codes verified, email swapped on the server and still verified
type EmailStep = "idle" | "input" | "otp" | "done"
const step = ref<EmailStep>("idle")

const emailInput = ref(props.user.email ?? "")
const error = ref("")
const sending = ref(false)
const showConfirm = ref(false)
const otp = ref("")
const newOtp = ref("")
const otpError = ref("")
const confirming = ref(false)

const isVerified = computed(() => !!props.user.verified_email_at)
const isFormatValid = computed(() => EMAIL_FORMAT_PATTERN.test(emailInput.value.trim()))
const isSameAsCurrent = computed(() => emailInput.value.trim().toLowerCase() === props.user.email?.toLowerCase())
const canSend = computed(() => !!emailInput.value && !isSameAsCurrent.value && !sending.value && isFormatValid.value)

watch(emailInput, () => (error.value = ""))
watch(() => props.user.email, email => step.value === "idle" && (emailInput.value = email ?? ""))

function reset() {
  step.value = "idle"
  otp.value = ""
  newOtp.value = ""
  otpError.value = ""
  error.value = ""
  emailInput.value = props.user.email ?? ""
}
const cancel = reset

// Step 1: unlock the field for a new address
function startChange() {
  step.value = "input"
  otpError.value = ""
  emailInput.value = ""
}

// Step 2: validate, then confirm before mailing
function requestOtp() {
  if (!emailInput.value) return
  error.value = ""
  if (!isFormatValid.value) {
    error.value = t("settings.email.errors.invalid")
    return
  }
  if (isSameAsCurrent.value) {
    error.value = t("settings.email.errors.same")
    return
  }
  showConfirm.value = true
}

// Step 3: POST /users/:id/email { new_email }
// The server stores the pending new_email + codes and mails one code to each address.
async function confirmRequestOtp() {
  showConfirm.value = false
  if (!emailInput.value) return
  sending.value = true
  error.value = ""
  try {
    await userApi.requestEmailChange(props.user.id, { new_email: emailInput.value })
    step.value = "otp"
    otp.value = ""
    newOtp.value = ""
  } catch (err: unknown) {
    error.value = t(emailErrorKey(err))
  } finally {
    sending.value = false
  }
}

// Step 4: PUT /users/:id/email { otp, new_email_otp }
// The server verifies both, then swaps the email and keeps it verified in one transaction.
async function verifyOtp() {
  if (otp.value.length !== EMAIL_OTP_LENGTH || newOtp.value.length !== EMAIL_OTP_LENGTH) return
  confirming.value = true
  otpError.value = ""
  try {
    const { data } = await userApi.verifyEmailChange(props.user.id, { otp: otp.value, new_email_otp: newOtp.value })
    emailInput.value = data.user.email ?? ""
    await authStore.fetchMe()
    step.value = "done"
    setTimeout(() => (step.value = "idle"), EMAIL_CHANGED_NOTICE_MS)
  } catch {
    otpError.value = t("settings.email.errors.invalidCode")
  } finally {
    confirming.value = false
  }
}

async function resendVerification() {
  if (!props.user.email) return
  try {
    await authApi.resendVerification(props.user.email)
    ui.toastSuccess(t("settings.email.verificationSent", { email: props.user.email }))
  } catch {
    ui.toastError(t("settings.email.verificationFailed"))
  }
}
</script>

<style scoped>
.row { display: flex; flex-wrap: wrap; align-items: flex-end; gap: 0.75rem; }
.row__input { flex: 1; min-width: 14rem; }
.row__actions { display: flex; gap: 0.5rem; padding-bottom: 0.25rem; }
.done { display: inline-flex; align-items: center; gap: 0.375rem; font-size: 0.8125rem; color: var(--color-success); }
.otp { display: flex; flex-direction: column; gap: 0.75rem; max-width: 20rem; }
.otp__error { font-size: 0.8125rem; color: var(--color-danger); }
</style>
