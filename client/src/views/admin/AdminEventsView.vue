<template>
  <AdminPage>
    <AdminPageHeader
      :title="t('admin.events.title')"
      :subtitle="t('admin.events.subtitle')"
    >
      <div class="admin-header-actions">
        <AppButton variant="outline" size="sm" @click="openGenerateModal">
          <Sparkles :size="16" aria-hidden="true" />
          {{ t("admin.events.generateMonthly") }}
        </AppButton>
        <AppButton variant="primary" size="sm" @click="openCreateModal">
          <Plus :size="16" aria-hidden="true" />
          {{ t("admin.events.newEvent") }}
        </AppButton>
      </div>
    </AdminPageHeader>

    <AdminFilters>
      <AdminSelect
        v-model="statusFilter"
        :label="t('admin.events.table.status')"
        :options="statusOptions"
      />
    </AdminFilters>

    <AdminTable
      :columns="columns"
      :label="t('admin.events.title')"
      :loading="loading"
      :error="error"
      :empty="events.length === 0"
      :filtered="statusFilter !== ''"
      :empty-title="t('admin.events.table.emptyTitle')"
      :empty-desc="t('admin.events.table.emptyDesc')"
      @retry="fetchEvents"
      @clear="statusFilter = ''"
    >
      <tr v-for="ev in events" :key="ev.id">
        <td>
          <div class="event-title-cell">
            <div class="event-thumb">
              <img
                v-if="ev.image_url"
                :src="ev.image_url"
                alt=""
                class="event-thumb__img"
                loading="lazy"
              />
              <span v-else class="event-thumb__empty" aria-hidden="true">
                <Megaphone :size="14" />
              </span>
            </div>
            <div class="event-title-info">
              <span class="event-title-primary">
                {{ eventTitle(ev, t, locale) }}
                <span class="inline-id">#{{ ev.id }}</span>
                <span v-if="ev.type !== 'announcement'" class="event-type-tag">
                  {{ t(`admin.events.types.${ev.type}`) }}
                </span>
              </span>
              <span v-if="locale === 'th' ? ev.title.en : ev.title.th" class="event-title-sub">
                {{ locale === "th" ? ev.title.en : ev.title.th }}
              </span>
            </div>
          </div>
        </td>
        <td>
          <AdminBadge :tone="statusBadgeTone(ev.status)">
            {{ t(`admin.events.status.${ev.status}`) }}
          </AdminBadge>
        </td>
        <td class="cell-dates">
          <time :datetime="ev.starts_at">{{ formatDate(ev.starts_at) }}</time>
          <span class="cell-dates__sep">{{ t("admin.events.table.dateSeparator") }}</span>
          <time :datetime="ev.ends_at">{{ formatDate(ev.ends_at) }}</time>
        </td>
        <td>
          <AdminBadge :tone="ev.show_popup ? 'success' : 'neutral'">
            {{ ev.show_popup ? t("admin.events.table.popupYes") : t("admin.events.table.popupNo") }}
          </AdminBadge>
        </td>
        <td class="cell-priority">
          {{ ev.popup_priority }}
        </td>
        <td class="is-end">
          <AdminRowMenu
            :label="t('admin.table.rowActions', { name: eventTitle(ev, t, locale) })"
            :items="getRowMenuItems(ev)"
            @select="(key) => onMenuAction(ev, key)"
          />
        </td>
      </tr>
    </AdminTable>

    <AdminPagination
      v-if="pagination"
      :page="pagination.page"
      :total-pages="pagination.total_pages"
      :total="pagination.total"
      :limit="pagination.limit"
      @change="onPageChange"
    />

    <!-- Create / Edit Modal -->
    <AppModal
      v-model="showFormModal"
      size="md"
      :title="editingEvent ? t('admin.events.editEvent') : t('admin.events.newEvent')"
    >
      <form class="event-form" @submit.prevent="submitForm">
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
              :class="{ 'is-invalid': formErrors.title_en }"
              :placeholder="generatedTitlePlaceholder('en') || t('admin.events.form.titleEnPlaceholder')"
              :maxlength="EVENT_TITLE_MAX"
              required
            />
            <span v-if="formErrors.title_en" class="form-error">
              {{ t(formErrors.title_en) }}
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
              :class="{ 'is-invalid': formErrors.title_th }"
              :placeholder="generatedTitlePlaceholder('th') || t('admin.events.form.titleThPlaceholder')"
              :maxlength="EVENT_TITLE_MAX"
              required
            />
            <span v-if="formErrors.title_th" class="form-error">
              {{ t(formErrors.title_th) }}
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
              :class="{ 'is-invalid': formErrors.body_en }"
              :placeholder="t('admin.events.form.bodyEnPlaceholder')"
              rows="3"
            />
            <span v-if="formErrors.body_en" class="form-error">
              {{ t(formErrors.body_en) }}
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
              :class="{ 'is-invalid': formErrors.body_th }"
              :placeholder="t('admin.events.form.bodyThPlaceholder')"
              rows="3"
            />
            <span v-if="formErrors.body_th" class="form-error">
              {{ t(formErrors.body_th) }}
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
            :class="{ 'is-invalid': formErrors.link_url }"
            :placeholder="t('admin.events.form.linkUrlPlaceholder')"
          />
          <span v-if="formErrors.link_url" class="form-error">
            {{ t(formErrors.link_url) }}
          </span>
        </div>

        <div class="form-row">
          <div class="form-field">
            <label class="form-label" for="event-starts-at">
              {{ t("admin.events.form.startsAt") }} *
            </label>
            <input
              id="event-starts-at"
              v-model="formStartsInput"
              type="datetime-local"
              class="form-input"
              :class="{ 'is-invalid': formErrors.starts_at }"
              :disabled="isVotingLocked"
              required
            />
            <span v-if="formErrors.starts_at" class="form-error">
              {{ t(formErrors.starts_at) }}
            </span>
          </div>

          <div class="form-field">
            <label class="form-label" for="event-ends-at">
              {{ t("admin.events.form.endsAt") }} *
            </label>
            <input
              id="event-ends-at"
              v-model="formEndsInput"
              type="datetime-local"
              class="form-input"
              :class="{ 'is-invalid': formErrors.ends_at }"
              :disabled="isVotingLocked"
              required
            />
            <span v-if="formErrors.ends_at" class="form-error">
              {{ t(formErrors.ends_at) }}
            </span>
          </div>
        </div>
        <p v-if="isVotingLocked" class="form-hint">
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
            @click="showFormModal = false"
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

    <!-- Hidden file input for image uploads -->
    <input
      ref="fileInputRef"
      type="file"
      class="sr-only"
      accept="image/jpeg,image/png,image/webp,image/gif"
      @change="onFileSelected"
    />

    <!-- Delete Confirmation Modal -->
    <AdminConfirmModal
      v-model="showDeleteModal"
      variant="danger"
      :title="t('admin.events.deleteTitle')"
      :description="t('admin.events.deleteDesc')"
      :confirm-text="t('admin.events.deleteConfirm')"
      :target-name="eventToDelete ? eventTitle(eventToDelete, t, locale) : undefined"
      :loading="deleting"
      @confirm="confirmDelete"
      @cancel="showDeleteModal = false"
    />

    <!-- Generate Monthly Modal -->
    <AppModal
      v-model="showGenerateModal"
      size="md"
      :title="t('admin.events.generateMonthlyModalTitle')"
    >
      <form class="event-form" @submit.prevent="handleGenerateMonthly">
        <div class="form-field">
          <label class="form-label">{{ t("admin.events.mediaType") }}</label>
          <div class="generate-radio-group">
            <label class="generate-radio">
              <input v-model="generateMediaType" type="radio" value="movie" />
              <span>{{ t("admin.events.movie") }}</span>
            </label>
            <label class="generate-radio">
              <input v-model="generateMediaType" type="radio" value="tv" />
              <span>{{ t("admin.events.tv") }}</span>
            </label>
          </div>
        </div>

        <div class="form-field">
          <label class="form-label" for="gen-period">{{ t("admin.events.period") }}</label>
          <input
            id="gen-period"
            v-model="generatePeriod"
            type="text"
            class="form-input"
            :placeholder="t('admin.events.periodPlaceholder')"
            required
          />
        </div>

        <div class="generate-preview-action">
          <AppButton
            variant="outline"
            size="sm"
            :loading="loadingSuggestions"
            @click.prevent="loadSuggestions"
          >
            {{ t("admin.events.previewNominees") }}
          </AppButton>
        </div>

        <div v-if="suggestedNominees.length" class="suggested-list">
          <div
            v-for="(nom, idx) in suggestedNominees"
            :key="nom.media_id"
            class="suggested-item"
          >
            <span class="suggested-item__idx">#{{ idx + 1 }}</span>
            <div class="suggested-item__info">
              <span class="suggested-item__title">{{ nom.media_title }}</span>
              <span class="suggested-item__source">{{ formatSource(nom.source) }}</span>
            </div>
          </div>
        </div>

        <div class="form-actions">
          <AppButton variant="ghost" @click="showGenerateModal = false">
            {{ t("admin.events.form.cancel") }}
          </AppButton>
          <AppButton variant="primary" type="submit" :loading="generating">
            {{ t("admin.events.generateRound") }}
          </AppButton>
        </div>
      </form>
    </AppModal>

    <!-- Finalize Confirm Modal -->
    <AdminConfirmModal
      v-model="showFinalizeModal"
      :title="t('admin.events.finalizeTitle')"
      :description="t('admin.events.finalizeDesc')"
      :confirm-text="t('admin.events.finalize')"
      :target-name="eventToFinalize ? eventTitle(eventToFinalize, t, locale) : undefined"
      :loading="finalizing"
      @confirm="confirmFinalize"
      @cancel="showFinalizeModal = false"
    />

    <!-- Manage Nominees Modal -->
    <AppModal
      v-model="showNomineesModal"
      size="md"
      :title="t('admin.events.nomineesModalTitle', { title: nomineesEvent ? eventTitle(nomineesEvent, t, locale) : '' })"
    >
      <div class="event-form">
        <div class="generate-preview-action">
          <AppButton
            variant="outline"
            size="sm"
            :loading="loadingNomineeSuggestions"
            @click.prevent="loadNomineeSuggestionsForRound"
          >
            {{ t("admin.events.loadSuggestedNominees") }}
          </AppButton>
        </div>

        <div v-if="editableNominees.length" class="suggested-list">
          <div
            v-for="(nom, idx) in editableNominees"
            :key="nom.media_id"
            class="suggested-item"
          >
            <span class="suggested-item__idx">#{{ idx + 1 }}</span>
            <div class="suggested-item__info">
              <span class="suggested-item__title">{{ nom.media_title }}</span>
            </div>
            <div class="suggested-item__actions">
              <AppButton
                variant="ghost"
                size="sm"
                icon-only
                :aria-label="t('admin.events.moveUp')"
                :disabled="idx === 0"
                @click="moveNominee(idx, -1)"
              >
                <ArrowUp :size="14" />
              </AppButton>
              <AppButton
                variant="ghost"
                size="sm"
                icon-only
                :aria-label="t('admin.events.moveDown')"
                :disabled="idx === editableNominees.length - 1"
                @click="moveNominee(idx, 1)"
              >
                <ArrowDown :size="14" />
              </AppButton>
              <AppButton
                variant="ghost"
                size="sm"
                icon-only
                :aria-label="t('admin.events.removeNominee')"
                @click="removeNomineeAt(idx)"
              >
                <Trash2 :size="14" />
              </AppButton>
            </div>
          </div>
        </div>
        <p v-else class="form-hint">{{ t("admin.events.noNomineesYet") }}</p>

        <div class="form-actions">
          <AppButton variant="ghost" @click="showNomineesModal = false">
            {{ t("admin.events.form.cancel") }}
          </AppButton>
          <AppButton
            variant="primary"
            :loading="savingNominees"
            @click="saveNomineesForRound"
          >
            {{ t("admin.events.saveNominees") }}
          </AppButton>
        </div>
      </div>
    </AppModal>
  </AdminPage>
