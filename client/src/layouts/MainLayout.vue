<template>
  <div class="app-shell">
    <header class="navbar" :class="{ 'navbar--scrolled': scrolled }">
      <RouterLink to="/" class="nav-logo" @click="closeMobileMenu">
        <img src="/popcorn-icon-192.png" alt="" class="nav-logo-icon" /><span class="logo-movie">POP</span><span class="logo-hub">CORN</span>
      </RouterLink>

      <!-- Desktop / Tablet nav links -->
      <nav class="nav-links">
        <RouterLink
          to="/"
          class="nav-link"
          :class="{ active: route.name === 'home' }"
          >{{ $t("navigation.home") }}</RouterLink
        >
        <RouterLink
          to="/movies"
          class="nav-link"
          :class="{ active: route.path.startsWith('/movies') }"
          >{{ $t("navigation.movies") }}</RouterLink
        >
        <RouterLink
          to="/tv"
          class="nav-link"
          :class="{ active: route.path.startsWith('/tv') }"
          >{{ $t("navigation.tvSeries") }}</RouterLink
        >
        <RouterLink
          to="/upcoming"
          class="nav-link"
          :class="{ active: route.name?.toString().startsWith('upcoming') }"
          >{{ $t("navigation.upcoming") }}</RouterLink
        >
        <RouterLink
          to="/about-us"
          class="nav-link"
          :class="{ active: route.name?.toString().startsWith('aboutUs') }"
          >{{ $t("navigation.aboutUs") }}</RouterLink
        >
      </nav>

      <div class="nav-right">
        <HeaderSearch />

        <!-- Language Switcher -->
        <button
          class="lang-btn"
          @click="localeStore.toggleLocale()"
          :disabled="localeStore.isSwitchingLocale.value"
          :aria-label="$t('navigation.switchLang')"
        >
          {{ localeStore.locale.value === "th" ? "EN" : "TH" }}
        </button>

        <div class="notification-wrapper">
          <button
            class="icon-btn"
            :class="{ 'has-unread': notificationStore.unreadCount > 0 }"
            @click.stop="toggleNotificationPanel"
          >
            <Bell :size="18" />
            <span v-if="notificationStore.unreadCount > 0" class="bell-badge">{{
              notificationStore.unreadCount > 99
                ? "99+"
                : notificationStore.unreadCount
            }}</span>
          </button>

          <Transition name="dropdown">
            <NotificationPanel
              v-if="notificationStore.isPanelOpen"
              @close="notificationStore.closePanel()"
            />
          </Transition>
        </div>

        <div class="user-menu" v-if="authStore.user" ref="userMenuRef">
          <button class="user-trigger" @click="toggleUserMenu">
            <div class="user-avatar">
              <img
                v-if="authStore.user.avatar_url"
                :src="authStore.user.avatar_url"
                :alt="authStore.user.username"
              />
              <UserIcon v-else :size="16" />
            </div>
            <span class="user-name">{{ authStore.user.username }}</span>
            <ChevronDown
              :size="14"
              :class="{ 'rotate-180': userMenuOpen }"
              class="transition-transform duration-200"
            />
          </button>

          <Transition name="dropdown">
            <div class="dropdown-menu" v-if="userMenuOpen">
              <RouterLink
                :to="`/users/${authStore.user.id}`"
                class="dropdown-item"
                @click="userMenuOpen = false"
              >
                <UserIcon :size="14" />{{ $t("navigation.profile") }}
              </RouterLink>
              <RouterLink
                :to="`/users/${authStore.user.id}/library`"
                class="dropdown-item"
                @click="userMenuOpen = false"
              >
                <BookMarked :size="14" />{{ $t("navigation.myLibrary") }}
              </RouterLink>
              <RouterLink
                :to="`/users/${authStore.user.id}/achievements`"
                class="dropdown-item"
                @click="userMenuOpen = false"
              >
                <Trophy :size="14" />{{ $t("navigation.achievement") }}
              </RouterLink>
              <RouterLink
                :to="{ name: 'feed' }"
                class="dropdown-item"
                @click="userMenuOpen = false"
              >
                <Rss :size="14" />{{ $t("navigation.feed") }}
              </RouterLink>
              <template v-if="authStore.isAdmin">
                <div class="dropdown-divider" />
                <button
                  class="dropdown-item dropdown-item--admin"
                  @click="handleEnterAdmin"
                >
                  <Shield :size="14" />{{ $t("navigation.adminMode") }}
                </button>
              </template>
              <div class="dropdown-divider" />
              <button class="dropdown-item" @click="handleSwitchToPlanzi">
                <img src="/planzi-icon-192.png" alt="" class="switch-platform-icon" :width="14" :height="14" />{{ $t("navigation.switchToPlanzi") }}
              </button>
              <button
                class="dropdown-item dropdown-item--danger"
                @click="handleLogout"
              >
                <LogOut :size="14" />{{ $t("navigation.logout") }}
              </button>
            </div>
          </Transition>
        </div>

        <!-- Hamburger toggle: shows on tablet/mobile -->
        <button
          class="hamburger-btn"
          :class="{ 'hamburger-btn--active': mobileMenuOpen }"
          @click="toggleMobileMenu"
          aria-label="Toggle menu"
        >
          <span class="hamburger-line" />
          <span class="hamburger-line" />
          <span class="hamburger-line" />
        </button>
      </div>
    </header>

    <!-- Mobile / Tablet slide-down menu -->
    <Transition name="mobile-menu">
      <div
        v-if="mobileMenuOpen"
        class="mobile-menu-overlay"
        @click.self="closeMobileMenu"
      >
        <nav class="mobile-menu-panel">
          <RouterLink
            to="/"
            class="mobile-nav-link"
            :class="{ active: route.name === 'home' }"
            @click="closeMobileMenu"
            >{{ $t("navigation.home") }}</RouterLink
          >
          <RouterLink
            to="/movies"
            class="mobile-nav-link"
            :class="{ active: route.path.startsWith('/movies') }"
            @click="closeMobileMenu"
            >{{ $t("navigation.movies") }}</RouterLink
          >
          <RouterLink
            to="/tv"
            class="mobile-nav-link"
            :class="{ active: route.path.startsWith('/tv') }"
            @click="closeMobileMenu"
            >{{ $t("navigation.tvSeries") }}</RouterLink
          >
          <RouterLink
            to="/upcoming"
            class="mobile-nav-link"
            :class="{ active: route.name?.toString().startsWith('upcoming') }"
            @click="closeMobileMenu"
            >{{ $t("navigation.upcoming") }}</RouterLink
          >
          <RouterLink
            to="/about-us"
            class="mobile-nav-link"
            :class="{ active: route.name?.toString().startsWith('aboutUs') }"
            @click="closeMobileMenu"
            >{{ $t("navigation.aboutUs") }}</RouterLink
          >

          <button
            class="mobile-nav-link"
            @click="localeStore.toggleLocale()"
            :disabled="localeStore.isSwitchingLocale.value"
          >
            {{ $t("navigation.switchLang") }} ({{ localeStore.locale.value === "th" ? "English" : "ไทย" }})
          </button>

          <div class="mobile-menu-divider" />

          <template v-if="authStore.user">
            <RouterLink
              :to="`/users/${authStore.user.id}`"
              class="mobile-nav-link mobile-nav-link--sub"
              @click="closeMobileMenu"
            >
              <UserIcon :size="16" />{{ $t("navigation.profile") }}
            </RouterLink>
            <RouterLink
              :to="`/users/${authStore.user.id}/library`"
              class="mobile-nav-link mobile-nav-link--sub"
              @click="closeMobileMenu"
            >
              <BookMarked :size="16" />{{ $t("navigation.myLibrary") }}
            </RouterLink>
            <RouterLink
              :to="`/users/${authStore.user.id}/achievements`"
              class="mobile-nav-link mobile-nav-link--sub"
              @click="closeMobileMenu"
            >
              <Trophy :size="16" />{{ $t("navigation.achievement") }}
            </RouterLink>
            <RouterLink
              :to="{ name: 'feed' }"
              class="mobile-nav-link mobile-nav-link--sub"
              @click="closeMobileMenu"
            >
              <Rss :size="16" />{{ $t("navigation.feed") }}
            </RouterLink>
            <button
              v-if="authStore.isAdmin"
              class="mobile-nav-link mobile-nav-link--sub mobile-nav-link--admin"
              @click="handleEnterAdmin"
            >
              <Shield :size="16" />{{ $t("navigation.adminMode") }}
            </button>
            <button
              class="mobile-nav-link mobile-nav-link--sub"
              @click="handleSwitchToPlanzi"
            >
              <img src="/planzi-icon-192.png" alt="" class="switch-platform-icon" :width="16" :height="16" />{{ $t("navigation.switchToPlanzi") }}
            </button>
            <button
              class="mobile-nav-link mobile-nav-link--sub mobile-nav-link--danger"
              @click="handleLogout"
            >
              <LogOut :size="16" />{{ $t("navigation.logout") }}
            </button>
          </template>
        </nav>
      </div>
    </Transition>

    <main class="page-content">
      <RouterView v-slot="{ Component }">
        <Transition name="page" mode="out-in">
          <component :is="Component" />
        </Transition>
      </RouterView>
    </main>
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted, onUnmounted, watch } from "vue";
import { useRouter, useRoute } from "vue-router";
import { useAuthStore } from "@/stores/auth";
import { useAdminStore } from "@/stores/admin";
import { useNotificationStore } from "@/stores/notification";
import { useLocale, i18n } from "@/i18n";
import NotificationPanel from "@/components/common/NotificationPanel.vue";
import HeaderSearch from "@/components/search/HeaderSearch.vue";
import { linkApi } from "@/api/endpoints/link";
import { playPlatformExit } from "@/utils/platformTransition";
import {
  Bell,
  User as UserIcon,
  ChevronDown,
  BookMarked,
  Trophy,
  Rss,
  LogOut,
  Shield,
  } from "lucide-vue-next";

