<template>
  <Teleport to="body">
  <div class="toast-container">
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
          <div v-if="toast.title" class="toast-title">{{ toast.title }}</div>
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
  import { ref, onMounted, onUnmounted } from "vue"
  import { CheckCircle, XCircle, AlertCircle, Info, X } from "lucide-vue-next"

  // กำหนดโครงสร้างข้อมูล Toast
  interface Toast {
    id: string
    title?: string
    message: string
    type: "success" | "error" | "info" | "warning"
    duration?: number
  }

  const toasts = ref<Toast[]>([])
  const emit = defineEmits(["toast-added"])

  // ฟังก์ชันสุ่มหรือเจน ID
  const generateId = () => Math.random().toString(36).substring(2, 9)

  // ฟังก์ชันหลักในการเพิ่ม Toast (สามารถเรียกใช้ผ่าน global window หรือ export สโตร์)
  function addToast(options: Omit<Toast, "id">) {
    const id = generateId()
    const newToast: Toast = { ...options, id }

    // ดันเข้าอาเรย์ (แบบ Queue ตัวใหม่จะสไลด์ลงมาต่อท้ายด้านบน หรือแทรกตัวที่ index 0)
    toasts.value.unshift(newToast)

    // ส่งสัญญาณไปให้กระดิ่งที่ MainLayout รับรู้เพื่อเพิ่มตัวเลข Badge
    emit("toast-added")

    // ตั้งเวลาลบตัวเองอัตโนมัติ
    const duration = options.duration ?? 4000
    setTimeout(() => {
      removeToast(id)
    }, duration)
  }

  function removeToast(id: string) {
    toasts.value = toasts.value.filter(t => t.id !== id)
  }

  // ผูกฟังก์ชันไว้กับ window object เพื่อให้เรียกใช้จากหน้าไหนในโปรเจกต์ก็ได้แบบง่ายๆ
  // (หรือคุณสามารถปรับไปใช้ Pinia Store ตามโครงสร้างระบบของคุณได้เช่นกัน)
  onMounted(() => {
    ;(window as any).$toast = {
      show: addToast,
      success: (msg: string, title?: string) =>
        addToast({ message: msg, title, type: "success" }),
      error: (msg: string, title?: string) =>
        addToast({ message: msg, title, type: "error" }),
      info: (msg: string, title?: string) =>
        addToast({ message: msg, title, type: "info" }),
      warning: (msg: string, title?: string) =>
        addToast({ message: msg, title, type: "warning" }),
    }
  })

  onUnmounted(() => {
    delete (window as any).$toast
  })
</script>

<style scoped>
  /* Container วางตำแหน่งอยู่ภายใต้ปุ่มกระดิ่งแบบ Absolute */
  .toast-container {
    /* Teleported to body: inside the fixed navbar it could never rise above a modal. */
    position: fixed;
    top: 76px;
    right: 16px;
    width: 320px;
    max-width: calc(100vw - 32px);
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
    background: rgba(26, 26, 26, 0.96);
    backdrop-filter: blur(12px);
    -webkit-backdrop-filter: blur(12px);
    border: 1px solid rgba(255, 255, 255, 0.06);
    border-radius: 10px;
    box-shadow: 0 10px 30px rgba(0, 0, 0, 0.5);
    color: #fff;
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
    background: #10b981;
  }
  .toast-item--success .toast-icon {
    color: #10b981;
  }

  .toast-item--error::before {
    background: var(--color-brand);
  } /* แดงคอร์ปอเรตเว็บหนัง */
  .toast-item--error .toast-icon {
    color: var(--color-brand);
  }

  .toast-item--warning::before {
    background: #f59e0b;
  }
  .toast-item--warning .toast-icon {
    color: #f59e0b;
  }

  .toast-item--info::before {
    background: #3b82f6;
  }
  .toast-item--info .toast-icon {
    color: #3b82f6;
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

  .toast-title {
    font-size: 0.85rem;
    font-weight: 700;
    letter-spacing: 0.3px;
    margin-bottom: 0.15rem;
    color: #ffffff;
  }

  .toast-message {
    font-size: 0.8rem;
    font-weight: 400;
    color: #cccccc;
    line-height: 1.4;
  }

  .toast-close-btn {
    background: none;
    border: none;
    cursor: pointer;
    color: #666;
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
    color: #fff;
    background: rgba(255, 255, 255, 0.08);
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
