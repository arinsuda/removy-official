<template>
  <div class="achv-page">
    <div class="achv-header">
      <div>
        <h1 class="achv-title">Achievements</h1>
        <p class="achv-sub">
          {{ store.unlockedCount }} / {{ store.pagination.total }} ปลดล็อกแล้ว
        </p>
      </div>
    </div>

    <div class="filter-bar">
      <div class="search-field">
        <Search :size="16" />
        <input
          v-model="store.filters.search"
          type="text"
          placeholder="ค้นหา achievement..."
        />
      </div>

      <div class="segmented">
        <button
          v-for="opt in statusOptions"
          :key="opt.value"
          class="segmented-btn"
          :class="{ active: store.filters.unlockedFilter === opt.value }"
          @click="changeStatus(opt.value)"
        >
          {{ opt.label }}
        </button>
      </div>

      <!-- ตัวกรองประเภท -->
      <div class="dropdown" ref="typeDropdownRef">
        <button
          type="button"
          class="dropdown-trigger"
          :class="{ active: isTypeDropdownOpen || !!store.filters.actionType }"
          @click="toggleDropdown('type')"
        >
          <component
            :is="
              store.filters.actionType
                ? getAchievementIcon(store.filters.actionType)
                : ListFilter
            "
            :size="14"
          />
          <span>
            {{
              store.filters.actionType
                ? getActionTypeLabel(store.filters.actionType)
                : "ทุกประเภท"
            }}
          </span>
          <ChevronDown
            :size="14"
            class="chevron"
            :class="{ open: isTypeDropdownOpen }"
          />
        </button>

        <Transition name="dropdown-fade">
          <div v-if="isTypeDropdownOpen" class="dropdown-menu">
            <button
              class="dropdown-item"
              :class="{ selected: !store.filters.actionType }"
              @click="selectActionType('')"
            >
              <ListFilter :size="14" />
              <span>ทุกประเภท</span>
            </button>
            <button
              v-for="t in store.actionTypes"
              :key="t"
              class="dropdown-item"
              :class="{ selected: store.filters.actionType === t }"
              @click="selectActionType(t)"
            >
              <component :is="getAchievementIcon(t)" :size="14" />
              <span>{{ getActionTypeLabel(t) }}</span>
            </button>
          </div>
        </Transition>
      </div>

      <!-- จำนวนต่อหน้า -->
      <div class="dropdown dropdown--sm" ref="limitDropdownRef">
        <button
          type="button"
          class="dropdown-trigger dropdown-trigger--sm"
          :class="{ active: isLimitDropdownOpen }"
          @click="toggleDropdown('limit')"
        >
          <span>{{ store.filters.limit }} / หน้า</span>
          <ChevronDown
            :size="14"
            class="chevron"
            :class="{ open: isLimitDropdownOpen }"
          />
        </button>

        <Transition name="dropdown-fade">
          <div
            v-if="isLimitDropdownOpen"
            class="dropdown-menu dropdown-menu--sm"
          >
            <button
              v-for="opt in limitOptions"
              :key="opt"
              class="dropdown-item"
              :class="{ selected: store.filters.limit === opt }"
              @click="selectLimit(opt)"
            >
              <span>{{ opt }} / หน้า</span>
            </button>
          </div>
        </Transition>
      </div>
    </div>

    <div v-if="store.loading" class="grid-skeleton">
      <SkeletonBlock v-for="n in 6" :key="n" height="110px" radius="12px" />
    </div>

    <div v-else-if="store.error" class="state-msg state-msg--error">
      {{ store.error }}
    </div>

    <div v-else-if="store.filteredBySearch.length === 0" class="state-msg">
      ไม่พบ achievement ที่ตรงกับเงื่อนไข
    </div>

    <div v-else class="achv-grid">
      <div
        v-for="(ua, i) in store.filteredBySearch"
        :key="ua.achievement_id"
        class="achv-card"
        :class="{ 'achv-card--locked': !ua.is_unlocked }"
        :style="{ '--i': i }"
      >
        <div class="achv-icon" :class="{ lit: ua.is_unlocked }">
          <component
            :is="getAchievementIcon(ua.achievement.action_type)"
            :size="18"
            :stroke-width="1.8"
          />
        </div>

        <div class="achv-body">
          <div class="achv-body-top">
            <p class="achv-name">{{ ua.achievement.name }}</p>
            <!-- <span class="exp-chip">+{{ ua.achievement.exp_reward }} EXP</span> -->
          </div>
          <p class="achv-desc">{{ ua.achievement.description }}</p>

          <div class="progress-row">
            <div class="progress-bar">
              <div
                class="progress-fill"
                :class="{ done: ua.is_unlocked }"
                :style="{ width: ua.progress_pct + '%' }"
              />
            </div>
            <span class="progress-text">
              {{ ua.current_count }} / {{ ua.achievement.target_count }}
            </span>
          </div>

          <p v-if="ua.is_unlocked && ua.unlocked_at" class="unlocked-at">
            <CheckCircle :size="10" /> ปลดล็อกเมื่อ
            {{ formatDate(ua.unlocked_at) }}
          </p>
        </div>

        <span
          class="badge"
          :class="ua.is_unlocked ? 'badge--done' : 'badge--locked'"
        >
          <component :is="ua.is_unlocked ? CheckCircle : Lock" :size="10" />
          {{ ua.is_unlocked ? "Unlocked" : "Locked" }}
        </span>
      </div>
    </div>

    <div
      v-if="!store.loading && store.pagination.total > 0"
      class="pagination-wrap"
    >
      <p class="pagination-info">
        แสดง
        <strong
          >{{ store.showingRange.from }}–{{ store.showingRange.to }}</strong
        >
        จากทั้งหมด <strong>{{ store.pagination.total }}</strong> รายการ
      </p>

      <div v-if="store.pagination.total_pages > 1" class="pagination">
        <button
          class="page-btn"
          :disabled="store.filters.page <= 1"
          title="หน้าแรก"
          @click="goToPage(1)"
        >
          <ChevronsLeft :size="15" />
        </button>
        <button
          class="page-btn"
          :disabled="store.filters.page <= 1"
          title="ก่อนหน้า"
          @click="goToPage(store.filters.page - 1)"
        >
          <ChevronLeft :size="15" />
        </button>

        <div class="page-numbers">
          <button
            v-for="p in pageList"
            :key="p === '...' ? `dots-${Math.random()}` : p"
            class="page-num"
            :class="{
              active: p === store.filters.page,
              dots: p === '...',
            }"
            :disabled="p === '...'"
            @click="p !== '...' && goToPage(p as number)"
          >
            {{ p }}
          </button>
        </div>

        <button
          class="page-btn"
          :disabled="store.filters.page >= store.pagination.total_pages"
          title="ถัดไป"
          @click="goToPage(store.filters.page + 1)"
        >
          <ChevronRight :size="15" />
        </button>
        <button
          class="page-btn"
          :disabled="store.filters.page >= store.pagination.total_pages"
          title="หน้าสุดท้าย"
          @click="goToPage(store.pagination.total_pages)"
        >
          <ChevronsRight :size="15" />
        </button>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import SkeletonBlock from "@/components/ui/skeleton/SkeletonBlock.vue"
  import { computed, onMounted, onUnmounted, ref } from "vue"
  import { useRoute } from "vue-router"
  import {
    Search,
    CheckCircle,
    Lock,
    ChevronLeft,
    ChevronRight,
    ChevronsLeft,
    ChevronsRight,
    ChevronDown,
    ListFilter,
  } from "lucide-vue-next"
  import { useAchievementStore } from "@/stores/achievement"
  import {
    getAchievementIcon,
    getActionTypeLabel,
  } from "@/utils/achievementIcons"
  import type { UnlockedFilter } from "@/types/achievement"

  const route = useRoute()
  const store = useAchievementStore()

  const userId = computed(() => Number(route.params.userId))

  const statusOptions: { label: string; value: UnlockedFilter }[] = [
    { label: "ทั้งหมด", value: "all" },
    { label: "ปลดล็อกแล้ว", value: "unlocked" },
    { label: "ยังไม่ปลดล็อก", value: "locked" },
  ]

  const limitOptions = [12, 24, 48]

  // สร้างรายการเลขหน้าแบบมี "..." เมื่อหน้าเยอะ เช่น 1 ... 4 5 6 ... 12
  const pageList = computed<(number | "...")[]>(() => {
    const total = store.pagination.total_pages
    const current = store.filters.page
    if (total <= 7) {
      return Array.from({ length: total }, (_, i) => i + 1)
    }

    const pages: (number | "...")[] = [1]

    if (current > 3) pages.push("...")

    const start = Math.max(2, current - 1)
    const end = Math.min(total - 1, current + 1)
    for (let i = start; i <= end; i++) pages.push(i)

    if (current < total - 2) pages.push("...")

    pages.push(total)
    return pages
  })

  function changeStatus(value: UnlockedFilter) {
    store.filters.unlockedFilter = value
    store.filters.page = 1
    load()
  }

  function onActionTypeChange() {
    store.filters.page = 1
    load()
  }

  function onLimitChange() {
    store.filters.page = 1
    load()
  }

  function goToPage(page: number) {
    store.setPage(page)
    load()
    window.scrollTo({ top: 0, behavior: "smooth" })
  }

  function load() {
    store.fetchUserAchievements(userId.value)
  }

  function formatDate(iso: string) {
    return new Date(iso).toLocaleDateString("th-TH", {
      day: "numeric",
      month: "short",
      year: "numeric",
    })
  }

  onMounted(async () => {
    store.resetFilters()
    await Promise.all([
      store.fetchCatalog(),
      store.fetchUserAchievements(userId.value),
    ])
  })

  // ── Dropdown: ประเภท ──
  const isTypeDropdownOpen = ref(false)
  const typeDropdownRef = ref<HTMLElement | null>(null)

  // ── Dropdown: จำนวนต่อหน้า ──
  const isLimitDropdownOpen = ref(false)
  const limitDropdownRef = ref<HTMLElement | null>(null)

  function toggleDropdown(which: "type" | "limit") {
    if (which === "type") {
      isTypeDropdownOpen.value = !isTypeDropdownOpen.value
      isLimitDropdownOpen.value = false
    } else {
      isLimitDropdownOpen.value = !isLimitDropdownOpen.value
      isTypeDropdownOpen.value = false
    }
  }

  function selectActionType(value: string) {
    store.filters.actionType = value
    isTypeDropdownOpen.value = false
    onActionTypeChange()
  }

  function selectLimit(value: number) {
    store.filters.limit = value
    isLimitDropdownOpen.value = false
    onLimitChange()
  }

  function handleClickOutside(e: MouseEvent) {
    const target = e.target as Node
    if (typeDropdownRef.value && !typeDropdownRef.value.contains(target)) {
      isTypeDropdownOpen.value = false
    }
    if (limitDropdownRef.value && !limitDropdownRef.value.contains(target)) {
      isLimitDropdownOpen.value = false
    }
  }

  onMounted(() => document.addEventListener("click", handleClickOutside))
  onUnmounted(() => document.removeEventListener("click", handleClickOutside))