const router = useRouter();
const route = useRoute();
const authStore = useAuthStore();
const adminStore = useAdminStore();
const notificationStore = useNotificationStore();
const localeStore = useLocale();

const scrolled = ref(false);
const userMenuOpen = ref(false);
const userMenuRef = ref<HTMLElement | null>(null);
const mobileMenuOpen = ref(false);


async function handleSwitchToPlanzi() {
  userMenuOpen.value = false;
  mobileMenuOpen.value = false;
  try {
    const { data } = await linkApi.createHandoff("planzi");
    await playPlatformExit();
    window.location.href = data.redirect_url;
  } catch {
    window.alert(i18n.global.t("navigation.link.switchFailed"));
  }
}

function toggleNotificationPanel() {
  userMenuOpen.value = false;
  mobileMenuOpen.value = false;
  notificationStore.togglePanel();
}

function onScroll() {
  scrolled.value = window.scrollY > 20;
}

function closeMobileMenu() {
  mobileMenuOpen.value = false;
}

// ปุ่ม avatar และปุ่ม hamburger ต้อง "เปิดพร้อมกันไม่ได้"
// ไม่งั้นจะเห็น dropdown เล็ก + mobile menu เต็มจอซ้อนกัน (ตามที่เจอในภาพ)
function toggleUserMenu() {
  userMenuOpen.value = !userMenuOpen.value;
  if (userMenuOpen.value) {
    mobileMenuOpen.value = false;
    notificationStore.closePanel();
  }
}

