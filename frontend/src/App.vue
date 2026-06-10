<script setup lang="ts">
import "vue-sonner/style.css"
import { onMounted } from "vue"
import { useAuthStore } from "./stores/auth"
import { Toaster } from "vue-sonner"

const auth = useAuthStore()

onMounted(() => {
  auth.init()
})
</script>

<template>
  <RouterView />
  <Toaster />
</template>

<style scoped></style>
