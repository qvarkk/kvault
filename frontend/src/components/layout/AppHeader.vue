<script setup lang="ts">
import { useSidebar } from "@/components/ui/sidebar"
import LocaleThemeSwitch from "@/components/common/LocaleThemeSwitch.vue"
import { SidebarOpen } from "lucide-vue-next"
import { Button } from "../ui/button"

const { isMobile, toggleSidebar } = useSidebar()
</script>

<template>
  <header class="flex items-center justify-between px-4 h-14 border-b">
    <Button v-if="isMobile" variant="ghost" @click="toggleSidebar">
      <SidebarOpen />
    </Button>
    <div class="flex items-center gap-2 ml-auto">
      <LocaleThemeSwitch />
    </div>
  </header>
</template>
