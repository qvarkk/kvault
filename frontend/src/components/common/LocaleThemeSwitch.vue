<script setup lang="ts">
import { useI18n } from "vue-i18n"
import { useDark } from "@vueuse/core"
import {
  DropdownMenu,
  DropdownMenuContent,
  DropdownMenuItem,
  DropdownMenuTrigger,
} from "../ui/dropdown-menu"
import { Button } from "../ui/button"
import { Moon, Sun } from "lucide-vue-next"

const { locale } = useI18n()
const isDark = useDark()
const toggleTheme = () => {
  isDark.value = !isDark.value
}

function switchLocale(lang: string) {
  locale.value = lang
  localStorage.setItem("locale", lang)
}
</script>

<template>
  <div class="flex items-center gap-2">
    <DropdownMenu>
      <DropdownMenuTrigger as-child>
        <Button variant="ghost" size="sm">
          {{ locale.toUpperCase() }}
        </Button>
      </DropdownMenuTrigger>
      <DropdownMenuContent>
        <DropdownMenuItem @click="switchLocale('en')">English</DropdownMenuItem>
        <DropdownMenuItem @click="switchLocale('ru')">Русский</DropdownMenuItem>
        <DropdownMenuItem @click="switchLocale('ja')">日本語</DropdownMenuItem>
      </DropdownMenuContent>
    </DropdownMenu>

    <Button variant="ghost" size="sm" @click="toggleTheme">
      <Sun v-if="isDark" class="w-4 h-4" />
      <Moon v-else class="w-4 h-4" />
    </Button>
  </div>
</template>
