<template>
  <Teleport to="body">
  <div class="toast-container" aria-live="polite">
    <TransitionGroup name="toast-list">
      <div
        v-for="toast in toasts"
        :key="toast.id"
        class="toast-item"
        :class="[`toast-item--${toast.type}`]"
      >
        <div class="toast-icon">
          <CheckCircle v-if="toast.type === 'success'" :size="18" />
          <XCircle v-else-if="toast.type === 'error'" :size="18" />
          <AlertCircle v-else-if="toast.type === 'warning'" :size="18" />
          <Info v-else :size="18" />
        </div>

        <div class="toast-content">
          <div class="toast-message">{{ toast.message }}</div>
        </div>

        <button class="toast-close-btn" @click="removeToast(toast.id)">
          <X :size="14" />
        </button>
      </div>
    </TransitionGroup>
  </div>
  </Teleport>
</template>

<script setup lang="ts">
  import { storeToRefs } from "pinia"
  import { CheckCircle, XCircle, AlertCircle, Info, X } from "lucide-vue-next"
  import { useUIStore } from "@/stores/ui"

  // Single renderer for the ui store's toasts; mounted once in App.vue so every layout shows them.
  const ui = useUIStore()
  const { toasts } = storeToRefs(ui)
  const removeToast = ui.removeToast
</script>

<style scoped>
  /* Container วางตำแหน่งอยู่ภายใต้ปุ่มกระดิ่งแบบ Absolute */
  .toast-container {
    /* Teleported to body: inside the fixed navbar it could never rise above a modal. */
    position: fixed;
    top: calc(76px + env(safe-area-inset-top));
    right: max(16px, env(safe-area-inset-right));
    width: 320px;
    max-width: calc(100vw - 32px - env(safe-area-inset-right));
    z-index: var(--z-toast);
    display: flex;
    flex-direction: column;
    gap: 0.5rem;
    pointer-events: none; /* ป้องกันการบล็อคเมาส์กรณีกล่องใสว่างเปล่า */
  }

  /* สไตล์ของตัว Toast Item แต่ละชิ้น */
  .toast-item {
    pointer-events: auto; /* ให้กดปุ่มบนตัวมันได้ปกติ */
    display: flex;
    align-items: flex-start;
    gap: 0.75rem;
    padding: 0.85rem 1rem;
    background: var(--color-surface-2);
    backdrop-filter: blur(12px);
    -webkit-backdrop-filter: blur(12px);
    border: 1px solid var(--color-border);
    border-radius: 10px;
    box-shadow: var(--shadow-raised);
    color: var(--color-text-primary);
    position: relative;
    overflow: hidden;
  }

  /* แยกความต่างตามประเภทงานด้วยขอบ Border สี Signature */
  .toast-item::before {
    content: "";
    position: absolute;
    left: 0;
    top: 0;
    bottom: 0;
    width: 4px;
  }

  .toast-item--success::before {
    background: var(--color-success);
  }
  .toast-item--success .toast-icon {
    color: var(--color-success);
  }

  .toast-item--error::before {
    background: var(--color-brand);
  } /* แดงคอร์ปอเรตเว็บหนัง */
  .toast-item--error .toast-icon {
    color: var(--color-brand);
  }

  .toast-item--warning::before {
    background: var(--color-warning);
  }
  .toast-item--warning .toast-icon {
    color: var(--color-warning);
  }

  .toast-item--info::before {
    background: var(--color-info);
  }
  .toast-item--info .toast-icon {
    color: var(--color-info);
  }

  /* ภายใน Toast */
  .toast-icon {
    display: flex;
    margin-top: 2px;
    flex-shrink: 0;
  }

  .toast-content {
    flex-grow: 1;
    padding-right: 0.75rem;
  }


  .toast-message {
    font-size: 0.8rem;
    font-weight: 400;
    color: var(--color-text-secondary);
    line-height: 1.4;
  }

  .toast-close-btn {
    background: none;
    border: none;
    cursor: pointer;
    color: var(--color-text-muted);
    display: flex;
    align-items: center;
    justify-content: center;
    padding: 0.2rem;
    border-radius: 4px;
    margin-top: 1px;
    transition:
      color 0.15s,
      background 0.15s;
  }

  .toast-close-btn:hover {
    color: var(--color-text-primary);
    background: var(--color-border);
  }

  /* ─── อนิเมชันการสไลด์เรียงตัว (Transition Group) ─── */
  .toast-list-enter-active,
  .toast-list-leave-active,
  .toast-list-move {
    transition: all 0.3s cubic-bezier(0.4, 0, 0.2, 1);
  }

  /* เด้งมาจากตำแหน่งกระดิ่งด้านบนและจางลงไป */
  .toast-list-enter-from {
    opacity: 0;
    transform: translateY(-15px) scale(0.95);
  }

  /* หายวับไปด้านข้าง หรือเลือนหายจุดเดิม */
  .toast-list-leave-to {
    opacity: 0;
    transform: translateX(30px);
  }

  /* เพื่อป้องกันชิ้นอื่นกระตุกตอนไอเทมก่อนหน้ากำลังสลายตัว */
  .toast-list-leave-active {
    position: absolute;
    width: 100%;
  }
</style>