function toggleMobileMenu() {
  mobileMenuOpen.value = !mobileMenuOpen.value;
  if (mobileMenuOpen.value) {
    userMenuOpen.value = false;
    notificationStore.closePanel();
  }
}

function onClickOutside(e: MouseEvent) {
  if (userMenuRef.value && !userMenuRef.value.contains(e.target as Node)) {
    userMenuOpen.value = false;
  }
}

async function handleLogout() {
  userMenuOpen.value = false;
  mobileMenuOpen.value = false;
  adminStore.reset();
  notificationStore.reset(); // ปิด socket + เคลียร์ notification ของ user เดิม
  await authStore.logout();
  router.push({ name: "login" });
}

function handleEnterAdmin() {
  userMenuOpen.value = false;
  mobileMenuOpen.value = false;
  adminStore.enterAdminMode();
  router.push("/admin");
}

// ปิดเมนูมือถืออัตโนมัติเมื่อขยายจอกลับมาเป็นเดสก์ท็อป
function onResize() {
  if (window.innerWidth > 900 && mobileMenuOpen.value) {
    mobileMenuOpen.value = false;
  }
}

// ล็อกการ scroll พื้นหลังตอนเปิดเมนูมือถือ
watch(mobileMenuOpen, (open) => {
  document.body.style.overflow = open ? "hidden" : "";
});

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

onMounted(() => {
  window.addEventListener("scroll", onScroll, { passive: true });
  document.addEventListener("click", onClickOutside);
  window.addEventListener("resize", onResize);
});
onUnmounted(() => {
  window.removeEventListener("scroll", onScroll);
  document.removeEventListener("click", onClickOutside);
  window.removeEventListener("resize", onResize);
  document.body.style.overflow = "";
});
</script>

<style scoped>
.app-shell {
  min-height: 100vh;
  background: var(--color-surface);
}

