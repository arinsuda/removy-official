<template>
  <div class="about-page" ref="pageRef">
    <!-- Animated background -->
    <div class="bg-glow"></div>
    <div class="bg-grid"></div>

    <!-- Hero -->
    <section class="hero">
      <div class="hero-badge">
        <span class="hero-badge-dot"></span>
        {{ $t("about.eyebrow") }}
      </div>
      <h1 class="hero-title">
        <span class="hero-title-line">{{ $t("about.title.line1") }}</span>
        <span class="hero-title-accent">{{ $t("about.title.line2") }}</span>
      </h1>
      <p class="hero-sub">
        {{ $t("about.sub") }}
      </p>
      <div class="hero-stats">
        <div class="stat-item" v-for="stat in heroStats" :key="stat.label">
          <span class="stat-value">{{ stat.value }}</span>
          <span class="stat-label">{{ stat.label }}</span>
        </div>
      </div>
    </section>

    <!-- Mission -->
    <section class="section" id="mission">
      <div class="section-label">
        <div class="section-label-line"></div>
        <span>{{ $t("about.mission.label") }}</span>
        <div class="section-label-line"></div>
      </div>
      <div class="story-card">
        <div class="story-icon-wrap">
          <Heart :size="24" />
        </div>
        <h2 class="story-title">{{ $t("about.mission.title") }}</h2>
        <p class="story-text">
          {{ $t("about.mission.paragraph1") }}
        </p>
        <p class="story-text">
          {{ $t("about.mission.paragraph2") }}
        </p>
      </div>
    </section>

    <!-- Features -->
    <section class="section" id="features">
      <div class="section-label">
        <div class="section-label-line"></div>
        <span>{{ $t("about.features.label") }}</span>
        <div class="section-label-line"></div>
      </div>
      <h2 class="section-heading">{{ $t("about.features.title") }}</h2>
      <div class="feature-grid">
        <div
          v-for="(f, i) in features"
          :key="f.title"
          class="feature-card"
          :style="{ '--delay': `${i * 0.08}s` }"
        >
          <div class="feature-icon" :style="{ '--accent': f.color }">
            <component :is="f.icon" :size="22" :stroke-width="1.8" />
          </div>
          <h3 class="feature-title">{{ f.title }}</h3>
          <p class="feature-desc">{{ f.desc }}</p>
          <div class="feature-shine"></div>
        </div>
      </div>
    </section>

    <!-- Tech Stack -->
    <section class="section" id="tech">
      <div class="section-label">
        <div class="section-label-line"></div>
        <span>{{ $t("about.tech.label") }}</span>
        <div class="section-label-line"></div>
      </div>
      <h2 class="section-heading">{{ $t("about.tech.title") }}</h2>
      <p class="section-desc">
        {{ $t("about.tech.desc") }}
      </p>
      <div class="stack-grid">
        <div
          v-for="(s, i) in stackItems"
          :key="s.name"
          class="stack-card"
          :style="{ '--delay': `${i * 0.05}s` }"
        >
          <div class="stack-icon" :style="{ background: s.color }">
            {{ s.emoji }}
          </div>
          <div class="stack-info">
            <span class="stack-name">{{ s.name }}</span>
            <span class="stack-role">{{ s.role }}</span>
          </div>
        </div>
      </div>
    </section>

    <!-- Creator -->
    <section class="section" id="creator">
      <div class="section-label">
        <div class="section-label-line"></div>
        <span>{{ $t("about.creator.label") }}</span>
        <div class="section-label-line"></div>
      </div>
      <div class="creator-card">
        <div class="creator-glow"></div>
        <div class="creator-top">
          <div class="creator-avatar-wrap">
            <div class="creator-avatar-ring"></div>
            <div class="creator-avatar">
              <img src="/arin.png" alt="ผู้พัฒนา" @error="onAvatarError" />
            </div>
          </div>
          <div class="creator-meta">
            <span class="creator-role">{{ $t("about.creator.role") }}</span>
            <h3 class="creator-name">{{ $t("about.creator.name") }}</h3>
            <p class="creator-bio">
              {{ $t("about.creator.bio") }}
            </p>
          </div>
        </div>
        <div class="creator-links">
          <a href="https://github.com/arinsuda" target="_blank" class="creator-link">
            <Github :size="16" />
            <span>GitHub</span>
          </a>
          <a href="https://linkedin.com/in/arinsuda" target="_blank" class="creator-link">
            <Linkedin :size="16" />
            <span>LinkedIn</span>
          </a>
          <a href="mailto:sixarin.thorn@gmail.com" class="creator-link">
            <Mail :size="16" />
            <span>Email</span>
          </a>
        </div>
      </div>
    </section>

    <!-- CTA -->
    <section class="cta-section">
      <div class="cta-card">
        <div class="cta-glow"></div>
        <div class="cta-content">
          <h3 class="cta-title">{{ $t("about.cta.title") }}</h3>
          <p class="cta-desc">{{ $t("about.cta.desc") }}</p>
        </div>
        <RouterLink :to="{ name: 'donate' }" class="cta-btn">
          <Heart :size="16" />
          <span>{{ $t("about.cta.btn") }}</span>
          <ArrowRight :size="16" class="cta-arrow" />
        </RouterLink>
      </div>
    </section>
  </div>