</template>

<script setup lang="ts">
import { computed, onMounted, ref, watch } from "vue"
import { useI18n } from "vue-i18n"
import { useRouter } from "vue-router"
import {
  ArrowDown,
  ArrowUp,
  Eye,
  Image as ImageIcon,
  Megaphone,
  Pencil,
  Plus,
  Sparkles,
  Trash2,
  Trophy,
} from "lucide-vue-next"
import AdminBadge from "@/components/admin/AdminBadge.vue"
import AdminConfirmModal from "@/components/admin/AdminConfirmModal.vue"
import AdminFilters from "@/components/admin/AdminFilters.vue"
import AdminPage from "@/components/admin/AdminPage.vue"
import AdminPageHeader from "@/components/admin/AdminPageHeader.vue"
import AdminPagination from "@/components/admin/AdminPagination.vue"
import AdminRowMenu, { type RowMenuItem } from "@/components/admin/AdminRowMenu.vue"
import AdminSelect from "@/components/admin/AdminSelect.vue"
import AdminTable from "@/components/admin/AdminTable.vue"
import AppButton from "@/components/ui/AppButton.vue"
import AppModal from "@/components/ui/AppModal.vue"
import { adminEventApi } from "@/api/endpoints/event"
import { useToast } from "@/composables/useToast"
import type {
  AdminEvent,
  AdminEventBody,
  EventPagination,
  EventStatus,
  NomineeInput,
  SuggestedNominee,
} from "@/types/event"
import {
  ADMIN_EVENTS_PAGE_SIZE,
  eventTitle,
  generatedRoundTitle,
  nextMonthlyPeriod,
  EVENT_DEFAULT_DURATION_MS,
  EVENT_IMAGE_MAX_BYTES,
  EVENT_IMAGE_TYPES,
  EVENT_POPUP_PRIORITY_MAX,
  EVENT_POPUP_PRIORITY_MIN,
  EVENT_TITLE_MAX,
  type EventFormErrors,
  eventErrorKey,
  isoToLocalInput,
  localInputToIso,
  validateEventForm,
} from "@/utils/events"