/* ── Navbar ─────────────────────────────────────────── */
.navbar {
  position: fixed;
  top: 0;
  left: 0;
  right: 0;
  z-index: 100;
  display: flex;
  align-items: center;
  gap: 2rem;
  padding: 0 2rem;
  height: 64px;
  transition:
    background 0.3s,
    backdrop-filter 0.3s,
    border-color 0.3s,
    padding 0.3s,
    height 0.3s;
  border-bottom: 1px solid transparent;
}
.navbar--scrolled {
  background: rgba(20, 20, 20, 0.92);
  backdrop-filter: blur(16px);
  -webkit-backdrop-filter: blur(16px);
  border-color: rgba(255, 255, 255, 0.06);
}

.switch-platform-icon {
  flex-shrink: 0;
  object-fit: contain;
}

.nav-logo-icon {
  display: inline-block;
  width: 28px;
  height: 28px;
  margin-right: 6px;
  object-fit: contain;
  vertical-align: middle;
}

.nav-logo {
  font-family: var(--font-sans);
  font-size: 1.5rem;
  font-weight: 900;
  text-decoration: none;
  letter-spacing: -0.5px;
  flex-shrink: 0;
  z-index: 1;
}
.logo-movie {
  color: #ffffff;
}
.logo-hub {
  color: var(--color-brand);
}

.nav-links {
  display: flex;
  align-items: center;
  gap: 0.25rem;
}
.nav-link {
  padding: 0.4rem 0.75rem;
  font-size: 0.82rem;
  font-weight: 600;
  letter-spacing: 0.8px;
  color: #a3a3a3;
  text-decoration: none;
  border-radius: 6px;
  white-space: nowrap;
  transition:
    color 0.2s,
    background 0.2s;
}
.nav-link:hover {
  color: #fff;
}
.nav-link.active {
  color: #fff;
}

.nav-right {
  margin-left: auto;
  display: flex;
  align-items: center;
  gap: 0.5rem;
}

/* Notification Wrapper & Badge */
.notification-wrapper {
  position: relative;
  display: flex;
  align-items: center;
}

.bell-badge {
  position: absolute;
  top: 2px;
  right: 2px;
  background: var(--color-brand);
  color: white;
  font-size: 0.65rem;
  font-weight: bold;
  min-width: 16px;
  height: 16px;
  border-radius: 50%;
  display: flex;
  align-items: center;
  justify-content: center;
  padding: 0 4px;
  border: 2px solid #141414;
  animation: pulse-badge 0.3s ease-out;
}

@keyframes pulse-badge {
  0% {
    transform: scale(0.6);
  }
  50% {
    transform: scale(1.2);
  }
  100% {
    transform: scale(1);
  }
}

.lang-btn {
  background: none;
  border: 1px solid rgba(255, 255, 255, 0.15);
  cursor: pointer;
  color: #a3a3a3;
  display: flex;
  align-items: center;
  justify-content: center;
  width: 36px;
  height: 36px;
  border-radius: 50%;
  font-size: 0.72rem;
  font-weight: 700;
  transition: color 0.2s, background 0.2s, border-color 0.2s;
  flex-shrink: 0;
}
.lang-btn:hover {
  color: #fff;
  border-color: #fff;
  background: rgba(255, 255, 255, 0.08);
}

/* Icon button */
.icon-btn {
  background: none;
  border: none;
  cursor: pointer;
  color: #a3a3a3;
  display: flex;
  align-items: center;
  justify-content: center;
  width: 36px;
  height: 36px;
  border-radius: 50%;
  transition:
    color 0.2s,
    background 0.2s;
  flex-shrink: 0;
  position: relative;
}
.icon-btn:hover,
.has-unread {
  color: #fff;
  background: rgba(255, 255, 255, 0.08);
}

/* User menu */
.user-menu {
  position: relative;
  flex-shrink: 0;
}
.user-trigger {
  display: flex;
  align-items: center;
  gap: 0.5rem;
  background: rgba(255, 255, 255, 0.08);
  border: 1px solid rgba(255, 255, 255, 0.1);
  border-radius: 9999px;
  padding: 0.3rem 0.75rem 0.3rem 0.3rem;
  cursor: pointer;
  color: #fff;
  transition: background 0.2s;
}
.user-trigger:hover {
  background: rgba(255, 255, 255, 0.12);
}
.user-avatar {
  width: 28px;
  height: 28px;
  border-radius: 50%;
  background: #2a2a2a;
  display: flex;
  align-items: center;
  justify-content: center;
  overflow: hidden;
  flex-shrink: 0;
}
.user-avatar img {
  width: 100%;
  height: 100%;
  object-fit: cover;
}
.user-name {
  font-size: 0.85rem;
  font-weight: 500;
  white-space: nowrap;
}

