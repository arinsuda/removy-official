<template>
  <AdminPage>
    <AdminPageHeader
      :title="t('admin.events.title')"
      :subtitle="t('admin.events.subtitle')"
    >
      <div class="admin-header-actions">
        <AppButton variant="outline" size="sm" @click="openGenerateModal">
          <Sparkles :size="16" aria-hidden="true" />
          {{ t("admin.events.generateRound") }}
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
    <p v-if="hasAutoDrafts" class="auto-hint">{{ t("admin.events.autoHint") }}</p>

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
                <AdminBadge v-if="ev.auto_generated" tone="info" :title="t('admin.events.autoHint')">
                  {{ t("admin.events.autoChip") }}
                </AdminBadge>
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

    <AdminEventFormModal
      v-model="showFormModal"
      v-model:form="form"
      v-model:starts-input="formStartsInput"
      v-model:ends-input="formEndsInput"
      :editing="editingEvent"
      :errors="formErrors"
      :saving="saving"
      :locked="isVotingLocked"
      @submit="submitForm"
    />

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

    <AdminGenerateRoundModal v-model="showGenerateModal" @generated="fetchEvents" />

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

    <AdminNomineesModal
      v-model="showNomineesModal"
      :event="nomineesEvent"
      @saved="fetchEvents"
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
  Sparkles,
  Trash2,
  Trophy,
} from "lucide-vue-next"
import AdminBadge from "@/components/admin/AdminBadge.vue"
import AdminConfirmModal from "@/components/admin/AdminConfirmModal.vue"
import AdminEventFormModal from "@/components/admin/events/AdminEventFormModal.vue"
import AdminFilters from "@/components/admin/AdminFilters.vue"
import AdminGenerateRoundModal from "@/components/admin/events/AdminGenerateRoundModal.vue"
import AdminNomineesModal from "@/components/admin/events/AdminNomineesModal.vue"
import AdminPage from "@/components/admin/AdminPage.vue"
import AdminPageHeader from "@/components/admin/AdminPageHeader.vue"
import AdminPagination from "@/components/admin/AdminPagination.vue"
import AdminRowMenu, { type RowMenuItem } from "@/components/admin/AdminRowMenu.vue"
import AdminSelect from "@/components/admin/AdminSelect.vue"
import AdminTable from "@/components/admin/AdminTable.vue"
import AppButton from "@/components/ui/AppButton.vue"
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
  type EventFormErrors,
  EVENT_MIN_NOMINEES_TO_PUBLISH,
  eventErrorKey,
  eventTitle,
  isoToLocalInput,
  localInputToIso,
  publishErrorKey,
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

const showGenerateModal = ref(false)

const showNomineesModal = ref(false)
const nomineesEvent = ref<AdminEvent | null>(null)

const showFinalizeModal = ref(false)
const eventToFinalize = ref<AdminEvent | null>(null)
const finalizing = ref(false)

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

/** Auto drafts publish on their own, so the page explains that while any are listed. */
const hasAutoDrafts = computed(() => events.value.some((ev) => ev.auto_generated && ev.status === "draft"))

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

// Filter and page can change faster than responses arrive; only the newest request may write state.
let latestRequest = 0

async function fetchEvents() {
  const request = ++latestRequest
  loading.value = true
  error.value = false
  try {
    const res = await adminEventApi.list({
      status: (statusFilter.value as EventStatus) || undefined,
      page: page.value,
      limit: ADMIN_EVENTS_PAGE_SIZE,
    })
    if (request !== latestRequest) return
    events.value = res.data.events
    pagination.value = res.data.pagination
  } catch {
    if (request === latestRequest) error.value = true
  } finally {
    if (request === latestRequest) loading.value = false
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
    toast.error(t(publishErrorKey(err), { min: EVENT_MIN_NOMINEES_TO_PUBLISH }))
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

function openGenerateModal() {
  showGenerateModal.value = true
}

function openNomineesModal(ev: AdminEvent) {
  nomineesEvent.value = ev
  showNomineesModal.value = true
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


.admin-header-actions {
  display: flex;
  align-items: center;
  gap: 0.5rem;
}

.auto-hint {
  font-size: var(--text-caption);
  color: var(--color-text-muted);
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
</style>
