<script setup lang="ts">
import { computed } from "vue"
import { useI18n } from "vue-i18n"
import { CircleHelp } from "lucide-vue-next"
import { Button } from "@/components/ui/button"
import {
  Popover,
  PopoverContent,
  PopoverTrigger,
} from "@/components/ui/popover"
import { Kbd, KbdGroup } from "../ui/kbd"
import { useEditorShortcuts } from "@/composables/useEditorShortcuts"
import { toast } from "vue-sonner"

const { t } = useI18n()
const { isHidden, hide, show } = useEditorShortcuts()

// Computed so labels re-translate on locale change.
const shortcuts = computed(() => [
  { keys: ["Ctrl", "B"], action: t("shortcuts.actions.bold") },
  { keys: ["Ctrl", "I"], action: t("shortcuts.actions.italic") },
  { keys: ["Ctrl", "U"], action: t("shortcuts.actions.underline") },
  { keys: ["Ctrl", "⇧", "S"], action: t("shortcuts.actions.strikethrough") },
  { keys: ["Ctrl", "↑"], action: t("shortcuts.actions.subscript") },
  { keys: ["Ctrl", "↓"], action: t("shortcuts.actions.superscript") },
  { keys: ["Ctrl", "Z"], action: t("shortcuts.actions.undo") },
  { keys: ["Ctrl", "Y"], action: t("shortcuts.actions.redo") },
  { keys: ["Ctrl", "⇧", "U"], action: t("shortcuts.actions.ul") },
  { keys: ["Ctrl", "O"], action: t("shortcuts.actions.ol") },
  { keys: ["Ctrl", "L"], action: t("shortcuts.actions.link") },
])

function handleHide() {
  hide()
  toast.info(t("shortcuts.hide.message.title"), {
    description: t("shortcuts.hide.message.description"),
    action: {
      label: t("shortcuts.hide.undo"),
      onClick: show,
    },
  })
}
</script>

<template>
  <div v-if="!isHidden" class="hidden md:flex fixed bottom-24 right-0 z-10">
    <Popover>
      <PopoverTrigger as-child>
        <Button
          variant="default"
          class="rounded-l-full rounded-r-none h-10 w-8 flex items-center justify-center"
        >
          <CircleHelp class="size-5 pointer-events-none" />
        </Button>
      </PopoverTrigger>
      <PopoverContent side="left" align="end" class="w-64 p-3">
        <div
          class="flex justify-between items-center text-xs uppercase tracking-widest mb-3"
        >
          <p class="text-muted-foreground">
            {{ t("shortcuts.title") }}
          </p>
          <p
            class="text-muted-foreground cursor-pointer hover:underline"
            @click="handleHide"
          >
            {{ t("shortcuts.hide.action") }}
          </p>
        </div>
        <div class="flex flex-col gap-1.5">
          <div
            v-for="shortcut in shortcuts"
            :key="shortcut.action"
            class="flex items-center justify-between"
          >
            <span class="text-xs text-muted-foreground">{{
              shortcut.action
            }}</span>
            <KbdGroup>
              <Kbd v-for="key in shortcut.keys" :key="key">{{ key }}</Kbd>
            </KbdGroup>
          </div>
        </div>
      </PopoverContent>
    </Popover>
  </div>
</template>
