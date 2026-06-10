<script setup lang="ts">
import type { Note } from "@/types"
import {
  Card,
  CardContent,
  CardFooter,
  CardHeader,
  CardTitle,
} from "../ui/card"
import { Badge } from "../ui/badge"
import { timeAgo } from "@/services/format"
import { useRouter } from "vue-router"
import {
  DropdownMenu,
  DropdownMenuContent,
  DropdownMenuItem,
  DropdownMenuSeparator,
  DropdownMenuTrigger,
} from "../ui/dropdown-menu"
import {
  AlertDialog,
  AlertDialogAction,
  AlertDialogCancel,
  AlertDialogContent,
  AlertDialogDescription,
  AlertDialogFooter,
  AlertDialogHeader,
  AlertDialogTitle,
} from "../ui/alert-dialog"
import { Button } from "../ui/button"
import { ArrowRight, Ellipsis, ExternalLink, Info, Loader2, RefreshCw, RotateCcw, TriangleAlert, Trash2 } from "lucide-vue-next"
import { useI18n } from "vue-i18n"
import { ref } from "vue"
import DeleteNoteAlertModal from "./DeleteNoteAlertModal.vue"
import NoteInfoModal from "./NoteInfoModal.vue"
import { notesService } from "@/services/notes"
import { toast } from "vue-sonner"
import { toastApiError } from "@/services/apiError"

const { t } = useI18n()
const router = useRouter()

const props = defineProps<{
  note: Note
  includeMenu: boolean
  trash?: boolean
}>()

const emit = defineEmits<{
  deleted: [id: string]
  restored: [id: string]
}>()

const deleteDialogOpen = ref(false)
const permanentDeleteDialogOpen = ref(false)
const infoOpen = ref(false)

function handleDeleted(id: string) {
  emit("deleted", id)
}

function openInNewTab() {
  window.open(`/notes/${props.note.id}`, "_blank")
}

function redirectToNote() {
  router.push({ name: "note", params: { id: props.note.id } })
}

function openSourceUrl() {
  if (props.note.sourceUrl) window.open(props.note.sourceUrl, "_blank")
}

async function handleRefetch() {
  try {
    await notesService.refetch(props.note.id)
    toast.success(t("notes.url.refetchStarted"))
  } catch (e) {
    toastApiError(e)
  }
}

async function handleRestore() {
  try {
    await notesService.restore(props.note.id)
    emit("restored", props.note.id)
    toast.success(t("notes.trash.restored"))
  } catch (e) {
    toastApiError(e)
  }
}

async function handlePermanentDelete() {
  try {
    await notesService.permanentDelete(props.note.id)
    emit("deleted", props.note.id)
    toast.success(t("notes.trash.deletedPermanently"))
    permanentDeleteDialogOpen.value = false
  } catch (e) {
    toastApiError(e)
  }
}
</script>

