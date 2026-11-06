<template>
  <div class="bol-share-card">
    <!-- Ambient Background Glows -->
    <div class="bg-glow bg-glow-1" :style="headerGlowStyle"></div>
    <div class="bg-glow bg-glow-2"></div>

    <!-- Header Section -->
    <div class="card-header">
      <div class="header-branding">
        <div class="brand-logo">
          <span class="logo-re">POP</span><span class="logo-mov">CORN</span>
        </div>
        <div class="header-badge">
          <span class="badge-dot"></span>
          BEST {{ mediaType === 'movie' ? 'MOVIES' : 'TV SERIES' }} OF LIFE
        </div>
      </div>
      <div v-if="showUsername && username" class="username-pill">
        <span class="by-prefix">curated by</span>
        <span class="username-text">@{{ username }}</span>
      </div>
    </div>

    <!-- Main Ranks Container (Fixed 5 Equal Rows Grid - Poster Grid Mode) -->
    <div class="ranks-container">
      <div
        v-for="group in processedRanks"
        :key="group.rank"
        class="rank-row"
      >
        <div class="row-left">
          <div class="rank-badge-grid" :class="[`rank-badge-${group.rank}`]">
            <span class="rank-hash">#</span>{{ group.rank }}
          </div>
        </div>
        <div class="row-content-grid">
          <div class="posters-grid">
            <div
              v-for="item in group.processedItems.slice(0, group.items.length > 7 ? 6 : 7)"
              :key="item.id"
              class="poster-item"
            >
              <img
                v-if="item.posterUrl"
                :src="item.posterUrl"
                :alt="item.displayTitle"
                class="poster-img"
              />
              <div v-else class="poster-fallback">
                <span>{{ item.displayTitle.charAt(0) }}</span>
              </div>
            </div>
            <div v-if="group.items.length > 7" class="more-count-badge">
              +{{ group.items.length - 6 }}
            </div>
          </div>
          <div class="titles-summary">
            {{ group.titlesSummary }}
          </div>
        </div>
      </div>

      <!-- Empty state if no ranks at all -->
      <div v-if="displayRanks.length === 0" class="empty-ranks">
        No {{ mediaType === 'movie' ? 'movies' : 'TV series' }} ranked yet
      </div>
    </div>

    <!-- Footer Section -->
    <div class="card-footer">
      <div class="footer-divider"></div>
      <div class="footer-content">
        <div class="footer-logo">
          <span class="logo-re">POP</span><span class="logo-mov">CORN</span>
        </div>
        <span class="footer-url">POPCORN</span>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { computed } from "vue"
import type { BMOLItemResponse } from "@/types/movie"
import type { MediaType } from "@/types/common"
import type { ExtractedColor } from "@/utils/extractDominantColor"
import { getTmdbImageUrl } from "@/utils/image"

export interface RankGroup {
  rank: number
  items: BMOLItemResponse[]
}

const props = defineProps<{
  mediaType: MediaType
  ranks: RankGroup[]
  username?: string
  showUsername?: boolean
  posterBlobUrls: Record<string, string>
  dominantColors: Record<string, ExtractedColor>
  englishTitles?: Record<string, string>
}>()

// Take top 5 ranks
const displayRanks = computed(() => props.ranks.slice(0, 5))

const DEFAULT_COLOR: ExtractedColor = {
  r: 229,
  g: 9,
  b: 20,
  hex: "#f5432c",
  rgbStr: "rgb(229, 9, 20)",
}

// Extract English title for item
function getItemEnglishTitle(item: BMOLItemResponse | undefined): string {
  if (!item || !item.media) return ""
  const key = `${item.media_type}:${item.media.id}`
  if (props.englishTitles && props.englishTitles[key]) {
    return props.englishTitles[key]
  }
  const m = item.media
  return m.english_title || m.original_title || m.original_name || m.title || m.name || ""
}

// Poster URL getter
function getPosterUrl(item: BMOLItemResponse | undefined): string | undefined {
  if (!item) return undefined
  const key = `${item.media_type}:${item.media.id}`
  const blob = props.posterBlobUrls[key]
  if (blob) return blob

  let rawUrl = item.media.poster_url
  if (rawUrl && !rawUrl.startsWith("http://") && !rawUrl.startsWith("https://") && !rawUrl.startsWith("data:")) {
    rawUrl = getTmdbImageUrl(rawUrl, "w500") || rawUrl
  }
  return rawUrl || undefined
}

