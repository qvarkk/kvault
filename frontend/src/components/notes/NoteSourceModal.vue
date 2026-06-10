<script setup lang="ts">
import { computed } from "vue"
import { useI18n } from "vue-i18n"
import type { Note } from "@/types"
import {
  Dialog,
  DialogScrollContent,
  DialogHeader,
  DialogTitle,
  DialogDescription,
} from "@/components/ui/dialog"
import { Link2 } from "lucide-vue-next"

const props = defineProps<{
  note: Note
}>()

const open = defineModel<boolean>("open", { default: false })

const { t } = useI18n()

// Build a YAML-frontmatter-like view of the URL metadata.
const frontmatter = computed(() => {
  const m = props.note.urlMetadata
  const rows: [string, string | undefined][] = [
    ["source_url", props.note.sourceUrl],
    ["title", m?.title],
    ["site_name", m?.siteName],
    ["description", m?.description],
    ["image", m?.image],
    ["status", props.note.urlStatus],
  ]
  const body = rows
    .filter(([, v]) => v && v.trim())
    .map(([k, v]) => `${k}: ${v}`)
    .join("\n")
  return `---\n${body}\n---`
})
</script>

<template>
  <Dialog v-model:open="open">
    <DialogScrollContent class="w-[calc(100vw-1rem)] max-w-xl">
      <DialogHeader>
        <DialogTitle>{{ t("notes.info.source.title") }}</DialogTitle>
        <DialogDescription>
          {{ t("notes.info.source.description") }}
        </DialogDescription>
      </DialogHeader>

      <div class="flex flex-col gap-3 min-w-0">
        <img
          v-if="note.urlMetadata?.image"
          :src="note.urlMetadata.image"
          :alt="note.urlMetadata.title || note.title"
          class="w-full max-h-56 object-cover rounded-md border"
          @error="(e: Event) => ((e.target as HTMLElement).style.display = 'none')"
        />

        <pre
          class="text-xs font-mono whitespace-pre-wrap break-all bg-muted rounded-md p-3 max-w-full"
        >{{ frontmatter }}</pre>

        <a
          v-if="note.sourceUrl"
          :href="note.sourceUrl"
          target="_blank"
          rel="noopener noreferrer"
          class="flex items-center gap-2 text-sm text-primary hover:underline break-all"
        >
          <Link2 class="w-4 h-4 shrink-0 pointer-events-none" />
          {{ note.sourceUrl }}
        </a>
      </div>
    </DialogScrollContent>
  </Dialog>
</template>