</template>

<script setup lang="ts">
  import { computed, ref } from "vue";
  import { useI18n } from "vue-i18n";
  import {
    Star,
    Bookmark,
    Trophy,
    Users,
    Github,
    Linkedin,
    Mail,
    Heart,
    ArrowRight,
    Film,
    Sparkles,
  } from "lucide-vue-next";

  const { t } = useI18n();

  const pageRef = ref<HTMLElement | null>(null);

  const heroStats = computed(() => [
    { value: "10K+", label: t("about.stats.movies") },
    { value: "5K+", label: t("about.stats.tv") },
    { value: "∞", label: t("about.stats.reviews") },
  ]);

  const features = computed(() => [
    {
      icon: Star,
      title: t("about.features.items.reviews.title"),
      desc: t("about.features.items.reviews.desc"),
      color: "#f59e0b",
    },
    {
      icon: Bookmark,
      title: t("about.features.items.watchlist.title"),
      desc: t("about.features.items.watchlist.desc"),
      color: "#3b82f6",
    },
    {
      icon: Trophy,
      title: t("about.features.items.achievements.title"),
      desc: t("about.features.items.achievements.desc"),
      color: "#a855f7",
    },
    {
      icon: Users,
      title: t("about.features.items.community.title"),
      desc: t("about.features.items.community.desc"),
      color: "#10b981",
    },
    {
      icon: Film,
      title: t("about.features.items.metadata.title"),
      desc: t("about.features.items.metadata.desc"),
      color: "#ec4899",
    },
    {
      icon: Sparkles,
      title: t("about.features.items.popcornScore.title"),
      desc: t("about.features.items.popcornScore.desc"),
      color: "#e1251b",
    },
  ]);

  const stackItems = computed(() => [
    { name: "Go", role: t("about.tech.roles.backendRuntime"), emoji: "🔷", color: "rgba(0, 173, 216, 0.15)" },
    { name: "Fiber v3", role: t("about.tech.roles.httpFramework"), emoji: "⚡", color: "rgba(52, 211, 153, 0.15)" },
    { name: "GORM", role: t("about.tech.roles.orm"), emoji: "🗃️", color: "rgba(168, 85, 247, 0.15)" },
    { name: "PostgreSQL", role: t("about.tech.roles.database"), emoji: "🐘", color: "rgba(59, 130, 246, 0.15)" },
    { name: "Vue 3", role: t("about.tech.roles.frontendFramework"), emoji: "💚", color: "rgba(66, 184, 131, 0.15)" },
    { name: "TypeScript", role: t("about.tech.roles.typeSystem"), emoji: "🔷", color: "rgba(49, 120, 198, 0.15)" },
    { name: "Pinia", role: t("about.tech.roles.stateManagement"), emoji: "🍍", color: "rgba(245, 158, 11, 0.15)" },
    { name: "TanStack Query", role: t("about.tech.roles.dataFetching"), emoji: "🔄", color: "rgba(239, 68, 68, 0.15)" },
  ]);

  function onAvatarError(e: Event) {
    ;(e.target as HTMLImageElement).style.display = "none";
  }
</script>