// Dominant color getter for a single item
function getItemColor(item: BMOLItemResponse | undefined): ExtractedColor {
  if (!item) return DEFAULT_COLOR
  const key = `${item.media_type}:${item.media.id}`
  return props.dominantColors[key] || DEFAULT_COLOR
}

// Pre-calculate processed items & English titles
const processedRanks = computed(() => {
  return displayRanks.value.map((group) => {
    const processedItems = group.items.map((item) => {
      const displayTitle = getItemEnglishTitle(item)
      return {
        ...item,
        displayTitle,
        posterUrl: getPosterUrl(item),
      }
    })

    const titlesSummary = formatTitles(processedItems, group.items.length > 7 ? 6 : 7)

    return {
      rank: group.rank,
      items: group.items,
      processedItems,
      titlesSummary,
    }
  })
})

// Format titles list using pre-extracted English titles
function formatTitles(items: { displayTitle: string }[], maxVisible: number = 7): string {
  if (!items || items.length === 0) return ""
  const visibleItems = items.slice(0, maxVisible)
  const titlesStr = visibleItems.map((i) => i.displayTitle).join("  ·  ")
  const remainingCount = items.length - maxVisible
  if (remainingCount > 0) {
    return `${titlesStr}  ·  ... [+${remainingCount}]`
  }
  return titlesStr
}

// Top header glow style based on #1 rank color if available
const headerGlowStyle = computed(() => {
  const firstRank = displayRanks.value[0]
  const firstItem = firstRank?.items[0]
  if (firstItem) {
    const col = getItemColor(firstItem)
    return {
      background: `radial-gradient(circle, rgba(${col.r}, ${col.g}, ${col.b}, 0.35) 0%, rgba(0, 0, 0, 0) 70%)`,
    }
  }
  return {
    background: "radial-gradient(circle, rgba(245, 67, 44, 0.3) 0%, rgba(0, 0, 0, 0) 70%)",
  }
})
</script>

<style scoped>
.bol-share-card {
  width: 1080px;
  height: 1920px;
  background-color: #0b0b0f;
  background-image:
    radial-gradient(at 0% 0%, rgba(30, 20, 50, 0.8) 0px, transparent 50%),
    radial-gradient(at 100% 100%, rgba(20, 15, 35, 0.8) 0px, transparent 50%);
  color: #ffffff;
  font-family: 'Inter', system-ui, -apple-system, BlinkMacSystemFont, 'Segoe UI', Roboto, sans-serif;
  position: relative;
  display: flex;
  flex-direction: column;
  justify-content: space-between;
  padding: 72px 64px 60px 64px;
  box-sizing: border-box;
  overflow: hidden;
  contain: content;
}

/* Ambient Glows */
.bg-glow {
  position: absolute;
  pointer-events: none;
  border-radius: 50%;
}
.bg-glow-1 {
  top: -150px;
  right: -150px;
  width: 800px;
  height: 800px;
}
.bg-glow-2 {
  bottom: -200px;
  left: -200px;
  width: 900px;
  height: 900px;
  background: radial-gradient(circle, rgba(245, 67, 44, 0.15) 0%, rgba(0, 0, 0, 0) 70%);
}

/* Header */
.card-header {
  position: relative;
  z-index: 2;
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding-bottom: 32px;
  border-bottom: 1px solid rgba(255, 255, 255, 0.1);
}

.header-branding {
  display: flex;
  flex-direction: column;
  gap: 12px;
}

.brand-logo {
  font-size: 52px;
  font-weight: 900;
  letter-spacing: 2px;
  line-height: 1;
}

.logo-re {
  color: #f5432c;
}

.logo-mov {
  color: #ffffff;
}

.header-badge {
  display: inline-flex;
  align-items: center;
  gap: 12px;
  font-size: 20px;
  font-weight: 800;
  letter-spacing: 4px;
  color: rgba(255, 255, 255, 0.7);
  text-transform: uppercase;
}

.badge-dot {
  width: 10px;
  height: 10px;
  border-radius: 50%;
  background-color: #f5432c;
  box-shadow: 0 0 12px #f5432c;
}

.username-pill {
  display: flex;
  flex-direction: column;
  align-items: flex-end;
  background: rgba(255, 255, 255, 0.06);
  border: 1px solid rgba(255, 255, 255, 0.12);
  padding: 14px 28px;
  border-radius: 40px;
}

