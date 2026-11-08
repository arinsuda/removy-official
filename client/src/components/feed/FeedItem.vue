<template>
  <article
    class="relative flex gap-4 rounded-2xl border border-white/5 bg-neutral-900/30 p-5 mb-5 transition-all duration-300 hover:border-white/10 hover:bg-neutral-900/50 shadow-md group"
  >
    <!-- Actor Avatar -->
    <RouterLink :to="actorLink" class="shrink-0">
      <AppAvatar
        :src="item.actor.avatar_url"
        :name="item.actor.display_name || item.actor.username"
        decorative
        class="transition-all duration-300 group-hover:ring-white/20"
      />
    </RouterLink>

    <div class="min-w-0 flex-1">
      <div class="flex items-start justify-between gap-2">
        <div>
          <!-- Actor Names & Date -->
          <div class="flex items-center flex-wrap gap-x-2 gap-y-1">
            <RouterLink
              :to="actorLink"
              class="font-semibold text-white hover:text-red-400 transition-colors"
            >
              {{ item.actor.display_name || item.actor.username }}
            </RouterLink>
            <span class="text-xs text-white/40">@{{ item.actor.username }}</span>
          </div>

          <!-- Activity Badge & Time -->
          <div class="mt-1 flex items-center gap-2">
            <span
              class="inline-flex items-center justify-center text-[10px] font-semibold border px-2.5 py-0.5 rounded-full"
              :class="badgeClasses"
            >
              <component :is="activityMeta.icon" class="mr-1 h-3 w-3 shrink-0" />
              {{ activityMeta.label }}
            </span>
            <span class="text-xs text-white/30" aria-hidden="true">·</span>
            <time :datetime="item.created_at" class="text-xs text-white/40">{{ relativeTime }}</time>
          </div>
        </div>

        <!-- Privacy & Delete Management Dropdown -->
        <div v-if="canManage" class="relative" ref="menuRef">
          <button
            type="button"
            class="flex h-7 w-7 items-center justify-center rounded-lg text-white/30 hover:bg-white/10 hover:text-white/70 transition-all cursor-pointer"
            :title="`ความเป็นส่วนตัว: ${visibilityLabel}`"
            @click="isMenuOpen = !isMenuOpen"
          >
            <component :is="activeVisibilityIcon" class="h-4 w-4" />
          </button>

          <!-- Dropdown Menu -->
          <Transition name="slide-up">
            <div
              v-if="isMenuOpen"
              class="absolute right-0 top-full z-30 mt-1.5 w-52 rounded-xl border border-white/10 bg-neutral-950 p-1.5 shadow-2xl backdrop-blur-xl"
            >
              <div
                class="px-2.5 py-1.5 text-[10px] font-bold uppercase tracking-wider text-white/40"
              >
                ตั้งค่าความส่วนตัว
              </div>
              <button
                v-for="opt in visibilityOptions"
                :key="opt.value"
                type="button"
                class="flex w-full items-center gap-2 rounded-lg px-2.5 py-2 text-left text-xs transition-colors hover:bg-white/5 cursor-pointer"
                :class="
                  item.visibility === opt.value
                    ? 'text-emerald-400 font-semibold bg-emerald-500/5'
                    : 'text-white/70'
                "
                @click="changeVisibility(opt.value)"
              >
                <component :is="opt.icon" class="h-3.5 w-3.5 shrink-0" />
                <span>{{ opt.label }}</span>
              </button>

              <div class="border-t border-white/5 my-1.5" />

              <button
                type="button"
                class="flex w-full items-center gap-2 rounded-lg px-2.5 py-2 text-left text-xs text-rose-400 transition-colors hover:bg-rose-500/10 cursor-pointer"
                @click="confirmDelete"
              >
                <Trash2 class="h-3.5 w-3.5 shrink-0" />
                <span>ลบกิจกรรมนี้</span>
              </button>
            </div>
          </Transition>
        </div>
      </div>

      <!-- Action status text (For feed clarity) -->
      <p class="mt-2.5 text-sm text-white/80 leading-relaxed">
        {{ actionDescription }}
      </p>

      <!-- Review content blockquote -->
      <div v-if="item.type === 'review_created' && item.message" class="mt-3 relative">
        <div
          class="border-l-2 border-emerald-500/40 bg-emerald-500/[0.02] px-4 py-3 rounded-r-xl text-sm leading-relaxed text-white/90 italic"
          :class="{ 'line-clamp-3 overflow-hidden': isTruncated && !isExpanded }"
        >
          "{{ item.message }}"
        </div>
        <button
          v-if="isTruncated"
          type="button"
          class="mt-2 text-xs font-semibold text-emerald-400 hover:text-emerald-300 transition-colors cursor-pointer"
          @click="isExpanded = !isExpanded"
        >
          {{ isExpanded ? "แสดงน้อยลง" : "อ่านรีวิวเพิ่มเติม..." }}
        </button>
      </div>

      <!-- Media Reference Card -->
      <RouterLink
        v-if="item.media"
        :to="mediaLink"
        class="mt-4 flex items-center gap-4 rounded-xl border border-white/5 bg-white/[0.01] p-3 transition-all duration-300 hover:border-red-500/20 hover:bg-white/[0.03] group/media shadow-sm"
      >
        <img
          v-if="item.media.poster_url"
          :src="item.media.poster_url"
          :alt="item.media.title"
          class="h-20 w-14 shrink-0 rounded-lg object-cover shadow-md transition-transform duration-300 group-hover/media:scale-105"
        />
        <div class="min-w-0 flex-1">
          <span class="text-[10px] uppercase tracking-wider text-red-500 font-semibold mb-1 block">
            {{ item.media.media_type === 'movie' ? 'ภาพยนตร์' : 'ซีรีส์' }}
          </span>
          <p class="truncate text-base font-semibold text-white group-hover/media:text-red-500 transition-colors duration-200">
            {{ item.media.title }}
          </p>
          <div class="mt-1">
            <MediaRating :ratings="item.media.ratings" :vote-average="item.media.vote_average" />
          </div>
        </div>
      </RouterLink>

      <!-- Target User Reference Card (For follow activities) -->
      <RouterLink
        v-if="item.type === 'user_followed' && item.target_user"
        :to="`/users/${item.target_user.id}`"
        class="mt-4 flex items-center gap-4 rounded-xl border border-white/5 bg-white/[0.01] p-3 transition-all duration-300 hover:border-white/10 hover:bg-white/[0.03] shadow-sm group/target"
      >
        <AppAvatar
          :src="item.target_user.avatar_url"
          :name="item.target_user.display_name || item.target_user.username"
          decorative
          class="transition-transform duration-300 group-hover/target:scale-105"
        />
        <div class="min-w-0 flex-1">
          <p class="truncate text-sm font-semibold text-white group-hover/target:text-red-400 transition-colors">
            {{ item.target_user.display_name || item.target_user.username }}
          </p>
          <p class="text-xs text-white/40">
            @{{ item.target_user.username }}
          </p>
        </div>
      </RouterLink>
    </div>
  </article>
