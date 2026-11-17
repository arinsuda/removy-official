<template>
  <main class="manga-detail">
    <RouterLink :to="{ name: 'manga' }" class="manga-detail__back">{{ t("manga.detail.back") }}</RouterLink>

    <SkeletonRegion v-if="detail.isLoading.value" class="manga-detail__loading">
      <SkeletonBlock height="12rem" radius="12px" />
      <SkeletonBlock width="60%" height="1.5rem" />
      <SkeletonBlock height="6rem" />
    </SkeletonRegion>

    <MangaAgeRestricted v-else-if="restricted" />

    <AppErrorState v-else-if="detail.isError.value || !manga" :title="t('manga.detail.error')" @retry="detail.refetch()" />

    <template v-else-if="manga && !adultGate">
      <MangaDetailHero :manga="manga" />
      <MangaActionBar
        :manga="manga"
        :user-id="auth.user?.id ?? null"
        :has-own-review="Boolean(myReview)"
        :activity="activity"
        @rate="onRate"
        @share="shareOpen = true"
      />
      <!-- Owner-only: the position belongs to the signed-in viewer, so guests never see the section. -->
      <MangaReadingProgress v-if="auth.user" :key="manga.id" :manga="manga" :activity="activity" />
      <MangaFacts :manga="manga" />
      <MangaTagList v-if="manga.genres.length" :label="t('manga.detail.genres')" :tags="manga.genres" />
      <MangaTagList v-if="manga.tags.length" :label="t('manga.detail.tags')" :tags="tagNames" />

      <section class="manga-detail__section" :aria-labelledby="descriptionId">
        <h2 :id="descriptionId" class="manga-detail__heading">{{ t("manga.detail.description") }}</h2>
        <p v-if="description" class="manga-detail__description">{{ description }}</p>
        <p v-else class="manga-detail__muted">{{ t("manga.detail.noDescription") }}</p>
      </section>

      <MangaReadLinks :links="manga.read_links" />
      <section id="reviews" class="manga-detail__section" :aria-labelledby="reviewsId">
        <h2 :id="reviewsId" class="manga-detail__heading">{{ t("manga.detail.reviews") }}</h2>
        <MovieReviews
          ref="reviews"
          :movie-id="manga.id"
          media-type="manga"
          :media-poster-path="manga.cover_url"
          :media-title="manga.display_title"
          :media-is-adult="manga.is_adult"
          :media-start-year="manga.start_year"
          :media-genre-names="manga.genres"
          @loaded="(list) => (reviewList = list)"
          @review-submitted="activity.refreshStats"
        />
      </section>
      <MangaSimilarRail :items="similar.data.value ?? []" />
      <MangaAttribution />
      <ShareImageModal v-if="shareMedia" v-model="shareOpen" share-type="poster" :media="shareMedia" />
    </template>

    <!-- Adult titles open only after the viewer confirms; leaving the modal without confirming goes back to the list. -->
    <MangaAdultConfirmModal v-if="manga?.is_adult" :model-value="adultGate" @confirm="acknowledge" @update:model-value="onAdultDialog" />
  </main>
</template>

<script setup lang="ts">
import { computed, ref, useId, useTemplateRef } from "vue"
import { useI18n } from "vue-i18n"
import { useRoute, useRouter } from "vue-router"
import MovieReviews from "@/components/movie/MovieReviews.vue"
import MangaActionBar from "@/components/manga/MangaActionBar.vue"
import MangaAdultConfirmModal from "@/components/manga/MangaAdultConfirmModal.vue"
import MangaAgeRestricted from "@/components/manga/MangaAgeRestricted.vue"
import MangaAttribution from "@/components/manga/MangaAttribution.vue"
import MangaDetailHero from "@/components/manga/MangaDetailHero.vue"
import MangaFacts from "@/components/manga/MangaFacts.vue"
import MangaReadingProgress from "@/components/manga/MangaReadingProgress.vue"
import MangaReadLinks from "@/components/manga/MangaReadLinks.vue"
import MangaSimilarRail from "@/components/manga/MangaSimilarRail.vue"
import MangaTagList from "@/components/manga/MangaTagList.vue"
import ShareImageModal from "@/components/share/ShareImageModal.vue"
import { isAgeRestricted } from "@/components/manga/mangaAdult"
import { plainDescription } from "@/components/manga/mangaFormat"
import AppErrorState from "@/components/ui/AppErrorState.vue"
import SkeletonBlock from "@/components/ui/skeleton/SkeletonBlock.vue"
import SkeletonRegion from "@/components/ui/skeleton/SkeletonRegion.vue"
import { useMangaAdult } from "@/composables/useMangaAdult"
import { useMangaDetail } from "@/composables/useMangaDetail"
import { useMediaActivity } from "@/composables/useMediaActivity"
import type { ReviewResponse } from "@/types/review"
import type { ShareMediaContext } from "@/types/share"
import { useAuthStore } from "@/stores/auth"