.dropdown-menu {
  position: absolute;
  top: calc(100% + 8px);
  right: 0;
  min-width: 160px;
  background: #1f1f1f;
  border: 1px solid rgba(255, 255, 255, 0.08);
  border-radius: 10px;
  padding: 0.4rem;
  box-shadow: 0 16px 48px rgba(0, 0, 0, 0.6);
}
.dropdown-item {
  display: flex;
  align-items: center;
  gap: 0.5rem;
  padding: 0.55rem 0.75rem;
  font-size: 0.875rem;
  color: #d1d1d1;
  text-decoration: none;
  border-radius: 6px;
  border: none;
  background: none;
  width: 100%;
  cursor: pointer;
  transition:
    background 0.15s,
    color 0.15s;
  text-align: left;
}
.dropdown-item:hover {
  background: rgba(255, 255, 255, 0.07);
  color: #fff;
}
.dropdown-item--danger:hover {
  background: rgba(245, 67, 44, 0.12);
  color: var(--color-brand);
}
.dropdown-divider {
  height: 1px;
  background: rgba(255, 255, 255, 0.07);
  margin: 0.3rem 0;
}

/* Dropdown transition */
.dropdown-enter-active,
.dropdown-leave-active {
  transition:
    opacity 0.15s,
    transform 0.15s;
}
.dropdown-enter-from,
.dropdown-leave-to {
  opacity: 0;
  transform: translateY(-6px);
}

/* ── Hamburger (ซ่อนเป็นค่าเริ่มต้นบนเดสก์ท็อป) ───────── */
.hamburger-btn {
  display: none;
  flex-direction: column;
  justify-content: center;
  align-items: center;
  gap: 5px;
  width: 36px;
  height: 36px;
  background: none;
  border: none;
  cursor: pointer;
  flex-shrink: 0;
  z-index: 1;
}
.hamburger-line {
  width: 20px;
  height: 2px;
  background: #fff;
  border-radius: 2px;
  transition:
    transform 0.25s ease,
    opacity 0.2s ease;
}
.hamburger-btn--active .hamburger-line:nth-child(1) {
  transform: translateY(7px) rotate(45deg);
}
.hamburger-btn--active .hamburger-line:nth-child(2) {
  opacity: 0;
}
.hamburger-btn--active .hamburger-line:nth-child(3) {
  transform: translateY(-7px) rotate(-45deg);
}

/* ── Mobile slide-down menu ───────────────────────────── */
.mobile-menu-overlay {
  position: fixed;
  inset: 64px 0 0 0;
  background: rgba(0, 0, 0, 0.55);
  z-index: 99;
}
.mobile-menu-panel {
  background: #181818;
  border-bottom: 1px solid rgba(255, 255, 255, 0.08);
  display: flex;
  flex-direction: column;
  padding: 0.5rem 1rem 1.25rem;
  max-height: calc(100vh - 64px);
  overflow-y: auto;
}
.mobile-nav-link {
  display: flex;
  align-items: center;
  gap: 0.6rem;
  padding: 0.85rem 0.5rem;
  font-size: 0.9rem;
  font-weight: 600;
  letter-spacing: 0.6px;
  color: #d1d1d1;
  text-decoration: none;
  border-bottom: 1px solid rgba(255, 255, 255, 0.05);
  background: none;
  border-left: none;
  border-right: none;
  border-top: none;
  width: 100%;
  text-align: left;
  cursor: pointer;
}
.mobile-nav-link.active {
  color: #fff;
}
.mobile-nav-link--sub {
  font-size: 0.85rem;
  font-weight: 500;
  color: #a3a3a3;
}
.mobile-nav-link--danger {
  color: var(--color-brand);
}
.mobile-menu-divider {
  height: 1px;
  background: rgba(255, 255, 255, 0.08);
  margin: 0.5rem 0;
}

.mobile-menu-enter-active,
.mobile-menu-leave-active {
  transition: opacity 0.2s ease;
}
.mobile-menu-enter-from,
.mobile-menu-leave-to {
  opacity: 0;
}

/* ── Page content ──────────────────────────────────── */
.page-content {
  padding-top: 64px;
}