const { t, d, locale } = useI18n()
const router = useRouter()
const toast = useToast()

const events = ref<AdminEvent[]>([])
const pagination = ref<EventPagination | null>(null)
const page = ref(1)
const statusFilter = ref<string>("")
const loading = ref(false)
const error = ref(false)

const showFormModal = ref(false)
const editingEvent = ref<AdminEvent | null>(null)
const saving = ref(false)
const formErrors = ref<EventFormErrors>({})

const form = ref<AdminEventBody>({
  title_en: "",
  title_th: "",
  body_en: "",
  body_th: "",
  link_url: "",
  starts_at: "",
  ends_at: "",
  show_popup: false,
  popup_priority: 0,
  status: "draft",
})
const formStartsInput = ref("")
const formEndsInput = ref("")

const showDeleteModal = ref(false)
const eventToDelete = ref<AdminEvent | null>(null)
const deleting = ref(false)

const fileInputRef = ref<HTMLInputElement | null>(null)
const targetUploadEvent = ref<AdminEvent | null>(null)

const columns = computed(() => [
  { key: "title", label: t("admin.events.table.title") },
  { key: "status", label: t("admin.events.table.status") },
  { key: "schedule", label: t("admin.events.table.schedule") },
  { key: "popup", label: t("admin.events.table.popup") },
  { key: "priority", label: t("admin.events.table.priority") },
  { key: "actions", label: t("admin.events.table.actions"), align: "end" as const },
])

