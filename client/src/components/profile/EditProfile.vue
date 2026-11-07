<template>
  <div class="edit-root">
    <header class="edit-header">
      <div class="edit-title-group">
        <span class="edit-eyebrow">Profile</span>
        <h2 class="edit-title">Edit Profile</h2>
      </div>
      <button class="close-btn" aria-label="Close" @click="handleClose">
        <X :size="14" />
      </button>
    </header>

    <form class="edit-form" @submit.prevent="handleSave">
      <!-- Avatar Upload -->
      <div class="avatar-upload">
        <div class="avatar-preview">
          <img
            v-if="form.avatar_url"
            :src="form.avatar_url"
            alt="Avatar"
            class="preview-img"
          />
          <div v-else class="preview-fallback"><UserIcon :size="24" /></div>
        </div>
        <div class="avatar-info">
          <span class="field-label">Profile Photo</span>
          <label class="upload-btn">
            <Upload :size="12" />
            <span>Upload new</span>
            <input
              type="file"
              accept="image/*"
              class="sr-only"
              @change="handleFileChange"
            />
          </label>
        </div>
      </div>

      <div class="field-divider" />

      <!-- Display Name -->
      <div class="field-group">
        <label class="field-label" for="ep-display-name">Display Name</label>
        <input
          id="ep-display-name"
          v-model="form.display_name"
          type="text"
          class="field-input"
          placeholder="Your display name"
          maxlength="60"
        />
      </div>

      <!-- Username -->
      <div class="field-group">
        <label class="field-label" for="ep-username">Username</label>
        <div class="input-prefix-wrap">
          <span class="input-prefix">@</span>
          <input
            id="ep-username"
            v-model="form.username"
            type="text"
            class="field-input field-input--prefixed"
            placeholder="username"
            maxlength="30"
          />
        </div>
      </div>

      <!-- ─── Email ─────────────────────────────────────────── -->
      <div class="field-group">
        <label class="field-label" for="ep-email">Email</label>
        <div class="email-row">
          <input
            id="ep-email"
            v-model="form.email"
            type="email"
            class="field-input email-input"
            :class="{ 'field-input--unlocked': emailStep === 'input' }"
            :disabled="emailStep !== 'input'"
            placeholder="your@email.com"
          />
          <button
            v-if="!isEmailVerified && emailStep === 'idle'"
            type="button"
            class="email-action-btn email-action-btn--verify"
            @click="handleResendVerification"
          >
            <Mail :size="12" /> Verify Email
          </button>
          <button
            v-else-if="emailStep === 'idle'"
            type="button"
            class="email-action-btn"
            @click="handleStartChangeEmail"
          >
            <Pencil :size="12" /> Change Email
          </button>
          <template v-else-if="emailStep === 'input'">
            <button
              type="button"
              class="email-action-btn email-action-btn--confirm"
              :disabled="
                !form.email ||
                form.email === props.user?.email ||
                emailSaving ||
                !isEmailFormatValid
              "
              @click="handleRequestOTP"
            >
              <AppSpinner v-if="emailSaving" />
              <template v-else><Mail :size="12" /> Send OTP</template>
            </button>
            <button
              type="button"
              class="email-action-btn email-action-btn--cancel"
              @click="cancelEmailChange"
            >
              <X :size="12" /> Cancel
            </button>
          </template>
          <button
            v-else-if="emailStep === 'otp'"
            type="button"
            class="email-action-btn email-action-btn--cancel"
            @click="cancelEmailChange"
          >
            <X :size="12" /> Cancel
          </button>
          <button
            v-else-if="emailStep === 'waiting'"
            type="button"
            class="email-action-btn email-action-btn--waiting"
            disabled
          >
            <AppSpinner /> Waiting…
          </button>
          <button
            v-else-if="emailStep === 'done'"
            type="button"
            class="email-action-btn email-action-btn--done"
            disabled
          >
            <CheckCircle :size="13" /> Verified!
          </button>
        </div>
        <transition name="inline-err">
          <p v-if="emailStep === 'input' && emailError" class="inline-error">
            <AlertCircle :size="11" /> {{ emailError }}
          </p>
        </transition>

        <!-- OTP sent to CURRENT email -->
        <transition name="otp-slide">
          <div v-if="emailStep === 'otp'" class="otp-box">
            <p class="otp-hint">
              {{ t("profile.editProfile.emailChange.currentCodeHint", { email: props.user?.email }) }}
            </p>
            <input
              v-model="otpValue"
              type="text"
              inputmode="numeric"
              maxlength="6"
              class="field-input otp-input"
              placeholder="000000"
              :aria-label="t('profile.editProfile.emailChange.currentCodeLabel')"
              @input="otpValue = otpValue.replace(/\D/g, '')"
            />
            <p class="otp-hint">
              {{ t("profile.editProfile.emailChange.newCodeHint", { email: form.email }) }}
            </p>
            <div class="otp-row">
              <input
                v-model="newOtpValue"
                type="text"
                inputmode="numeric"
                maxlength="6"
                class="field-input otp-input"
                placeholder="000000"
                :aria-label="t('profile.editProfile.emailChange.newCodeLabel')"
                @input="newOtpValue = newOtpValue.replace(/\D/g, '')"
              />
              <button
                type="button"
                class="email-action-btn email-action-btn--confirm"
                :disabled="otpValue.length !== 6 || newOtpValue.length !== 6 || otpConfirming"
                @click="handleVerifyOTP"
              >
                <AppSpinner v-if="otpConfirming" /><span v-else
                  >Confirm</span
                >
              </button>
            </div>
            <p v-if="otpError" class="otp-error">{{ otpError }}</p>
          </div>
        </transition>

        <!-- waiting for click on link sent to NEW email -->
        <transition name="otp-slide">
          <div v-if="emailStep === 'waiting'" class="otp-box otp-box--waiting">
            <div class="waiting-row">
              <AppSpinner />
              <p class="otp-hint" style="margin: 0">
                Verification link sent to <strong>{{ form.email }}</strong
                >.<br />Click the link in your email to confirm.
              </p>
            </div>
            <button
              type="button"
              class="resend-btn"
              @click="handleResendEmailVerification"
            >
              Resend link
            </button>
          </div>
        </transition>

        <p v-if="emailStep === 'input'" class="email-hint">
          <Info :size="11" />
          {{ t("profile.editProfile.emailChange.inputHint", { current: props.user?.email }) }}
        </p>
      </div>
      <!-- ─────────────────────────────────────────────────────── -->

      <!-- Bio -->
      <div class="field-group">
        <label class="field-label" for="ep-bio">Bio</label>
        <textarea
          id="ep-bio"
          v-model="form.bio"
          class="field-textarea"
          placeholder="Say something about yourself…"
          rows="3"
          maxlength="200"
        />
        <span class="char-count">{{ form.bio?.length ?? 0 }} / 200</span>
      </div>

      <!-- Date of Birth -->
      <div class="field-group">
        <label class="field-label" for="ep-dob">Date of Birth</label>
        <input
          id="ep-dob"
          v-model="form.date_of_birth"
          type="date"
          class="field-input field-input--date"
        />
      </div>

      <!-- Gender -->
      <div class="field-group">
        <label class="field-label">Gender</label>
        <div class="radio-group">
          <label
            v-for="opt in genderOptions"
            :key="opt.value"
            class="radio-option"
            :class="{ 'radio-option--active': form.gender === opt.value }"
          >
            <input
              type="radio"
              :value="opt.value"
              v-model="form.gender"
              class="sr-only"
            />
            {{ opt.label }}
          </label>
        </div>
      </div>

      <!-- Private Account -->
      <div class="field-group">
        <div class="toggle-row">
          <div class="toggle-info">
            <span class="toggle-label">Private Account</span>
            <span class="toggle-desc"
              >Only followers can see your activity</span
            >
          </div>
          <button
            type="button"
            class="toggle-btn"
            :class="{ 'toggle-btn--on': form.is_private }"
            @click="form.is_private = !form.is_private"
            :aria-pressed="form.is_private"
          >
            <span class="toggle-knob" />
          </button>
        </div>
      </div>

      <!-- ─── Change Password ──────────────────────────────────── -->
      <div class="field-divider" />
      <div class="section-header" @click="pwOpen = !pwOpen">
        <div class="section-header__left">
          <KeyRound :size="14" class="section-header__icon" />
          <span class="section-header__title">Change Password</span>
        </div>
        <ChevronDown
          :size="14"
          class="section-header__chevron"
          :class="{ 'section-header__chevron--open': pwOpen }"
        />
      </div>

      <transition name="pw-slide">
        <div v-if="pwOpen" class="pw-section">
          <div class="field-group">
            <label class="field-label" for="ep-old-pw">Current Password</label>
            <div class="pw-input-wrap">
              <input
                id="ep-old-pw"
                v-model="pwForm.old_password"
                :type="showOld ? 'text' : 'password'"
                class="field-input"
                autocomplete="current-password"
                placeholder="Your current password"
                :disabled="pwLoading"
              />
              <button
                type="button"
                class="pw-toggle"
                @click="showOld = !showOld"
                tabindex="-1"
              >
                <EyeOff v-if="showOld" :size="14" /><Eye v-else :size="14" />
              </button>
            </div>
          </div>

          <div class="field-group">
            <label class="field-label" for="ep-new-pw">New Password</label>
            <div class="pw-input-wrap">
              <input
                id="ep-new-pw"
                v-model="pwForm.new_password"
                :type="showNew ? 'text' : 'password'"
                class="field-input"
                autocomplete="new-password"
                placeholder="At least 8 characters"
                :disabled="pwLoading"
              />
              <button
                type="button"
                class="pw-toggle"
                @click="showNew = !showNew"
                tabindex="-1"
              >
                <EyeOff v-if="showNew" :size="14" /><Eye v-else :size="14" />
              </button>
            </div>
            <!-- strength bar -->
            <div v-if="pwForm.new_password" class="strength-bar">
              <div
                class="strength-bar__fill"
                :class="`strength-bar__fill--${pwStrength.level}`"
                :style="{ width: pwStrength.width }"
              />
            </div>
            <span
              v-if="pwForm.new_password"
              class="strength-label"
              :class="`strength-label--${pwStrength.level}`"
              >{{ pwStrength.label }}</span
            >
          </div>

          <div class="field-group">
            <label class="field-label" for="ep-confirm-pw"
              >Confirm New Password</label
            >
            <div class="pw-input-wrap">
              <input
                id="ep-confirm-pw"
                v-model="pwForm.confirm_password"
                :type="showConfirm ? 'text' : 'password'"
                class="field-input"
                autocomplete="new-password"
                placeholder="Repeat your new password"
                :disabled="pwLoading"
              />
              <button
                type="button"
                class="pw-toggle"
                @click="showConfirm = !showConfirm"
                tabindex="-1"
              >
                <EyeOff v-if="showConfirm" :size="14" /><Eye
                  v-else
                  :size="14"
                />
              </button>
            </div>
            <transition name="inline-err">
              <span
                v-if="
                  pwForm.confirm_password &&
                  pwForm.new_password !== pwForm.confirm_password
                "
                class="inline-error"
              >
                <AlertCircle :size="11" /> Passwords do not match
              </span>
            </transition>
          </div>

          <!-- feedback -->
          <transition name="otp-slide">
            <div v-if="pwError" class="pw-alert pw-alert--error">
              <AlertCircle :size="13" /> {{ pwError }}
            </div>
          </transition>
          <transition name="otp-slide">
            <div v-if="pwSuccess" class="pw-alert pw-alert--success">
              <CheckCircle :size="13" /> Password changed successfully!
            </div>
          </transition>

          <button
            type="button"
            class="pw-submit-btn"
            :disabled="pwLoading || !pwCanSubmit"
            @click="handleChangePassword"
          >
            <AppSpinner v-if="pwLoading" />
            <span v-else>Update Password</span>
          </button>
        </div>
      </transition>
      <!-- ─────────────────────────────────────────────────────── -->

      <!-- ─── Connected Accounts (Google) ──────────────────────── -->
      <div class="field-divider" />
      <div class="section-header" @click="googleOpen = !googleOpen">
        <div class="section-header__left">
          <svg class="google-logo" viewBox="0 0 24 24" width="16" height="16">
            <path fill="#4285F4" d="M22.56 12.25c0-.78-.07-1.53-.2-2.25H12v4.26h5.92c-.26 1.37-1.04 2.53-2.21 3.31v2.77h3.57c2.08-1.92 3.28-4.74 3.28-8.09z"/>
            <path fill="#34A853" d="M12 23c2.97 0 5.46-.98 7.28-2.66l-3.57-2.77c-.98.66-2.23 1.06-3.71 1.06-2.86 0-5.29-1.93-6.16-4.53H2.18v2.84C3.99 20.53 7.7 23 12 23z"/>
            <path fill="#FBBC05" d="M5.84 14.1c-.22-.66-.35-1.36-.35-2.1s.13-1.44.35-2.1V7.06H2.18C1.43 8.55 1 10.22 1 12s.43 3.45 1.18 4.94l2.85-2.22.81-.62z"/>
            <path fill="#EA4335" d="M12 5.38c1.62 0 3.06.56 4.21 1.64l3.15-3.15C17.45 2.09 14.97 1 12 1 7.7 1 3.99 3.47 2.18 7.06l3.66 2.84c.87-2.6 3.3-4.52 6.16-4.52z"/>
          </svg>
          <span class="section-header__title">Connected Accounts</span>
        </div>
        <ChevronDown
          :size="14"
          class="section-header__chevron"
          :class="{ 'section-header__chevron--open': googleOpen }"
        />
      </div>

      <transition name="pw-slide">
        <div v-if="googleOpen" class="pw-section">
          <div class="google-status-row">
            <div class="google-status-info">
              <span class="google-title">Google Account</span>
              <span v-if="googleStatus.google_connected" class="google-desc text-green">
                Connected: {{ googleStatus.google_email }}
              </span>
              <span v-else class="google-desc text-muted">Not connected</span>
            </div>
            <button
              v-if="googleStatus.google_connected"
              type="button"
              class="email-action-btn email-action-btn--cancel"
              :disabled="!googleStatus.can_disconnect || googleLoading"
              @click="handleDisconnectGoogle"
              :title="!googleStatus.can_disconnect ? 'Cannot disconnect your only authentication method' : ''"
            >
              <AppSpinner v-if="googleLoading" />
              <span v-else>Disconnect</span>
            </button>
            <button
              v-else
              type="button"
              class="email-action-btn email-action-btn--confirm"
              :disabled="googleLoading"
              @click="handleConnectGoogle"
            >
              <AppSpinner v-if="googleLoading" />
              <span v-else>Connect Google</span>
            </button>
          </div>
          <p v-if="googleError" class="inline-error mt-2">
            <AlertCircle :size="11" /> {{ googleError }}
          </p>
        </div>
      </transition>
      <!-- ─────────────────────────────────────────────────────── -->
      <LinkedAccounts />
    </form>

    <div class="edit-footer">
      <button
        type="button"
        class="footer-btn footer-btn--cancel"
        @click="handleClose"
      >
        Cancel
      </button>
      <button
        type="submit"
        class="footer-btn footer-btn--save"
        :disabled="saving || !isDirty"
        @click="handleSave"
      >
        <AppSpinner v-if="saving" /><span v-else
          >Save Changes</span
        >
      </button>
    </div>

    <ConfirmModal
      v-model="showEmailConfirmModal"
      list-type="email_change"
      :item-name="props.user?.email"
      @confirm="confirmRequestOTP"
      @cancel="showEmailConfirmModal = false"
    />
  </div>
