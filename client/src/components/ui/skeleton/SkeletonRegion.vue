<template>
  <div role="status" aria-busy="true" :aria-label="$t('common.loading')">
    <slot />
  </div>
</template>
