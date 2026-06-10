<script setup lang="ts">
import AppLayout from "@/components/layout/AppLayout.vue"
import CreateNoteModal from "@/components/notes/CreateNoteModal.vue"
import NoteCard from "@/components/notes/NoteCard.vue"
import NoteSkeleton from "@/components/notes/NoteSkeleton.vue"
import NotesToolbar from "@/components/notes/NotesToolbar.vue"
import { Button } from "@/components/ui/button"
import { Spinner } from "@/components/ui/spinner"
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
import { useInfiniteEntities } from "@/composables/useInfiniteEntities"
import { notesService } from "@/services/notes"
import type { ApiError, Note } from "@/types"
import { useIntersectionObserver, watchDebounced } from "@vueuse/core"
import { useHead } from "@vueuse/head"
import { NotebookPen, Plus, RefreshCw, Trash2 } from "lucide-vue-next"
import { computed, onMounted, ref, watch } from "vue"
import { useI18n } from "vue-i18n"
import { toast } from "vue-sonner"
import { toastApiError } from "@/services/apiError"

const PAGE_SIZE = 20

const { t } = useI18n()
useHead({ title: t("head.notes") })

const query = ref("")
const sortBy = ref("updated_at")
const orderBy = ref("DESC")
const tagIds = ref<string[]>([])
const createModalOpen = ref(false)
const trashMode = ref(false)
const emptyTrashDialogOpen = ref(false)

const serviceProxy = {
  list: (params?: Record<string, unknown>) =>
    trashMode.value
      ? notesService.listDeleted(params)
      : notesService.list(params),
}

const { entities, loading, loadingMore, error, fetchEntities, loadMore } =
  useInfiniteEntities<Note>(serviceProxy)

const sentinel = ref<HTMLElement | null>(null)

const params = computed(() => ({
  q: query.value,
  sort_by: sortBy.value,
  order_by: orderBy.value,
  tag_ids: tagIds.value,
  page_size: PAGE_SIZE,
}))

async function loadNotes() {
  await fetchEntities(params.value)
}

function openCreateModal() {
  createModalOpen.value = true
}

function handleNoteDeleted(id: string) {
  entities.value = entities.value.filter((n) => n.id !== id)
}

function handleNoteRestored(id: string) {
  entities.value = entities.value.filter((n) => n.id !== id)
}

async function handleEmptyTrash() {
  try {
    await notesService.emptyTrash()
    toast.success(t("notes.trash.emptied"))
    emptyTrashDialogOpen.value = false
    await loadNotes()
  } catch (e) {
    toastApiError(e)
  }
}

useIntersectionObserver(sentinel, ([entry]) => {
  if (entry?.isIntersecting) loadMore(params.value)
})

watchDebounced(query, loadNotes, { debounce: 400 })
watch([sortBy, orderBy, tagIds], loadNotes)
watch(trashMode, loadNotes)
onMounted(loadNotes)
</script>

<template>
  <AppLayout>
    <div class="flex flex-col gap-4 container">
      <div class="flex items-center justify-between gap-2">
        <div class="flex gap-1">
          <Button
            :variant="trashMode ? 'outline' : 'secondary'"
            size="sm"
            @click="trashMode = false"
          >
            <NotebookPen class="w-3.5 h-3.5 mr-1.5 pointer-events-none" />
            {{ t("nav.notes") }}
          </Button>
          <Button
            :variant="trashMode ? 'secondary' : 'outline'"
            size="sm"
            @click="trashMode = true"
          >
            <Trash2 class="w-3.5 h-3.5 mr-1.5 pointer-events-none" />
            {{ t("notes.trash.toggle") }}
          </Button>
        </div>
        <Button
          v-if="trashMode && entities.length > 0 && !loading"
          variant="destructive"
          size="sm"
          @click="emptyTrashDialogOpen = true"
        >
          {{ t("notes.trash.emptyAction") }}
        </Button>
      </div>

      <NotesToolbar
        v-model:query="query"
        v-model:sortBy="sortBy"
        v-model:orderBy="orderBy"
        v-model:tagIds="tagIds"
        v-model:loading="loading"
      />

      <div
        v-if="loading"
        class="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-4"
      >
        <NoteSkeleton v-for="n in 20" :key="n" />
      </div>

      <div
        v-else-if="error"
        class="flex flex-col items-center justify-center gap-4 pt-24 pb-8 text-muted-foreground"
      >
        <p class="text-destructive">{{ error.detail }}</p>
        <Button variant="outline" @click="loadNotes">
          <RefreshCw class="w-4 h-4 mr-2" />
          {{ t("common.retry") }}
        </Button>
      </div>

      <div v-else>
        <div
          v-if="entities.length === 0"
          class="flex flex-col items-center justify-center gap-4 pt-24 pb-8 text-muted-foreground"
        >
          <NotebookPen class="w-12 h-12" />
          <p>{{ trashMode ? t("notes.trash.empty") : t("notes.empty") }}</p>
          <Button v-if="!trashMode" @click="openCreateModal">
            {{ t("notes.create.first") }}
          </Button>
        </div>

        <div
          v-else
          class="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-4"
        >
          <NoteCard
            v-for="note in entities"
            :key="note.id"
            :note="note"
            :include-menu="true"
            :trash="trashMode"
            @deleted="handleNoteDeleted"
            @restored="handleNoteRestored"
          />
        </div>

        <div ref="sentinel" class="h-1" />

        <div v-if="loadingMore" class="flex justify-center py-4">
          <Spinner />
        </div>
      </div>
    </div>

    <Button
      v-if="!trashMode"
      class="fixed bottom-6 right-6 rounded-full w-12 h-12 shadow-lg"
      size="icon"
      @click="openCreateModal"
    >
      <Plus class="w-5 h-5 pointer-events-none" />
    </Button>
  </AppLayout>

  <CreateNoteModal v-if="!trashMode" v-model:open="createModalOpen" />

  <AlertDialog v-model:open="emptyTrashDialogOpen">
    <AlertDialogContent>
      <AlertDialogHeader>
        <AlertDialogTitle>{{ t("notes.trash.emptyConfirm.title") }}</AlertDialogTitle>
        <AlertDialogDescription>
          {{ t("notes.trash.emptyConfirm.description") }}
        </AlertDialogDescription>
      </AlertDialogHeader>
      <AlertDialogFooter>
        <AlertDialogCancel>{{ t("common.cancel") }}</AlertDialogCancel>
        <AlertDialogAction
          @click="handleEmptyTrash"
          class="bg-destructive text-destructive-foreground hover:bg-destructive/90"
        >
          {{ t("notes.trash.emptyConfirm.action") }}
        </AlertDialogAction>
      </AlertDialogFooter>
    </AlertDialogContent>
  </AlertDialog>
</template>
