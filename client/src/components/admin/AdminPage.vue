<template>
  <div class="admin-page"><slot /></div>
</template>

<style scoped>
.admin-page {
  display: flex;
  flex-direction: column;
  gap: 1rem;
  min-width: 0;
}
</style>
