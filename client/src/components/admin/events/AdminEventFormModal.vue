<template>
  <AppModal
    v-model="open"
    size="md"
    :title="editing ? t('admin.events.editEvent') : t('admin.events.newEvent')"
  >
    <form class="event-form" @submit.prevent="emit('submit')">
      <div class="form-row">
        <div class="form-field">
          <label class="form-label" for="event-title-en">
            {{ t("admin.events.form.titleEn") }} *
          </label>
          <input
            id="event-title-en"
            v-model="form.title_en"
            type="text"
            class="form-input"
            :class="{ 'is-invalid': errors.title_en }"
            :placeholder="generatedTitlePlaceholder(editing, 'en') || t('admin.events.form.titleEnPlaceholder')"
            :maxlength="EVENT_TITLE_MAX"
            required
          />
          <span v-if="errors.title_en" class="form-error">
            {{ t(errors.title_en) }}
          </span>
        </div>

        <div class="form-field">
          <label class="form-label" for="event-title-th">
            {{ t("admin.events.form.titleTh") }} *
          </label>
          <input
            id="event-title-th"
            v-model="form.title_th"
            type="text"
            class="form-input"
            :class="{ 'is-invalid': errors.title_th }"
            :placeholder="generatedTitlePlaceholder(editing, 'th') || t('admin.events.form.titleThPlaceholder')"
            :maxlength="EVENT_TITLE_MAX"
            required
          />
          <span v-if="errors.title_th" class="form-error">
            {{ t(errors.title_th) }}
          </span>
        </div>
      </div>

      <div class="form-row">
        <div class="form-field">
          <label class="form-label" for="event-body-en">
            {{ t("admin.events.form.bodyEn") }}
          </label>
          <textarea
            id="event-body-en"
            v-model="form.body_en"
            class="form-textarea"
            :class="{ 'is-invalid': errors.body_en }"
            :placeholder="t('admin.events.form.bodyEnPlaceholder')"
            rows="3"
          />
          <span v-if="errors.body_en" class="form-error">
            {{ t(errors.body_en) }}
          </span>
        </div>

        <div class="form-field">
          <label class="form-label" for="event-body-th">
            {{ t("admin.events.form.bodyTh") }}
          </label>
          <textarea
            id="event-body-th"
            v-model="form.body_th"
            class="form-textarea"
            :class="{ 'is-invalid': errors.body_th }"
            :placeholder="t('admin.events.form.bodyThPlaceholder')"
            rows="3"
          />
          <span v-if="errors.body_th" class="form-error">
            {{ t(errors.body_th) }}
          </span>
        </div>
      </div>

      <div class="form-field">
        <label class="form-label" for="event-link-url">
          {{ t("admin.events.form.linkUrl") }}
        </label>
        <input
          id="event-link-url"
          v-model="form.link_url"
          type="text"
          class="form-input"
          :class="{ 'is-invalid': errors.link_url }"
          :placeholder="t('admin.events.form.linkUrlPlaceholder')"
        />
        <span v-if="errors.link_url" class="form-error">
          {{ t(errors.link_url) }}
        </span>
      </div>

      <div class="form-row">
        <div class="form-field">
          <label class="form-label" for="event-starts-at">
            {{ t("admin.events.form.startsAt") }} *
          </label>
          <input
            id="event-starts-at"
            v-model="startsInput"
            type="datetime-local"
            class="form-input"
            :class="{ 'is-invalid': errors.starts_at }"
            :disabled="locked"
            required
          />
          <span v-if="errors.starts_at" class="form-error">
            {{ t(errors.starts_at) }}
          </span>
        </div>

        <div class="form-field">
          <label class="form-label" for="event-ends-at">
            {{ t("admin.events.form.endsAt") }} *
          </label>
          <input
            id="event-ends-at"
            v-model="endsInput"
            type="datetime-local"
            class="form-input"
            :class="{ 'is-invalid': errors.ends_at }"
            :disabled="locked"
            required
          />
          <span v-if="errors.ends_at" class="form-error">
            {{ t(errors.ends_at) }}
          </span>
        </div>
      </div>
      <p v-if="locked" class="form-hint">
        {{ t("admin.events.lockedScheduleHint") }}
      </p>

      <div class="form-row">
        <div class="form-field">
          <label class="form-label" for="event-status">
            {{ t("admin.events.form.status") }}
          </label>
          <select id="event-status" v-model="form.status" class="form-input">
            <option value="draft">{{ t("admin.events.status.draft") }}</option>
            <option value="published">{{ t("admin.events.status.published") }}</option>
            <option value="cancelled">{{ t("admin.events.status.cancelled") }}</option>
          </select>
        </div>

        <div class="form-field">
          <label class="form-label" for="event-priority">
            {{ t("admin.events.form.popupPriority") }}
          </label>
          <input
            id="event-priority"
            v-model.number="form.popup_priority"
            type="number"
            class="form-input"
            :min="EVENT_POPUP_PRIORITY_MIN"
            :max="EVENT_POPUP_PRIORITY_MAX"
          />
        </div>
      </div>

      <div class="form-field form-field--checkbox">
        <label class="form-checkbox-label">
          <input
            v-model="form.show_popup"
            type="checkbox"
            class="form-checkbox"
          />
          <span>{{ t("admin.events.form.showPopup") }}</span>
        </label>
      </div>

      <div class="form-actions">
        <AppButton
          variant="outline"
          size="md"
          type="button"
          :disabled="saving"
          @click="open = false"
        >
          {{ t("admin.events.form.cancel") }}
        </AppButton>
        <AppButton
          variant="primary"
          size="md"
          type="submit"
          :loading="saving"
        >
          {{ t("admin.events.form.save") }}
        </AppButton>
      </div>
    </form>
  </AppModal>
</template>

<script setup lang="ts">
import { useI18n } from "vue-i18n"
import AppButton from "@/components/ui/AppButton.vue"
import AppModal from "@/components/ui/AppModal.vue"
import type { AdminEvent, AdminEventBody } from "@/types/event"
import {
  EVENT_POPUP_PRIORITY_MAX,
  EVENT_POPUP_PRIORITY_MIN,
  EVENT_TITLE_MAX,
  type EventFormErrors,
  generatedRoundTitle,
} from "@/utils/events"
import "./adminEventForm.css"

defineProps<{
  /** Event being edited, or null when creating one. */
  editing: AdminEvent | null
  errors: EventFormErrors
  saving: boolean
  /** Schedule is read-only once voting has opened. */
  locked: boolean
}>()

const emit = defineEmits<{ submit: [] }>()

const open = defineModel<boolean>({ required: true })
const form = defineModel<AdminEventBody>("form", { required: true })
const startsInput = defineModel<string>("startsInput", { required: true })
const endsInput = defineModel<string>("endsInput", { required: true })

const { t } = useI18n()

/** Placeholder for a blank title input on a monthly round: the copy the server would generate. */
function generatedTitlePlaceholder(editing: AdminEvent | null, lang: "en" | "th"): string {
  if (!editing) return ""
  return generatedRoundTitle(editing, t, lang) ?? ""
}
</script>