</template>

<script setup lang="ts">
import AppSpinner from "@/components/ui/AppSpinner.vue"
import { reactive, ref, computed, onMounted, onUnmounted, watch } from "vue";
import { useI18n } from "vue-i18n";
import {
  X,
  Upload,
  User as UserIcon,
  Pencil,
  Mail,
  Info,
  CheckCircle,
  KeyRound,
  ChevronDown,
  Eye,
  EyeOff,
  AlertCircle,
} from "lucide-vue-next";
import type { UserProfile } from "@/types/user";
import { authApi, userApi } from "@/api/api";
import { useAuthStore } from "@/stores/auth";
import LinkedAccounts from "@/components/profile/LinkedAccounts.vue";
import ConfirmModal from "@/components/profile/components/ConfirmModal.vue";

const { t } = useI18n();
const props = defineProps<{ user: UserProfile | null }>();
const emit = defineEmits<{ close: [] }>();

const authStore = useAuthStore();

// ── email step machine ────────────────────────────────────────────────
// idle     -> nothing happening, showing current email
// input    -> user is typing the new email
// otp      -> codes sent to the CURRENT and the NEW email, waiting for both
// waiting  -> legacy step (email change used to need a verification link); unused
// done     -> codes verified, email swapped on BE and still verified
type EmailStep = "idle" | "input" | "otp" | "waiting" | "done";
const emailStep = ref<EmailStep>("idle");
const emailError = ref("");
const EMAIL_REGEX = /^[^\s@]+@[^\s@]+\.[^\s@]+$/;

