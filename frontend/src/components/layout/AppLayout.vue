<script setup lang="ts">
import { SidebarProvider } from "../ui/sidebar"
import AppSidebar from "./AppSidebar.vue"
import AppHeader from "./AppHeader.vue"
</script>

<template>
  <SidebarProvider>
    <AppSidebar />
    <div class="flex flex-col flex-1 min-w-0">
      <AppHeader />
      <main class="flex justify-center flex-1 p-4 overflow-x-hidden">
        <slot />
      </main>
    </div>
  </SidebarProvider>
</template>
