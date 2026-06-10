<script setup lang="ts">
import { ref } from "vue"
import { useI18n } from "vue-i18n"
import type { File } from "@/types"
import {
  Dialog,
  DialogContent,
  DialogHeader,
  DialogTitle,
  DialogDescription,
  DialogFooter,
} from "@/components/ui/dialog"
import { Badge } from "@/components/ui/badge"
import { Button } from "@/components/ui/button"
import { FileText } from "lucide-vue-next"
import ExtractedTextModal from "@/components/notes/ExtractedTextModal.vue"
import { formatBytes, formatDateTime, timeAgo } from "@/services/format"

const props = defineProps<{
  file: File
}>()

const open = defineModel<boolean>("open", { default: false })

const { t } = useI18n()

const extractedTextOpen = ref(false)

const statusVariant: Record<string, "default" | "secondary" | "destructive"> = {
  ready: "default",
  uploading: "secondary",
  processing: "secondary",
  error: "destructive",
}
</script>

<template>
  <Dialog v-model:open="open">
    <DialogContent class="w-[calc(100vw-1rem)] max-w-md">
      <DialogHeader class="min-w-0">
        <DialogTitle class="min-w-0 break-all pr-6">
          {{ file.originalName }}
        </DialogTitle>
        <DialogDescription class="hidden">
          {{ t("files.info.title") }}
        </DialogDescription>
      </DialogHeader>

      <div class="flex flex-col gap-3">
        <div
          class="flex items-center gap-2 rounded-md border p-3"
          :class="
            file.status === 'error'
              ? 'border-destructive/40 bg-destructive/5'
              : 'bg-muted/40'
          "
        >
          <Badge :variant="statusVariant[file.status] ?? 'secondary'">
            {{ t(`files.status.${file.status}`) }}
          </Badge>
          <p class="text-xs text-muted-foreground">
            {{
              file.status === "ready" && !file.textContent
                ? t("files.status.description.readyNoText")
                : t(`files.status.description.${file.status}`)
            }}
          </p>
        </div>

        <dl class="flex flex-col divide-y text-sm">
          <div class="flex items-center justify-between py-2">
            <dt class="text-muted-foreground">{{ t("files.info.size") }}</dt>
            <dd>{{ formatBytes(file.size) }}</dd>
          </div>
          <div class="flex items-center justify-between gap-4 py-2">
            <dt class="text-muted-foreground shrink-0">
              {{ t("files.info.mimeType") }}
            </dt>
            <dd class="font-mono text-xs truncate">{{ file.mimeType }}</dd>
          </div>
          <div class="flex items-center justify-between gap-4 py-2">
            <dt class="text-muted-foreground shrink-0">
              {{ t("files.info.uploaded") }}
            </dt>
            <dd class="text-right">{{ formatDateTime(file.createdAt) }}</dd>
          </div>
          <div class="flex items-center justify-between gap-4 py-2">
            <dt class="text-muted-foreground shrink-0">
              {{ t("files.info.updated") }}
            </dt>
            <dd class="text-right" :title="formatDateTime(file.updatedAt)">
              {{ timeAgo(file.updatedAt) }}
            </dd>
          </div>
        </dl>
      </div>

      <DialogFooter v-if="file.textContent">
        <Button variant="outline" class="w-full" @click="extractedTextOpen = true">
          <FileText class="w-4 h-4 mr-2 pointer-events-none" />
          {{ t("files.info.viewExtractedText") }}
        </Button>
      </DialogFooter>
    </DialogContent>
  </Dialog>

  <ExtractedTextModal v-model:open="extractedTextOpen" :text="file.textContent" />
</template>
