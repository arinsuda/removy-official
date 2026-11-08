<template>
  <div class="actor-detail-page">
    <!-- Navigation Back Button -->
    <button type="button" class="btn-back" @click="goBack">
      <i class="pi pi-arrow-left" />
      <span>{{ $t("actor.back") }}</span>
    </button>

    <!-- Loading State -->
    <SkeletonProfile v-if="isLoading" />

    <!-- 404 Not Found State -->
    <div v-else-if="isNotFound" class="state-container error-state">
      <div class="error-icon"><i class="pi pi-user-minus" /></div>
      <h2>{{ $t("actor.notFoundTitle") }}</h2>
      <p>{{ $t("actor.notFoundDetail") }}</p>
      <RouterLink to="/" class="btn-primary">
        {{ $t("actor.backToHome") }}
      </RouterLink>
    </div>

    <!-- API Error State -->
    <div v-else-if="isError" class="state-container error-state">
      <div class="error-icon"><i class="pi pi-exclamation-triangle" /></div>
      <h2>{{ $t("actor.errorTitle") }}</h2>
      <p>{{ $t("actor.errorDetail") }}</p>
      <button type="button" class="btn-primary" @click="fetchData">
        {{ $t("actor.retry") }}
      </button>
    </div>

    <!-- Loaded Content -->
    <div v-else-if="actor" class="actor-content">
      <!-- Actor Profile Header -->
      <section class="actor-header-card">
        <div class="avatar-column">
          <div class="avatar-wrap">
            <img
              v-if="actor.profile_path"
              :src="`https://image.tmdb.org/t/p/w500${actor.profile_path}`"
              :alt="actor.name"
              class="actor-avatar"
            />
            <div v-else class="avatar-placeholder">
              <i class="pi pi-user" />
            </div>
          </div>
        </div>

        <div class="info-column">
          <h1 class="actor-name">{{ actor.name }}</h1>

          <div class="meta-pills">
            <span v-if="actor.known_for_department" class="meta-pill department">
              <i class="pi pi-briefcase" />
              {{ actor.known_for_department }}
            </span>

            <span v-if="birthMeta" class="meta-pill birth">
              <i class="pi pi-calendar" />
              {{ birthMeta }}
            </span>

            <span v-if="actor.place_of_birth" class="meta-pill birthplace">
              <i class="pi pi-map-marker" />
              {{ actor.place_of_birth }}
            </span>

            <a
              v-if="actor.imdb_id"
              :href="`https://www.imdb.com/name/${actor.imdb_id}`"
              target="_blank"
              rel="noopener noreferrer"
              class="meta-pill imdb-link"
              aria-label="IMDb Profile"
            >
              <i class="pi pi-external-link" />
              IMDb
            </a>
          </div>

          <!-- Biography Section -->
          <div class="bio-section">
            <h2 class="bio-title"><i class="pi pi-align-left" /> {{ $t("actor.biography") }}</h2>
            <p v-if="actor.biography" class="bio-text" :class="{ collapsed: isBioCollapsed }">
              {{ actor.biography }}
            </p>
            <p v-else class="bio-empty">
              {{ $t("actor.noBiography") }}
            </p>

            <button
              v-if="actor.biography && actor.biography.length > 300"
              type="button"
              class="btn-toggle-bio"
              @click="isBioCollapsed = !isBioCollapsed"
            >
              {{ isBioCollapsed ? $t("actor.showMore") : $t("actor.showLess") }}
              <i :class="isBioCollapsed ? 'pi pi-chevron-down' : 'pi pi-chevron-up'" />
            </button>
          </div>
        </div>
      </section>

      <!-- Workography Section -->
      <section class="workography-section">
        <div class="workography-header">
          <h2 class="section-title">
            <i class="pi pi-film" /> {{ $t("actor.workography") }}
          </h2>

          <div class="tab-controls" role="tablist">
            <button
              type="button"
              role="tab"
              :aria-selected="activeTab === 'movies'"
              class="tab-btn"
              :class="{ active: activeTab === 'movies' }"
              @click="activeTab = 'movies'"
            >
              {{ $t("actor.moviesTab") }}
              <span class="tab-badge">{{ processedMovies.length }}</span>
            </button>

            <button
              type="button"
              role="tab"
              :aria-selected="activeTab === 'tv'"
              class="tab-btn"
              :class="{ active: activeTab === 'tv' }"
              @click="activeTab = 'tv'"
            >
              {{ $t("actor.tvTab") }}
              <span class="tab-badge">{{ processedTVSeries.length }}</span>
            </button>
          </div>
        </div>

        <!-- Movies Tab Grid -->
        <div v-if="activeTab === 'movies'" class="media-grid-container">
          <div v-if="processedMovies.length === 0" class="empty-credits">
            <i class="pi pi-inbox" />
            <p>{{ $t("actor.noMovieCredits") }}</p>
          </div>

          <div v-else class="media-grid">
            <RouterLink
              v-for="movie in processedMovies"
              :key="movie.id"
              :to="`/movies/${movie.id}`"
              class="media-card"
              :aria-label="movie.title"
            >
              <div class="poster-wrap">
                <img
                  v-if="movie.poster_path"
                  :src="`https://image.tmdb.org/t/p/w342${movie.poster_path}`"
                  :alt="movie.title"
                  class="poster-img"
                  loading="lazy"
                />
                <div v-else class="poster-placeholder">
                  <i class="pi pi-image" />
                </div>
                <div class="rating-badge">
                  <MediaRating :ratings="movie.ratings" :vote-average="movie.vote_average" />
                </div>
              </div>

              <div class="card-info">
                <h3 class="media-title">{{ movie.title }}</h3>
                <p class="character-role">
                  <i class="pi pi-user" />
                  <span>{{ formatCharacter(movie.character) }}</span>
                </p>
                <p class="release-year">{{ formatYear(movie.release_date) }}</p>
              </div>
            </RouterLink>
          </div>
        </div>

        <!-- TV Series Tab Grid -->
        <div v-if="activeTab === 'tv'" class="media-grid-container">
          <div v-if="processedTVSeries.length === 0" class="empty-credits">
            <i class="pi pi-inbox" />
            <p>{{ $t("actor.noTVCredits") }}</p>
          </div>

          <div v-else class="media-grid">
            <RouterLink
              v-for="show in processedTVSeries"
              :key="show.id"
              :to="`/tv/${show.id}`"
              class="media-card"
              :aria-label="show.name"
            >
              <div class="poster-wrap">
                <img
                  v-if="show.poster_path"
                  :src="`https://image.tmdb.org/t/p/w342${show.poster_path}`"
                  :alt="show.name"
                  class="poster-img"
                  loading="lazy"
                />
                <div v-else class="poster-placeholder">
                  <i class="pi pi-image" />
                </div>
                <div class="rating-badge">
                  <MediaRating :ratings="show.ratings" :vote-average="show.vote_average" />
                </div>
              </div>

              <div class="card-info">
                <h3 class="media-title">{{ show.name }}</h3>
                <p class="character-role">
                  <i class="pi pi-user" />
                  <span>{{ formatCharacter(show.character) }}</span>
                </p>
                <p class="release-year">{{ formatYear(show.first_air_date) }}</p>
              </div>
            </RouterLink>
          </div>
        </div>
      </section>
    </div>
  </div>