const statusOptions = computed(() => [
  { value: "", label: t("admin.events.status.all") },
  { value: "draft", label: t("admin.events.status.draft") },
  { value: "published", label: t("admin.events.status.published") },
  { value: "cancelled", label: t("admin.events.status.cancelled") },
])

function statusBadgeTone(status: EventStatus): "neutral" | "success" | "danger" {
  switch (status) {
    case "published":
      return "success"
    case "cancelled":
      return "danger"
    default:
      return "neutral"
  }
}

function formatDate(iso: string): string {
  try {
    return d(new Date(iso), "short")
  } catch {
    return iso.slice(0, 10)
  }
}

async function fetchEvents() {
  loading.value = true
  error.value = false
  try {
    const res = await adminEventApi.list({
      status: (statusFilter.value as EventStatus) || undefined,
      page: page.value,
      limit: ADMIN_EVENTS_PAGE_SIZE,
    })
    events.value = res.data.events
    pagination.value = res.data.pagination
  } catch {
    error.value = true
  } finally {
    loading.value = false
  }
}

function onPageChange(newPage: number) {
  page.value = newPage
  void fetchEvents()
}

watch(statusFilter, () => {
  page.value = 1
  void fetchEvents()
})

function getRowMenuItems(ev: AdminEvent): RowMenuItem[] {
  const items: RowMenuItem[] = [
    { key: "edit", label: t("common.edit"), icon: Pencil },
    {
      key: "upload-image",
      label: ev.image_url ? t("admin.events.form.changeImage") : t("admin.events.form.uploadImage"),
      icon: ImageIcon,
    },
  ]

  if (ev.type !== "announcement" && ev.state === "upcoming") {
    items.push({
      key: "edit-nominees",
      label: t("admin.events.editNominees"),
      icon: Sparkles,
    })
  }

  if (ev.type !== "announcement" && ev.state === "past" && !ev.is_finalized) {
    items.push({
      key: "finalize",
      label: t("admin.events.finalize"),
      icon: Trophy,
    })
  }

  // Drafts and cancelled events are not served publicly, so there is no page to open.
  if (ev.status === "published") {
    items.unshift({ key: "view", label: t("events.card.viewDetails"), icon: Eye })
  }

  if (ev.image_url) {
    items.push({
      key: "remove-image",
      label: t("admin.events.form.removeImage"),
      icon: ImageIcon,
      tone: "danger",
    })
  }

  items.push({
    key: "delete",
    label: t("common.delete"),
    icon: Trash2,
    tone: "danger",
  })

  return items
}

