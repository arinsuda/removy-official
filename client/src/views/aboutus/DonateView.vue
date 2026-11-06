<script setup lang="ts">
import { ref } from 'vue'
import {
  QrCode,
  Landmark,
  Copy,
  Check,
  Coffee,
  Server,
  Zap,
  ShieldCheck
} from 'lucide-vue-next'

const qrError = ref(false)
const bankLogoError = ref(false)
const copied = ref(false)

const handleQrError = () => {
  qrError.value = true
}

const handleBankLogoError = () => {
  bankLogoError.value = true
}

const copyAccountNumber = async () => {
  try {
    await navigator.clipboard.writeText('0453315436')
    copied.value = true
    setTimeout(() => {
      copied.value = false
    }, 2000)
  } catch (err) {
    console.error('Failed to copy text: ', err)
  }
}
</script>

<template>
  <div class="donate-page">
    <!-- Animated Background -->
    <div class="background-animation">
      <div class="gradient-sphere sphere-1"></div>
      <div class="gradient-sphere sphere-2"></div>
      <div class="gradient-sphere sphere-3"></div>
      <div class="grid-overlay"></div>
    </div>

    <div class="content-container">
      <!-- Hero Section -->
      <section class="hero-section fade-in-up">
        <div class="hero-icon-wrapper">
          <Coffee class="hero-icon" :size="48" stroke-width="1.5" />
        </div>
        <h1 class="hero-title">{{ $t("donate.title") }}</h1>
        <p class="hero-subtitle">
          {{ $t("donate.subtitle") }}
        </p>
      </section>

      <!-- Main Donation Cards -->
      <section class="donation-grid fade-in-up delay-1">
        <!-- PromptPay Card -->
        <div class="glass-card premium-card">
          <div class="card-glow"></div>
          <div class="card-header">
            <div class="icon-box qr-theme">
              <QrCode :size="24" />
            </div>
            <h2>{{ $t("donate.promptpay.title") }}</h2>
          </div>
          <div class="card-body flex-center">
            <div class="qr-container">
              <div class="qr-corners">
                <span></span><span></span><span></span><span></span>
              </div>
              <img
                v-if="!qrError"
                src="/mePromptPay.png"
                alt="PromptPay QR Code"
                class="qr-image"
                @error="handleQrError"
              />
              <div v-else class="fallback-box">
                <QrCode :size="48" class="text-muted" />
                <p>{{ $t("donate.promptpay.fallback") }}</p>
              </div>
            </div>
            <p class="promptpay-name">{{ $t("donate.promptpay.name") }}</p>
            <p class="qr-hint">{{ $t("donate.promptpay.hint") }}</p>
          </div>
        </div>

        <!-- Bank Transfer Card -->
        <div class="glass-card premium-card">
          <div class="card-glow"></div>
          <div class="card-header">
            <div class="icon-box bank-theme">
              <Landmark :size="24" />
            </div>
            <h2>{{ $t("donate.bank.title") }}</h2>
          </div>
          <div class="card-body">
            <div class="bank-details">
              <div class="bank-logo-wrapper">
                <img
                  v-if="!bankLogoError"
                  src="/kbank-logo.png"
                  alt="Kasikorn Bank"
                  class="bank-logo"
                  @error="handleBankLogoError"
                />
                <Landmark v-else :size="40" class="fallback-bank-icon" />
              </div>
              <div class="bank-info">
                <h3>{{ $t("donate.bank.name") }}</h3>
                <p class="account-name">{{ $t("donate.bank.accountName") }}</p>
              </div>
            </div>
            <div class="account-number-box">
              <div class="account-number">
                <span class="number-digit">045</span>
                <span class="separator">-</span>
                <span class="number-digit">3</span>
                <span class="separator">-</span>
                <span class="number-digit">31543</span>
                <span class="separator">-</span>
                <span class="number-digit">6</span>
              </div>
              <button
                @click="copyAccountNumber"
                class="copy-btn"
                :class="{ 'copied': copied }"
                :aria-label="$t('donate.bank.copy')"
              >
                <span class="btn-content">
                  <Check v-if="copied" :size="18" class="icon-check" />
                  <Copy v-else :size="18" class="icon-copy" />
                  <span class="btn-text">{{ copied ? $t("donate.bank.copied") : $t("donate.bank.copy") }}</span>
                </span>
              </button>
            </div>
          </div>
        </div>
      </section>

      <!-- Impact Section -->
      <section class="impact-section fade-in-up delay-2">
        <h3 class="section-title">{{ $t("donate.impact.title") }}</h3>
        <div class="impact-grid">
          <div class="impact-card">
            <div class="impact-icon server-icon">
              <Server :size="28" />
            </div>
            <h4>{{ $t("donate.impact.server.title") }}</h4>
            <p>{{ $t("donate.impact.server.desc") }}</p>
          </div>
          <div class="impact-card">
            <div class="impact-icon feature-icon">
              <Zap :size="28" />
            </div>
            <h4>{{ $t("donate.impact.features.title") }}</h4>
            <p>{{ $t("donate.impact.features.desc") }}</p>
          </div>
          <div class="impact-card">
            <div class="impact-icon bug-icon">
              <ShieldCheck :size="28" />
            </div>
            <h4>{{ $t("donate.impact.maintenance.title") }}</h4>
            <p>{{ $t("donate.impact.maintenance.desc") }}</p>
          </div>
        </div>
      </section>

      <!-- Footer Thank You -->
      <footer class="donate-footer fade-in-up delay-3">
        <h2 class="thank-you-text">{{ $t("donate.thanks.title") }}</h2>
        <p>{{ $t("donate.thanks.subtitle") }}</p>
      </footer>
    </div>
  </div>
