<template>
  <div class="session-bootstrap-container">
    <div class="session-bootstrap-card">
      <div class="brand-logo">
        <img src="/popcorn-full-512.png" alt="POPCORN" style="display:block;width:140px;margin:0 auto" />
      </div>

      <!-- Loading State -->
      <div v-if="isLoading" class="bootstrap-status">
        <div class="bootstrap-spinner" />
        <p class="bootstrap-text">{{ $t("session.checking") }}</p>
      </div>

      <!-- Error State (Network / Server unavailable) -->
      <div v-else-if="hasError" class="bootstrap-error">
        <div class="error-icon">
          <svg
            xmlns="http://www.w3.org/2000/svg"
            width="32"
            height="32"
            viewBox="0 0 24 24"
            fill="none"
            stroke="currentColor"
            stroke-width="2"
            stroke-linecap="round"
            stroke-linejoin="round"
          >
            <path
              d="m21.73 18-8-14a2 2 0 0 0-3.48 0l-8 14A2 2 0 0 0 4 21h16a2 2 0 0 0 1.73-3Z"
            />
            <line x1="12" y1="9" x2="12" y2="13" />
            <line x1="12" y1="17" x2="12.01" y2="17" />
          </svg>
        </div>
        <h3 class="error-title">{{ $t("session.error") }}</h3>
        <p class="error-detail">{{ $t("session.errorDetail") }}</p>
        <button
          type="button"
          class="btn-retry"
          :disabled="isRetrying"
          @click="handleRetry"
        >
          <span v-if="isRetrying" class="btn-spinner" />
          <span v-else>{{ $t("session.retry") }}</span>
        </button>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
  import { ref, computed } from "vue"
  import { useAuthStore } from "@/stores/auth"

  const authStore = useAuthStore()
  const isRetrying = ref(false)

  const hasError = computed(
    () =>
      authStore.initError === "network" || authStore.initError === "server",
  )

  const isLoading = computed(() => !authStore.isInitialized || isRetrying.value)

  async function handleRetry() {
    isRetrying.value = true
    try {
      await authStore.initialize()
    } finally {
      isRetrying.value = false
    }
  }
</script>

<style scoped>
  .session-bootstrap-container {
    min-height: 100vh;
    display: flex;
    align-items: center;
    justify-content: center;
    background: #0a0a0a;
    color: #ffffff;
    font-family: inherit;
    padding: 1.5rem;
  }

  .session-bootstrap-card {
    display: flex;
    flex-direction: column;
    align-items: center;
    justify-content: center;
    text-align: center;
    max-width: 400px;
    width: 100%;
    gap: 2rem;
  }

  .brand-logo {
    font-size: 2.5rem;
    font-weight: 800;
    letter-spacing: -0.025em;
  }

  .brand-movie {
    color: #ffffff;
  }

  .brand-hub {
    color: var(--color-brand);
  }

  .bootstrap-status {
    display: flex;
    flex-direction: column;
    align-items: center;
    gap: 1.25rem;
  }

  .bootstrap-spinner {
    width: 44px;
    height: 44px;
    border: 3px solid rgba(255, 255, 255, 0.1);
    border-top-color: var(--color-brand);
    border-radius: 50%;
    animation: bootstrap-spin 0.8s linear infinite;
  }

  @keyframes bootstrap-spin {
    to {
      transform: rotate(360deg);
    }
  }

  .bootstrap-text {
    font-size: 1rem;
    color: #a3a3a3;
    margin: 0;
  }

  .bootstrap-error {
    display: flex;
    flex-direction: column;
    align-items: center;
    gap: 1rem;
    animation: fadeIn 0.3s ease-out;
  }

  .error-icon {
    color: var(--color-brand);
    margin-bottom: 0.25rem;
  }

  .error-title {
    font-size: 1.25rem;
    font-weight: 600;
    margin: 0;
    color: #ffffff;
  }

  .error-detail {
    font-size: 0.875rem;
    color: #a3a3a3;
    margin: 0;
    line-height: 1.5;
  }

  .btn-retry {
    margin-top: 0.5rem;
    padding: 0.625rem 1.75rem;
    background-color: var(--color-brand);
    color: #ffffff;
    border: none;
    border-radius: 6px;
    font-size: 0.9375rem;
    font-weight: 600;
    cursor: pointer;
    transition: background-color 0.2s ease;
    display: flex;
    align-items: center;
    justify-content: center;
    min-width: 120px;
    min-height: 40px;
  }

  .btn-retry:hover:not(:disabled) {
    background-color: var(--color-brand-hover);
  }

  .btn-retry:disabled {
    opacity: 0.7;
    cursor: not-allowed;
  }

  .btn-spinner {
    width: 18px;
    height: 18px;
    border: 2px solid rgba(255, 255, 255, 0.3);
    border-top-color: #ffffff;
    border-radius: 50%;
    animation: bootstrap-spin 0.8s linear infinite;
  }

  @keyframes fadeIn {
    from {
      opacity: 0;
      transform: translateY(8px);
    }
    to {
      opacity: 1;
      transform: translateY(0);
    }
  }
</style>