function onMenuAction(ev: AdminEvent, actionKey: string) {
  switch (actionKey) {
    case "view":
      void router.push(`/events/${ev.id}`)
      break
    case "edit":
      openEditModal(ev)
      break
    case "upload-image":
      targetUploadEvent.value = ev
      fileInputRef.value?.click()
      break
    case "remove-image":
      void handleRemoveImage(ev)
      break
    case "edit-nominees":
      openNomineesModal(ev)
      break
    case "finalize":
      eventToFinalize.value = ev
      showFinalizeModal.value = true
      break
    case "delete":
      eventToDelete.value = ev
      showDeleteModal.value = true
      break
  }
}

function openCreateModal() {
  editingEvent.value = null
  const now = new Date()
  const later = new Date(now.getTime() + EVENT_DEFAULT_DURATION_MS)

  formStartsInput.value = isoToLocalInput(now.toISOString())
  formEndsInput.value = isoToLocalInput(later.toISOString())

  form.value = {
    title_en: "",
    title_th: "",
    body_en: "",
    body_th: "",
    link_url: "",
    starts_at: localInputToIso(formStartsInput.value),
    ends_at: localInputToIso(formEndsInput.value),
    show_popup: false,
    popup_priority: 0,
    status: "draft",
  }
  formErrors.value = {}
  showFormModal.value = true
}

const isVotingLocked = computed(() =>
  Boolean(
    editingEvent.value &&
      editingEvent.value.type !== "announcement" &&
      editingEvent.value.state !== "upcoming"
  )
)

function openEditModal(ev: AdminEvent) {
  editingEvent.value = ev
  formStartsInput.value = isoToLocalInput(ev.starts_at)
  formEndsInput.value = isoToLocalInput(ev.ends_at)

  form.value = {
    type: ev.type,
    media_type: ev.media_type,
    period: ev.period,
    title_en: ev.title.en,
    title_th: ev.title.th,
    body_en: ev.body.en,
    body_th: ev.body.th,
    link_url: ev.link_url || "",
    starts_at: ev.starts_at,
    ends_at: ev.ends_at,
    show_popup: ev.show_popup,
    popup_priority: ev.popup_priority,
    status: ev.status,
  }
  formErrors.value = {}
  showFormModal.value = true
}

