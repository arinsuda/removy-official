<template>
  <div
    class="w-76 rounded-2xl border border-white/10 bg-neutral-950/90 p-5 shadow-2xl backdrop-blur-xl"
  >
    <h3 class="mb-1 text-sm font-semibold text-white">การแสดงกิจกรรมในฟีด</h3>
    <p class="mb-4 text-xs text-white/40 leading-relaxed">
      เลือกว่ากิจกรรมประเภทไหนของคุณจะไปแสดงในฟีดของคนที่ติดตามคุณอยู่
    </p>

    <SkeletonList v-if="loading" :count="3" />

    <p v-else-if="settingsError" class="py-6 text-center text-xs text-red-400">
      {{ settingsError }}
    </p>

    <ul v-else-if="settings" class="space-y-1">
      <li
        v-for="opt in options"
        :key="opt.key"
        class="flex items-center justify-between gap-3 hover:bg-white/[0.03] -mx-2 px-2.5 py-2.5 rounded-xl transition-all duration-200 group"
      >
        <div class="flex items-center gap-3">
          <component
            :is="opt.icon"
            class="h-4 w-4 text-white/30 group-hover:text-red-400 transition-colors duration-200"
          />
          <span class="text-xs font-semibold text-white/80 select-none group-hover:text-white transition-colors duration-200">
            {{ opt.label }}
          </span>
        </div>

        <!-- Custom Styled Toggle Switch -->
        <button
          type="button"
          role="switch"
          :aria-checked="settings[opt.key]"
          class="relative h-5.5 w-10 shrink-0 rounded-full transition-all duration-300 border focus:outline-none cursor-pointer"
          :class="settings[opt.key]
            ? 'bg-emerald-500/20 border-emerald-500/40 shadow-[0_0_10px_rgba(16,185,129,0.2)]'
            : 'bg-white/5 border-white/10'"
          @click="toggle(opt.key)"
        >
          <span
            class="absolute top-0.5 left-0.5 h-3.5 w-3.5 rounded-full transition-all duration-300 shadow-md"
            :class="settings[opt.key] ? 'translate-x-4.5 bg-emerald-400' : 'translate-x-0 bg-white/40'"
          />
        </button>
      </li>
    </ul>
  </div>
</template>

<script setup lang="ts">
import SkeletonList from "@/components/ui/skeleton/SkeletonList.vue"
import { onMounted } from "vue";
import { storeToRefs } from "pinia";
import { useFeedStore } from "@/stores/feed";
import type { ActivitySettingsResponse } from "@/types/feed";
import {
  PenTool,
  MessageSquare,
  Heart,
  Film,
  Bookmark,
  CheckCircle2,
  Trophy,
  UserPlus
} from "lucide-vue-next";

const store = useFeedStore();
const {
  settings,
  settingsLoading: loading,
  settingsError,
} = storeToRefs(store);

const options = [
  { key: "review_created" as const, label: "สร้างรีวิวใหม่", icon: PenTool },
  { key: "review_commented" as const, label: "แสดงความคิดเห็นในรีวิว", icon: MessageSquare },
  { key: "review_liked" as const, label: "ถูกใจรีวิว", icon: Heart },
  { key: "media_liked" as const, label: "ถูกใจหนัง/ซีรีส์", icon: Film },
  { key: "watchlist_added" as const, label: "เพิ่มลงวอทช์ลิสต์", icon: Bookmark },
  { key: "watched_added" as const, label: "ทำเครื่องหมายว่าดูแล้ว", icon: CheckCircle2 },
  { key: "achievement_unlocked" as const, label: "ปลดล็อกความสำเร็จ", icon: Trophy },
  { key: "user_followed" as const, label: "ติดตามผู้ใช้คนอื่น", icon: UserPlus },
];

function toggle(key: keyof ActivitySettingsResponse) {
  if (!settings.value) return;
  store.updateSettings({ [key]: !settings.value[key] });
}

onMounted(() => {
  if (!settings.value) store.fetchSettings();
});
</script>