const saving = ref(false);
const avatarFile = ref<File | null>(null);
const emailSaving = ref(false);
const showEmailConfirmModal = ref(false);
const otpValue = ref("");
const newOtpValue = ref("");
const otpError = ref("");
const otpConfirming = ref(false);

let pollTimer: ReturnType<typeof setInterval> | null = null;
const POLL_INTERVAL = 3000;

const genderOptions = [
  { value: "male", label: "Male" },
  { value: "female", label: "Female" },
  { value: "other", label: "Other" },
];

const initialForm = {
  display_name: props.user?.display_name ?? "",
  username: props.user?.username ?? "",
  bio: props.user?.bio ?? "",
  avatar_url: props.user?.avatar_url ?? "",
  date_of_birth: props.user?.date_of_birth
    ? props.user.date_of_birth.slice(0, 10)
    : "",
  gender: props.user?.gender ?? "",
  is_private: props.user?.is_private ?? false,
  email: props.user?.email ?? "",
};
const form = reactive({ ...initialForm });

const isEmailVerified = computed(() => !!props.user?.verified_email_at);

const isDirty = computed(
  () =>
    avatarFile.value !== null ||
    form.display_name !== initialForm.display_name ||
    form.username !== initialForm.username ||
    form.bio !== initialForm.bio ||
    form.date_of_birth !== initialForm.date_of_birth ||
    form.gender !== initialForm.gender ||
    form.is_private !== initialForm.is_private,
);

