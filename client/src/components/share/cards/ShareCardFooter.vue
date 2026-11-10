<template>
  <footer class="sc-footer">
    <!-- hardcode-ok: brand wordmark -->
    <span class="sc-wordmark"><span class="sc-wordmark__pop">POP</span><span>CORN</span></span>
    <span class="sc-footer__url">{{ url }}</span>
  </footer>
</template>

<script setup lang="ts">
defineProps<{ url: string }>()
</script>
