import { ref, onUnmounted } from "vue";
import { movieApi } from "@/api/api";
import {
  resolveTrailerCandidates,
  useTrailerPreview,
  type ResolvedTrailer,
} from "@/composables/useTrailerPreview";

export interface HoverPreviewItem {
  id: number;
  mediaType?: "movie" | "tv";
}

interface Options {
  showDelay?: number;
  mountDelay?: number;
  columns?: number;
}

export function useHoverPreviewGrid(options: Options = {}) {
  const { showDelay = 200, mountDelay = 500, columns = 5 } = options;

  const hoveredId = ref<number | null>(null);
  const cardStates = new Map<number, ReturnType<typeof useTrailerPreview>>();
  const cardTrailers = new Map<number, ResolvedTrailer | null>();
  let insidePopup = false;
  let showTimer: ReturnType<typeof setTimeout> | null = null;

  function getState(id: number) {
    if (!cardStates.has(id)) {
      cardStates.set(id, useTrailerPreview({ mountDelay }));
    }
    return cardStates.get(id)!;
  }

  function getTrailer(id: number): ResolvedTrailer | null {
    return cardTrailers.get(id) ?? null;
  }

  async function fetchAndCacheTrailer(item: HoverPreviewItem) {
    if (cardTrailers.has(item.id)) return;
    cardTrailers.set(item.id, null);
    try {
      const fetchVideos =
        item.mediaType === "tv" ? movieApi.getTVVideos : movieApi.getVideos;
      const res = await fetchVideos(item.id);
      const videos = res.data?.results ?? [];
      const candidates = resolveTrailerCandidates(videos);
      getState(item.id).setCandidates(candidates);
      cardTrailers.set(item.id, null);
      if (hoveredId.value === item.id && candidates.length > 0) {
        getState(item.id).scheduleMount();
      }
    } catch {
      cardTrailers.set(item.id, null);
    }
  }

  function onCardEnter(item: HoverPreviewItem) {
    clearTimeout(showTimer ?? undefined);
    insidePopup = false;
    fetchAndCacheTrailer(item);
    showTimer = setTimeout(() => {
      hoveredId.value = item.id;
      const trailer = getTrailer(item.id);
      if (trailer) getState(item.id).scheduleMount();
    }, showDelay);
  }

  function onCardLeave(id: number) {
    clearTimeout(showTimer ?? undefined);
    setTimeout(() => {
      if (!insidePopup) closeCard(id);
    }, 80);
  }

  function onPopupEnter() {
    insidePopup = true;
  }

  function onPopupLeave(id: number) {
    insidePopup = false;
    closeCard(id);
  }

  function closeCard(id: number) {
    if (hoveredId.value !== id) return;
    hoveredId.value = null;
    cardStates.get(id)?.unmount();
  }

  function getPopupPos(index: number) {
    const col = index % columns;
    if (col === 0) return "popup--right";
    if (col === columns - 1) return "popup--left";
    return "popup--center";
  }

  onUnmounted(() => {
    clearTimeout(showTimer ?? undefined);
    cardStates.forEach((s) => s.unmount());
    cardStates.clear();
    cardTrailers.clear();
  });

  return {
    hoveredId,
    getState,
    getTrailer,
    onCardEnter,
    onCardLeave,
    onPopupEnter,
    onPopupLeave,
    getPopupPos,
  };
}