async function submitForm() {
  if (isVotingLocked.value && editingEvent.value) {
    form.value.starts_at = editingEvent.value.starts_at
    form.value.ends_at = editingEvent.value.ends_at
  } else {
    form.value.starts_at = localInputToIso(formStartsInput.value)
    form.value.ends_at = localInputToIso(formEndsInput.value)
  }

  const errors = validateEventForm(form.value)
  formErrors.value = errors
  if (Object.keys(errors).length > 0) return

  saving.value = true
  try {
    if (editingEvent.value) {
      await adminEventApi.update(editingEvent.value.id, form.value)
      toast.success(t("admin.events.updated"))
    } else {
      await adminEventApi.create(form.value)
      toast.success(t("admin.events.created"))
    }
    showFormModal.value = false
    void fetchEvents()
  } catch (err) {
    toast.error(t(eventErrorKey(err)))
  } finally {
    saving.value = false
  }
}

async function onFileSelected(event: Event) {
  const target = event.target as HTMLInputElement
  const file = target.files?.[0]
  target.value = ""

  if (!file || !targetUploadEvent.value) return

  if (file.size > EVENT_IMAGE_MAX_BYTES) {
    toast.error(t("admin.events.errors.imageTooLarge"))
    return
  }

  if (!EVENT_IMAGE_TYPES.includes(file.type as typeof EVENT_IMAGE_TYPES[number])) {
    toast.error(t("admin.events.errors.imageType"))
    return
  }

  const evId = targetUploadEvent.value.id
  try {
    await adminEventApi.uploadImage(evId, file)
    toast.success(t("admin.events.form.imageUpdated"))
    void fetchEvents()
  } catch (err) {
    toast.error(t(eventErrorKey(err)))
  } finally {
    targetUploadEvent.value = null
  }
}

async function handleRemoveImage(ev: AdminEvent) {
  try {
    await adminEventApi.removeImage(ev.id)
    toast.success(t("admin.events.form.imageRemoved"))
    void fetchEvents()
  } catch (err) {
    toast.error(t(eventErrorKey(err)))
  }
}

async function confirmDelete() {
  if (!eventToDelete.value) return

  deleting.value = true
  try {
    await adminEventApi.remove(eventToDelete.value.id)
    toast.success(t("admin.events.deleted"))
    showDeleteModal.value = false
    eventToDelete.value = null
    void fetchEvents()
  } catch (err) {
    toast.error(t(eventErrorKey(err)))
  } finally {
    deleting.value = false
  }
}

const showGenerateModal = ref(false)
const generateMediaType = ref<"movie" | "tv">("movie")
const generatePeriod = ref("")
const suggestedNominees = ref<SuggestedNominee[]>([])
const loadingSuggestions = ref(false)
const generating = ref(false)

const showFinalizeModal = ref(false)
const eventToFinalize = ref<AdminEvent | null>(null)
const finalizing = ref(false)

const showNomineesModal = ref(false)
const nomineesEvent = ref<AdminEvent | null>(null)
const editableNominees = ref<NomineeInput[]>([])
const loadingNomineeSuggestions = ref(false)
const savingNominees = ref(false)

/** Placeholder for a blank title input on a monthly round: the copy the server would generate. */
function generatedTitlePlaceholder(lang: "en" | "th"): string {
  if (!editingEvent.value) return ""
  return generatedRoundTitle(editingEvent.value, t, lang) ?? ""
}

function openGenerateModal() {
  generateMediaType.value = "movie"
  generatePeriod.value = nextMonthlyPeriod()
  suggestedNominees.value = []
  showGenerateModal.value = true
}

function openNomineesModal(ev: AdminEvent) {
  nomineesEvent.value = ev
  editableNominees.value = (ev.nominees ?? []).map((n, idx) => ({
    id: n.id,
    media_id: n.media_id,
    media_type: n.media_type,
    position: idx + 1,
    media_title: n.media_title,
    poster_path: n.poster_path ?? "",
    backdrop_path: n.backdrop_path ?? "",
    overview: n.overview ?? "",
  }))
  showNomineesModal.value = true
}

function moveNominee(idx: number, delta: -1 | 1) {
  const target = idx + delta
  if (target < 0 || target >= editableNominees.value.length) return
  const list = [...editableNominees.value]
  const [item] = list.splice(idx, 1)
  if (!item) return
  list.splice(target, 0, item)
  editableNominees.value = list.map((n, i) => ({ ...n, position: i + 1 }))
}