</template>

<script setup lang="ts">
import MediaRating from "@/components/movie/MediaRating.vue"
import { computed, ref } from "vue";
import type { FeedItemResponse } from "@/types/feed";
import { getActivityMeta, mediaRoutePrefix } from "@/utils/feedActivity";
import { formatRelativeTime } from "@/utils/relativeTime";
import { useClickOutside } from "@/composables/useClickOutside";
import AppAvatar from "@/components/ui/AppAvatar.vue";
import { Globe, Users, Lock, Compass, Trash2 } from "lucide-vue-next";

const props = defineProps<{
  item: FeedItemResponse;
  canManage?: boolean;
}>();

const emit = defineEmits<{
  "update-visibility": [
    id: number,
    visibility: "default" | "public" | "followers" | "private",
  ];
  "delete-activity": [id: number];
}>();

const isMenuOpen = ref(false);
const menuRef = ref<HTMLElement | null>(null);
const isExpanded = ref(false);

useClickOutside(menuRef, () => (isMenuOpen.value = false));

const activityMeta = computed(() => getActivityMeta(props.item.type));
const relativeTime = computed(() => formatRelativeTime(props.item.created_at));
const actorLink = computed(() => `/users/${props.item.actor.id}`);

const mediaLink = computed(() => {
  if (!props.item.media) return "/";
  return `/${mediaRoutePrefix(props.item.media.media_type)}/${props.item.media.id}`;
});