const isEmailFormatValid = computed(() => EMAIL_REGEX.test(form.email.trim()));

const isEmailSameAsCurrent = computed(
  () => form.email.trim().toLowerCase() === props.user?.email?.toLowerCase(),
);

// ── change password state ─────────────────────────────────────────────
const pwOpen = ref(false);
const pwLoading = ref(false);
const pwError = ref("");
const pwSuccess = ref(false);
const showOld = ref(false);
const showNew = ref(false);
const showConfirm = ref(false);

// ── Google OAuth state ────────────────────────────────────────────────
const googleOpen = ref(false);
const googleLoading = ref(false);
const googleError = ref("");
const googleStatus = reactive({
  enabled: false,
  google_connected: false,
  can_disconnect: false,
  google_email: "",
});

async function fetchGoogleStatus() {
  try {
    const { data } = await authApi.getGoogleStatus();
    googleStatus.enabled = data.enabled;
    googleStatus.google_connected = data.google_connected;
    googleStatus.can_disconnect = data.can_disconnect;
    googleStatus.google_email = data.google_email ?? "";
  } catch {
    /* ignore */
  }
}

async function handleConnectGoogle() {
  googleLoading.value = true;
  googleError.value = "";
  try {
    const { data } = await authApi.startGoogleLink("/profile");
    if (data.authorization_url) {
      window.location.assign(data.authorization_url);
    }
  } catch (e: unknown) {
    const err = e as { response?: { data?: { error?: string } } };
    googleError.value = err?.response?.data?.error ?? "Failed to initiate Google link.";
  } finally {
    googleLoading.value = false;
  }
}

async function handleDisconnectGoogle() {
  googleLoading.value = true;
  googleError.value = "";
  try {
    await authApi.disconnectGoogle();
    await fetchGoogleStatus();
  } catch (e: unknown) {
    const err = e as { response?: { data?: { error?: string } } };
    googleError.value = err?.response?.data?.error ?? "Failed to disconnect Google account.";
  } finally {
    googleLoading.value = false;
  }
}

const pwForm = reactive({
  old_password: "",
  new_password: "",
  confirm_password: "",
});

const pwStrength = computed(() => {
  const p = pwForm.new_password;
  if (!p) return { level: "none", label: "", width: "0%" };
  let score = 0;
  if (p.length >= 8) score++;
  if (p.length >= 12) score++;
  if (/[A-Z]/.test(p)) score++;
  if (/[0-9]/.test(p)) score++;
  if (/[^A-Za-z0-9]/.test(p)) score++;
  if (score <= 1) return { level: "weak", label: "Weak", width: "25%" };
  if (score <= 2) return { level: "fair", label: "Fair", width: "50%" };
  if (score <= 3) return { level: "good", label: "Good", width: "75%" };
  return { level: "strong", label: "Strong", width: "100%" };
});