const { t } = useI18n()
const route = useRoute()
const router = useRouter()
const id = computed(() => Number(route.params.id))
const { acknowledged, canShowAdult, acknowledge } = useMangaAdult()
const { detail, similar } = useMangaDetail(id, canShowAdult)

const manga = computed(() => detail.data.value ?? null)
const restricted = computed(() => isAgeRestricted(detail.error.value))
const adultGate = computed(() => !!manga.value?.is_adult && !acknowledged.value)
const description = computed(() => plainDescription(manga.value?.description ?? ""))
const tagNames = computed(() => (manga.value?.tags ?? []).map((tag) => tag.name))
const descriptionId = "manga-description-heading"
const reviewsId = useId()

const auth = useAuthStore()
const activity = useMediaActivity("manga", computed(() => manga.value?.id ?? 0))

// Reviews: the rate button scrolls to the section and focuses its composer.
const reviewList = ref<ReviewResponse[]>([])
const reviewsRef = useTemplateRef<InstanceType<typeof MovieReviews>>("reviews")
const myReview = computed(() => reviewList.value.find((r) => r.user.id === auth.user?.id) ?? null)

// Share as image: the cover is the poster; adult covers stay hidden in the modal until the viewer may see them.
const shareOpen = ref(false)
const shareMedia = computed<ShareMediaContext | null>(() => {
  const m = manga.value
  if (!m) return null
  return {
    id: m.id,
    mediaType: "manga",
    title: m.display_title,
    posterPath: null,
    posterUrl: m.cover_url,
    releaseYear: m.start_year ? String(m.start_year) : null,
    genres: [],
    genreNames: m.genres,
    voteAverage: 0,
    popcornRating: m.ratings.popcorn?.average ?? null,
    popcornCount: m.ratings.popcorn?.count ?? null,
    isAdult: m.is_adult,
  }
})

function onRate() {
  document.getElementById("reviews")?.scrollIntoView({
    behavior: window.matchMedia("(prefers-reduced-motion: reduce)").matches ? "auto" : "smooth",
    block: "start",
  })
  reviewsRef.value?.focusComposer()
}

/** Closing the 18+ dialog without confirming leaves the adult title. */
function onAdultDialog(open: boolean) {
  if (!open && !acknowledged.value) void router.push({ name: "manga" })
}
</script>

<style scoped>
.manga-detail {
  display: flex;
  flex-direction: column;
  gap: 1.25rem;
  width: 100%;
  max-width: 960px;
  margin: 0 auto;
  padding: 1rem 1rem 3rem;
}
.manga-detail__back {
  align-self: flex-start;
  font-size: var(--text-small);
  color: var(--color-text-secondary);
  text-decoration: none;
}
.manga-detail__loading {
  display: flex;
  flex-direction: column;
  gap: 0.75rem;
}
.manga-detail__section {
  display: flex;
  flex-direction: column;
  gap: 0.5rem;
}
.manga-detail__heading {
  margin: 0;
  font-size: var(--text-body);
  font-weight: 600;
  color: var(--color-text-primary);
}
.manga-detail__description {
  margin: 0;
  font-size: var(--text-small);
  line-height: 1.7;
  white-space: pre-line;
  color: var(--color-text-secondary);
  overflow-wrap: anywhere;
}
.manga-detail__muted {
  margin: 0;
  font-size: var(--text-small);
  color: var(--color-text-muted);
}
@media (min-width: 768px) {
  .manga-detail {
    padding: 2rem 1.5rem 4rem;
  }
}
</style>
