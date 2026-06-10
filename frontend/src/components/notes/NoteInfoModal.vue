<script setup lang="ts">
import { ref } from "vue"
import { useI18n } from "vue-i18n"
import type { Note } from "@/types"
import {
  Dialog,
  DialogContent,
  DialogHeader,
  DialogTitle,
  DialogDescription,
  DialogFooter,
} from "@/components/ui/dialog"
import { Button } from "@/components/ui/button"
import { Badge } from "@/components/ui/badge"
import { FileText } from "lucide-vue-next"
import { formatDateTime, timeAgo } from "@/services/format"
import ExtractedTextModal from "./ExtractedTextModal.vue"

const props = defineProps<{
  note: Note
}>()

const open = defineModel<boolean>("open", { default: false })

const { t } = useI18n()

const extractedTextOpen = ref(false)

const urlStatusVariant: Record<string, "default" | "secondary" | "destructive"> =
  {
    ready: "default",
    pending: "secondary",
    error: "destructive",
  }
</script>

<template>
  <Dialog v-model:open="open">
    <DialogContent class="w-[calc(100vw-1rem)] max-w-md">
      <DialogHeader>
        <DialogTitle>{{ t("notes.info.title") }}</DialogTitle>
        <DialogDescription class="hidden">
          {{ t("notes.info.title") }}
        </DialogDescription>
      </DialogHeader>

      <dl class="flex flex-col divide-y text-sm">
        <div class="flex items-center justify-between py-2">
          <dt class="text-muted-foreground">{{ t("notes.info.type") }}</dt>
          <dd>
            <Badge variant="outline">{{ t(`notes.info.types.${note.type}`) }}</Badge>
          </dd>
        </div>

        <div
          v-if="note.type === 'url'"
          class="flex items-center justify-between py-2"
        >
          <dt class="text-muted-foreground">{{ t("notes.info.urlStatus") }}</dt>
          <dd>
            <Badge
              v-if="note.urlStatus"
              :variant="urlStatusVariant[note.urlStatus] ?? 'secondary'"
            >
              {{ t(`notes.url.status.${note.urlStatus}`) }}
            </Badge>
            <span v-else class="text-muted-foreground">—</span>
          </dd>
        </div>

        <div class="flex items-center justify-between py-2">
          <dt class="text-muted-foreground">{{ t("notes.info.tags") }}</dt>
          <dd>{{ note.tags.length }}</dd>
        </div>

        <div class="flex items-center justify-between gap-4 py-2">
          <dt class="text-muted-foreground shrink-0">
            {{ t("notes.info.created") }}
          </dt>
          <dd class="text-right" :title="formatDateTime(note.createdAt)">
            {{ formatDateTime(note.createdAt) }}
          </dd>
        </div>

        <div class="flex items-center justify-between gap-4 py-2">
          <dt class="text-muted-foreground shrink-0">
            {{ t("notes.info.updated") }}
          </dt>
          <dd class="text-right" :title="formatDateTime(note.updatedAt)">
            {{ timeAgo(note.updatedAt) }}
          </dd>
        </div>
      </dl>

      <DialogFooter v-if="note.type === 'url'">
        <Button variant="outline" class="w-full" @click="extractedTextOpen = true">
          <FileText class="w-4 h-4 mr-2 pointer-events-none" />
          {{ t("notes.info.viewExtractedText") }}
        </Button>
      </DialogFooter>
    </DialogContent>
  </Dialog>

  <ExtractedTextModal
    v-model:open="extractedTextOpen"
    :text="note.extractedContent"
  />
</template>
