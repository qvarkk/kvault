<script setup lang="ts">
import { computed, ref, watch } from "vue"
import { useI18n } from "vue-i18n"
import {
  Dialog,
  DialogScrollContent,
  DialogHeader,
  DialogTitle,
  DialogDescription,
} from "@/components/ui/dialog"
import { FileText } from "lucide-vue-next"

const props = defineProps<{
  text?: string
}>()

const open = defineModel<boolean>("open", { default: false })

const { t } = useI18n()

// Large extracted text crashes the page if rendered at once. Split into
// fixed-size chunks and reveal them incrementally as the user scrolls.
const CHUNK_SIZE = 5000
const CHUNK_STEP = 5

const chunks = computed(() => {
  const value = props.text ?? ""
  const result: string[] = []
  for (let i = 0; i < value.length; i += CHUNK_SIZE) {
    result.push(value.slice(i, i + CHUNK_SIZE))
  }
  return result
})

const visibleCount = ref(CHUNK_STEP)
const visibleChunks = computed(() => chunks.value.slice(0, visibleCount.value))
const hasMore = computed(() => visibleCount.value < chunks.value.length)

watch(open, (isOpen) => {
  if (isOpen) visibleCount.value = CHUNK_STEP
})

function onScroll(e: Event) {
  const el = e.target as HTMLElement
  if (!hasMore.value) return
  // Reveal next batch when scrolled near the bottom.
  if (el.scrollHeight - el.scrollTop - el.clientHeight < 600) {
    visibleCount.value += CHUNK_STEP
  }
}
</script>

<template>
  <Dialog v-model:open="open">
    <DialogScrollContent class="w-[calc(100vw-1rem)] max-w-2xl">
      <DialogHeader>
        <DialogTitle>{{ t("notes.info.extractedText.title") }}</DialogTitle>
        <DialogDescription>
          {{ t("notes.info.extractedText.description") }}
        </DialogDescription>
      </DialogHeader>

      <div
        v-if="text && text.trim()"
        class="text-sm whitespace-pre-wrap wrap-break-word leading-relaxed text-muted-foreground max-h-[60vh] overflow-y-auto"
        @scroll="onScroll"
      >
        <span v-for="(chunk, i) in visibleChunks" :key="i">{{ chunk }}</span>
      </div>
      <div
        v-else
        class="flex flex-col items-center justify-center gap-2 py-12 text-muted-foreground"
      >
        <FileText class="w-8 h-8" />
        <p class="text-sm">{{ t("notes.info.extractedText.empty") }}</p>
      </div>
    </DialogScrollContent>
  </Dialog>
</template>
