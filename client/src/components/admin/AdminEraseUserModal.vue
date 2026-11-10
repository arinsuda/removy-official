<template>
  <AppModal
    :model-value="modelValue"
    role="alertdialog"
    size="sm"
    :show-close="false"
    :dismiss-disabled="loading"
    :aria-labelledby="titleId"
    :aria-describedby="descId"
    @update:model-value="emit('update:modelValue', $event)"
  >
    <div class="erase">
      <div class="erase__badge"><Flame :size="24" :stroke-width="1.8" aria-hidden="true" /></div>
      <div class="erase__text">
        <h3 :id="titleId" class="erase__title">{{ t("admin.userModal.eraseTitle") }}</h3>
        <p :id="descId" class="erase__desc">{{ t("admin.userModal.eraseDesc", { username: handle }) }}</p>
        <p class="erase__warning">{{ t("admin.userModal.eraseWarning") }}</p>
      </div>

      <form :id="formId" class="erase__form" @submit.prevent="submit">
        <AppInput
          v-model="typed"
          :label="t('admin.userModal.eraseConfirmLabel', { username: user?.username ?? '' })"
          :hint="t('admin.userModal.eraseConfirmHint')"
          :disabled="loading"
          autocomplete="off"
          autocapitalize="off"
          spellcheck="false"
        />
        <div class="erase__reason">
          <label :for="reasonId" class="erase__label">{{ t("admin.userModal.eraseReasonLabel") }}</label>
          <textarea :id="reasonId" v-model="reason" class="erase__textarea" rows="2" :disabled="loading" />
          <p class="erase__note">{{ t("admin.userModal.eraseReasonNote") }} {{ t("admin.userModal.reasonHint") }}</p>
        </div>
        <p v-if="errorKey" class="erase__error" role="alert">{{ t(errorKey) }}</p>
      </form>
    </div>

    <template #footer>
      <div class="erase__actions">
        <AppButton variant="outline" :disabled="loading" @click="close">{{ t("common.cancel") }}</AppButton>
        <AppButton variant="danger" :loading="loading" :disabled="!matches" @click="submit">
          {{ t("admin.userModal.eraseConfirm") }}
        </AppButton>
      </div>
    </template>
  </AppModal>
</template>

<script setup lang="ts">
import { computed, ref, useId, watch } from "vue"
import { useI18n } from "vue-i18n"
import { Flame } from "lucide-vue-next"
import AppButton from "@/components/ui/AppButton.vue"
import AppInput from "@/components/ui/AppInput.vue"
import AppModal from "@/components/ui/AppModal.vue"
import { useAdminStore } from "@/stores/admin"
import type { AdminUserRow } from "@/types"
import { adminErrorKey } from "@/utils/adminError"

const props = defineProps<{ modelValue: boolean; user: AdminUserRow | null }>()
const emit = defineEmits<{
  "update:modelValue": [value: boolean]
  /** The account is gone; the parent refreshes its list and toasts. */
  erased: [user: AdminUserRow]
  /** The server says the account no longer exists or is not erasable in this state. */
  stale: []
}>()

const { t } = useI18n()
const adminStore = useAdminStore()
const titleId = useId()
const descId = useId()
const reasonId = useId()
const formId = useId()

const typed = ref("")
const reason = ref("")
const loading = ref(false)
const errorKey = ref("")

const handle = computed(() => `@${props.user?.username ?? ""}`)
// Exact, case-sensitive match: the server compares the same way.
const matches = computed(() => !!props.user && typed.value === props.user.username)

watch(
  () => props.modelValue,
  (open) => {
    if (!open) return
    typed.value = ""
    reason.value = ""
    errorKey.value = ""
  },
  { immediate: true },
)

function close() {
  if (!loading.value) emit("update:modelValue", false)
}

async function submit() {
  const user = props.user
  if (!user || !matches.value || loading.value) return
  loading.value = true
  errorKey.value = ""
  try {
    await adminStore.eraseUser(user.id, {
      confirm_username: typed.value,
      reason: reason.value.trim() || undefined,
    })
    emit("update:modelValue", false)
    emit("erased", user)
  } catch (err: unknown) {
    errorKey.value = adminErrorKey(err)
    const status = (err as { response?: { status?: number } }).response?.status
    if (status === 404 || status === 409) emit("stale")
  } finally {
    loading.value = false
  }
}
</script>

<style scoped>
.erase {
  display: flex;
  flex-direction: column;
  align-items: center;
  gap: 1.25rem;
  padding-top: 0.5rem;
}

.erase__badge {
  display: flex;
  align-items: center;
  justify-content: center;
  width: 3.5rem;
  height: 3.5rem;
  border-radius: var(--radius-card);
  color: var(--color-danger);
  background: color-mix(in srgb, var(--color-danger) 15%, transparent);
  border: 1px solid color-mix(in srgb, var(--color-danger) 30%, transparent);
}

.erase__text {
  display: flex;
  flex-direction: column;
  align-items: center;
  gap: 0.5rem;
  text-align: center;
}

.erase__title {
  margin: 0;
  font-size: var(--text-title);
  font-weight: 700;
  color: var(--color-text-primary);
}

.erase__desc,
.erase__note {
  margin: 0;
  font-size: var(--text-small);
  line-height: 1.5;
  color: var(--color-text-secondary);
}

.erase__warning {
  margin: 0;
  font-size: var(--text-small);
  font-weight: 700;
  color: var(--color-danger);
}

.erase__form,
.erase__reason {
  display: flex;
  flex-direction: column;
  gap: 0.75rem;
  width: 100%;
}

.erase__reason {
  gap: 0.5rem;
}

.erase__label {
  font-size: var(--text-caption);
  font-weight: 600;
  color: var(--color-text-secondary);
}

.erase__textarea {
  width: 100%;
  min-height: 3.5rem;
  padding: 0.625rem 0.875rem;
  box-sizing: border-box;
  background: var(--color-surface);
  border: 1px solid var(--color-border-strong);
  border-radius: var(--radius-btn);
  color: var(--color-text-primary);
  font-family: var(--font-sans);
  font-size: var(--text-small);
  resize: vertical;
}

.erase__error {
  margin: 0;
  font-size: var(--text-small);
  color: var(--color-danger);
}

.erase__actions {
  display: flex;
  gap: 0.75rem;
  width: 100%;
}

.erase__actions > * {
  flex: 1;
}
</style>