const isTruncated = computed(() => {
  return props.item.message && props.item.message.length > 200;
});

const actionDescription = computed(() => {
  const mediaTitle = props.item.media?.title ? `"${props.item.media.title}"` : "";
  const targetName = props.item.target_user
    ? `${props.item.target_user.display_name || props.item.target_user.username}`
    : "";

  switch (props.item.type) {
    case "review_created":
      return `ได้เขียนบทวิจารณ์ให้กับภาพยนตร์ ${mediaTitle}`;
    case "review_commented":
      return `ได้แสดงความคิดเห็นในบทวิจารณ์ของภาพยนตร์ ${mediaTitle}`;
    case "review_liked":
      return `ได้ถูกใจบทวิจารณ์ของภาพยนตร์ ${mediaTitle}`;
    case "media_liked":
      return `ได้ถูกใจภาพยนตร์ ${mediaTitle}`;
    case "watchlist_added":
      return `ได้เพิ่มภาพยนตร์ ${mediaTitle} ลงในรายการภาพยนตร์ที่อยากดู (Watchlist)`;
    case "watched_added":
      return `ได้รับชมภาพยนตร์ ${mediaTitle} แล้ว`;
    case "achievement_unlocked":
      return `ได้รับความสำเร็จระดับใหม่: "${props.item.message}"`;
    case "user_followed":
      return `ได้เริ่มติดตามผู้ใช้ ${targetName}`;
    case "watch_log_created":
      return `ได้บันทึกการดูภาพยนตร์ ${mediaTitle}`;
    case "rewatched":
      return `ได้ดูภาพยนตร์ ${mediaTitle} ซ้ำ`;
    default:
      return props.item.message;
  }
});

const badgeClasses = computed(() => {
  switch (props.item.type) {
    case "review_created":
    case "review_commented":
      return "bg-emerald-500/10 text-emerald-400 border-emerald-500/20";
    case "review_liked":
    case "media_liked":
      return "bg-rose-500/10 text-rose-400 border-rose-500/20";
    case "watchlist_added":
      return "bg-blue-500/10 text-blue-400 border-blue-500/20";
    case "watched_added":
    case "watch_log_created":
    case "rewatched":
      return "bg-indigo-500/10 text-indigo-400 border-indigo-500/20";
    case "achievement_unlocked":
      return "bg-amber-500/10 text-amber-400 border-amber-500/20";
    case "user_followed":
      return "bg-purple-500/10 text-purple-400 border-purple-500/20";
    default:
      return "bg-white/5 text-white/50 border-white/10";
  }
});

const visibilityLabel = computed(() => {
  switch (props.item.visibility) {
    case "public":
      return "สาธารณะ";
    case "followers":
      return "ผู้ติดตามเท่านั้น";
    case "private":
      return "เฉพาะฉัน";
    case "default":
    default:
      return "ค่าเริ่มต้น";
  }
});

const activeVisibilityIcon = computed(() => {
  switch (props.item.visibility) {
    case "public":
      return Globe;
    case "followers":
      return Users;
    case "private":
      return Lock;
    case "default":
    default:
      return Compass;
  }
});

const visibilityOptions = [
  { value: "default" as const, label: "ค่าเริ่มต้น (ตามประเภทบัญชี)", icon: Compass },
  { value: "public" as const, label: "สาธารณะ", icon: Globe },
  { value: "followers" as const, label: "ผู้ติดตามเท่านั้น", icon: Users },
  { value: "private" as const, label: "เฉพาะฉัน", icon: Lock },
];

function changeVisibility(
  visibility: "default" | "public" | "followers" | "private"
) {
  emit("update-visibility", props.item.id, visibility);
  isMenuOpen.value = false;
}

function confirmDelete() {
  if (confirm("คุณแน่ใจหรือไม่ที่จะลบกิจกรรมนี้ออกจากฟีดอย่างถาวร?")) {
    emit("delete-activity", props.item.id);
  }
  isMenuOpen.value = false;
}
</script>

<style scoped>
.slide-up-enter-active,
.slide-up-leave-active {
  transition:
    transform 0.15s ease,
    opacity 0.15s ease;
}
.slide-up-enter-from,
.slide-up-leave-to {
  transform: translateY(4px);
  opacity: 0;
}
</style>