.by-prefix {
  font-size: 14px;
  font-weight: 600;
  color: rgba(255, 255, 255, 0.5);
  text-transform: uppercase;
  letter-spacing: 2px;
}

.username-text {
  font-size: 26px;
  font-weight: 700;
  color: #ffffff;
}

/* Ranks Container (Fixed 5 Equal Rows Grid Layout) */
.ranks-container {
  position: relative;
  z-index: 2;
  flex: 1;
  display: grid;
  grid-template-rows: repeat(5, 1fr);
  margin: 28px 0;
  gap: 16px;
}

.rank-row {
  width: 100%;
  height: 100%;
  display: flex;
  align-items: center;
  border-radius: 20px;
  box-sizing: border-box;
  background: rgba(255, 255, 255, 0.035);
  border: 1px solid rgba(255, 255, 255, 0.08);
  padding: 16px 24px;
  gap: 24px;
}

.rank-badge-grid {
  width: 68px;
  height: 68px;
  border-radius: 18px;
  background: rgba(255, 255, 255, 0.08);
  border: 1px solid rgba(255, 255, 255, 0.15);
  display: flex;
  align-items: center;
  justify-content: center;
  font-size: 34px;
  font-weight: 900;
  color: #ffffff;
  flex-shrink: 0;
}

.rank-badge-1 {
  background: linear-gradient(135deg, #ffd700 0%, #ff8c00 100%);
  color: #000000;
  box-shadow: 0 8px 24px rgba(255, 215, 0, 0.35);
  border: none;
}

.rank-badge-2 {
  background: linear-gradient(135deg, #e0e0e0 0%, #9e9e9e 100%);
  color: #000000;
  box-shadow: 0 6px 20px rgba(220, 220, 220, 0.25);
  border: none;
}

.rank-badge-3 {
  background: linear-gradient(135deg, #cd7f32 0%, #8b4513 100%);
  color: #ffffff;
  box-shadow: 0 6px 20px rgba(205, 127, 50, 0.25);
  border: none;
}

.rank-hash {
  font-size: 22px;
  opacity: 0.8;
  margin-right: 2px;
}

.row-content-grid {
  flex: 1;
  display: flex;
  flex-direction: column;
  gap: 12px;
  min-width: 0;
}

.posters-grid {
  display: flex;
  align-items: center;
  gap: 12px;
}

.poster-item {
  width: 72px;
  height: 108px;
  border-radius: 12px;
  overflow: hidden;
  box-shadow: 0 6px 16px rgba(0, 0, 0, 0.4);
  border: 1px solid rgba(255, 255, 255, 0.15);
  flex-shrink: 0;
}

.poster-img {
  width: 100%;
  height: 100%;
  object-fit: cover;
}

.poster-fallback {
  width: 100%;
  height: 100%;
  background: #1e1e24;
  display: flex;
  align-items: center;
  justify-content: center;
  font-size: 26px;
  font-weight: 700;
  color: rgba(255, 255, 255, 0.4);
}

.more-count-badge {
  background: rgba(255, 255, 255, 0.1);
  border: 1px solid rgba(255, 255, 255, 0.2);
  border-radius: 12px;
  padding: 8px 16px;
  font-size: 18px;
  font-weight: 700;
  color: #ffffff;
}

.titles-summary {
  font-size: 20px;
  font-weight: 600;
  color: rgba(255, 255, 255, 0.9);
  line-height: 1.35;
  display: -webkit-box;
  -webkit-line-clamp: 3;
  -webkit-box-orient: vertical;
  overflow: hidden;
  word-break: break-word;
}

.empty-ranks {
  text-align: center;
  padding: 80px 0;
  font-size: 28px;
  color: rgba(255, 255, 255, 0.4);
}

/* Footer */
.card-footer {
  position: relative;
  z-index: 2;
  display: flex;
  flex-direction: column;
  gap: 24px;
}

.footer-divider {
  width: 100%;
  height: 1px;
  background: linear-gradient(90deg, transparent 0%, rgba(255, 255, 255, 0.15) 50%, transparent 100%);
}

.footer-content {
  display: flex;
  align-items: center;
  justify-content: space-between;
}

.footer-logo {
  font-size: 36px;
  font-weight: 900;
  letter-spacing: 2px;
}

.footer-url {
  font-size: 22px;
  font-weight: 700;
  color: rgba(255, 255, 255, 0.5);
  letter-spacing: 2px;
}
</style>