<style scoped>

  .about-page {
    --c-bg: var(--color-surface);
    --c-surface: var(--color-surface-2);
    --c-card: var(--color-surface-2);
    --c-card-hover: var(--color-surface-3);
    --c-border: var(--color-border);
    --c-border-hover: var(--color-border-strong);
    --c-red: var(--color-brand);
    --c-red-glow: color-mix(in srgb, var(--color-brand) 30%, transparent);
    --c-text: var(--color-text-primary);
    --c-text-secondary: var(--color-text-secondary);
    --c-text-muted: var(--color-text-muted);
    --radius-lg: 20px;
    --radius-md: 14px;
    --radius-sm: 10px;
    --ease-out: cubic-bezier(0.16, 1, 0.3, 1);

    font-family: var(--font-sans);
    color: var(--c-text);
    max-width: 960px;
    margin: 0 auto;
    padding: 0 24px 100px;
    position: relative;
    overflow: hidden;
  }

  /* Animated background effects */
  .bg-glow {
    position: fixed;
    top: -200px;
    left: 50%;
    transform: translateX(-50%);
    width: 800px;
    height: 600px;
    background: radial-gradient(
      ellipse at center,
      rgba(225, 37, 27, 0.08) 0%,
      rgba(225, 37, 27, 0.02) 40%,
      transparent 70%
    );
    pointer-events: none;
    z-index: 0;
    will-change: transform;
    contain: layout style paint;
    animation: glowPulse 8s ease-in-out infinite;
  }

  .bg-grid {
    position: fixed;
    inset: 0;
    background-image:
      linear-gradient(rgba(255, 255, 255, 0.02) 1px, transparent 1px),
      linear-gradient(90deg, rgba(255, 255, 255, 0.02) 1px, transparent 1px);
    background-size: 60px 60px;
    pointer-events: none;
    z-index: 0;
    will-change: transform;
    contain: layout style paint;
    mask-image: radial-gradient(ellipse at center, black 30%, transparent 70%);
    -webkit-mask-image: radial-gradient(ellipse at center, black 30%, transparent 70%);
  }

  @keyframes glowPulse {
    0%, 100% { opacity: 0.6; transform: translateX(-50%) scale(1); }
    50% { opacity: 1; transform: translateX(-50%) scale(1.1); }
  }

  /* Hero */
  .hero {
    position: relative;
    z-index: 1;
    padding: 80px 0 64px;
    text-align: center;
    animation: fadeInUp 0.8s var(--ease-out) both;
  }

  .hero-badge {
    display: inline-flex;
    align-items: center;
    gap: 8px;
    font-size: 0.72rem;
    font-weight: 600;
    letter-spacing: 0.1em;
    text-transform: uppercase;
    color: var(--c-red);
    background: rgba(225, 37, 27, 0.08);
    border: 1px solid rgba(225, 37, 27, 0.15);
    padding: 6px 16px;
    border-radius: 100px;
    margin-bottom: 28px;
  }

  .hero-badge-dot {
    width: 6px;
    height: 6px;
    border-radius: 50%;
    background: var(--c-red);
    position: relative;
  }

  .hero-badge-dot::after {
    content: "";
    position: absolute;
    inset: -6px;
    border-radius: 50%;
    background: rgba(225, 37, 27, 0.4);
    will-change: transform, opacity;
    animation: dotPulseNew 2s ease-in-out infinite;
  }

  @keyframes dotPulseNew {
    0%, 100% { transform: scale(0); opacity: 0.4; }
    50%      { transform: scale(1); opacity: 0; }
  }

  .hero-title {
    font-size: 3.2rem;
    font-weight: 900;
    line-height: 1.15;
    letter-spacing: -0.03em;
    margin: 0 0 20px;
  }

  .hero-title-line {
    display: block;
    color: var(--c-text);
  }

  .hero-title-accent {
    display: block;
    background: linear-gradient(135deg, #e1251b 0%, #ff6b6b 50%, #e1251b 100%);
    background-size: 200% auto;
    -webkit-background-clip: text;
    -webkit-text-fill-color: transparent;
    background-clip: text;
    will-change: background-position;
    animation: gradientShift 4s ease-in-out infinite;
  }

  @keyframes gradientShift {
    0%, 100% { background-position: 0% center; }
    50% { background-position: 100% center; }
  }

  .hero-sub {
    font-size: 1.05rem;
    color: var(--c-text-secondary);
    line-height: 1.7;
    max-width: 500px;
    margin: 0 auto 40px;
    font-weight: 400;
  }

  .hero-stats {
    display: flex;
    justify-content: center;
    gap: 48px;
  }

  .stat-item {
    display: flex;
    flex-direction: column;
    align-items: center;
    gap: 4px;
  }

  .stat-value {
    font-size: 1.6rem;
    font-weight: 800;
    letter-spacing: -0.02em;
    background: linear-gradient(180deg, var(--c-text) 0%, var(--c-text-secondary) 100%);
    -webkit-background-clip: text;
    -webkit-text-fill-color: transparent;
    background-clip: text;
  }

  .stat-label {
    font-size: 0.72rem;
    color: var(--c-text-muted);
    font-weight: 500;
    text-transform: uppercase;
    letter-spacing: 0.08em;
  }

  /* Sections */
  .section {
    position: relative;
    z-index: 1;
    margin-bottom: 72px;
    animation: fadeInUp 0.6s var(--ease-out) both;
  }

  .section-label {
    display: flex;
    align-items: center;
    justify-content: center;
    gap: 16px;
    margin-bottom: 24px;
  }

  .section-label span {
    font-size: 0.68rem;
    font-weight: 700;
    letter-spacing: 0.15em;
    text-transform: uppercase;
    color: var(--c-text-muted);
    white-space: nowrap;
  }

  .section-label-line {
    height: 1px;
    width: 48px;
    background: linear-gradient(90deg, transparent, var(--c-border), transparent);
  }

  .section-heading {
    text-align: center;
    font-size: 1.8rem;
    font-weight: 800;
    margin: 0 0 12px;
    letter-spacing: -0.02em;
  }

  .section-desc {
    text-align: center;
    font-size: 0.9rem;
    color: var(--c-text-secondary);
    line-height: 1.7;
    max-width: 560px;
    margin: 0 auto 32px;
  }

  /* Story card */
  .story-card {
    position: relative;
    background: var(--c-card);
    border: 1px solid var(--c-border);
    border-radius: var(--radius-lg);
    padding: 40px 36px;
    backdrop-filter: blur(8px);
    -webkit-backdrop-filter: blur(8px);
    overflow: hidden;
    transition: border-color 0.3s var(--ease-out);
  }

  .story-card:hover {
    border-color: var(--c-border-hover);
  }

  .story-card::before {
    content: "";
    position: absolute;
    top: 0;
    left: 0;
    right: 0;
    height: 1px;
    background: linear-gradient(90deg, transparent, rgba(225, 37, 27, 0.3), transparent);
  }

  .story-icon-wrap {
    display: inline-flex;
    align-items: center;
    justify-content: center;
    width: 48px;
    height: 48px;
    border-radius: 14px;
    background: rgba(225, 37, 27, 0.1);
    color: var(--c-red);
    margin-bottom: 20px;
  }

  .story-title {
    font-size: 1.3rem;
    font-weight: 700;
    margin: 0 0 16px;
  }

  .story-text {
    font-size: 0.88rem;
    color: var(--c-text-secondary);
    line-height: 1.8;
    margin: 0 0 12px;
  }

  .story-text:last-child {
    margin-bottom: 0;
  }

  /* Feature grid */
  .feature-grid {
    display: grid;
    grid-template-columns: repeat(3, 1fr);
    gap: 16px;
  }

  .feature-card {
    position: relative;
    background: var(--c-card);
    border: 1px solid var(--c-border);
    border-radius: var(--radius-md);
    padding: 24px 20px;
    transition: transform 0.35s var(--ease-out), background 0.35s var(--ease-out), border-color 0.35s var(--ease-out);
    overflow: hidden;
    animation: fadeInUp 0.6s var(--ease-out) both;
    animation-delay: var(--delay);
  }

  .feature-card:hover {
    border-color: var(--c-border-hover);
    transform: translateY(-4px);
    background: var(--c-card-hover);
  }

  .feature-shine {
    position: absolute;
    top: 0;
    left: 0;
    width: 100%;
    height: 100%;
    background: linear-gradient(
      90deg,
      transparent,
      rgba(255, 255, 255, 0.03),
      transparent
    );
    transform: translateX(-100%);
    transition: transform 0.6s var(--ease-out);
    will-change: transform;
    pointer-events: none;
  }

  .feature-card:hover .feature-shine {
    transform: translateX(100%);
  }

  .feature-icon {
    display: inline-flex;
    align-items: center;
    justify-content: center;
    width: 42px;
    height: 42px;
    border-radius: 12px;
    background: rgba(255, 255, 255, 0.04);
    color: var(--accent);
    margin-bottom: 14px;
    transition: transform 0.3s var(--ease-out);
  }

  .feature-card:hover .feature-icon {
    transform: scale(1.1);
  }

  .feature-title {
    font-size: 0.92rem;
    font-weight: 700;
    margin: 0 0 6px;
  }

  .feature-desc {
    font-size: 0.78rem;
    color: var(--c-text-secondary);
    line-height: 1.6;
    margin: 0;
  }

  /* Stack grid */
  .stack-grid {
    display: grid;
    grid-template-columns: repeat(4, 1fr);
    gap: 12px;
  }

  .stack-card {
    display: flex;
    align-items: center;
    gap: 12px;
    background: var(--c-card);
    border: 1px solid var(--c-border);
    border-radius: var(--radius-sm);
    padding: 14px 16px;
    transition: transform 0.3s var(--ease-out), border-color 0.3s var(--ease-out);
    animation: fadeInUp 0.5s var(--ease-out) both;
    animation-delay: var(--delay);
  }

  .stack-card:hover {
    border-color: var(--c-border-hover);
    transform: translateY(-2px);
  }

  .stack-icon {
    display: flex;
    align-items: center;
    justify-content: center;
    width: 36px;
    height: 36px;
    border-radius: 10px;
    font-size: 1rem;
    flex-shrink: 0;
  }

  .stack-info {
    display: flex;
    flex-direction: column;
    gap: 1px;
    min-width: 0;
  }

  .stack-name {
    font-size: 0.82rem;
    font-weight: 700;
    white-space: nowrap;
    overflow: hidden;
    text-overflow: ellipsis;
  }

  .stack-role {
    font-size: 0.66rem;
    color: var(--c-text-muted);
    font-weight: 500;
  }

  /* Creator */
  .creator-card {
    position: relative;
    background: var(--c-card);
    border: 1px solid var(--c-border);
    border-radius: var(--radius-lg);
    padding: 36px;
    overflow: hidden;
    backdrop-filter: blur(8px);
    -webkit-backdrop-filter: blur(8px);
    transition: border-color 0.3s var(--ease-out);
  }

  .creator-card:hover {
    border-color: var(--c-border-hover);
  }

  .creator-glow {
    position: absolute;
    top: -80px;
    right: -80px;
    width: 200px;
    height: 200px;
    background: radial-gradient(circle, rgba(225, 37, 27, 0.08) 0%, transparent 70%);
    pointer-events: none;
  }

  .creator-top {
    display: flex;
    align-items: center;
    gap: 24px;
    margin-bottom: 24px;
  }

  .creator-avatar-wrap {
    position: relative;
    flex-shrink: 0;
  }

  .creator-avatar-ring {
    position: absolute;
    inset: -3px;
    border-radius: 50%;
    border: 2px solid transparent;
    background: linear-gradient(135deg, var(--c-red), #ff6b6b, var(--c-red)) border-box;
    -webkit-mask: linear-gradient(#fff 0 0) padding-box, linear-gradient(#fff 0 0);
    -webkit-mask-composite: xor;
    mask-composite: exclude;
    will-change: transform;
    animation: ringRotate 6s linear infinite;
  }

  @keyframes ringRotate {
    from { transform: rotate(0deg); }
    to { transform: rotate(360deg); }
  }

  .creator-avatar {
    width: 80px;
    height: 80px;
    border-radius: 50%;
    overflow: hidden;
    background: var(--c-surface);
  }

  .creator-avatar img {
    width: 100%;
    height: 100%;
    object-fit: cover;
  }

  .creator-meta {
    flex: 1;
    min-width: 0;
  }

  .creator-role {
    font-size: 0.68rem;
    font-weight: 700;
    letter-spacing: 0.1em;
    text-transform: uppercase;
    color: var(--c-red);
  }

  .creator-name {
    font-size: 1.4rem;
    font-weight: 800;
    margin: 4px 0 8px;
    letter-spacing: -0.01em;
  }

  .creator-bio {
    font-size: 0.85rem;
    color: var(--c-text-secondary);
    line-height: 1.6;
    margin: 0;
  }

  .creator-links {
    display: flex;
    gap: 10px;
    flex-wrap: wrap;
  }

  .creator-link {
    display: inline-flex;
    align-items: center;
    gap: 8px;
    font-size: 0.78rem;
    font-weight: 600;
    color: var(--c-text-secondary);
    text-decoration: none;
    padding: 8px 16px;
    border-radius: 10px;
    background: rgba(255, 255, 255, 0.04);
    border: 1px solid var(--c-border);
    transition: color 0.25s var(--ease-out), border-color 0.25s var(--ease-out), background-color 0.25s var(--ease-out), transform 0.25s var(--ease-out);
  }

  .creator-link:hover {
    color: var(--c-text);
    border-color: var(--c-border-hover);
    background: rgba(255, 255, 255, 0.08);
    transform: translateY(-2px);
  }

  /* CTA */
  .cta-section {
    position: relative;
    z-index: 1;
  }

  .cta-card {
    position: relative;
    background: linear-gradient(135deg, rgba(225, 37, 27, 0.06), rgba(225, 37, 27, 0.02));
    border: 1px solid rgba(225, 37, 27, 0.12);
    border-radius: var(--radius-lg);
    padding: 40px 36px;
    text-align: center;
    overflow: hidden;
  }

  .cta-glow {
    position: absolute;
    bottom: -60px;
    left: 50%;
    transform: translateX(-50%);
    width: 300px;
    height: 200px;
    background: radial-gradient(circle, rgba(225, 37, 27, 0.1) 0%, transparent 70%);
    pointer-events: none;
  }

  .cta-content {
    position: relative;
    margin-bottom: 24px;
  }

  .cta-title {
    font-size: 1.5rem;
    font-weight: 800;
    margin: 0 0 8px;
  }

  .cta-desc {
    font-size: 0.88rem;
    color: var(--c-text-secondary);
    margin: 0;
    line-height: 1.6;
  }

  .cta-btn {
    position: relative;
    display: inline-flex;
    align-items: center;
    gap: 8px;
    background: var(--c-red);
    color: #fff;
    font-size: 0.88rem;
    font-weight: 700;
    padding: 12px 28px;
    border-radius: 12px;
    text-decoration: none;
    transition: transform 0.3s var(--ease-out);
  }

  .cta-btn::after {
    content: "";
    position: absolute;
    inset: 0;
    border-radius: 12px;
    box-shadow: 0 8px 32px rgba(225, 37, 27, 0.35);
    opacity: 0;
    transition: opacity 0.3s;
    pointer-events: none;
  }

  .cta-btn:hover {
    transform: translateY(-2px);
  }

  .cta-btn:hover::after {
    opacity: 1;
  }

  .cta-arrow {
    transition: transform 0.3s var(--ease-out);
  }

  .cta-btn:hover .cta-arrow {
    transform: translateX(4px);
  }

  /* Animations */
  @keyframes fadeInUp {
    from {
      opacity: 0;
      transform: translateY(24px);
    }
    to {
      opacity: 1;
      transform: translateY(0);
    }
  }

  /* Accessibility & Mobile Fallbacks */
  @media (max-width: 768px) {
    .bg-glow, .bg-grid, .creator-avatar-ring {
      animation: none !important;
    }
    .story-card, .creator-card {
      backdrop-filter: none !important;
      -webkit-backdrop-filter: none !important;
      background: var(--c-surface);
    }

    .hero-title {
      font-size: 2.2rem;
    }

    .hero-sub {
      font-size: 0.92rem;
    }

    .hero-stats {
      gap: 32px;
    }

    .feature-grid {
      grid-template-columns: repeat(2, 1fr);
    }

    .stack-grid {
      grid-template-columns: repeat(2, 1fr);
    }

    .story-card {
      padding: 28px 24px;
    }

    .creator-card {
      padding: 28px 24px;
    }

    .cta-card {
      padding: 32px 24px;
    }
  }

  @media (max-width: 480px) {
    .about-page {
      padding: 0 16px 80px;
    }

    .hero {
      padding: 56px 0 48px;
    }

    .hero-title {
      font-size: 1.8rem;
    }

    .hero-stats {
      gap: 24px;
    }

    .stat-value {
      font-size: 1.3rem;
    }

    .section-heading {
      font-size: 1.4rem;
    }

    .feature-grid {
      grid-template-columns: 1fr;
    }

    .stack-grid {
      grid-template-columns: 1fr;
    }

    .creator-top {
      flex-direction: column;
      text-align: center;
    }

    .creator-links {
      justify-content: center;
    }
  }

  @media (prefers-reduced-motion: reduce) {
    .bg-glow, .bg-grid, .creator-avatar-ring, .hero-title-accent, .hero-badge-dot::after {
      animation: none !important;
      transition: none !important;
    }
    .feature-shine {
      display: none !important;
    }
  }
</style>
