<template>
  <main class="home-page min-h-screen bg-accent-50 px-5 py-10 text-primary-900 dark:bg-dark-950 dark:text-dark-100 sm:px-8 sm:py-14">
    <header class="mx-auto flex w-full max-w-3xl items-center gap-3 border-b border-accent-200 pb-5 dark:border-dark-700">
      <img
        :src="siteLogo || '/logo.svg'"
        :alt="siteName"
        class="h-9 w-9 object-contain"
      />
      <span class="text-lg font-semibold tracking-wide">{{ siteName }}</span>
    </header>

    <section class="mx-auto flex min-h-[calc(100vh-10rem)] w-full max-w-3xl flex-col items-center justify-center text-center">
      <p
        data-testid="daily-quote"
        class="mb-8 max-w-2xl text-xl leading-relaxed tracking-wide text-primary-800 sm:mb-10 sm:text-2xl dark:text-dark-100"
      >
        {{ quote || '正在加载一言…' }}
      </p>

      <div class="image-frame w-full max-w-2xl overflow-hidden bg-accent-100 dark:bg-dark-900">
        <img
          :src="wallpaperUrl"
          alt="Snoopy 随机壁纸"
          class="block h-auto max-h-[58vh] w-full object-contain"
        />
      </div>

      <div class="mt-8 flex w-full max-w-md flex-col gap-3 sm:mt-10 sm:flex-row sm:justify-center">
        <router-link to="/login" class="home-button home-button-primary">登录</router-link>
        <router-link to="/register" class="home-button home-button-secondary">注册</router-link>
      </div>
    </section>
  </main>
</template>

<script setup lang="ts">
import { computed, onMounted, ref } from 'vue'
import { useAppStore } from '@/stores'
import { sanitizeUrl } from '@/utils/url'

const appStore = useAppStore()
const siteName = computed(() => appStore.cachedPublicSettings?.site_name || appStore.siteName || 'Sub2API')
const siteLogo = computed(() => sanitizeUrl(appStore.cachedPublicSettings?.site_logo || appStore.siteLogo || '', { allowRelative: true, allowDataUrl: true }))

const wallpapers = [
  '/wallpapers/snoopy/1194837-min.webp',
  '/wallpapers/snoopy/1220422-min.webp',
  '/wallpapers/snoopy/1220424-min.webp',
  '/wallpapers/snoopy/1220427-min.webp',
  '/wallpapers/snoopy/1220428-min.webp',
  '/wallpapers/snoopy/1220429-min.webp',
  '/wallpapers/snoopy/1220430-min.webp',
  '/wallpapers/snoopy/1220431-min.webp',
  '/wallpapers/snoopy/1220433-min.webp',
  '/wallpapers/snoopy/1220436-min.webp',
  '/wallpapers/snoopy/1220441-min.webp',
  '/wallpapers/snoopy/1220442-min.webp',
  '/wallpapers/snoopy/1220446-min.webp',
  '/wallpapers/snoopy/1220447-min.webp',
  '/wallpapers/snoopy/1220450-min.webp',
  '/wallpapers/snoopy/1220451-min.webp',
  '/wallpapers/snoopy/1220452-min.webp',
  '/wallpapers/snoopy/436482-min.webp',
  '/wallpapers/snoopy/468478-min.webp',
  '/wallpapers/snoopy/469505-min.webp',
]

const quote = ref('')
const wallpaperUrl = ref(wallpapers[Math.floor(Math.random() * wallpapers.length)])

async function loadQuote() {
  try {
    const response = await fetch('/api/v1/public/quote', { cache: 'no-store' })
    if (!response.ok) throw new Error(`Quote request failed: ${response.status}`)
    const nextQuote = (await response.text()).trim()
    if (!nextQuote) throw new Error('Quote response is empty')
    quote.value = nextQuote
  } catch (error) {
    console.error('Failed to load daily quote:', error)
    quote.value = '暂时无法获取一言，请稍后刷新。'
  }
}

onMounted(() => {
  if (!appStore.publicSettingsLoaded) appStore.fetchPublicSettings()
  loadQuote()
})
</script>

<style scoped>
.home-page {
  font-family: 'LXGW WenKai', '霞鹜文楷', cursive;
}

.image-frame {
  border: 1px solid theme('colors.accent.200');
  border-radius: 0 !important;
  box-shadow: 0 8px 24px rgb(36 37 34 / 0.08);
}

.home-button {
  display: inline-flex;
  min-height: 46px;
  align-items: center;
  justify-content: center;
  border-radius: 0 !important;
  padding: 0.7rem 1.5rem;
  font-size: 1rem;
  font-weight: 700;
  transition: background-color 160ms ease, color 160ms ease, box-shadow 160ms ease;
}

.home-button:focus-visible {
  outline: 3px solid rgb(120 113 108 / 0.45);
  outline-offset: 3px;
}

.home-button-primary {
  background: #292524;
  color: #fff;
  box-shadow: 0 6px 16px rgb(68 64 60 / 0.16);
}

.home-button-primary:hover {
  background: #44403c;
}

.home-button-secondary {
  border: 1px solid #a8a29e;
  background: #fff;
  color: #44403c;
}

.home-button-secondary:hover {
  background: #e7e5e4;
}

@media (prefers-reduced-motion: reduce) {
  .home-button {
    transition: none;
  }
}
</style>
