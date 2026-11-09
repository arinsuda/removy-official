<template>
  <div class="oauth-divider form-item">
    <span class="oauth-divider__line" />
    <span class="oauth-divider__text">{{ t("auth.common.or") }}</span>
    <span class="oauth-divider__line" />
  </div>

  <AppButton
    v-bind="$attrs"
    variant="outline"
    size="lg"
    block
    class="form-item"
    :disabled="disabled"
  >
    <svg viewBox="0 0 24 24" width="18" height="18" aria-hidden="true">
      <!-- hardcode-ok: Google brand colors -->
      <path fill="#4285F4" d="M22.56 12.25c0-.78-.07-1.53-.2-2.25H12v4.26h5.92c-.26 1.37-1.04 2.53-2.21 3.31v2.77h3.57c2.08-1.92 3.28-4.74 3.28-8.09z"/>
      <!-- hardcode-ok: Google brand colors -->
      <path fill="#34A853" d="M12 23c2.97 0 5.46-.98 7.28-2.66l-3.57-2.77c-.98.66-2.23 1.06-3.71 1.06-2.86 0-5.29-1.93-6.16-4.53H2.18v2.84C3.99 20.53 7.7 23 12 23z"/>
      <!-- hardcode-ok: Google brand colors -->
      <path fill="#FBBC05" d="M5.84 14.1c-.22-.66-.35-1.36-.35-2.1s.13-1.44.35-2.1V7.06H2.18C1.43 8.55 1 10.22 1 12s.43 3.45 1.18 4.94l2.85-2.22.81-.62z"/>
      <!-- hardcode-ok: Google brand colors -->
      <path fill="#EA4335" d="M12 5.38c1.62 0 3.06.56 4.21 1.64l3.15-3.15C17.45 2.09 14.97 1 12 1 7.7 1 3.99 3.47 2.18 7.06l3.66 2.84c.87-2.6 3.3-4.52 6.16-4.52z"/>
    </svg>
    {{ t("auth.common.google") }}
  </AppButton>
</template>

<script setup lang="ts">
import { useI18n } from "vue-i18n"
import AppButton from "@/components/ui/AppButton.vue"

defineProps<{ disabled?: boolean }>()
const { t } = useI18n()
// The template has two roots, so listeners are forwarded to the button explicitly.
defineOptions({ inheritAttrs: false })
</script>

<style scoped>
.oauth-divider {
  display: flex;
  align-items: center;
  gap: 12px;
}

.oauth-divider__line {
  flex: 1;
  height: 1px;
  background: var(--color-border-strong);
}

.oauth-divider__text {
  font-size: 0.75rem;
  font-weight: 600;
  letter-spacing: 1px;
  color: var(--color-text-muted);
}
</style>