const pwCanSubmit = computed(
  () =>
    pwForm.old_password.length > 0 &&
    pwForm.new_password.length >= 8 &&
    pwForm.new_password === pwForm.confirm_password,
);

async function handleChangePassword() {
  if (!props.user?.id || !pwCanSubmit.value) return;
  pwError.value = "";
  pwSuccess.value = false;
  pwLoading.value = true;
  try {
    const { data } = await userApi.changePassword(props.user.id, {
      old_password: pwForm.old_password,
      new_password: pwForm.new_password,
      confirm_password: pwForm.confirm_password,
    });
    // The change ended every session, this one included; the server sends a fresh one.
    // Keep it, or the next request fails to refresh and signs the user out.
    if (data.access_token) localStorage.setItem("access_token", data.access_token);
    if (data.refresh_token) localStorage.setItem("refresh_token", data.refresh_token);
    pwSuccess.value = true;
    pwForm.old_password = "";
    pwForm.new_password = "";
    pwForm.confirm_password = "";
    setTimeout(() => {
      pwSuccess.value = false;
    }, 4000);
  } catch (e: any) {
    const status = e?.response?.status;
    if (status === 401) pwError.value = "Current password is incorrect.";
    else if (status === 400)
      pwError.value = e?.response?.data?.error ?? "Invalid request.";
    else pwError.value = "Something went wrong. Please try again.";
  } finally {
    pwLoading.value = false;
  }
}

// ── email helpers ─────────────────────────────────────────────────────
function stopPolling() {
  if (pollTimer) {
    clearInterval(pollTimer);
    pollTimer = null;
  }
}

function resetEmailFlow() {
  stopPolling();
  emailStep.value = "idle";
  otpValue.value = "";
  newOtpValue.value = "";
  otpError.value = "";
  emailError.value = "";
  form.email = initialForm.email;
}

function startPolling(newEmail: string) {
  stopPolling();
  pollTimer = setInterval(async () => {
    try {
      if (!props.user?.id) return;
      const { data } = await userApi.getProfile(props.user.id);
      const fresh = data.user;
      if (fresh.email === newEmail && fresh.verified_email_at) {
        stopPolling();
        await authStore.fetchMe();
        emailStep.value = "done";
        setTimeout(() => {
          emailStep.value = "idle";
        }, 2500);
      }
    } catch {
      /* ignore */
    }
  }, POLL_INTERVAL);
}

function handleFileChange(e: Event) {
  const file = (e.target as HTMLInputElement).files?.[0];
  if (!file) return;
  avatarFile.value = file;
  const reader = new FileReader();
  reader.onload = (ev) => {
    form.avatar_url = ev.target?.result as string;
  };
  reader.readAsDataURL(file);
}

// Step 1: user clicks "Change Email" -> unlock the field for a new address
function handleStartChangeEmail() {
  emailStep.value = "input";
  otpError.value = "";
  form.email = "";
}

// Step 2: user typed the new email and clicks "Send OTP" -> confirm first
function handleRequestOTP() {
  if (!props.user?.id || !form.email) return;
  emailError.value = "";

  if (!isEmailFormatValid.value) {
    emailError.value = "รูปแบบอีเมลไม่ถูกต้อง";
    return;
  }
  if (isEmailSameAsCurrent.value) {
    emailError.value = "อีเมลนี้เป็นอีเมลปัจจุบันของคุณอยู่แล้ว";
    return;
  }

  showEmailConfirmModal.value = true;
}

// Step 3: confirmed -> POST /users/:id/email { new_email }
// BE stores pending new_email + OTP, and sends the OTP to the CURRENT email.
async function confirmRequestOTP() {
  showEmailConfirmModal.value = false;
  if (!props.user?.id || !form.email) return;
  emailSaving.value = true;
  emailError.value = "";
  try {
    await userApi.requestEmailChange(props.user.id, {
      new_email: form.email,
    });
    emailStep.value = "otp";
    otpValue.value = "";
    newOtpValue.value = "";
  } catch (err: any) {
    const msg = err?.response?.data?.error;
    emailError.value =
      msg === "email already in use"
        ? "อีเมลนี้ถูกใช้งานแล้วในระบบ กรุณาใช้อีเมลอื่น"
        : (msg ?? "ไม่สามารถส่ง OTP ได้ กรุณาลองใหม่อีกครั้ง");
  } finally {
    emailSaving.value = false;
  }
}

// Step 4: user enters both codes -> PUT /users/:id/email { otp, new_email_otp }
// BE verifies both, then swaps the email and keeps it verified in one transaction.
async function handleVerifyOTP() {
  if (!props.user?.id || otpValue.value.length !== 6 || newOtpValue.value.length !== 6) return;
  otpConfirming.value = true;
  otpError.value = "";
  try {
    const { data } = await userApi.verifyEmailChange(props.user.id, {
      otp: otpValue.value,
      new_email_otp: newOtpValue.value,
    });
    form.email = data.user.email;
    await authStore.fetchMe();
    emailStep.value = "done";
    setTimeout(() => {
      emailStep.value = "idle";
    }, 2500);
  } catch (err: any) {
    otpError.value = err?.response?.data?.error ?? "Invalid OTP.";
  } finally {
    otpConfirming.value = false;
  }
}

