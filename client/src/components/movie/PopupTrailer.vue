<template>
  <div class="trailer-popup__overlay" @click.self="$emit('close')">
    <div class="trailer-popup__container">
      <button
        class="trailer-popup__close-btn"
        @click="$emit('close')"
        title="ปิดหน้านี้"
        aria-label="Close trailer"
      >
        <X :size="24" />
      </button>

      <div class="trailer-popup__media-wrapper">
        <iframe
          v-if="cleanTrailerUrl"
          :src="cleanTrailerUrl"
          frameborder="0"
          allow="autoplay; encrypted-media; gyroscope; picture-in-picture"
          allowfullscreen
          class="trailer-popup__iframe"
          referrerpolicy="strict-origin-when-cross-origin"
        />

        <div v-else class="trailer-popup__no-trailer">
          <VolumeX :size="48" class="trailer-popup__no-icon" />
          <p>ขออภัย ไม่พบตัวอย่างภาพยนตร์เรื่องนี้</p>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
  import { computed } from "vue"
  import { X, VolumeX } from "lucide-vue-next"

  // ─── Props & Emits ──────────────────────────────────────────────

  const props = defineProps<{
    // รับลิงก์ Embed ของเทรลเลอร์มาจากหน้า MovieDetailView
    trailerUrl: string | null | undefined
  }>()

  defineEmits<{
    (e: "close"): void
  }>()

  // ─── Computed ───────────────────────────────────────────────────

  // จัดการปรับแต่ง URL เพื่อให้เล่นอัตโนมัติ (Autoplay) และปล่อยเสียง (Unmute) เสมอเมื่อเปิดป็อปอัปตัวนี้ขึ้นมา
  const cleanTrailerUrl = computed(() => {
    if (!props.trailerUrl) return ""

    try {
      const url = new URL(props.trailerUrl)
      // ลบการบังคับ Mute ออกไปทั้งหมดเพื่อให้เสียงดังตั้งแต่เริ่มต้น
      url.searchParams.delete("mute")
      url.searchParams.delete("muted")

      // ตั้งค่า Autoplay เป็น 1 และเปิดใช้งานส่วนเสริมที่จำเป็น
      url.searchParams.set("autoplay", "1")
      url.searchParams.set("enablejsapi", "1")

      return url.toString()
    } catch (e) {
      // กรณีที่ส่งมาเป็น string ธรรมดา (fallback เผื่อไว้)
      return props.trailerUrl
        .replace(/[?&]mute=1/, "")
        .replace(/[?&]muted=1/, "")
    }
  })
</script>

<style scoped>
  /* ── Overlay คลุมทั้งหน้าจอ ───────────────────────────────────────── */
  .trailer-popup__overlay {
    position: fixed;
    inset: 0;
    background-color: rgba(0, 0, 0, 0.85);
    backdrop-filter: blur(8px);
    display: flex;
    align-items: center;
    justify-content: center;
    z-index: 999; /* มั่นใจว่าจะอยู่เหนือทุกสิ่งบนหน้าจอ */
    padding: 1rem;
    animation: fadeIn 0.25s ease-out;
  }

  /* ── กล่องกรอบของ Trailer (สัดส่วน 16:9 อัตโนมัติ) ────────────────────── */
  .trailer-popup__container {
    position: relative;
    width: 100%;
    max-width: 960px; /* ความกว้างยอดนิยมกำลังพอดีตา */
    background: #000;
    border-radius: 12px;
    overflow: hidden;
    box-shadow: 0 25px 50px -12px rgba(0, 0, 0, 0.5);
    border: 1px solid rgba(255, 255, 255, 0.1);
  }

  .trailer-popup__media-wrapper {
    position: relative;
    width: 100%;
    aspect-ratio: 16 / 9;
  }

  /* ── สไตล์ของ Iframe ────────────────────────────────────────────── */
  .trailer-popup__iframe {
    position: absolute;
    inset: 0;
    width: 100%;
    height: 100%;
    border: none;
  }

  /* ── ปุ่มกากบาทปิด (Close Button) ─────────────────────────────────── */
  .trailer-popup__close-btn {
    position: absolute;
    top: 16px;
    right: 16px;
    width: 40px;
    height: 40px;
    border-radius: 50%;
    background: rgba(0, 0, 0, 0.6);
    border: 1px solid rgba(255, 255, 255, 0.2);
    color: #fff;
    display: flex;
    align-items: center;
    justify-content: center;
    cursor: pointer;
    z-index: 10;
    transition:
      background 0.2s,
      transform 0.2s;
  }

  .trailer-popup__close-btn:hover {
    background: var(--color-brand); /* แดงสไตล์เน็ตฟลิกซ์ */
    border-color: transparent;
    transform: scale(1.05);
  }

  .trailer-popup__close-btn:active {
    transform: scale(0.95);
  }

  /* ── กรณีที่ไม่มีตัวอย่างหนังให้เล่น ───────────────────────────────────── */
  .trailer-popup__no-trailer {
    position: absolute;
    inset: 0;
    display: flex;
    flex-direction: column;
    align-items: center;
    justify-content: center;
    gap: 12px;
    color: #666;
    font-size: 1rem;
    background: #111;
  }

  .trailer-popup__no-icon {
    color: #333;
  }

  /* ── Animation เอฟเฟกต์เด้งเบาๆ ตอนเปิดหน้า ─────────────────────────────── */
  @keyframes fadeIn {
    from {
      opacity: 0;
    }
    to {
      opacity: 1;
    }
  }
</style>
