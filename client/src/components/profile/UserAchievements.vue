<template>
  <div class="ach-root">
    <div class="section-head">
      <span class="eyebrow">Achievements</span>
      <span class="count-chip">{{ store.pagination.total }} unlocked</span>
      <div class="rule" />
      <RouterLink
        :to="{ name: 'user-achievements', params: { userId } }"
        class="view-all"
      >
        ดูทั้งหมด
      </RouterLink>
    </div>

    <SkeletonList v-if="store.loading" avatar :count="5" />

    <div v-else-if="store.error" class="state-empty">{{ store.error }}</div>

    <div v-else-if="store.userAchievements.length === 0" class="state-empty">
      ยังไม่มี Achievement ที่ปลดล็อก
    </div>

    <div v-else class="ach-list">
      <div
        v-for="(ua, i) in store.userAchievements"
        :key="ua.achievement_id"
        class="ach-card"
        :style="{ '--i': i }"
      >
        <div class="ach-icon-box lit">
          <component
            :is="getAchievementIcon(ua.achievement.action_type)"
            :size="15"
            :stroke-width="1.8"
          />
        </div>

        <div class="ach-info">
          <p class="ach-name">{{ ua.achievement.name }}</p>
          <p class="ach-desc">{{ ua.achievement.description }}</p>
          <p v-if="ua.unlocked_at" class="ach-unlocked-at">
            <CheckCircle :size="9" /> {{ formatDate(ua.unlocked_at) }}
          </p>
        </div>

        <div class="ach-badge">
          <span class="badge badge--done">
            <CheckCircle :size="9" /> Unlocked
          </span>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import SkeletonList from "@/components/ui/skeleton/SkeletonList.vue"
  import { onMounted } from "vue"
  import { CheckCircle } from "lucide-vue-next"
  import { useAchievementStore } from "@/stores/achievement"
  import { getAchievementIcon } from "@/utils/achievementIcons"

  const props = defineProps<{ userId: number }>()
  const store = useAchievementStore()

  function formatDate(iso: string) {
    return new Date(iso).toLocaleDateString("th-TH", {
      day: "numeric",
      month: "short",
      year: "numeric",
    })
  }

  onMounted(() => {
    store.fetchUnlockedOnly(props.userId)
  })
</script>

<style scoped>
  .ach-root {
    --c-card: #161616;
    --c-border: rgba(255, 255, 255, 0.06);
    --c-border-h: rgba(255, 255, 255, 0.12);
    --c-red: #e1251b;
    --c-text: #f0f0f0;
    --c-sub: #8a8a8e;
    --c-muted: #3a3a3c;
    --font: var(--font-sans);
    --ease: cubic-bezier(0.16, 1, 0.3, 1);
    font-family: var(--font-sans);
    color: var(--c-text);
  }

  .section-head {
    display: flex;
    align-items: center;
    gap: 10px;
    margin-bottom: 20px;
  }
  .eyebrow {
    font-size: 0.6rem;
    font-weight: 700;
    letter-spacing: 0.1em;
    text-transform: uppercase;
    color: var(--c-muted);
    white-space: nowrap;
  }
  .count-chip {
    font-size: 0.65rem;
    font-weight: 700;
    background: var(--c-red);
    color: #fff;
    padding: 2px 7px;
    border-radius: 4px;
  }
  .rule {
    flex: 1;
    height: 1px;
    background: var(--c-border);
  }
  .view-all {
    font-size: 0.68rem;
    font-weight: 600;
    color: var(--c-sub);
    text-decoration: none;
    white-space: nowrap;
    transition: color 0.15s;
  }
  .view-all:hover {
    color: #fff;
  }

  .state-empty {
    padding: 32px 0;
    text-align: center;
    color: var(--c-sub);
    font-size: 0.78rem;
  }

  .ach-list {
    display: flex;
    flex-direction: column;
    gap: 8px;
  }

  .ach-card {
    display: grid;
    grid-template-columns: 40px 1fr auto;
    align-items: center;
    gap: 14px;
    padding: 14px 16px;
    background: #101010;
    border: 1px solid var(--c-border);
    border-radius: 10px;
    position: relative;
    overflow: hidden;
    transition: all 0.22s var(--ease);
    animation: cardIn 0.4s var(--ease) calc(var(--i) * 50ms) both;
  }
  @keyframes cardIn {
    from {
      opacity: 0;
      transform: translateY(8px);
    }
    to {
      opacity: 1;
      transform: translateY(0);
    }
  }
  .ach-card:hover {
    background: var(--c-card);
    border-color: var(--c-border-h);
  }
  .ach-card::before {
    content: "";
    position: absolute;
    left: 0;
    top: 20%;
    bottom: 20%;
    width: 2px;
    background: #fff;
    border-radius: 0 2px 2px 0;
  }

  .ach-icon-box {
    width: 40px;
    height: 40px;
    background: var(--c-card);
    border: 1px solid var(--c-border);
    border-radius: 8px;
    display: flex;
    align-items: center;
    justify-content: center;
    color: var(--c-sub);
    transition: all 0.2s;
  }
  .ach-icon-box.lit {
    color: #fff;
    background: rgba(255, 255, 255, 0.04);
  }

  .ach-info {
    display: flex;
    flex-direction: column;
    gap: 2px;
    min-width: 0;
  }
  .ach-name {
    font-size: 0.84rem;
    font-weight: 500;
    color: #fff;
    margin: 0;
  }
  .ach-desc {
    font-size: 0.73rem;
    color: var(--c-sub);
    line-height: 1.45;
    margin: 0;
  }
  .ach-unlocked-at {
    display: inline-flex;
    align-items: center;
    gap: 4px;
    font-size: 0.65rem;
    color: var(--c-sub);
    margin: 3px 0 0;
    opacity: 0.75;
  }

  .badge {
    display: inline-flex;
    align-items: center;
    gap: 4px;
    font-size: 0.62rem;
    font-weight: 600;
    padding: 4px 9px;
    border-radius: 5px;
    white-space: nowrap;
  }
  .badge--done {
    background: rgba(255, 255, 255, 0.05);
    color: #fff;
    border: 1px solid rgba(255, 255, 255, 0.1);
  }

  @media (max-width: 480px) {
    .ach-card {
      grid-template-columns: 40px 1fr;
      row-gap: 8px;
    }
    .ach-badge {
      grid-column: 2;
      justify-self: start;
    }
  }
</style>