async function handleResendEmailVerification() {
  if (!form.email) return;
  try {
    await authApi.resendVerification(form.email);
  } catch {
    /* ignore */
  }
}

function cancelEmailChange() {
  resetEmailFlow();
}

function handleClose() {
  if (emailStep.value !== "idle" && emailStep.value !== "done")
    resetEmailFlow();
  emit("close");
}

async function handleResendVerification() {
  if (!props.user?.email) return;
  try {
    await authApi.resendVerification(props.user.email);
    alert(`Verification email sent to ${props.user.email}.`);
  } catch {
    alert("Failed to resend verification email.");
  }
}

async function handleSave() {
  saving.value = true;
  try {
    const payload = new FormData();
    if (form.display_name !== initialForm.display_name)
      payload.append("display_name", form.display_name);
    if (form.bio !== initialForm.bio) payload.append("bio", form.bio);
    if (form.date_of_birth !== initialForm.date_of_birth)
      payload.append("date_of_birth", form.date_of_birth);
    if (form.gender !== initialForm.gender)
      payload.append("gender", form.gender);
    if (form.is_private !== initialForm.is_private)
      payload.append("is_private", String(form.is_private));
    if (avatarFile.value) payload.append("avatar", avatarFile.value);
    await userApi.updateProfile(props.user!.id, payload);
    await authStore.fetchMe();
    emit("close");
  } catch (err) {
    console.error("Save profile failed:", err);
  } finally {
    saving.value = false;
  }
}

onMounted(() => {
  fetchGoogleStatus();
});

onUnmounted(() => stopPolling());

watch(
  () => form.email,
  () => {
    emailError.value = "";
  },
);
</script>

<style scoped>
.edit-root {
  --c-surface: #111111;
  --c-card: #161616;
  --c-border: rgba(255, 255, 255, 0.07);
  --c-border-h: rgba(255, 255, 255, 0.14);
  --c-red: #e1251b;
  --c-green: #30d158;
  --c-yellow: #ffd60a;
  --c-text: #f0f0f0;
  --c-sub: #8a8a8e;
  --c-muted: #3a3a3c;
  --font:
    "Inter", -apple-system, BlinkMacSystemFont, "SF Pro Text", "Helvetica Neue",
    system-ui, sans-serif;
  --ease: cubic-bezier(0.16, 1, 0.3, 1);
  padding: 24px;
  display: flex;
  flex-direction: column;
  max-height: 90vh;
  overflow: hidden;
  font-family: var(--font);
  color: var(--c-text);
  background: var(--c-surface);
  border-radius: 14px;
}

/* Header */
.edit-header {
  display: flex;
  align-items: flex-start;
  justify-content: space-between;
  margin-bottom: 24px;
}
.edit-title-group {
  display: flex;
  flex-direction: column;
  gap: 2px;
}
.edit-eyebrow {
  font-size: 0.58rem;
  font-weight: 700;
  letter-spacing: 0.1em;
  text-transform: uppercase;
  color: var(--c-muted);
}
.edit-title {
  font-size: 1.05rem;
  font-weight: 600;
  margin: 0;
  letter-spacing: -0.01em;
  color: #fff;
}
.close-btn {
  width: 28px;
  height: 28px;
  border-radius: 7px;
  border: 1px solid var(--c-border);
  background: var(--c-card);
  color: var(--c-sub);
  display: flex;
  align-items: center;
  justify-content: center;
  cursor: pointer;
  transition: all 0.2s;
  flex-shrink: 0;
}
.close-btn:hover {
  color: var(--c-text);
  border-color: var(--c-border-h);
}

/* Avatar */
.avatar-upload {
  display: flex;
  align-items: center;
  gap: 16px;
  margin-bottom: 20px;
}
.avatar-preview {
  width: 64px;
  height: 64px;
  border-radius: 50%;
  background: var(--c-card);
  border: 1px solid var(--c-border);
  overflow: hidden;
  flex-shrink: 0;
  display: flex;
  align-items: center;
  justify-content: center;
  color: var(--c-muted);
}
.preview-img {
  width: 100%;
  height: 100%;
  object-fit: cover;
}
.avatar-info {
  display: flex;
  flex-direction: column;
  gap: 8px;
}
.upload-btn {
  display: inline-flex;
  align-items: center;
  gap: 6px;
  font-size: 0.75rem;
  font-weight: 500;
  color: var(--c-text);
  background: var(--c-card);
  border: 1px solid var(--c-border);
  padding: 6px 12px;
  border-radius: 7px;
  cursor: pointer;
  transition: all 0.2s;
}
.upload-btn:hover {
  border-color: var(--c-border-h);
}
.sr-only {
  position: absolute;
  width: 1px;
  height: 1px;
  overflow: hidden;
  clip: rect(0, 0, 0, 0);
}
.field-divider {
  height: 1px;
  background: var(--c-border);
  margin-bottom: 20px;
}