function removeNomineeAt(idx: number) {
  editableNominees.value = editableNominees.value
    .filter((_, i) => i !== idx)
    .map((n, i) => ({ ...n, position: i + 1 }))
}

async function loadNomineeSuggestionsForRound() {
  if (!nomineesEvent.value) return
  loadingNomineeSuggestions.value = true
  try {
    const mediaType = nomineesEvent.value.media_type ?? "movie"
    const period = nomineesEvent.value.period ?? nextMonthlyPeriod()
    const res = await adminEventApi.suggestNominees(mediaType, period)
    editableNominees.value = res.data.nominees.map((s, idx) => ({
      media_id: s.media_id,
      media_type: s.media_type,
      position: idx + 1,
      media_title: s.media_title,
      poster_path: s.poster_path,
      backdrop_path: s.backdrop_path,
      overview: s.overview,
    }))
  } catch (err) {
    toast.error(t(eventErrorKey(err)))
  } finally {
    loadingNomineeSuggestions.value = false
  }
}

async function saveNomineesForRound() {
  if (!nomineesEvent.value) return
  savingNominees.value = true
  try {
    await adminEventApi.setNominees(nomineesEvent.value.id, editableNominees.value)
    toast.success(t("admin.events.nomineesSaved"))
    showNomineesModal.value = false
    void fetchEvents()
  } catch (err) {
    toast.error(t(eventErrorKey(err)))
  } finally {
    savingNominees.value = false
  }
}

async function loadSuggestions() {
  loadingSuggestions.value = true
  try {
    const res = await adminEventApi.suggestNominees(generateMediaType.value, generatePeriod.value)
    suggestedNominees.value = res.data.nominees
  } catch (err) {
    toast.error(t(eventErrorKey(err)))
  } finally {
    loadingSuggestions.value = false
  }
}

async function handleGenerateMonthly() {
  generating.value = true
  try {
    await adminEventApi.generateMonthly(generateMediaType.value, generatePeriod.value)
    toast.success(t("admin.events.generateSuccess"))
    showGenerateModal.value = false
    void fetchEvents()
  } catch (err) {
    toast.error(t(eventErrorKey(err)))
  } finally {
    generating.value = false
  }
}

async function confirmFinalize() {
  if (!eventToFinalize.value) return
  finalizing.value = true
  try {
    await adminEventApi.finalize(eventToFinalize.value.id)
    toast.success(t("admin.events.finalized"))
    showFinalizeModal.value = false
    eventToFinalize.value = null
    void fetchEvents()
  } catch (err) {
    toast.error(t(eventErrorKey(err)))
  } finally {
    finalizing.value = false
  }
}

function formatSource(source: string): string {
  switch (source) {
    case "community":
      return t("admin.events.sourceCommunity")
    case "release":
      return t("admin.events.sourceRelease")
    case "popular":
      return t("admin.events.sourcePopular")
    default:
      return t("admin.events.sourceFallback")
  }
}

onMounted(() => {
  void fetchEvents()
})
</script>

<style scoped>
.event-title-cell {
  display: flex;
  align-items: center;
  gap: 0.75rem;
}

.event-thumb {
  width: 2.75rem;
  height: 2rem;
  border-radius: var(--radius-sm);
  background-color: var(--color-surface-3);
  overflow: hidden;
  flex-shrink: 0;
}

.event-thumb__img {
  width: 100%;
  height: 100%;
  object-fit: cover;
  display: block;
}

.event-thumb__empty {
  display: flex;
  align-items: center;
  justify-content: center;
  width: 100%;
  height: 100%;
  color: var(--color-text-muted);
}

.event-title-info {
  display: flex;
  flex-direction: column;
  min-width: 0;
}

.event-title-primary {
  font-size: var(--text-small);
  font-weight: 600;
  color: var(--color-text-primary);
}

.inline-id {
  margin-left: 0.25rem;
  font-size: 0.6875rem;
  color: var(--color-text-muted);
}

.event-title-sub {
  font-size: 0.6875rem;
  color: var(--color-text-muted);
}

.cell-dates {
  font-size: var(--text-small);
  color: var(--color-text-secondary);
  white-space: nowrap;
}

.cell-dates__sep {
  margin: 0 0.375rem;
  color: var(--color-text-muted);
}

