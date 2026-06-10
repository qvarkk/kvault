<script setup lang="ts">
import { ref, computed } from "vue"
import { useI18n } from "vue-i18n"
import { useRouter } from "vue-router"
import { FileText, Download, Trash2, Ellipsis, RotateCcw, ExternalLink, ArrowRight, Info } from "lucide-vue-next"
import {
  Card,
  CardHeader,
  CardTitle,
  CardContent,
  CardFooter,
} from "@/components/ui/card"
import { Button } from "@/components/ui/button"
import FileInfoModal from "@/components/files/FileInfoModal.vue"
import {
  DropdownMenu,
  DropdownMenuTrigger,
  DropdownMenuContent,
  DropdownMenuItem,
  DropdownMenuSeparator,
} from "@/components/ui/dropdown-menu"
import {
  AlertDialog,
  AlertDialogAction,
  AlertDialogCancel,
  AlertDialogContent,
  AlertDialogDescription,
  AlertDialogFooter,
  AlertDialogHeader,
  AlertDialogTitle,
} from "@/components/ui/alert-dialog"
import type { File } from "@/types"
import { timeAgo } from "@/services/format"

const { t } = useI18n()
const router = useRouter()

const props = defineProps<{
  file: File
  trash?: boolean
}>()

const emit = defineEmits<{
  download: [id: string]
  delete: [id: string]
  restore: [id: string]
}>()

const deleteOpen = ref(false)
const infoOpen = ref(false)

const clickable = computed(() => !props.trash)

function openFile() {
  router.push({ name: "file", params: { id: props.file.id } })
}

function openFileInNewTab() {
  window.open(`/files/${props.file.id}`, "_blank")
}

function formatSize(bytes: number): string {
  if (bytes < 1024) return `${bytes} B`
  if (bytes < 1024 * 1024) return `${(bytes / 1024).toFixed(1)} KB`
  return `${(bytes / (1024 * 1024)).toFixed(1)} MB`
}
</script>

<template>
  <Card
    class="flex flex-col gap-2 p-4 transition-shadow hover:shadow-md dark:transition-colors dark:hover:bg-accent"
    :class="clickable ? 'cursor-pointer' : 'cursor-default'"
    @click.prevent="clickable && openFile()"
    @mousedown.middle.prevent="clickable && openFileInNewTab()"
    @auxclick.middle="clickable && openFileInNewTab()"
  >
    <CardHeader class="p-0 flex flex-row items-start justify-between pointer-events-none">
      <div class="flex items-center gap-2 min-w-0 flex-1">
        <FileText
          class="w-4 h-4 shrink-0 text-muted-foreground pointer-events-none"
        />
        <CardTitle class="text-sm truncate">{{ file.originalName }}</CardTitle>
      </div>
      <DropdownMenu>
        <DropdownMenuTrigger as-child @click.stop>
          <Button variant="ghost" size="icon" class="w-6 h-6 shrink-0 -mt-1 -mr-1 pointer-events-auto">
            <Ellipsis class="w-4 h-4 pointer-events-none" />
          </Button>
        </DropdownMenuTrigger>
        <DropdownMenuContent @click.stop>
          <template v-if="trash">
            <DropdownMenuItem @click="emit('restore', file.id)">
              <RotateCcw class="w-4 h-4 mr-2 pointer-events-none" />
              {{ t("files.trash.restore") }}
            </DropdownMenuItem>
            <DropdownMenuSeparator />
            <DropdownMenuItem @click="deleteOpen = true" class="text-destructive">
              <Trash2 class="w-4 h-4 mr-2 pointer-events-none" />
              {{ t("files.trash.deletePermanently") }}
            </DropdownMenuItem>
          </template>
          <template v-else>
            <DropdownMenuItem @click="openFileInNewTab" :disabled="!clickable">
              <ExternalLink class="w-4 h-4 mr-2 pointer-events-none" />
              {{ t("common.openInNewTab") }}
            </DropdownMenuItem>
            <DropdownMenuItem @click="openFile" :disabled="!clickable">
              <ArrowRight class="w-4 h-4 mr-2 pointer-events-none" />
              {{ t("common.open") }}
            </DropdownMenuItem>
            <DropdownMenuItem @click="infoOpen = true">
              <Info class="w-4 h-4 mr-2 pointer-events-none" />
              {{ t("common.additionalInfo") }}
            </DropdownMenuItem>
            <DropdownMenuSeparator />
            <DropdownMenuItem
              @click="emit('download', file.id)"
            >
              <Download class="w-4 h-4 mr-2 pointer-events-none" />
              {{ t("files.download") }}
            </DropdownMenuItem>
            <DropdownMenuSeparator />
            <DropdownMenuItem @click="deleteOpen = true" class="text-destructive">
              <Trash2 class="w-4 h-4 mr-2 pointer-events-none" />
              {{ t("common.delete") }}
            </DropdownMenuItem>
          </template>
        </DropdownMenuContent>
      </DropdownMenu>
    </CardHeader>

    <CardContent class="p-0 flex flex-col gap-1 pointer-events-none">
      <span class="text-xs text-muted-foreground">
        {{ formatSize(file.size) }}
      </span>
    </CardContent>

    <CardFooter class="p-0 flex items-center justify-between pointer-events-none">
      <span class="text-xs text-muted-foreground">{{
        timeAgo(file.createdAt)
      }}</span>
    </CardFooter>
  </Card>

  <AlertDialog v-model:open="deleteOpen">
    <AlertDialogContent>
      <AlertDialogHeader>
        <AlertDialogTitle>
          {{ trash ? t("files.trash.deleteConfirm.title") : t("files.delete.title") }}
        </AlertDialogTitle>
        <AlertDialogDescription>
          {{ trash ? t("files.trash.deleteConfirm.description") : t("files.delete.description") }}
        </AlertDialogDescription>
      </AlertDialogHeader>
      <AlertDialogFooter>
        <AlertDialogCancel>{{ t("common.cancel") }}</AlertDialogCancel>
        <AlertDialogAction
          @click="emit('delete', file.id)"
          class="bg-destructive text-destructive-foreground hover:bg-destructive/90"
        >
          {{ trash ? t("files.trash.deleteConfirm.action") : t("common.delete") }}
        </AlertDialogAction>
      </AlertDialogFooter>
    </AlertDialogContent>
  </AlertDialog>

  <FileInfoModal v-model:open="infoOpen" :file="file" />
</template>
