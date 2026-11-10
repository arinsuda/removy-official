<template>
  <SettingsCard id="profile" :title="t('settings.profile.title')" :description="t('settings.profile.desc')">
    <AvatarPicker :src="user.avatar_url" :preview="avatarPreview" :name="user.display_name || user.username" @pick="emit('avatar', $event)" />

    <div class="row">
      <AppInput v-model="displayName" :label="t('settings.profile.displayName')" :maxlength="DISPLAY_NAME_MAX" :hint="`${len(displayName)} / ${DISPLAY_NAME_MAX}`" :placeholder="user.username" />
      <AppInput :model-value="`@${user.username}`" :label="t('settings.profile.username')" :hint="t('settings.profile.usernameHint')" disabled />
    </div>

    <div class="field">
      <label class="field__label" :for="bioId">{{ t("settings.profile.bio") }}</label>
      <textarea :id="bioId" data-focus-custom v-model="bio" class="field__area" rows="3" :maxlength="BIO_MAX" :placeholder="t('settings.profile.bioPlaceholder')" />
      <p class="field__hint">{{ len(bio) }} / {{ BIO_MAX }}</p>
    </div>

    <div class="field">
      <label class="field__label" :for="dobId">{{ t("settings.profile.dob") }}</label>
      <input :id="dobId" v-model="dateOfBirth" type="date" class="field__input" min="1900-01-01" :max="today" />
    </div>

    <fieldset class="field">
      <legend class="field__label">{{ t("settings.profile.gender") }}</legend>
      <div class="chips">
        <label v-for="g in genders" :key="g" class="chip" :class="{ 'is-on': gender === g }">
          <input v-model="gender" type="radio" class="sr-only" :value="g" />{{ t(`settings.profile.genders.${g}`) }}
        </label>
      </div>
      <AppInput v-if="gender === 'other'" v-model="genderOther" :label="t('settings.profile.genderOther')" maxlength="100" />
    </fieldset>
  </SettingsCard>
</template>

<script setup lang="ts">
import { useId } from "vue"
import { useI18n } from "vue-i18n"
import AppInput from "@/components/ui/AppInput.vue"
import type { UserProfile } from "@/types/user"
import AvatarPicker from "./AvatarPicker.vue"
import SettingsCard from "./SettingsCard.vue"
import { BIO_MAX, DISPLAY_NAME_MAX } from "./useProfileForm"

defineProps<{ user: UserProfile; avatarPreview: string | null }>()
const displayName = defineModel<string>("displayName", { required: true })
const bio = defineModel<string>("bio", { required: true })
const dateOfBirth = defineModel<string>("dateOfBirth", { required: true })
const gender = defineModel<string>("gender", { required: true })
const genderOther = defineModel<string>("genderOther", { required: true })
const emit = defineEmits<{ avatar: [blob: Blob] }>()
const { t } = useI18n()
const bioId = useId()
const dobId = useId()
const genders = ["male", "female", "other"] as const
const today = new Date().toISOString().slice(0, 10)
const len = (s: string) => Array.from(s).length
</script>

<style scoped>
.row { display: grid; gap: 1rem; }
@media (min-width: 640px) { .row { grid-template-columns: repeat(2, minmax(0, 1fr)); } }
.field { display: flex; flex-direction: column; gap: 0.375rem; min-width: 0; padding: 0; border: 0; }
.field__label { font-size: 0.8125rem; font-weight: 600; color: var(--color-text-secondary); }
.field__input, .field__area { width: 100%; padding: 0.5rem 0.75rem; color: var(--color-text-primary); background: var(--color-surface-3); border: 1px solid var(--color-border); border-radius: var(--radius-btn); }
.field__input { max-width: 14rem; color-scheme: dark; }
.field__area { resize: vertical; border-radius: var(--radius-card); }
.field__area:focus-visible { outline: none; border-color: var(--color-focus); box-shadow: var(--focus-input-shadow); }
.field__hint { font-size: var(--text-caption); color: var(--color-text-muted); text-align: right; }
.chips { display: flex; flex-wrap: wrap; gap: 0.5rem; }
.chip { padding: 0.375rem 0.875rem; font-size: 0.8125rem; color: var(--color-text-secondary); cursor: pointer; border: 1px solid var(--color-border-strong); border-radius: 999px; }
.chip:focus-within { outline: 2px solid var(--color-focus); outline-offset: 2px; }
.chip.is-on { color: var(--color-surface); background: var(--color-text-primary); border-color: var(--color-text-primary); }
.sr-only { position: absolute; width: 1px; height: 1px; overflow: hidden; clip-path: inset(50%); }
</style>