/* ════════════════════════════════════════════════════════════
   RESPONSIVE BREAKPOINTS
   - ≥1440px : จอใหญ่ / iMac / desktop monitor
   - 1200–1439px : notebook / laptop มาตรฐาน (เช่น MacBook Pro 14"/16")
   - 1025–1199px : notebook ขนาดเล็ก / iPad Pro landscape
   - 901–1024px  : iPad / Tablet (landscape เล็ก, portrait ใหญ่)
   - 769–900px   : iPad Mini / Tablet portrait, mobile landscape ใหญ่
   - 577–768px   : Tablet แนวตั้งเล็ก / mobile landscape
   - 421–576px   : มือถือจอใหญ่ (เช่น iPhone Pro Max, Android ใหญ่)
   - ≤420px      : มือถือจอมาตรฐาน/เล็ก (iPhone SE ฯลฯ)
   ════════════════════════════════════════════════════════════ */

@media (min-width: 1440px) {
  .navbar {
    padding: 0 3rem;
    gap: 2.5rem;
  }
  .nav-link {
    font-size: 0.85rem;
    padding: 0.45rem 0.9rem;
  }
}

@media (max-width: 1439px) {
  .navbar {
    padding: 0 1.75rem;
    gap: 1.5rem;
  }
}

@media (max-width: 1199px) {
  .navbar {
    gap: 1rem;
    padding: 0 1.5rem;
  }
  .nav-links {
    gap: 0;
  }
  .nav-link {
    font-size: 0.78rem;
    padding: 0.4rem 0.55rem;
  }
  .user-name {
    max-width: 90px;
    overflow: hidden;
    text-overflow: ellipsis;
  }
}

@media (max-width: 1024px) {
  .nav-links {
    display: none;
  }
  .hamburger-btn {
    display: flex;
  }
  .navbar {
    gap: 0.75rem;
  }
  .user-name {
    display: none;
  }
  .user-trigger {
    padding: 0.3rem;
    gap: 0;
  }
  .user-trigger .rotate-180,
  .user-trigger > svg:last-child {
    display: none;
  }
  .dropdown-menu {
    display: none !important;
  }
}

@media (max-width: 900px) {
  .navbar {
    padding: 0 1.25rem;
    height: 60px;
  }
  .page-content {
    padding-top: 60px;
  }
  .mobile-menu-overlay {
    inset: 60px 0 0 0;
  }
  .mobile-menu-panel {
    max-height: calc(100vh - 60px);
  }
  .nav-logo {
    font-size: 1.35rem;
  }
}

@media (max-width: 768px) {
  .navbar {
    padding: 0 1rem;
    gap: 0.5rem;
  }
  .nav-right {
    gap: 0.25rem;
  }
  .icon-btn {
    width: 32px;
    height: 32px;
  }
}

@media (max-width: 576px) {
  .navbar {
    padding: 0 0.75rem;
    height: 56px;
  }
  .page-content {
    padding-top: 56px;
  }
  .mobile-menu-overlay {
    inset: 56px 0 0 0;
  }
  .mobile-menu-panel {
    max-height: calc(100vh - 56px);
    padding: 0.5rem 0.75rem 1rem;
  }
  .nav-logo {
    font-size: 1.2rem;
  }
  .notification-wrapper {
    display: flex;
  }
  .bell-badge {
    min-width: 14px;
    height: 14px;
    font-size: 0.6rem;
  }
  .hamburger-btn {
    width: 32px;
    height: 32px;
  }
}

@media (max-width: 420px) {
  .navbar {
    gap: 0.35rem;
    padding: 0 0.6rem;
  }
  .nav-logo {
    font-size: 1.05rem;
  }
  .icon-btn {
    width: 30px;
    height: 30px;
  }
  .user-avatar {
    width: 26px;
    height: 26px;
  }
  .mobile-nav-link {
    font-size: 0.85rem;
    padding: 0.75rem 0.4rem;
  }
}

@media (-webkit-min-device-pixel-ratio: 2), (min-resolution: 192dpi) {
  .navbar {
    border-bottom-width: 0.5px;
  }
}

@media (max-height: 420px) and (orientation: landscape) {
  .mobile-menu-panel {
    max-height: calc(100vh - 56px);
    overflow-y: auto;
  }
  .navbar {
    height: 52px;
  }
  .page-content {
    padding-top: 52px;
  }
}

@media (prefers-reduced-motion: reduce) {
  .navbar,
  .hamburger-line,
  .dropdown-enter-active,
  .dropdown-leave-active,
  .mobile-menu-enter-active,
  .mobile-menu-leave-active {
    transition: none !important;
  }
}
</style>