</script>

<style scoped>
  .achv-page {
    --c-bg: var(--color-surface);
    --c-card: var(--color-surface-2);
    --c-border: var(--color-border);
    --c-border-h: var(--color-border-strong);
    --c-red: var(--color-brand);
    --c-text: var(--color-text-primary);
    --c-sub: var(--color-text-secondary);
    --c-muted: var(--color-text-muted);
    --ease: cubic-bezier(0.16, 1, 0.3, 1);
    color: var(--c-text);
    max-width: 1100px;
    margin: 0 auto;
    padding: 32px 24px 64px;
  }

  .achv-header {
    display: flex;
    align-items: flex-end;
    justify-content: space-between;
    margin-bottom: 20px;
  }
  .achv-title {
    font-size: 1.6rem;
    font-weight: 800;
    margin: 0;
  }
  .achv-sub {
    color: var(--c-sub);
    font-size: 0.85rem;
    margin: 4px 0 0;
  }

  .filter-bar {
    display: flex;
    flex-wrap: wrap;
    gap: 10px;
    align-items: center;
    margin-bottom: 24px;
    padding: 12px;
    background: var(--c-card);
    border: 1px solid var(--c-border);
    border-radius: 12px;
  }

  .search-field {
    display: flex;
    align-items: center;
    gap: 8px;
    background: #0d0d0d;
    border: 1px solid var(--c-border);
    border-radius: 8px;
    padding: 8px 12px;
    color: var(--c-sub);
    flex: 1;
    min-width: 200px;
  }
  .search-field input {
    background: none;
    border: none;
    outline: none;
    color: #fff;
    font-size: 0.85rem;
    width: 100%;
  }

  .segmented {
    display: flex;
    background: #0d0d0d;
    border: 1px solid var(--c-border);
    border-radius: 8px;
    padding: 3px;
    gap: 2px;
  }
  .segmented-btn {
    background: none;
    border: none;
    color: var(--c-sub);
    font-size: 0.78rem;
    font-weight: 600;
    padding: 6px 12px;
    border-radius: 6px;
    cursor: pointer;
    transition: all 0.15s;
  }
  .segmented-btn.active {
    background: var(--c-red);
    color: #fff;
  }

  /* ── Custom dropdown (ใช้แทน native <select> เพื่อคุมธีมได้เต็มที่) ── */
  .dropdown {
    position: relative;
  }
  .dropdown--sm {
    flex-shrink: 0;
  }

  .dropdown-trigger {
    display: flex;
    align-items: center;
    gap: 8px;
    background: #0d0d0d;
    border: 1px solid var(--c-border);
    border-radius: 8px;
    color: var(--c-sub);
    font-size: 0.8rem;
    font-weight: 500;
    padding: 8px 10px;
    cursor: pointer;
    min-width: 170px;
    transition: all 0.15s var(--ease);
  }
  .dropdown-trigger:hover {
    border-color: var(--c-border-h);
  }
  .dropdown-trigger.active {
    color: #fff;
    border-color: var(--c-border-h);
  }
  .dropdown-trigger span {
    flex: 1;
    text-align: left;
    white-space: nowrap;
    overflow: hidden;
    text-overflow: ellipsis;
  }
  .dropdown-trigger .chevron {
    transition: transform 0.2s var(--ease);
    flex-shrink: 0;
    color: var(--c-sub);
  }
  .dropdown-trigger .chevron.open {
    transform: rotate(180deg);
  }

  .dropdown-trigger--sm {
    min-width: 104px;
  }

  .dropdown-menu {
    position: absolute;
    top: calc(100% + 6px);
    left: 0;
    min-width: 220px;
    max-height: 320px;
    overflow-y: auto;
    background: #161616;
    border: 1px solid var(--c-border-h);
    border-radius: 10px;
    padding: 6px;
    box-shadow: 0 12px 32px rgba(0, 0, 0, 0.5);
    z-index: 20;
  }
  .dropdown-menu--sm {
    min-width: 104px;
  }

  .dropdown-item {
    display: flex;
    align-items: center;
    gap: 8px;
    width: 100%;
    background: none;
    border: none;
    border-radius: 7px;
    color: var(--c-sub);
    font-size: 0.8rem;
    text-align: left;
    padding: 8px 10px;
    cursor: pointer;
    transition: all 0.12s;
  }
  .dropdown-item:hover {
    background: rgba(255, 255, 255, 0.05);
    color: #fff;
  }
  .dropdown-item.selected {
    background: var(--c-red);
    color: #fff;
  }

  .dropdown-fade-enter-active,
  .dropdown-fade-leave-active {
    transition: all 0.15s var(--ease);
  }
  .dropdown-fade-enter-from,
  .dropdown-fade-leave-to {
    opacity: 0;
    transform: translateY(-4px);
  }

  .grid-skeleton,
  .achv-grid {
    display: grid;
    grid-template-columns: repeat(auto-fill, minmax(300px, 1fr));
    gap: 12px;
  }
  .state-msg {
    padding: 48px 0;
    text-align: center;
    color: var(--c-sub);
    font-size: 0.85rem;
  }
  .state-msg--error {
    color: var(--c-red);
  }

  .achv-card {
    display: flex;
    gap: 14px;
    padding: 16px;
    background: var(--c-card);
    border: 1px solid var(--c-border);
    border-radius: 12px;
    position: relative;
    transition: all 0.2s var(--ease);
    animation: cardIn 0.4s var(--ease) calc(var(--i) * 40ms) both;
  }
  @keyframes cardIn {
    from {
      opacity: 0;
      transform: translateY(6px);
    }
    to {
      opacity: 1;
      transform: translateY(0);
    }
  }
  .achv-card:hover {
    border-color: var(--c-border-h);
  }
  .achv-card--locked {
    opacity: 0.55;
  }

  .achv-icon {
    width: 40px;
    height: 40px;
    flex-shrink: 0;
    border-radius: 8px;
    background: #0d0d0d;
    border: 1px solid var(--c-border);
    display: flex;
    align-items: center;
    justify-content: center;
    color: var(--c-sub);
  }
  .achv-icon.lit {
    color: #fff;
    background: rgba(255, 255, 255, 0.04);
  }

  .achv-body {
    flex: 1;
    min-width: 0;
  }
  .achv-body-top {
    display: flex;
    align-items: center;
    justify-content: space-between;
    gap: 8px;
  }
  .achv-name {
    font-size: 0.88rem;
    font-weight: 600;
    color: #fff;
    margin: 0;
  }
  .exp-chip {
    font-size: 0.62rem;
    font-weight: 700;
    color: var(--c-red);
    white-space: nowrap;
  }
  .achv-desc {
    font-size: 0.74rem;
    color: var(--c-sub);
    line-height: 1.45;
    margin: 2px 0 8px;
  }

  .progress-row {
    display: flex;
    align-items: center;
    gap: 8px;
  }
  .progress-bar {
    flex: 1;
    height: 5px;
    background: #0d0d0d;
    border-radius: 3px;
    overflow: hidden;
  }
  .progress-fill {
    height: 100%;
    background: var(--c-sub);
    border-radius: 3px;
    transition: width 0.3s var(--ease);
  }
  .progress-fill.done {
    background: #fff;
  }
  .progress-text {
    font-size: 0.65rem;
    color: var(--c-sub);
    white-space: nowrap;
  }

  .unlocked-at {
    display: inline-flex;
    align-items: center;
    gap: 4px;
    font-size: 0.65rem;
    color: var(--c-sub);
    margin: 6px 0 0;
  }

  .badge {
    position: absolute;
    top: 14px;
    right: 14px;
    display: inline-flex;
    align-items: center;
    gap: 4px;
    font-size: 0.6rem;
    font-weight: 600;
    padding: 3px 8px;
    border-radius: 5px;
  }
  .badge--done {
    background: rgba(255, 255, 255, 0.06);
    color: #fff;
    border: 1px solid rgba(255, 255, 255, 0.1);
  }
  .badge--locked {
    background: transparent;
    color: var(--c-muted);
    border: 1px solid var(--c-border);
  }

  .pagination-wrap {
    display: flex;
    flex-direction: column;
    align-items: center;
    gap: 14px;
    margin-top: 32px;
    padding-top: 20px;
    border-top: 1px solid var(--c-border);
  }

  .pagination-info {
    font-size: 0.78rem;
    color: var(--c-sub);
    margin: 0;
  }
  .pagination-info strong {
    color: #fff;
    font-weight: 700;
  }

  .pagination {
    display: flex;
    align-items: center;
    gap: 6px;
  }

  .page-btn {
    width: 32px;
    height: 32px;
    border-radius: 8px;
    background: var(--c-card);
    border: 1px solid var(--c-border);
    color: #fff;
    display: flex;
    align-items: center;
    justify-content: center;
    cursor: pointer;
    transition: all 0.15s;
    flex-shrink: 0;
  }
  .page-btn:hover:not(:disabled) {
    border-color: var(--c-border-h);
    background: #1c1c1c;
  }
  .page-btn:disabled {
    opacity: 0.3;
    cursor: not-allowed;
  }

  .page-numbers {
    display: flex;
    align-items: center;
    gap: 4px;
    margin: 0 4px;
  }

  .page-num {
    min-width: 32px;
    height: 32px;
    padding: 0 8px;
    border-radius: 8px;
    background: transparent;
    border: 1px solid transparent;
    color: var(--c-sub);
    font-size: 0.8rem;
    font-weight: 600;
    cursor: pointer;
    transition: all 0.15s;
  }
  .page-num:hover:not(:disabled):not(.active) {
    background: var(--c-card);
    color: #fff;
    border-color: var(--c-border);
  }
  .page-num.active {
    background: var(--c-red);
    color: #fff;
  }
  .page-num.dots {
    cursor: default;
    color: var(--c-muted);
  }
  .page-btn:disabled {
    opacity: 0.35;
    cursor: not-allowed;
  }
</style>
