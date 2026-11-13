<template>
  <AdminPage>
    <AdminPageHeader
      :title="t('admin.events.title')"
      :subtitle="t('admin.events.subtitle')"
    >
      <AppButton variant="primary" size="sm" @click="openCreateModal">
        <Plus :size="16" aria-hidden="true" />
        {{ t("admin.events.newEvent") }}
      </AppButton>
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
                {{ localizedText(ev.title, locale) }}
                <span class="inline-id">#{{ ev.id }}</span>
              </span>
              <span class="event-title-sub">
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
            :label="t('admin.table.rowActions', { name: localizedText(ev.title, locale) })"
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
              :placeholder="t('admin.events.form.titleEnPlaceholder')"
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
              :placeholder="t('admin.events.form.titleThPlaceholder')"
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
              required
            />
            <span v-if="formErrors.ends_at" class="form-error">
              {{ t(formErrors.ends_at) }}
            </span>
          </div>
        </div>

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
      :target-name="eventToDelete ? localizedText(eventToDelete.title, locale) : undefined"
      :loading="deleting"
      @confirm="confirmDelete"
      @cancel="showDeleteModal = false"
    />
  </AdminPage>
</template>

<script setup lang="ts">
import { computed, onMounted, ref, watch } from "vue"
import { useI18n } from "vue-i18n"
import { useRouter } from "vue-router"
import {
  Eye,
  Image as ImageIcon,
  Megaphone,
  Pencil,
  Plus,
  Trash2,
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
} from "@/types/event"
import {
  ADMIN_EVENTS_PAGE_SIZE,
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
  localizedText,
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

function openEditModal(ev: AdminEvent) {
  editingEvent.value = ev
  formStartsInput.value = isoToLocalInput(ev.starts_at)
  formEndsInput.value = isoToLocalInput(ev.ends_at)

  form.value = {
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
  form.value.starts_at = localInputToIso(formStartsInput.value)
  form.value.ends_at = localInputToIso(formEndsInput.value)

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
</style>