<template>
  <Card
    class="group flex flex-col gap-3 p-4 h-full hover:shadow-md transition-shadow dark:hover:bg-accent dark:transition-colors relative overflow-hidden"
    :class="trash ? 'cursor-default' : 'cursor-pointer'"
    @click.prevent="!trash && redirectToNote()"
    @mousedown.middle.prevent="!trash && openInNewTab()"
    @auxclick.middle="!trash && openInNewTab()"
  >
    <div
      v-if="note.type === 'url' && note.urlMetadata?.image"
      class="absolute inset-0 bg-cover bg-center opacity-10 pointer-events-none"
      :style="{ backgroundImage: `url(${note.urlMetadata.image})` }"
    />

    <CardHeader
      class="p-0 flex flex-row items-start justify-between pointer-events-none"
    >
      <CardTitle class="text-base">{{ note.title }}</CardTitle>
      <DropdownMenu v-if="includeMenu">
        <DropdownMenuTrigger as-child @click.stop>
          <Button
            variant="ghost"
            size="icon"
            class="w-6 h-6 shrink-0 -mt-1 -mr-1 dark:hover:brightness-200 transition-all pointer-events-auto"
          >
            <Ellipsis class="w-4 h-4 pointer-events-none" />
          </Button>
        </DropdownMenuTrigger>
        <DropdownMenuContent @click.stop>
          <template v-if="trash">
            <DropdownMenuItem @click="handleRestore">
              <RotateCcw class="w-4 h-4 mr-2 pointer-events-none" />
              {{ t("notes.trash.restore") }}
            </DropdownMenuItem>
            <DropdownMenuSeparator />
            <DropdownMenuItem
              @click="permanentDeleteDialogOpen = true"
              class="text-destructive"
            >
              <Trash2 class="w-4 h-4 mr-2 pointer-events-none" />
              {{ t("notes.trash.deletePermanently") }}
            </DropdownMenuItem>
          </template>
          <template v-else>
            <DropdownMenuItem @click="openInNewTab">
              <ExternalLink class="w-4 h-4 mr-2 pointer-events-none" />
              {{ t("common.openInNewTab") }}
            </DropdownMenuItem>
            <DropdownMenuItem @click="redirectToNote">
              <ArrowRight class="w-4 h-4 mr-2 pointer-events-none" />
              {{ t("common.open") }}
            </DropdownMenuItem>
            <DropdownMenuItem @click="infoOpen = true">
              <Info class="w-4 h-4 mr-2 pointer-events-none" />
              {{ t("common.additionalInfo") }}
            </DropdownMenuItem>
            <template v-if="note.type === 'url'">
              <DropdownMenuSeparator />
              <DropdownMenuItem v-if="note.sourceUrl" @click="openSourceUrl">
                <ExternalLink class="w-4 h-4 mr-2 pointer-events-none" />
                {{ t("notes.url.openUrl") }}
              </DropdownMenuItem>
              <DropdownMenuItem @click="handleRefetch">
                <RefreshCw class="w-4 h-4 mr-2 pointer-events-none" />
                {{ t("notes.url.refetch") }}
              </DropdownMenuItem>
            </template>
            <DropdownMenuSeparator />
            <DropdownMenuItem @click="deleteDialogOpen = true" class="text-destructive">
              <Trash2 class="w-4 h-4 mr-2 pointer-events-none" />
              {{ t("common.delete") }}
            </DropdownMenuItem>
          </template>
        </DropdownMenuContent>
      </DropdownMenu>
    </CardHeader>

    <CardContent
      class="p-0 text-sm text-muted-foreground flex-1 pointer-events-none break-words"
    >
      <template v-if="note.type === 'url'">
        <div
          v-if="note.urlStatus && note.urlStatus !== 'ready'"
          class="flex items-center gap-1 mb-1 text-xs"
          :class="note.urlStatus === 'error' ? 'text-destructive' : 'text-muted-foreground'"
        >
          <Loader2
            v-if="note.urlStatus === 'pending'"
            class="w-3 h-3 animate-spin pointer-events-none"
          />
          <TriangleAlert v-else class="w-3 h-3 pointer-events-none" />
          <span>{{ t(`notes.url.status.${note.urlStatus}`) }}</span>
        </div>
        <p v-if="note.urlMetadata?.siteName" class="font-medium text-foreground">
          {{ note.urlMetadata.siteName }}
        </p>
        <p v-if="note.sourceUrl">{{ t("notes.url.source") }}: {{ note.sourceUrl }}</p>
        <p v-if="note.content">{{ note.content.slice(0, 50) + (note.content.length > 50 ? "..." : "") }}</p>
      </template>
      <template v-else>
        {{ note.content.slice(0, 100) + (note.content.length > 100 ? "..." : "") }}
      </template>
    </CardContent>

    <CardFooter
      class="p-0 flex items-center justify-between pointer-events-none"
    >
      <div class="flex gap-1 flex-wrap">
        <Badge
          v-for="tag in note.tags"
          :key="tag.id"
          variant="secondary"
          class="group-hover:dark:brightness-125 transition-all"
        >
          {{ tag.name }}
        </Badge>
      </div>
      <span class="text-xs text-muted-foreground">{{
        timeAgo(note.updatedAt)
      }}</span>
    </CardFooter>
  </Card>

  <DeleteNoteAlertModal
    v-if="!trash"
    v-model:open="deleteDialogOpen"
    :note-id="note.id"
    @deleted="handleDeleted"
  />

  <NoteInfoModal v-model:open="infoOpen" :note="note" />

  <AlertDialog v-if="trash" v-model:open="permanentDeleteDialogOpen">
    <AlertDialogContent @click.stop>
      <AlertDialogHeader>
        <AlertDialogTitle>{{ t("notes.trash.deleteConfirm.title") }}</AlertDialogTitle>
        <AlertDialogDescription>
          {{ t("notes.trash.deleteConfirm.description") }}
        </AlertDialogDescription>
      </AlertDialogHeader>
      <AlertDialogFooter>
        <AlertDialogCancel>{{ t("common.cancel") }}</AlertDialogCancel>
        <AlertDialogAction
          @click="handlePermanentDelete"
          class="bg-destructive text-destructive-foreground hover:bg-destructive/90"
        >
          {{ t("notes.trash.deleteConfirm.action") }}
        </AlertDialogAction>
      </AlertDialogFooter>
    </AlertDialogContent>
  </AlertDialog>
</template>