</template>

<script setup lang="ts">
import SkeletonProfile from "@/components/ui/skeleton/SkeletonProfile.vue"
  import { ref, computed, onMounted, watch } from "vue"
  import { useRoute, useRouter, RouterLink } from "vue-router"
  import { useI18n } from "vue-i18n"
  import { actorApi } from "@/api/api"
  import MediaRating from "@/components/movie/MediaRating.vue"
  import type {
    PersonDetail,
    ActorMovieCredit,
    ActorTVCredit,
  } from "@/types"

  const route = useRoute()
  const router = useRouter()
  const { t } = useI18n()

  function goBack() {
    if (window.history.length > 1) {
      router.back()
    } else {
      router.push("/")
    }
  }

  const actor = ref<PersonDetail | null>(null)
  const movieCredits = ref<ActorMovieCredit[]>([])
  const tvCredits = ref<ActorTVCredit[]>([])

  const isLoading = ref(true)
  const isError = ref(false)
  const isNotFound = ref(false)
  const isBioCollapsed = ref(true)
  const activeTab = ref<"movies" | "tv">("movies")

  async function fetchData() {
    const actorId = Number(route.params.id)
    if (!actorId || isNaN(actorId)) {
      isNotFound.value = true
      isLoading.value = false
      return
    }

    isLoading.value = true
    isError.value = false
    isNotFound.value = false

    try {
      const [detailRes, movieRes, tvRes] = await Promise.all([
        actorApi.getById(actorId),
        actorApi.getMovies(actorId),
        actorApi.getTVSeries(actorId),
      ])

      actor.value = detailRes.data
      movieCredits.value = movieRes.data?.cast || []
      tvCredits.value = tvRes.data?.cast || []
    } catch (err: unknown) {
      const status = (err as { response?: { status?: number } })?.response?.status
      if (status === 404) {
        isNotFound.value = true
      } else {
        isError.value = true
      }
    } finally {
      isLoading.value = false
    }
  }

  onMounted(fetchData)
  watch(() => route.params.id, fetchData)

  // ── Birthday & Age Calculation ──
  const birthMeta = computed(() => {
    if (!actor.value || !actor.value.birthday) return null

    const bdate = new Date(actor.value.birthday)
    if (isNaN(bdate.getTime())) return actor.value.birthday

    if (actor.value.deathday) {
      const ddate = new Date(actor.value.deathday)
      if (!isNaN(ddate.getTime())) {
        let ageAtDeath = ddate.getFullYear() - bdate.getFullYear()
        const mDiff = ddate.getMonth() - bdate.getMonth()
        if (mDiff < 0 || (mDiff === 0 && ddate.getDate() < bdate.getDate())) {
          ageAtDeath--
        }
        return `${t("actor.died")} ${actor.value.deathday} (${t("actor.agedAtDeath")} ${ageAtDeath})`
      }
      return `${t("actor.died")} ${actor.value.deathday}`
    }

    const today = new Date()
    let age = today.getFullYear() - bdate.getFullYear()
    const mDiff = today.getMonth() - bdate.getMonth()
    if (mDiff < 0 || (mDiff === 0 && today.getDate() < bdate.getDate())) {
      age--
    }
    return `${t("actor.born")} ${actor.value.birthday} (${age} ${t("actor.age")})`
  })

  // ── Character Formatting ──
  function formatCharacter(character?: string): string {
    if (!character || !character.trim()) {
      return t("actor.unknownCharacter")
    }
    return t("actor.asCharacter", { character: character.trim() })
  }

  function formatYear(dateStr?: string): string {
    if (!dateStr || !dateStr.trim()) return "-"
    const parts = dateStr.split("-")
    return parts[0] || "-"
  }

  // ── Deterministic Sorting & Deduplication ──
  const processedMovies = computed(() => {
    const map = new Map<number, ActorMovieCredit>()

    for (const item of movieCredits.value) {
      if (!map.has(item.id)) {
        map.set(item.id, { ...item })
      } else {
        const existing = map.get(item.id)!
        if (item.character && item.character !== existing.character) {
          const charA = existing.character ? existing.character.trim() : ""
          const charB = item.character.trim()
          if (charA && charB && !charA.includes(charB)) {
            existing.character = `${charA} / ${charB}`
          }
        }
      }
    }

    const list = Array.from(map.values())

    list.sort((a, b) => {
      const dateA = a.release_date ? new Date(a.release_date).getTime() : 0
      const dateB = b.release_date ? new Date(b.release_date).getTime() : 0
      return dateB - dateA
    })

    return list
  })

  const processedTVSeries = computed(() => {
    const map = new Map<number, ActorTVCredit>()

    for (const item of tvCredits.value) {
      if (!map.has(item.id)) {
        map.set(item.id, { ...item })
      } else {
        const existing = map.get(item.id)!
        if (item.character && item.character !== existing.character) {
          const charA = existing.character ? existing.character.trim() : ""
          const charB = item.character.trim()
          if (charA && charB && !charA.includes(charB)) {
            existing.character = `${charA} / ${charB}`
          }
        }
      }
    }

    const list = Array.from(map.values())

    list.sort((a, b) => {
      const dateA = a.first_air_date ? new Date(a.first_air_date).getTime() : 0
      const dateB = b.first_air_date ? new Date(b.first_air_date).getTime() : 0
      return dateB - dateA
    })

    return list
  })