</template>

<style scoped>

.donate-page {
  font-family: 'Inter', 'Noto Sans Thai', sans-serif;
  min-height: 100vh;
  background-color: #0a0a0a;
  color: #f5f5f5;
  position: relative;
  overflow: hidden;
  padding: 4rem 1.5rem;
  display: flex;
  justify-content: center;
}

/* Background */
.background-animation {
  position: fixed;
  top: 0;
  left: 0;
  width: 100%;
  height: 100%;
  z-index: 0;
  overflow: hidden;
  pointer-events: none;
  contain: layout style paint;
}

.gradient-sphere {
  position: absolute;
  border-radius: 50%;
  opacity: 0.8;
  will-change: transform;
  transform: translate3d(0, 0, 0);
  pointer-events: none;
  animation: float-gpu 20s infinite ease-in-out alternate;
}

.sphere-1 {
  background: radial-gradient(circle, rgba(225, 37, 27, 0.3) 0%, rgba(225, 37, 27, 0.1) 40%, rgba(225, 37, 27, 0.02) 70%, transparent 100%);
  width: 50vw;
  height: 50vw;
  top: -10%;
  left: -10%;
  animation-delay: 0s;
}

.sphere-2 {
  background: radial-gradient(circle, rgba(255, 184, 0, 0.2) 0%, rgba(255, 184, 0, 0.07) 40%, rgba(255, 184, 0, 0.01) 70%, transparent 100%);
  width: 40vw;
  height: 40vw;
  bottom: -10%;
  right: -5%;
  animation-delay: -5s;
  animation-duration: 25s;
}

.sphere-3 {
  background: radial-gradient(circle, rgba(138, 22, 18, 0.25) 0%, rgba(138, 22, 18, 0.08) 40%, rgba(138, 22, 18, 0.01) 70%, transparent 100%);
  width: 45vw;
  height: 45vw;
  top: 40%;
  left: 30%;
  animation-delay: -10s;
  animation-duration: 30s;
}

@keyframes float-gpu {
  0% { transform: translate3d(0, 0, 0) scale(1); }
  50% { transform: translate3d(6vw, 4vh, 0) scale(1.05); }
  100% { transform: translate3d(0, 0, 0) scale(1); }
}

.grid-overlay {
  position: absolute;
  inset: 0;
  background-image: 
    linear-gradient(to right, rgba(255, 255, 255, 0.03) 1px, transparent 1px),
    linear-gradient(to bottom, rgba(255, 255, 255, 0.03) 1px, transparent 1px);
  background-size: 40px 40px;
  mask-image: radial-gradient(circle at center, black 40%, transparent 100%);
  -webkit-mask-image: radial-gradient(circle at center, black 40%, transparent 100%);
}

/* Layout */
.content-container {
  position: relative;
  z-index: 1;
  max-width: 1000px;
  width: 100%;
  display: flex;
  flex-direction: column;
  gap: 4rem;
}

/* Typography & Hero */
.hero-section {
  text-align: center;
  max-width: 700px;
  margin: 0 auto;
}

.hero-icon-wrapper {
  display: inline-flex;
  padding: 1.25rem;
  background: rgba(225, 37, 27, 0.1);
  border-radius: 50%;
  margin-bottom: 1.5rem;
  border: 1px solid rgba(225, 37, 27, 0.2);
  box-shadow: 0 0 30px rgba(225, 37, 27, 0.25);
  color: #e1251b;
}

