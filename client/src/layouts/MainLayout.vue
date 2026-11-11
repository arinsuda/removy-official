<template>
  <div class="app-shell">
    <AppHeader />

    <main class="page-content">
      <RouterView v-slot="{ Component }">
        <Transition name="page" mode="out-in">
          <component :is="Component" />
        </Transition>
      </RouterView>
    </main>

    <AuthPromptModal />
  </div>
</template>

<script setup lang="ts">
import { watch } from "vue";
import { useAuthStore } from "@/stores/auth";
import { useNotificationStore } from "@/stores/notification";
import AuthPromptModal from "@/components/auth/AuthPromptModal.vue";
import AppHeader from "@/components/layout/AppHeader.vue";

const authStore = useAuthStore();
const notificationStore = useNotificationStore();

// ต่อ/ตัด socket ตามสถานะล็อกอิน (ครอบทั้งกรณีโหลดหน้าแรกตอน user
// ยังไม่พร้อม เพราะ fetchMe() ยังทำงานอยู่ใน router guard)
watch(
  () => authStore.user?.id,
  (userId) => {
    if (userId) {
      notificationStore.fetchUnreadCount();
      notificationStore.bindSocket();
    } else {
      notificationStore.reset();
    }
  },
  { immediate: true },
);
</script>

<style scoped>
.app-shell {
  min-height: 100vh; min-height: 100dvh;
  background: var(--color-surface);
}

/* Matches the header height in AppHeader (56 mobile, 64 from 768). */
.page-content {
  padding-top: 56px;
}
@media (min-width: 768px) {
  .page-content {
    padding-top: 64px;
  }
}
</style>
