import { defineStore } from "pinia";
import { ref, computed } from "vue";
import { feedApi } from "@/api/endpoints/feed";
import { useSocket } from "@/composables/useSocket";
import { logApiError } from "@/utils/logApiError";
import type {
  FeedItemResponse,
  FeedPaginationMeta,
  ActivitySettingsResponse,
  UpdateActivitySettingsRequest,
} from "@/types/feed";

type SocketHandler = (...args: unknown[]) => void;

const DEFAULT_PAGINATION: FeedPaginationMeta = {
  page: 1,
  limit: 20,
  total: 0,
  total_pages: 0,
};

export const useFeedStore = defineStore("feed", () => {
  const items = ref<FeedItemResponse[]>([]);
  const pagination = ref<FeedPaginationMeta>({ ...DEFAULT_PAGINATION });
  const loading = ref(false);
  const loadingMore = ref(false);
  const error = ref<string | null>(null);
  const selectedCategory = ref<string>("");

  const settings = ref<ActivitySettingsResponse | null>(null);
  const settingsLoading = ref(false);
  const settingsError = ref<string | null>(null);

  const newItemsCount = ref(0);
  let socketBound = false;

  const hasMore = computed(
    () => pagination.value.page < pagination.value.total_pages
  );
  const isEmpty = computed(() => !loading.value && items.value.length === 0);

  const topActivityId = computed(() => {
    const first = items.value[0];
    return first ? first.id : null;
  });


  async function checkNewItems() {
    const topId = topActivityId.value;
    if (!topId) return;
    try {
      const res = await feedApi.getNewCount(topId, selectedCategory.value || undefined);
      newItemsCount.value = res.data.count;
    } catch (err) {
      logApiError("checkNewItems failed:", err);
    }
  }

  function handleActivityUpdated(payload: { id: number; visibility: string }) {
    const item = items.value.find((i) => i.id === payload.id);
    if (item) {
      item.visibility = payload.visibility as FeedItemResponse["visibility"];
    }
  }

  function handleActivityRemoved(payload: { id: number }) {
    const idx = items.value.findIndex((i) => i.id === payload.id);
    if (idx !== -1) {
      items.value.splice(idx, 1);
      pagination.value.total = Math.max(0, pagination.value.total - 1);
    }
  }

  function bindSocket() {
    if (socketBound) return;
    const { connect, on } = useSocket();
    connect();

    on("feed:refresh_required", checkNewItems);
    on<{ id: number; visibility: string }>(
      "feed:activity_updated",
      handleActivityUpdated
    );
    on<{ id: number }>("feed:activity_removed", handleActivityRemoved);

    socketBound = true;
  }

  function unbindSocket() {
    if (!socketBound) return;
    const { off } = useSocket();
    off("feed:refresh_required", checkNewItems as SocketHandler);
    off("feed:activity_updated", handleActivityUpdated as SocketHandler);
    off("feed:activity_removed", handleActivityRemoved as SocketHandler);
    socketBound = false;
  }

  async function showNewItems() {
    newItemsCount.value = 0;
    await refresh();
    window.scrollTo({ top: 0, behavior: "smooth" });
  }

  async function fetchFeed(page = 1) {
    if (loading.value || loadingMore.value) return;

    const limit = pagination.value.limit;
    if (page === 1) {
      loading.value = true;
      newItemsCount.value = 0;
    } else {
      loadingMore.value = true;
    }
    error.value = null;

    try {
      const res = await feedApi.getFeed({
        page,
        limit,
        category: selectedCategory.value || undefined,
      });
      const next = res.data.items ?? [];
      items.value = page === 1 ? next : [...items.value, ...next];
      pagination.value = {
        ...DEFAULT_PAGINATION,
        ...res.data.pagination,
      };
    } catch (err) {
      logApiError("fetchFeed failed:", err);
      error.value = "load_failed";
    } finally {
      loading.value = false;
      loadingMore.value = false;
    }
  }

  async function setCategory(cat: string) {
    selectedCategory.value = cat;
    items.value = [];
    pagination.value = { ...DEFAULT_PAGINATION };
    await fetchFeed(1);
  }

  async function loadMore() {
    if (!hasMore.value || loadingMore.value || loading.value) return;
    await fetchFeed(pagination.value.page + 1);
  }

  async function refresh() {
    await fetchFeed(1);
  }

  async function updateVisibility(
    id: number,
    visibility: "default" | "public" | "followers" | "private"
  ) {
    const item = items.value.find((i) => i.id === id);
    if (!item) return;

    const prevVisibility = item.visibility;
    item.visibility = visibility;

    try {
      await feedApi.updateVisibility(id, visibility);
    } catch (err) {
      item.visibility = prevVisibility;
      logApiError("updateVisibility failed:", err);
      throw err;
    }
  }

  async function deleteActivity(id: number) {
    const idx = items.value.findIndex((i) => i.id === id);
    if (idx === -1) return;

    const [removed] = items.value.splice(idx, 1);
    if (!removed) return;
    pagination.value.total = Math.max(0, pagination.value.total - 1);

    try {
      await feedApi.deleteActivity(id);
    } catch (err) {
      items.value.splice(idx, 0, removed);
      pagination.value.total++;
      logApiError("deleteActivity failed:", err);
      throw err;
    }
  }

  async function fetchSettings() {
    if (settingsLoading.value) return;
    settingsLoading.value = true;
    settingsError.value = null;
    try {
      const res = await feedApi.getSettings();
      settings.value = res.data;
    } catch (err) {
      logApiError("fetchSettings failed:", err);
      settingsError.value = "load_failed";
    } finally {
      settingsLoading.value = false;
    }
  }

  async function updateSettings(patch: UpdateActivitySettingsRequest) {
    const prev = settings.value;
    if (settings.value) {
      settings.value = {
        ...settings.value,
        ...patch,
      } as ActivitySettingsResponse;
    }

    try {
      const res = await feedApi.updateSettings(patch);
      settings.value = res.data;
      return true;
    } catch (err) {
      settings.value = prev;
      logApiError("updateSettings failed:", err);
      return false;
    }
  }

  function reset() {
    items.value = [];
    pagination.value = { ...DEFAULT_PAGINATION };
    loading.value = false;
    loadingMore.value = false;
    error.value = null;
    newItemsCount.value = 0;
  }

  return {
    items,
    pagination,
    loading,
    loadingMore,
    error,
    hasMore,
    isEmpty,
    newItemsCount,

    settings,
    settingsLoading,
    settingsError,
    selectedCategory,

    fetchFeed,
    loadMore,
    refresh,
    updateVisibility,
    deleteActivity,
    fetchSettings,
    updateSettings,
    reset,
    bindSocket,
    unbindSocket,
    showNewItems,
    setCategory,
  };
});

