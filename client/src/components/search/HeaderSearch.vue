<template>
  <div ref="root" class="hs" :class="{ 'hs--open': open }">
    <button v-if="!open" type="button" class="hs-toggle" :aria-label="t('navigation.search.open')" @click="openSearch">
      <Search :size="18" aria-hidden="true" />
    </button>

    <div v-else class="hs-field">
      <Search :size="16" class="hs-icon" aria-hidden="true" />
      <input
        ref="input"
        v-model="query"
        data-focus-custom
        type="text"
        class="hs-input"
        role="combobox"
        autocomplete="off"
        autocapitalize="off"
        spellcheck="false"
        enterkeyhint="search"
        aria-autocomplete="list"
        :aria-label="t('navigation.searchPlaceholder')"
        :aria-expanded="showPanel"
        :aria-controls="LISTBOX_ID"
        :aria-activedescendant="activeIndex >= 0 ? optionId(activeIndex) : undefined"
        :placeholder="t('navigation.searchPlaceholder')"
        @keydown="onKeydown"
      />
      <button v-if="query" type="button" class="hs-btn hs-clear" :aria-label="t('navigation.search.clear')" @click="clear">
        <X :size="16" aria-hidden="true" />
      </button>
      <button type="button" class="hs-btn hs-close" :aria-label="t('navigation.search.close')" @click="close">
        <span>{{ t("navigation.search.cancel") }}</span>
      </button>
    </div>

    <div v-if="showPanel" class="hs-panel">
      <p class="sr-only" role="status">{{ statusText }}</p>

      <ul v-if="status === 'loading'" class="hs-list" aria-hidden="true">
        <li v-for="n in 3" :key="n" class="hs-skel">
          <SkeletonBlock width="40px" height="60px" radius="4px" />
          <span class="hs-skel-text">
            <SkeletonBlock width="70%" height="14px" radius="4px" />
            <SkeletonBlock width="45%" height="12px" radius="4px" />
          </span>
        </li>
      </ul>

      <div v-else-if="status === 'error'" class="hs-state" role="alert">
        <p>{{ t("navigation.search.error") }}</p>
        <button type="button" class="hs-retry" @click="retry">{{ t("navigation.search.retry") }}</button>
      </div>

      <div v-else-if="results.length === 0" class="hs-state">
        <p>{{ t("navigation.search.noResults", { query: trimmed }) }}</p>
      </div>

      <ul v-else :id="LISTBOX_ID" class="hs-list" role="listbox" :aria-label="t('navigation.search.results')">
        <SearchSuggestionItem
          v-for="(item, i) in results"
          :id="optionId(i)"
          :key="`${item.type}-${item.id}`"
          :item="item"
          :query="trimmed"
          :active="activeIndex === i"
          @select="goToItem(item)"
          @hover="activeIndex = i"
        />
        <li
          :id="optionId(results.length)"
          role="option"
          :aria-selected="activeIndex === results.length"
          class="hs-viewall"
          :class="{ 'hs-viewall--active': activeIndex === results.length }"
          @mousedown.prevent
          @click="viewAll"
          @mouseenter="activeIndex = results.length"
        >
          <span class="hs-viewall-text">{{ t("navigation.viewAllResults", { query: trimmed }) }}</span>
          <kbd class="hs-kbd" aria-hidden="true">{{ t("navigation.search.enter") }}</kbd>
          <ArrowRight :size="16" aria-hidden="true" />
        </li>
      </ul>
    </div>
  </div>
</template>

<script setup lang="ts">
import { computed, nextTick, onBeforeUnmount, onMounted, ref } from "vue"
import { useI18n } from "vue-i18n"
import { ArrowRight, Search, X } from "lucide-vue-next"
import SkeletonBlock from "@/components/ui/skeleton/SkeletonBlock.vue"
import { useHeaderSearch } from "@/composables/useHeaderSearch"
import SearchSuggestionItem from "./SearchSuggestionItem.vue"

const LISTBOX_ID = "header-search-listbox"
const optionId = (i: number) => `header-search-option-${i}`

const { t } = useI18n()
const { open, query, status, results, activeIndex, trimmed, showPanel, retry, close, goToItem, viewAll, onKeydown } =
  useHeaderSearch()

const root = ref<HTMLElement | null>(null)
const input = ref<HTMLInputElement | null>(null)

const statusText = computed(() => {
  if (status.value === "loading") return t("navigation.searching")
  if (status.value === "error") return t("navigation.search.error")
  return t("navigation.search.resultCount", { count: results.value.length }, results.value.length)
})

function openSearch() {
  open.value = true
  nextTick(() => input.value?.focus())
}

function clear() {
  query.value = ""
  input.value?.focus()
}

function onPointerDown(e: PointerEvent) {
  if (open.value && root.value && !root.value.contains(e.target as Node)) close()
}

onMounted(() => document.addEventListener("pointerdown", onPointerDown))
onBeforeUnmount(() => document.removeEventListener("pointerdown", onPointerDown))
</script>

