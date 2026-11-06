<template>
  <SessionBootstrapView v-if="showBootstrap" />
  <RouterView v-else v-slot="{ Component }">
    <Transition name="auth-flow" mode="out-in">
      <component v-if="Component" :is="Component" />
      <div v-else key="app-loading" class="app-loading-screen">
        <div class="app-loading-spinner" />
      </div>
    </Transition>
  </RouterView>
</template>

<script setup lang="ts">
  import { computed } from "vue"
  import { RouterView } from "vue-router"
  import { useAuthStore } from "@/stores/auth"
  import SessionBootstrapView from "@/components/common/SessionBootstrapView.vue"

  const authStore = useAuthStore()

  const showBootstrap = computed(
    () =>
      !authStore.isInitialized ||
      authStore.initError === "network" ||
      authStore.initError === "server",
  )
</script>

<style>
  /* ─── Auth Flow Smooth Page Transition ─── */
  .auth-flow-enter-active,
  .auth-flow-leave-active {
    transition:
      opacity 0.3s cubic-bezier(0.4, 0, 0.2, 1),
      transform 0.3s cubic-bezier(0.4, 0, 0.2, 1);
  }

  .auth-flow-leave-to {
    opacity: 0;
    transform: translateY(-12px);
  }

  .auth-flow-enter-from {
    opacity: 0;
    transform: translateY(12px);
  }

  .auth-flow-enter-to,
  .auth-flow-leave-from {
    opacity: 1;
    transform: translateY(0);
  }

  /* ─── Loading Screen for Protected Routes ─── */
  .app-loading-screen {
    min-height: 100vh;
    display: flex;
    align-items: center;
    justify-content: center;
    background: #0a0a0a;
  }

  .app-loading-spinner {
    width: 40px;
    height: 40px;
    border: 3px solid rgba(255, 255, 255, 0.1);
    border-top-color: var(--color-brand);
    border-radius: 50%;
    animation: app-spin 0.8s linear infinite;
  }

  @keyframes app-spin {
    to {
      transform: rotate(360deg);
    }
  }
</style>