/* Form */
.edit-form {
  display: flex;
  flex-direction: column;
  gap: 16px;
  overflow-y: auto;
  padding-bottom: 8px;
}
.field-group {
  display: flex;
  flex-direction: column;
  gap: 6px;
}
.field-label {
  font-size: 0.68rem;
  font-weight: 700;
  letter-spacing: 0.08em;
  text-transform: uppercase;
  color: var(--c-muted);
}
.field-input,
.field-textarea {
  background: var(--c-card);
  border: 1px solid var(--c-border);
  border-radius: 8px;
  padding: 10px 14px;
  font-family: var(--font);
  font-size: 0.875rem;
  color: var(--c-text);
  outline: none;
  transition: border-color 0.2s;
  width: 100%;
  box-sizing: border-box;
}
.field-input:focus,
.field-textarea:focus {
  border-color: rgba(255, 255, 255, 0.2);
}
.field-input:disabled {
  opacity: 0.45;
  cursor: not-allowed;
}
.field-input--unlocked {
  border-color: rgba(255, 255, 255, 0.2);
  background: #1a1a1a;
}
.field-textarea {
  resize: none;
  line-height: 1.55;
}
.input-prefix-wrap {
  position: relative;
}
.input-prefix {
  position: absolute;
  left: 12px;
  top: 50%;
  transform: translateY(-50%);
  color: var(--c-sub);
  font-size: 0.875rem;
  pointer-events: none;
}
.field-input--prefixed {
  padding-left: 26px;
}
.field-input--date {
  color-scheme: dark;
}
.char-count {
  font-size: 0.65rem;
  color: var(--c-muted);
  text-align: right;
  margin-top: -4px;
}

/* Email */
.email-row {
  display: flex;
  gap: 8px;
  align-items: stretch;
}
.email-input {
  flex: 1;
  min-width: 0;
}
.email-action-btn {
  display: inline-flex;
  align-items: center;
  gap: 5px;
  font-family: var(--font);
  font-size: 0.72rem;
  font-weight: 600;
  white-space: nowrap;
  padding: 0 12px;
  border-radius: 8px;
  border: 1px solid var(--c-border);
  background: var(--c-card);
  color: var(--c-text);
  cursor: pointer;
  transition: all 0.18s;
  flex-shrink: 0;
}
.email-action-btn:hover:not(:disabled) {
  border-color: var(--c-border-h);
}
.email-action-btn:disabled {
  opacity: 0.45;
  cursor: not-allowed;
}
.email-action-btn--verify {
  border-color: rgba(255, 214, 10, 0.3);
  color: var(--c-yellow);
}
.email-action-btn--verify:hover:not(:disabled) {
  background: rgba(255, 214, 10, 0.08);
  border-color: rgba(255, 214, 10, 0.5);
}
.email-action-btn--cancel {
  color: var(--c-sub);
}
.email-action-btn--confirm {
  background: var(--c-red);
  border-color: var(--c-red);
  color: #fff;
  padding: 0 16px;
}
.email-action-btn--confirm:hover:not(:disabled) {
  background: #ff3b30;
}
.email-action-btn--waiting {
  color: var(--c-sub);
  gap: 6px;
}
.email-action-btn--done {
  border-color: rgba(48, 209, 88, 0.4);
  color: var(--c-green);
  background: rgba(48, 209, 88, 0.08);
}

/* OTP */
.otp-box {
  background: var(--c-card);
  border: 1px solid var(--c-border);
  border-radius: 10px;
  padding: 14px;
  display: flex;
  flex-direction: column;
  gap: 10px;
}
.otp-hint {
  font-size: 0.78rem;
  color: var(--c-sub);
  margin: 0;
  line-height: 1.5;
}
.otp-hint strong {
  color: var(--c-text);
  font-weight: 500;
}
.otp-row {
  display: flex;
  gap: 8px;
}
.otp-input {
  flex: 1;
  letter-spacing: 0.25em;
  font-size: 1rem;
  text-align: center;
}
.otp-error {
  font-size: 0.72rem;
  color: #ff453a;
  margin: 0;
}
.otp-box--waiting {
  border-color: rgba(255, 255, 255, 0.1);
}
.waiting-row {
  display: flex;
  align-items: flex-start;
  gap: 12px;
}
.resend-btn {
  font-family: var(--font);
  font-size: 0.7rem;
  color: var(--c-sub);
  background: none;
  border: none;
  cursor: pointer;
  padding: 0;
  text-decoration: underline;
  text-underline-offset: 2px;
  align-self: flex-start;
}
.resend-btn:hover {
  color: var(--c-text);
}
.email-hint {
  display: flex;
  align-items: center;
  gap: 5px;
  font-size: 0.7rem;
  color: var(--c-sub);
  margin: 0;
}

/* ── Change Password section ──────────────────────────────────────── */
.section-header {
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding: 10px 14px;
  background: var(--c-card);
  border: 1px solid var(--c-border);
  border-radius: 9px;
  cursor: pointer;
  user-select: none;
  transition: border-color 0.18s;
  margin-bottom: 4px;
}
.section-header:hover {
  border-color: var(--c-border-h);
}
.section-header__left {
  display: flex;
  align-items: center;
  gap: 8px;
}
.section-header__icon {
  color: var(--c-sub);
  flex-shrink: 0;
}
.section-header__title {
  font-size: 0.8rem;
  font-weight: 600;
  color: var(--c-text);
}
.section-header__chevron {
  color: var(--c-sub);
  transition: transform 0.25s var(--ease);
  flex-shrink: 0;
}
.section-header__chevron--open {
  transform: rotate(180deg);
}

.pw-section {
  display: flex;
  flex-direction: column;
  gap: 12px;
  padding: 14px;
  background: var(--c-card);
  border: 1px solid var(--c-border);
  border-radius: 9px;
  margin-bottom: 4px;
}