<style scoped>
.hs {
  position: relative;
  display: flex;
  align-items: center;
}
.hs-toggle {
  display: flex;
  align-items: center;
  justify-content: center;
  min-width: 2.75rem;
  min-height: 2.75rem;
  padding: 0.5rem 0.6rem;
  border: none;
  background: none;
  color: var(--color-text-secondary);
  cursor: pointer;
  transition: color 120ms;
}
.hs-toggle:hover {
  color: var(--color-text-primary);
}
.hs-field {
  display: flex;
  align-items: center;
  gap: 0.25rem;
  width: clamp(12rem, 24vw, 17.5rem);
  height: 40px;
  padding: 0 0.5rem 0 0.9rem;
  background: var(--color-surface-2);
  border: 1px solid var(--color-border-strong);
  border-radius: 9999px;
  transition: border-color 120ms, box-shadow 120ms;
}
/* One ring on the pill; the input opts out of the global focus outline. */
.hs-field:focus-within {
  border-color: var(--color-focus);
  box-shadow: 0 0 0 3px color-mix(in srgb, var(--color-focus) 22%, transparent);
}
.hs-icon {
  flex-shrink: 0;
  color: var(--color-text-muted);
}
.hs-input {
  flex: 1;
  min-width: 0;
  height: 100%;
  padding: 0 0.25rem 0 0.4rem;
  border: none;
  outline: none;
  background: none;
  color: var(--color-text-primary);
  font-size: 0.875rem;
}
.hs-input::placeholder {
  color: var(--color-text-muted);
}
.hs-btn {
  display: flex;
  align-items: center;
  justify-content: center;
  width: 28px;
  height: 28px;
  flex-shrink: 0;
  border: none;
  border-radius: 9999px;
  background: none;
  color: var(--color-text-secondary);
  cursor: pointer;
}
.hs-btn:hover {
  background: var(--color-surface-3);
  color: var(--color-text-primary);
}
.hs-close {
  display: none;
}

.hs-panel {
  position: absolute;
  top: calc(100% + 8px);
  right: 0;
  width: min(420px, 92vw);
  max-height: min(520px, calc(100dvh - 90px));
  overflow-y: auto;
  padding: 6px;
  background: var(--color-surface-2);
  border: 1px solid var(--color-border-strong);
  border-radius: 14px;
  box-shadow: 0 16px 48px color-mix(in srgb, black 55%, transparent);
  z-index: var(--z-popover);
  overscroll-behavior: contain;
}
.hs-list {
  margin: 0;
  padding: 0;
  list-style: none;
}
.hs-skel {
  display: flex;
  align-items: center;
  gap: 12px;
  min-height: 72px;
  padding: 6px 12px;
}
.hs-skel-text {
  display: flex;
  flex: 1;
  flex-direction: column;
  gap: 8px;
}
.hs-state {
  padding: 24px 16px;
  text-align: center;
  font-size: 0.875rem;
  color: var(--color-text-secondary);
  overflow-wrap: anywhere;
}
.hs-state p {
  margin: 0;
}
.hs-retry {
  margin-top: 12px;
  min-height: 44px;
  padding: 0 20px;
  border: 1px solid var(--color-border-strong);
  border-radius: 9999px;
  background: none;
  color: var(--color-text-primary);
  font-size: 0.8125rem;
  font-weight: 600;
  cursor: pointer;
}
.hs-retry:hover {
  background: var(--color-surface-3);
}
.hs-viewall {
  display: flex;
  align-items: center;
  gap: 8px;
  min-height: 48px;
  margin-top: 4px;
  padding: 0 14px;
  border-top: 1px solid var(--color-border);
  border-radius: 8px;
  color: var(--color-text-primary);
  font-size: 0.8125rem;
  font-weight: 600;
  cursor: pointer;
}
.hs-viewall:hover,
.hs-viewall--active {
  background: var(--color-surface-3);
}
.hs-viewall-text {
  flex: 1;
  min-width: 0;
  overflow: hidden;
  white-space: nowrap;
  text-overflow: ellipsis;
}
.hs-kbd {
  display: none;
  padding: 2px 6px;
  border: 1px solid var(--color-border-strong);
  border-radius: 4px;
  font-family: inherit;
  font-size: 0.6875rem;
  color: var(--color-text-muted);
}
@media (hover: hover) and (min-width: 768px) {
  .hs-kbd {
    display: inline-block;
  }
}

/* Mobile: the open box becomes a bar over the header; the panel is a full-width sheet under it. */
@media (max-width: 767.98px) {
  .hs--open {
    position: absolute;
    inset: 0;
    padding: 0 0.75rem;
    background: var(--color-surface);
    z-index: 2;
  }
  .hs-field {
    flex: 1;
    width: auto;
    height: 44px;
  }
  .hs-input {
    font-size: 1rem; /* 16px stops iOS zooming on focus */
  }
  .hs-btn {
    width: 44px;
    height: 44px;
  }
  .hs-close {
    display: flex;
    width: auto;
    padding: 0 0.5rem;
    font-size: 0.8125rem;
    font-weight: 600;
  }
  .hs-panel {
    top: 100%;
    left: 0;
    right: 0;
    width: auto;
    max-height: calc(100dvh - var(--hs-header, 60px));
    border-width: 1px 0 0;
    border-radius: 0 0 14px 14px;
  }
}
@media (max-width: 639.98px) {
  .hs {
    --hs-header: 56px;
  }
}
</style>
