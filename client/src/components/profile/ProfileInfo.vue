<template>
  <div class="pinfo-root">
    <div class="section-head">
      <span class="eyebrow">Personal Information</span>
      <div class="rule" />
    </div>

    <div class="info-card">
      <div
        v-for="(row, i) in rows"
        :key="i"
        class="info-row"
        :style="{ '--delay': i * 40 + 'ms' }"
      >
        <span class="row-icon"><component :is="row.icon" :size="13" /></span>
        <span class="row-key">{{ row.key }}</span>
        <span class="row-val" :class="row.class">{{ row.val }}</span>
      </div>
    </div>

    <template v-if="user.bio">
      <div class="section-head section-head--mt">
        <span class="eyebrow">Biography</span>
        <div class="rule" />
      </div>
      <p class="bio-text">{{ user.bio }}</p>
    </template>
  </div>
</template>

<script setup lang="ts">
  import { computed } from "vue"
  import { User, AtSign, Shield, CalendarDays } from "lucide-vue-next"
  import type { UserProfile } from "@/types"

  const props = defineProps<{ user: UserProfile }>()

  const joinDate = computed(() => {
    if (!props.user.created_at) return "—"
    return new Date(props.user.created_at).toLocaleDateString("en-US", {
      year: "numeric",
      month: "long",
      day: "numeric",
    })
  })

  const rows = computed(() => [
    {
      icon: User,
      key: "Display Name",
      val: props.user.display_name || "—",
      class: "",
    },
    {
      icon: AtSign,
      key: "Username",
      val: `@${props.user.username}`,
      class: "",
    },
    {
      icon: Shield,
      key: "Account Role",
      val: "Standard Member",
      class: "val-badge",
    },
    { icon: CalendarDays, key: "Joined", val: joinDate.value, class: "" },
  ])
</script>

<style scoped>
  .pinfo-root {
    --c-card: var(--color-surface-2);
    --c-border: var(--color-border);
    --c-border-h: var(--color-border-strong);
    --c-red: var(--color-brand);
    --c-text: var(--color-text-primary);
    --c-sub: var(--color-text-secondary);
    --c-muted: var(--color-text-muted);
    --font: var(--font-sans);
    --ease: cubic-bezier(0.16, 1, 0.3, 1);
    font-family: var(--font-sans);
    color: var(--c-text);
  }

  .section-head {
    display: flex;
    align-items: center;
    gap: 12px;
    margin-bottom: 16px;
  }
  .section-head--mt {
    margin-top: 32px;
  }

  .eyebrow {
    font-size: 0.6rem;
    font-weight: 700;
    letter-spacing: 0.1em;
    text-transform: uppercase;
    color: var(--c-muted);
    white-space: nowrap;
  }
  .rule {
    flex: 1;
    height: 1px;
    background: var(--c-border);
  }

  .info-card {
    background: var(--c-card);
    border: 1px solid var(--c-border);
    border-radius: 12px;
    overflow: hidden;
  }

  .info-row {
    display: grid;
    grid-template-columns: 18px 150px 1fr;
    align-items: center;
    gap: 12px;
    padding: 15px 18px;
    border-bottom: 1px solid var(--c-border);
    animation: rowIn 0.4s var(--ease) var(--delay, 0ms) both;
    transition: background 0.15s;
  }
  .info-row:last-child {
    border-bottom: none;
  }
  .info-row:hover {
    background: rgba(255, 255, 255, 0.025);
  }

  @keyframes rowIn {
    from {
      opacity: 0;
      transform: translateX(-8px);
    }
    to {
      opacity: 1;
      transform: translateX(0);
    }
  }

  .row-icon {
    color: var(--c-muted);
    display: flex;
    align-items: center;
  }
  .row-key {
    font-size: 0.78rem;
    color: var(--c-sub);
    font-weight: 500;
  }
  .row-val {
    font-size: 0.85rem;
    color: #fff;
    font-weight: 400;
  }

  .val-badge {
    display: inline-flex;
    align-items: center;
    background: rgba(255, 255, 255, 0.05);
    border: 1px solid rgba(255, 255, 255, 0.08);
    padding: 2px 10px;
    border-radius: 20px;
    font-size: 0.72rem;
    font-weight: 500;
    color: var(--c-text);
  }

  .bio-text {
    font-size: 0.875rem;
    color: var(--c-sub);
    line-height: 1.7;
    max-width: 600px;
    margin: 0;
    padding: 0 2px;
  }

  @media (max-width: 540px) {
    .info-row {
      grid-template-columns: 18px 1fr;
      row-gap: 3px;
      padding: 12px 14px;
    }
    .row-val {
      grid-column: 2;
    }
  }
</style>