</script>

<style scoped>
  .actor-detail-page {
    min-height: 100vh;
    padding: 2.5rem 1.5rem 4rem;
    max-width: 1280px;
    margin: 0 auto;
    color: #ffffff;
  }

  /* ── Back Button ── */
  .btn-back {
    display: inline-flex;
    align-items: center;
    gap: 0.5rem;
    padding: 0.6rem 1.25rem;
    background: rgba(255, 255, 255, 0.05);
    border: 1px solid rgba(255, 255, 255, 0.1);
    border-radius: 8px;
    color: #ffffff;
    font-size: 0.9375rem;
    font-weight: 600;
    cursor: pointer;
    margin-bottom: 1.75rem;
    transition: all 0.2s ease;
    backdrop-filter: blur(8px);
  }

  .btn-back:hover {
    background: rgba(255, 255, 255, 0.12);
    border-color: rgba(255, 255, 255, 0.25);
    transform: translateX(-4px);
  }

  .btn-back:focus-visible {
    outline: 2px solid var(--color-brand);
    outline-offset: 2px;
  }

  /* ── State Containers ── */
  .state-container {
    min-height: 60vh;
    display: flex;
    flex-direction: column;
    align-items: center;
    justify-content: center;
    text-align: center;
    gap: 1.5rem;
  }

  .error-icon {
    font-size: 3.5rem;
    color: var(--color-brand);
  }

  .btn-primary {
    display: inline-flex;
    align-items: center;
    gap: 0.5rem;
    padding: 0.75rem 1.75rem;
    background-color: var(--color-brand);
    color: #ffffff;
    border: none;
    border-radius: 8px;
    font-weight: 600;
    text-decoration: none;
    cursor: pointer;
    transition: all 0.2s ease;
  }

  .btn-primary:hover {
    background-color: var(--color-brand-hover);
    transform: translateY(-2px);
  }

  /* ── Actor Content ── */
  .actor-header-card {
    display: flex;
    gap: 2.5rem;
    background: rgba(20, 20, 20, 0.7);
    backdrop-filter: blur(12px);
    border: 1px solid rgba(255, 255, 255, 0.08);
    border-radius: 16px;
    padding: 2rem;
    margin-bottom: 3rem;
  }

  .avatar-wrap {
    width: 220px;
    height: 330px;
    border-radius: 12px;
    overflow: hidden;
    background: #141414;
    border: 2px solid rgba(255, 255, 255, 0.1);
    box-shadow: 0 12px 32px rgba(0, 0, 0, 0.5);
    flex-shrink: 0;
  }

  .actor-avatar {
    width: 100%;
    height: 100%;
    object-fit: cover;
  }

  .avatar-placeholder {
    width: 100%;
    height: 100%;
    display: flex;
    align-items: center;
    justify-content: center;
    font-size: 4rem;
    color: #404040;
    background: #1a1a1a;
  }

  .info-column {
    flex: 1;
    display: flex;
    flex-direction: column;
  }

  .actor-name {
    font-size: 2.5rem;
    font-weight: 800;
    margin: 0 0 1rem;
    letter-spacing: -0.02em;
    background: linear-gradient(135deg, #ffffff 0%, #a3a3a3 100%);
    -webkit-background-clip: text;
    -webkit-text-fill-color: transparent;
  }

  .meta-pills {
    display: flex;
    flex-wrap: wrap;
    gap: 0.75rem;
    margin-bottom: 1.5rem;
  }

  .meta-pill {
    display: inline-flex;
    align-items: center;
    gap: 0.5rem;
    padding: 0.4rem 0.9rem;
    background: rgba(255, 255, 255, 0.06);
    border: 1px solid rgba(255, 255, 255, 0.1);
    border-radius: 20px;
    font-size: 0.875rem;
    color: #d4d4d4;
    text-decoration: none;
    transition: all 0.2s ease;
  }

  .meta-pill i {
    color: var(--color-brand);
  }

  .meta-pill.imdb-link:hover {
    background: #f5c518;
    color: #000000;
    border-color: #f5c518;
  }

  .meta-pill.imdb-link:hover i {
    color: #000000;
  }

  /* ── Biography ── */
  .bio-section {
    margin-top: auto;
  }

  .bio-title {
    font-size: 1.125rem;
    font-weight: 600;
    margin: 0 0 0.75rem;
    display: flex;
    align-items: center;
    gap: 0.5rem;
    color: #e5e5e5;
  }

  .bio-text {
    font-size: 0.9375rem;
    line-height: 1.6;
    color: #a3a3a3;
    margin: 0;
    white-space: pre-line;
  }

  .bio-text.collapsed {
    display: -webkit-box;
    -webkit-line-clamp: 4;
    -webkit-box-orient: vertical;
    overflow: hidden;
  }

  .bio-empty {
    font-size: 0.9375rem;
    color: #737373;
    font-style: italic;
  }

  .btn-toggle-bio {
    background: none;
    border: none;
    color: var(--color-brand);
    font-size: 0.875rem;
    font-weight: 600;
    cursor: pointer;
    margin-top: 0.5rem;
    display: inline-flex;
    align-items: center;
    gap: 0.35rem;
    padding: 0;
  }

  .btn-toggle-bio:hover {
    text-decoration: underline;
  }

  /* ── Workography ── */
  .workography-header {
    display: flex;
    justify-content: space-between;
    align-items: center;
    margin-bottom: 1.75rem;
    flex-wrap: wrap;
    gap: 1rem;
  }

  .section-title {
    font-size: 1.5rem;
    font-weight: 700;
    margin: 0;
    display: flex;
    align-items: center;
    gap: 0.625rem;
  }

  .section-title i {
    color: var(--color-brand);
  }

  .tab-controls {
    display: flex;
    gap: 0.5rem;
    background: rgba(255, 255, 255, 0.05);
    padding: 0.25rem;
    border-radius: 10px;
    border: 1px solid rgba(255, 255, 255, 0.08);
  }

  .tab-btn {
    display: inline-flex;
    align-items: center;
    gap: 0.5rem;
    padding: 0.5rem 1.25rem;
    background: transparent;
    border: none;
    border-radius: 8px;
    color: #a3a3a3;
    font-weight: 600;
    font-size: 0.9375rem;
    cursor: pointer;
    transition: all 0.2s ease;
  }

  .tab-btn.active {
    background: var(--color-brand);
    color: #ffffff;
  }

  .tab-badge {
    padding: 0.15rem 0.5rem;
    border-radius: 12px;
    background: rgba(0, 0, 0, 0.3);
    font-size: 0.75rem;
  }

  /* ── Media Grid ── */
  .media-grid {
    display: grid;
    grid-template-columns: repeat(auto-fill, minmax(180px, 1fr));
    gap: 1.5rem;
  }

  .media-card {
    display: flex;
    flex-direction: column;
    background: rgba(26, 26, 26, 0.6);
    border: 1px solid rgba(255, 255, 255, 0.06);
    border-radius: 12px;
    overflow: hidden;
    text-decoration: none;
    color: inherit;
    transition: all 0.3s cubic-bezier(0.4, 0, 0.2, 1);
  }

  .media-card:hover,
  .media-card:focus-visible {
    transform: translateY(-6px);
    border-color: rgba(245, 67, 44, 0.5);
    box-shadow: 0 12px 24px rgba(0, 0, 0, 0.5);
    outline: none;
  }

  .media-card:focus-visible {
    box-shadow: 0 0 0 2px var(--color-brand), 0 12px 24px rgba(0, 0, 0, 0.5);
  }

  .poster-wrap {
    position: relative;
    width: 100%;
    aspect-ratio: 2 / 3;
    background: #141414;
    overflow: hidden;
  }

  .poster-img {
    width: 100%;
    height: 100%;
    object-fit: cover;
    transition: transform 0.3s ease;
  }

  .media-card:hover .poster-img {
    transform: scale(1.05);
  }

  .poster-placeholder {
    width: 100%;
    height: 100%;
    display: flex;
    align-items: center;
    justify-content: center;
    font-size: 2.5rem;
    color: #404040;
  }

  .rating-badge {
    position: absolute;
    top: 0.5rem;
    right: 0.5rem;
    background: rgba(0, 0, 0, 0.75);
    backdrop-filter: blur(4px);
    padding: 0.25rem 0.5rem;
    border-radius: 6px;
    font-size: 0.75rem;
    font-weight: 700;
    color: var(--color-text-muted);
    display: flex;
    align-items: center;
    gap: 0.25rem;
  }

  .card-info {
    padding: 0.875rem;
    display: flex;
    flex-direction: column;
    gap: 0.35rem;
    flex: 1;
  }

  .media-title {
    font-size: 0.9375rem;
    font-weight: 600;
    margin: 0;
    line-height: 1.3;
    display: -webkit-box;
    -webkit-line-clamp: 2;
    -webkit-box-orient: vertical;
    overflow: hidden;
  }

  .character-role {
    font-size: 0.8125rem;
    color: var(--color-brand);
    margin: 0;
    display: flex;
    align-items: flex-start;
    gap: 0.35rem;
    line-height: 1.3;
  }

  .character-role i {
    font-size: 0.75rem;
    margin-top: 0.15rem;
    flex-shrink: 0;
  }

  .release-year {
    font-size: 0.75rem;
    color: #737373;
    margin: auto 0 0;
  }

  .empty-credits {
    text-align: center;
    padding: 4rem 1.5rem;
    background: rgba(255, 255, 255, 0.02);
    border-radius: 12px;
    border: 1px dashed rgba(255, 255, 255, 0.1);
    color: #737373;
  }

  .empty-credits i {
    font-size: 3rem;
    margin-bottom: 1rem;
    display: block;
  }

  /* ── Responsiveness ── */
  @media (max-width: 768px) {
    .actor-header-card {
      flex-direction: column;
      align-items: center;
      text-align: center;
      padding: 1.5rem;
    }

    .avatar-wrap {
      width: 180px;
      height: 270px;
    }

    .meta-pills {
      justify-content: center;
    }

    .bio-title {
      justify-content: center;
    }

    .workography-header {
      flex-direction: column;
      align-items: flex-start;
    }

    .media-grid {
      grid-template-columns: repeat(auto-fill, minmax(140px, 1fr));
      gap: 1rem;
    }
  }
</style>