.cell-priority {
  font-variant-numeric: tabular-nums;
  font-size: var(--text-small);
  color: var(--color-text-secondary);
}

.event-form {
  display: flex;
  flex-direction: column;
  gap: 1rem;
}

.form-row {
  display: grid;
  grid-template-columns: 1fr;
  gap: 1rem;
}

@media (min-width: 640px) {
  .form-row {
    grid-template-columns: repeat(2, 1fr);
  }
}

.form-field {
  display: flex;
  flex-direction: column;
  gap: 0.375rem;
}

.form-label {
  font-size: var(--text-small);
  font-weight: 600;
  color: var(--color-text-primary);
}

.form-input,
.form-textarea {
  width: 100%;
  padding: 0.5rem 0.75rem;
  background-color: var(--color-surface-2);
  border: 1px solid var(--color-border-strong);
  border-radius: var(--radius-md);
  font-size: var(--text-small);
  color: var(--color-text-primary);
  font-family: inherit;
}

.form-input:focus,
.form-textarea:focus {
  outline: none;
  border-color: var(--color-focus, var(--color-brand));
}

.form-input.is-invalid,
.form-textarea.is-invalid {
  border-color: var(--color-danger);
}

.form-textarea {
  resize: vertical;
}

.form-error {
  font-size: 0.6875rem;
  color: var(--color-danger);
}

.form-field--checkbox {
  padding-top: 0.25rem;
}

.form-checkbox-label {
  display: flex;
  align-items: center;
  gap: 0.5rem;
  font-size: var(--text-small);
  color: var(--color-text-primary);
  cursor: pointer;
  user-select: none;
}

.form-checkbox {
  accent-color: var(--color-brand);
  cursor: pointer;
}

.form-actions {
  display: flex;
  justify-content: flex-end;
  gap: 0.75rem;
  margin-top: 1rem;
  padding-top: 1rem;
  border-top: 1px solid var(--color-border);
}

.admin-header-actions {
  display: flex;
  align-items: center;
  gap: 0.5rem;
}

.event-type-tag {
  display: inline-flex;
  padding: 0.125rem 0.375rem;
  font-size: 0.6875rem;
  font-weight: 500;
  border-radius: var(--radius-sm);
  background-color: var(--color-surface-3);
  color: var(--color-text-secondary);
  margin-left: 0.375rem;
}

.generate-radio-group {
  display: flex;
  gap: 1.5rem;
  padding-top: 0.25rem;
}

.generate-radio {
  display: flex;
  align-items: center;
  gap: 0.5rem;
  font-size: var(--text-small);
  color: var(--color-text-primary);
  cursor: pointer;
}

.generate-radio input {
  accent-color: var(--color-brand);
}

.generate-preview-action {
  margin: 0.5rem 0 1rem;
}

.suggested-list {
  display: flex;
  flex-direction: column;
  gap: 0.5rem;
  margin-bottom: 1rem;
  max-height: 220px;
  overflow-y: auto;
  border: 1px solid var(--color-border);
  border-radius: var(--radius-sm);
  padding: 0.5rem;
  background-color: var(--color-surface);
}

.suggested-item {
  display: flex;
  align-items: center;
  gap: 0.75rem;
  padding: 0.5rem;
  background-color: var(--color-surface-2);
  border-radius: var(--radius-sm);
}

.suggested-item__idx {
  font-weight: 700;
  font-size: 0.75rem;
  color: var(--color-text-muted);
}

.suggested-item__info {
  display: flex;
  flex: 1;
  align-items: center;
  justify-content: space-between;
  gap: 0.5rem;
}

.suggested-item__title {
  font-size: 0.8125rem;
  font-weight: 600;
  color: var(--color-text-primary);
}

.suggested-item__source {
  font-size: 0.6875rem;
  color: var(--color-brand);
  background-color: color-mix(in srgb, var(--color-brand) 12%, transparent);
  padding: 0.125rem 0.375rem;
  border-radius: var(--radius-sm);
}

.suggested-item__actions {
  display: flex;
  align-items: center;
  gap: 0.25rem;
}

.form-hint {
  font-size: 0.75rem;
  color: var(--color-text-muted);
  margin: 0;
}
</style>