.hero-title {
  font-size: clamp(2rem, 5vw, 3.5rem);
  font-weight: 800;
  margin-bottom: 1rem;
  background: linear-gradient(135deg, #ffffff 0%, #d1d5db 100%);
  -webkit-background-clip: text;
  -webkit-text-fill-color: transparent;
  letter-spacing: -0.02em;
}

.hero-subtitle {
  font-size: 1.125rem;
  color: #a3a3a3;
  line-height: 1.6;
}

/* Glass Cards */
.donation-grid {
  display: grid;
  grid-template-columns: 1fr;
  gap: 2rem;
}

@media (min-width: 768px) {
  .donation-grid {
    grid-template-columns: 1fr 1fr;
  }
}

.glass-card {
  background: rgba(255, 255, 255, 0.03);
  backdrop-filter: blur(8px);
  -webkit-backdrop-filter: blur(8px);
  border: 1px solid rgba(255, 255, 255, 0.08);
  border-radius: 24px;
  padding: 2rem;
  position: relative;
  overflow: hidden;
  transition: transform 0.4s cubic-bezier(0.175, 0.885, 0.32, 1.275), border-color 0.4s ease;
  display: flex;
  flex-direction: column;
}

.premium-card:hover {
  transform: translateY(-8px);
  border-color: rgba(255, 255, 255, 0.2);
}

.card-glow {
  position: absolute;
  top: 0;
  left: 0;
  width: 100%;
  height: 100%;
  background: radial-gradient(circle at 100% 0%, rgba(255,255,255,0.03) 0%, transparent 50%);
  pointer-events: none;
}

.card-header {
  display: flex;
  align-items: center;
  gap: 1rem;
  margin-bottom: 2rem;
}

.icon-box {
  display: flex;
  align-items: center;
  justify-content: center;
  width: 48px;
  height: 48px;
  border-radius: 16px;
  background: rgba(255, 255, 255, 0.05);
}

.qr-theme {
  color: #10b981;
  background: rgba(16, 185, 129, 0.1);
}

.bank-theme {
  color: #3b82f6;
  background: rgba(59, 130, 246, 0.1);
}

.card-header h2 {
  font-size: 1.25rem;
  font-weight: 700;
  margin: 0;
}

.card-body {
  flex: 1;
  display: flex;
  flex-direction: column;
}

.flex-center {
  align-items: center;
  text-align: center;
}

/* QR Code Container styling */
.qr-container {
  position: relative;
  width: 220px;
  height: 220px;
  background: #ffffff;
  padding: 10px;
  border-radius: 16px;
  box-shadow: 0 10px 30px rgba(0, 0, 0, 0.25);
  margin-bottom: 1.5rem;
  display: flex;
  align-items: center;
  justify-content: center;
}

.qr-corners {
  position: absolute;
  inset: -8px;
  pointer-events: none;
}

.qr-corners span {
  position: absolute;
  width: 20px;
  height: 20px;
  border-color: #e1251b;
  border-style: solid;
  transition: transform 0.3s ease, border-color 0.3s ease;
  z-index: 2;
  will-change: transform;
}

.qr-corners span:nth-child(1) { top: -4px; left: -4px; border-width: 3px 0 0 3px; border-top-left-radius: 8px; }
.qr-corners span:nth-child(2) { top: -4px; right: -4px; border-width: 3px 3px 0 0; border-top-right-radius: 8px; }
.qr-corners span:nth-child(3) { bottom: -4px; left: -4px; border-width: 0 0 3px 3px; border-bottom-left-radius: 8px; }
.qr-corners span:nth-child(4) { bottom: -4px; right: -4px; border-width: 0 3px 3px 0; border-bottom-right-radius: 8px; }

.qr-container:hover .qr-corners span {
  transform: scale(1.2);
  border-color: #ff473e;
}

.qr-container::before {
  content: '';
  position: absolute;
  top: 0;
  left: 0;
  width: 100%;
  height: 4px;
  background: rgba(225, 37, 27, 0.5);
  box-shadow: 0 0 10px rgba(225, 37, 27, 0.8);
  z-index: 3;
  opacity: 0;
  border-radius: 4px;
  will-change: transform;
}

.qr-container:hover::before {
  opacity: 1;
  animation: scan 3s infinite linear;
}

@keyframes scan {
  0% { transform: translateY(0); }
  50% { transform: translateY(220px); }
  100% { transform: translateY(0); }
}

.qr-image {
  width: 200px;
  height: 200px;
  object-fit: contain;
  border-radius: 8px;
}

.promptpay-name {
  font-size: 1.125rem;
  font-weight: 700;
  margin-bottom: 0.25rem;
}

.qr-hint {
  font-size: 0.875rem;
  color: #a3a3a3;
}

.fallback-box {
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  color: #52525b;
}

.fallback-box p {
  font-size: 0.825rem;
  margin-top: 0.5rem;
}

/* Bank Details Styling */
.bank-details {
  display: flex;
  align-items: center;
  gap: 1rem;
  margin-bottom: 2rem;
}

.bank-logo-wrapper {
  width: 56px;
  height: 56px;
  border-radius: 14px;
  background: #ffffff;
  padding: 4px;
  display: flex;
  align-items: center;
  justify-content: center;
}

.bank-logo {
  width: 100%;
  height: 100%;
  object-fit: contain;
}

.fallback-bank-icon {
  color: #10b981;
}

.bank-info h3 {
  font-size: 1.125rem;
  font-weight: 700;
  margin-bottom: 0.25rem;
}

.account-name {
  font-size: 0.875rem;
  color: #a3a3a3;
}

.account-number-box {
  background: rgba(0, 0, 0, 0.3);
  border: 1px solid rgba(255, 255, 255, 0.05);
  border-radius: 16px;
  padding: 1rem 1.5rem;
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 1rem;
}

.account-number {
  font-size: 1.5rem;
  font-weight: 700;
  letter-spacing: 0.05em;
  font-family: monospace;
}

.separator {
  color: #52525b;
  margin: 0 0.25rem;
}

.copy-btn {
  background: #ffffff;
  color: #0a0a0a;
  border: none;
  border-radius: 12px;
  padding: 0.75rem 1.25rem;
  font-weight: 600;
  cursor: pointer;
  transition: transform 0.2s, background-color 0.2s;
}

.copy-btn:hover {
  background: #e5e7eb;
}

.copy-btn:active {
  transform: scale(0.95);
}

.copy-btn.copied {
  background: #10b981;
  color: #ffffff;
}

.btn-content {
  display: flex;
  align-items: center;
  gap: 0.5rem;
}

.btn-text {
  font-size: 0.875rem;
}

/* Impact Section */
.impact-section {
  text-align: center;
}

.section-title {
  font-size: 1.5rem;
  font-weight: 700;
  margin-bottom: 2.5rem;
}

.impact-grid {
  display: grid;
  grid-template-columns: 1fr;
  gap: 2rem;
}

@media (min-width: 768px) {
  .impact-grid {
    grid-template-columns: repeat(3, 1fr);
  }
}

.impact-card {
  background: rgba(255, 255, 255, 0.02);
  border: 1px solid rgba(255, 255, 255, 0.05);
  border-radius: 20px;
  padding: 2rem 1.5rem;
  transition: transform 0.3s ease, border-color 0.3s ease, background-color 0.3s ease;
}

.impact-card:hover {
  transform: translateY(-4px);
  border-color: rgba(255, 255, 255, 0.1);
  background: rgba(255, 255, 255, 0.04);
}

.impact-icon {
  width: 56px;
  height: 56px;
  border-radius: 16px;
  display: flex;
  align-items: center;
  justify-content: center;
  margin: 0 auto 1.5rem;
}

.server-icon { color: #3b82f6; background: rgba(59, 130, 246, 0.1); }
.feature-icon { color: #a855f7; background: rgba(168, 85, 247, 0.1); }
.bug-icon { color: #10b981; background: rgba(16, 185, 129, 0.1); }

.impact-card h4 {
  font-size: 1.125rem;
  font-weight: 700;
  margin-bottom: 0.5rem;
}

.impact-card p {
  font-size: 0.875rem;
  color: #a3a3a3;
  line-height: 1.5;
}

/* Footer styling */
.donate-footer {
  text-align: center;
  max-width: 600px;
  margin: 0 auto;
}

.thank-you-text {
  font-size: 1.75rem;
  font-weight: 800;
  background: linear-gradient(135deg, #ff6b6b 0%, #e1251b 100%);
  -webkit-background-clip: text;
  -webkit-text-fill-color: transparent;
  margin-bottom: 0.5rem;
}

.donate-footer p {
  color: #a3a3a3;
}

/* Entrance Animations */
.fade-in-up {
  opacity: 0;
  transform: translateY(30px);
  animation: fadeInUp 0.8s cubic-bezier(0.16, 1, 0.3, 1) forwards;
}

.delay-1 { animation-delay: 0.2s; }
.delay-2 { animation-delay: 0.4s; }
.delay-3 { animation-delay: 0.6s; }

@keyframes fadeInUp {
  from {
    opacity: 0;
    transform: translateY(30px);
  }
  to {
    opacity: 1;
    transform: translateY(0);
  }
}

/* Responsive adjustments */
@media (max-width: 768px) {
  .gradient-sphere {
    animation: none !important;
  }
  .glass-card, .impact-card {
    backdrop-filter: none !important;
    -webkit-backdrop-filter: none !important;
    background: #161616;
  }
}

@media (max-width: 480px) {
  .donate-page {
    padding: 2rem 1rem;
  }
  
  .account-number {
    font-size: 1.25rem;
  }
  
  .bank-details {
    flex-direction: column;
    text-align: center;
  }
}

@media (prefers-reduced-motion: reduce) {
  .gradient-sphere, .hero-icon-wrapper::after, .qr-container::before {
    animation: none !important;
    transition: none !important;
  }
  .fade-in-up {
    opacity: 1 !important;
    transform: none !important;
    animation: none !important;
  }
}
</style>
