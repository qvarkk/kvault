<script setup lang="ts">
import LocaleThemeSwitch from "../common/LocaleThemeSwitch.vue"
</script>

<template>
  <main>
    <div class="min-h-screen flex items-center justify-center bg-background">
      <div class="absolute top-4 right-4">
        <LocaleThemeSwitch />
      </div>
      <div class="w-full max-w-sm px-4">
        <slot />
      </div>
    </div>
  </main>
</template>