.pw-input-wrap {
  position: relative;
}
.pw-input-wrap .field-input {
  padding-right: 36px;
}
.pw-toggle {
  position: absolute;
  right: 10px;
  top: 50%;
  transform: translateY(-50%);
  background: none;
  border: none;
  color: var(--c-sub);
  cursor: pointer;
  padding: 2px;
  display: flex;
  align-items: center;
  transition: color 0.15s;
}
.pw-toggle:hover {
  color: var(--c-text);
}

/* strength bar */
.strength-bar {
  height: 3px;
  background: rgba(255, 255, 255, 0.08);
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
  font-size: 0.68rem;
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

/* inline error */
.inline-error {
  display: flex;
  align-items: center;
  gap: 4px;
  font-size: 0.7rem;
  color: #ff453a;
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

/* pw feedback alerts */
.pw-alert {
  display: flex;
  align-items: center;
  gap: 6px;
  font-size: 0.75rem;
  padding: 8px 12px;
  border-radius: 8px;
}
.pw-alert--error {
  background: rgba(245, 67, 44, 0.1);
  border: 1px solid rgba(245, 67, 44, 0.25);
  color: #ff6b6b;
}
.pw-alert--success {
  background: rgba(48, 209, 88, 0.1);
  border: 1px solid rgba(48, 209, 88, 0.25);
  color: #30d158;
}

.pw-submit-btn {
  align-self: flex-end;
  display: inline-flex;
  align-items: center;
  gap: 6px;
  font-family: var(--font);
  font-size: 0.8rem;
  font-weight: 600;
  padding: 9px 18px;
  background: var(--c-red);
  color: #fff;
  border: none;
  border-radius: 8px;
  cursor: pointer;
  transition:
    background 0.2s,
    opacity 0.2s;
}
.pw-submit-btn:hover:not(:disabled) {
  background: #ff3b30;
}
.pw-submit-btn:disabled {
  opacity: 0.45;
  cursor: not-allowed;
}

/* animations */
.pw-slide-enter-active {
  transition: all 0.28s var(--ease);
}
.pw-slide-leave-active {
  transition: all 0.2s ease-in;
}
.pw-slide-enter-from,
.pw-slide-leave-to {
  opacity: 0;
  transform: translateY(-8px);
}

.otp-slide-enter-active,
.otp-slide-leave-active {
  transition: all 0.22s var(--ease);
}
.otp-slide-enter-from,
.otp-slide-leave-to {
  opacity: 0;
  transform: translateY(-6px);
}

/* Gender */
.radio-group {
  display: flex;
  gap: 8px;
}
.radio-option {
  flex: 1;
  text-align: center;
  padding: 8px 0;
  font-size: 0.8rem;
  font-weight: 500;
  border-radius: 8px;
  border: 1px solid var(--c-border);
  background: var(--c-card);
  color: var(--c-sub);
  cursor: pointer;
  transition: all 0.18s;
  user-select: none;
}
.radio-option:hover {
  color: var(--c-text);
  border-color: var(--c-border-h);
}
.radio-option--active {
  border-color: var(--c-red);
  color: var(--c-text);
  background: rgba(225, 37, 27, 0.1);
}

/* Toggle */
.toggle-row {
  display: flex;
  align-items: center;
  justify-content: space-between;
  background: var(--c-card);
  border: 1px solid var(--c-border);
  border-radius: 10px;
  padding: 13px 14px;
}
.toggle-info {
  display: flex;
  flex-direction: column;
  gap: 2px;
}
.toggle-label {
  font-size: 0.875rem;
  font-weight: 500;
  color: var(--c-text);
}
.toggle-desc {
  font-size: 0.72rem;
  color: var(--c-sub);
}
.toggle-btn {
  width: 42px;
  height: 24px;
  border-radius: 99px;
  border: none;
  background: var(--c-muted);
  cursor: pointer;
  position: relative;
  transition: background 0.2s;
  flex-shrink: 0;
}
.toggle-btn--on {
  background: var(--c-red);
}
.toggle-knob {
  position: absolute;
  top: 3px;
  left: 3px;
  width: 18px;
  height: 18px;
  border-radius: 50%;
  background: #fff;
  transition: transform 0.2s;
  display: block;
}
.toggle-btn--on .toggle-knob {
  transform: translateX(18px);
}

/* Footer */
.edit-footer {
  display: flex;
  gap: 10px;
  justify-content: flex-end;
  margin-top: 8px;
  padding-top: 16px;
  border-top: 1px solid var(--c-border);
  flex-shrink: 0;
}
.footer-btn {
  font-family: var(--font);
  font-size: 0.82rem;
  font-weight: 500;
  padding: 9px 18px;
  border-radius: 8px;
  cursor: pointer;
  transition: all 0.2s;
  border: none;
  display: flex;
  align-items: center;
  gap: 6px;
}
.footer-btn--cancel {
  background: var(--c-card);
  color: var(--c-sub);
  border: 1px solid var(--c-border);
}
.footer-btn--cancel:hover {
  color: var(--c-text);
  border-color: var(--c-border-h);
}
.footer-btn--save {
  background: var(--c-red);
  color: #fff;
  min-width: 110px;
  justify-content: center;
}
.footer-btn--save:hover:not(:disabled) {
  background: #ff3b30;
  box-shadow: 0 4px 14px rgba(225, 37, 27, 0.35);
}
.footer-btn--save:disabled {
  opacity: 0.5;
  cursor: not-allowed;
}
</style>
