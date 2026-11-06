<template>
  <Teleport to="body">
    <Transition name="slide">
      <div class="panel-backdrop" @click.self="$emit('close')">
        <div class="panel" role="dialog" aria-modal="true" :aria-label="t('watchLog.watchHistory')">
          <div class="panel-header">
            <h2 class="panel-title">{{ t('watchLog.watchHistory') }}</h2>
            <button class="panel-close" :aria-label="t('common.close')" @click="$emit('close')">
              <X class="icon" />
            </button>
          </div>

          <div class="panel-content">
            <SkeletonList v-if="loading" :count="3" />
            <div v-else-if="!logs || !logs.length" class="empty">
              {{ t('watchLog.noHistory') }}
            </div>
            <div v-else class="logs-list">
              <div v-for="log in logs" :key="log.id" class="log-item">
                <div class="log-info">
                  <div class="log-date">{{ formatDate(log.watched_on) }}</div>
                  <div class="log-visibility">
                    <Globe v-if="log.visibility === 'public'" class="vis-icon" />
                    <Users v-else-if="log.visibility === 'followers'" class="vis-icon" />
                    <Lock v-else class="vis-icon" />
                    {{ getVisibilityText(log.visibility) }}
                  </div>
                </div>
                <button class="btn-delete" :aria-label="t('watchLog.deleteWatch')" :title="t('watchLog.deleteWatch')" @click="handleDelete(log.id)">
                  <Trash2 class="icon" />
                </button>
              </div>
            </div>
          </div>
        </div>
      </div>
    </Transition>
  </Teleport>
</template>

<script setup lang="ts">
import SkeletonList from "@/components/ui/skeleton/SkeletonList.vue"
import { ref, onMounted, onBeforeUnmount } from 'vue'
import { useI18n } from 'vue-i18n'
import { X, Trash2, Globe, Users, Lock } from 'lucide-vue-next'
import { watchLogApi } from '@/api/endpoints/watchLog'
import type { WatchLogResponse, Visibility } from '@/types'
import { useToast } from '@/composables/useToast'

const props = defineProps<{
  mediaId: number
  mediaType: "movie" | "tv"
}>()

const emit = defineEmits<{
  close: []
  changed: []
}>()

const { t, locale } = useI18n()
const { success: toastSuccess, error: toastError } = useToast()

const logs = ref<WatchLogResponse[]>([])
const loading = ref(true)

function getVisibilityText(visibility: Visibility): string {
  if (visibility === 'public') return t('visibility.public')
  if (visibility === 'followers') return t('visibility.followers')
  return t('visibility.private')
}

async function loadHistory() {
  try {
    const res = await watchLogApi.getMyWatchLogs(props.mediaType, props.mediaId)
    logs.value = res.data?.logs || []
  } catch (err) {
    console.error('Failed to load watch history:', err)
    toastError(t('watchLog.loadFailed'))
  } finally {
    loading.value = false
  }
}

async function handleDelete(logId: number) {
  if (!confirm(t('watchLog.deleteConfirm'))) return
  try {
    await watchLogApi.deleteWatchLog(logId)
    logs.value = logs.value.filter(l => l.id !== logId)
    toastSuccess(t('watchLog.deleteSuccess'))
    emit('changed')
  } catch (err) {
    console.error('Failed to delete watch log:', err)
    toastError(t('watchLog.deleteFailed'))
  }
}

function formatDate(dateStr: string) {
  if (!dateStr) return ''
  const d = new Date(dateStr)
  return d.toLocaleDateString(locale.value === 'th' ? 'th-TH' : 'en-US', {
    year: 'numeric',
    month: 'short',
    day: 'numeric'
  })
}

onMounted(() => {
  document.body.style.overflow = 'hidden'
  loadHistory()
})
onBeforeUnmount(() => { document.body.style.overflow = '' })
</script>

<style scoped>
.panel-backdrop {
  position: fixed;
  inset: 0;
  background: rgba(0, 0, 0, 0.5);
  display: flex;
  justify-content: flex-end;
  z-index: 200;
}
.panel {
  width: 100%;
  max-width: 400px;
  background: #111;
  border-left: 1px solid rgba(255, 255, 255, 0.08);
  height: 100vh;
  display: flex;
  flex-direction: column;
}
.panel-header {
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding: 20px;
  border-bottom: 1px solid rgba(255, 255, 255, 0.08);
}
.panel-title {
  font-size: 1.1rem;
  font-weight: 600;
  color: #fff;
  margin: 0;
}
.panel-close {
  background: transparent;
  border: none;
  color: #8a8a8e;
  cursor: pointer;
}
.panel-close:hover {
  color: #fff;
}
.icon {
  width: 1.25rem;
  height: 1.25rem;
}
.panel-content {
  flex: 1;
  overflow-y: auto;
  padding: 20px;
}
.loading, .empty {
  color: #8a8a8e;
  text-align: center;
  margin-top: 2rem;
}
.logs-list {
  display: flex;
  flex-direction: column;
  gap: 12px;
}
.log-item {
  display: flex;
  align-items: center;
  justify-content: space-between;
  background: rgba(255, 255, 255, 0.03);
  border: 1px solid rgba(255, 255, 255, 0.08);
  padding: 12px 16px;
  border-radius: 8px;
}
.log-info {
  display: flex;
  flex-direction: column;
  gap: 4px;
}
.log-date {
  color: #fff;
  font-weight: 500;
  font-size: 0.9rem;
}
.log-visibility {
  display: flex;
  align-items: center;
  gap: 4px;
  color: #8a8a8e;
  font-size: 0.75rem;
  text-transform: capitalize;
}
.vis-icon {
  width: 0.85rem;
  height: 0.85rem;
}
.btn-delete {
  background: transparent;
  border: none;
  color: #ef4444;
  cursor: pointer;
  padding: 4px;
  border-radius: 4px;
}
.btn-delete:hover {
  background: rgba(239, 68, 68, 0.1);
}
.slide-enter-active, .slide-leave-active {
  transition: transform 0.3s ease;
}
.slide-enter-from, .slide-leave-to {
  transform: translateX(100%);
}
</style>
