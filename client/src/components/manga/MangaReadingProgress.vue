<template>
  <section class="manga-progress" :aria-labelledby="headingId" data-test="manga-progress">
    <div class="manga-progress__head">
      <h2 :id="headingId" class="manga-progress__heading">{{ t("manga.progress.title") }}</h2>
      <div v-if="isLoaded" class="manga-progress__tools">
        <AppButton variant="outline" size="sm" data-test="manga-progress-edit" @click="editOpen = true">
          <Pencil :size="16" aria-hidden="true" />
          <span>{{ t("manga.progress.edit") }}</span>
        </AppButton>
        <AppButton v-if="progress" variant="ghost" size="sm" data-test="manga-progress-clear" @click="clearProgress">
          <Eraser :size="16" aria-hidden="true" />
          <span>{{ t("manga.progress.clear") }}</span>
        </AppButton>
      </div>
    </div>

    <SkeletonRegion v-if="isPending">
      <SkeletonBlock height="3rem" />
    </SkeletonRegion>

    <AppErrorState v-else-if="loadFailed" :title="t('manga.progress.errors.loadFailed')" @retry="query.refetch()" />

    <template v-else>
      <p v-if="!progress" class="manga-progress__muted">{{ t("manga.progress.notStarted") }}</p>

      <div class="manga-progress__steppers">
        <MangaProgressStepper
          :label="t('manga.progress.chapter')"
          :value="values.chapter"
          :decrease-label="t('manga.progress.decrease', { field: t('manga.progress.chapter') })"
          :increase-label="t('manga.progress.increase', { field: t('manga.progress.chapter') })"
          :can-decrement="canStep('chapter', -1)"
          :can-increment="canStep('chapter', 1)"
          @step="(delta) => step('chapter', delta)"
        />
        <MangaProgressStepper
          :label="t('manga.progress.volume')"
          :value="values.volume"
          :decrease-label="t('manga.progress.decrease', { field: t('manga.progress.volume') })"
          :increase-label="t('manga.progress.increase', { field: t('manga.progress.volume') })"
          :can-decrement="canStep('volume', -1)"
          :can-increment="canStep('volume', 1)"
          @step="(delta) => step('volume', delta)"
        />
      </div>

      <div
        v-if="chapterTotal"
        class="manga-progress__bar"
        role="progressbar"
        :aria-label="t('manga.progress.chapter')"
        aria-valuemin="0"
        :aria-valuemax="chapterTotal"
        :aria-valuenow="Math.min(values.chapter, chapterTotal)"
        :aria-valuetext="t('manga.progress.chaptersRead', { count: values.chapter })"
      >
        <span class="manga-progress__fill" :style="{ width: `${chapterFraction * 100}%` }" />
      </div>

      <div v-if="markReadVisible" class="manga-progress__finish">
        <p class="manga-progress__muted">{{ t("manga.progress.finishedHint") }}</p>
        <AppButton
          variant="outline"
          size="sm"
          data-test="manga-progress-mark-read"
          :loading="activity.pending.value.has('watched')"
          @click="markRead"
        >
          <BookCheck :size="16" aria-hidden="true" />
          <span>{{ t("manga.progress.markRead") }}</span>
        </AppButton>
      </div>
    </template>

    <MangaProgressEditModal v-model="editOpen" :values="values" @save="commitProgress" />
  </section>
</template>

<script setup lang="ts">
import { computed, ref, useId } from "vue"
import { useI18n } from "vue-i18n"
import { BookCheck, Eraser, Pencil } from "lucide-vue-next"
import AppButton from "@/components/ui/AppButton.vue"
import AppErrorState from "@/components/ui/AppErrorState.vue"
import SkeletonBlock from "@/components/ui/skeleton/SkeletonBlock.vue"
import SkeletonRegion from "@/components/ui/skeleton/SkeletonRegion.vue"
import MangaProgressEditModal from "@/components/manga/MangaProgressEditModal.vue"
import MangaProgressStepper from "@/components/manga/MangaProgressStepper.vue"
import { stepProgress, type ProgressField, type ProgressTotals } from "@/components/manga/mangaProgress"
import { useReadingProgress } from "@/composables/useReadingProgress"
import type { useMediaActivity } from "@/composables/useMediaActivity"
import type { MangaItem } from "@/types/manga"

const props = defineProps<{
  manga: MangaItem
  activity: ReturnType<typeof useMediaActivity>
}>()

const { t } = useI18n()
const headingId = useId()
const editOpen = ref(false)

// Limits follow the title: AniList's counts cap the steppers when it knows them.
const totals = computed<ProgressTotals>(() => ({ chapters: props.manga.chapters, volumes: props.manga.volumes }))
const { query, values, step, commit: commitProgress, clearProgress } = useReadingProgress(props.manga.id, totals)

const progress = computed(() => query.data.value ?? null)
const isPending = computed(() => query.isPending.value)
const loadFailed = computed(() => query.isError.value)
const isLoaded = computed(() => query.isSuccess.value)

const chapterTotal = computed(() => (props.manga.chapters ? props.manga.chapters : null))
const chapterFraction = computed(() =>
  chapterTotal.value ? Math.min(values.value.chapter / chapterTotal.value, 1) : 0,
)
// Only a title with a known chapter count can be finished, so the prompt waits for that.
const markReadVisible = computed(
  () => chapterTotal.value !== null && values.value.chapter >= chapterTotal.value && !props.activity.isWatched.value,
)

function canStep(field: ProgressField, delta: number): boolean {
  return stepProgress(values.value, field, delta, totals.value) !== null
}

/** Mark-as-read is never automatic: the viewer confirms it here. */
function markRead() {
  void props.activity.toggleWatched()
}
</script>

<style scoped>
.manga-progress {
  display: flex;
  flex-direction: column;
  gap: 0.875rem;
  padding: 1rem;
  background: var(--color-surface-2);
  border: 1px solid var(--color-border);
  border-radius: var(--radius-card);
}
.manga-progress__head {
  display: flex;
  flex-wrap: wrap;
  align-items: center;
  justify-content: space-between;
  gap: 0.5rem;
}
.manga-progress__heading {
  margin: 0;
  font-size: var(--text-body);
  font-weight: 600;
  color: var(--color-text-primary);
}
.manga-progress__tools {
  display: flex;
  flex-wrap: wrap;
  gap: 0.5rem;
}
.manga-progress__steppers {
  display: grid;
  grid-template-columns: repeat(2, minmax(0, 1fr));
  gap: 1rem;
}
.manga-progress__bar {
  height: 0.375rem;
  overflow: hidden;
  background: var(--color-surface-3);
  border-radius: var(--radius-full);
}
.manga-progress__fill {
  display: block;
  height: 100%;
  background: var(--color-brand);
  transition: width var(--motion-fast) var(--ease-out-soft);
}
.manga-progress__finish {
  display: flex;
  flex-wrap: wrap;
  align-items: center;
  justify-content: space-between;
  gap: 0.5rem;
}
.manga-progress__muted {
  margin: 0;
  font-size: var(--text-small);
  color: var(--color-text-muted);
}
</style>
