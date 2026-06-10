<script setup lang="ts">
import { computed, onMounted, ref, watch } from "vue"
import { useRoute, useRouter } from "vue-router"
import { useI18n } from "vue-i18n"
import { useIntersectionObserver, watchDebounced } from "@vueuse/core"
import { useHead } from "@vueuse/head"
import { Files, NotebookPen, RefreshCw, SearchIcon } from "lucide-vue-next"
import AppLayout from "@/components/layout/AppLayout.vue"
import NoteCard from "@/components/notes/NoteCard.vue"
import FileCard from "@/components/files/FileCard.vue"
import NoteSkeleton from "@/components/notes/NoteSkeleton.vue"
import TagFilter from "@/components/notes/TagFilter.vue"
import SortControl from "@/components/common/SortControl.vue"
import { Button } from "@/components/ui/button"
import { Spinner } from "@/components/ui/spinner"
import { Card } from "@/components/ui/card"
import { Skeleton } from "@/components/ui/skeleton"
import {
  InputGroup,
  InputGroupAddon,
  InputGroupInput,
} from "@/components/ui/input-group"
import { useInfiniteEntities } from "@/composables/useInfiniteEntities"
import { useSearchHistory } from "@/composables/useSearchHistory"
import { notesService } from "@/services/notes"
import { filesService } from "@/services/files"
import { toast } from "vue-sonner"
import { toastApiError } from "@/services/apiError"
import type { File, Note } from "@/types"

const PAGE_SIZE = 20

const { t } = useI18n()
const route = useRoute()
const router = useRouter()
const { add: addHistory } = useSearchHistory()

useHead({ title: t("head.search") })

const query = ref(typeof route.query.q === "string" ? route.query.q : "")

// Notes section state
const noteSortBy = ref("updated_at")
const noteOrderBy = ref("DESC")
const tagIds = ref<string[]>([])
const notesSentinel = ref<HTMLElement | null>(null)

// Files section state
const fileSortBy = ref("created_at")
const fileOrderBy = ref("DESC")
const filesSentinel = ref<HTMLElement | null>(null)

const notes = useInfiniteEntities<Note>(notesService)
const files = useInfiniteEntities<File>(filesService)

const noteSortOptions = computed(() => [
  { value: "updated_at", label: t("notes.sort.updatedAt") },
  { value: "created_at", label: t("notes.sort.createdAt") },
  { value: "title", label: t("notes.sort.title") },
])

const fileSortOptions = computed(() => [
  { value: "created_at", label: t("files.sort.createdAt") },
  { value: "original_name", label: t("files.sort.name") },
  { value: "size", label: t("files.sort.size") },
])

const noteParams = computed(() => ({
  q: query.value,
  sort_by: noteSortBy.value,
  order_by: noteOrderBy.value,
  tag_ids: tagIds.value,
  page_size: PAGE_SIZE,
}))

const fileParams = computed(() => ({
  q: query.value,
  sort_by: fileSortBy.value,
  order_by: fileOrderBy.value,
  page_size: PAGE_SIZE,
}))

function loadNotes() {
  notes.fetchEntities(noteParams.value)
}

function loadFiles() {
  files.fetchEntities(fileParams.value)
}

function loadAll() {
  loadNotes()
  loadFiles()
}

function syncUrl() {
  router.replace({ query: query.value ? { q: query.value } : {} })
}

function handleNoteRemoved(id: string) {
  notes.entities.value = notes.entities.value.filter((n) => n.id !== id)
}

async function handleFileDownload(id: string) {
  try {
    const response = await filesService.getDownloadUrl(id)
    window.open(response.data.url, "_blank")
  } catch (e) {
    toastApiError(e)
  }
}

async function handleFileDelete(id: string) {
  try {
    await filesService.remove(id)
    files.entities.value = files.entities.value.filter((f) => f.id !== id)
    toast.success(t("files.delete.deleted"))
  } catch (e) {
    toastApiError(e)
  }
}

useIntersectionObserver(notesSentinel, ([entry]) => {
  if (entry?.isIntersecting) notes.loadMore(noteParams.value)
})
useIntersectionObserver(filesSentinel, ([entry]) => {
  if (entry?.isIntersecting) files.loadMore(fileParams.value)
})

watchDebounced(
  query,
  () => {
    addHistory(query.value)
    syncUrl()
    loadAll()
  },
  { debounce: 400 },
)
watch([noteSortBy, noteOrderBy, tagIds], loadNotes)
watch([fileSortBy, fileOrderBy], loadFiles)

onMounted(() => {
  if (query.value) addHistory(query.value)
  loadAll()
})
</script>

<template>
  <AppLayout>
    <div class="flex flex-col gap-6 container">
      <InputGroup>
        <InputGroupInput
          v-model="query"
          autofocus
          :placeholder="t('search.placeholder')"
        />
        <InputGroupAddon>
          <SearchIcon />
        </InputGroupAddon>
      </InputGroup>

      <!-- Notes -->
      <section class="flex flex-col gap-3">
        <div class="flex items-center gap-2">
          <NotebookPen class="w-4 h-4 text-muted-foreground" />
          <h2 class="text-sm font-medium">{{ t("search.notes") }}</h2>
          <Spinner v-if="notes.loading.value" class="ml-1" />
        </div>

        <div class="flex items-center gap-2 md:gap-4">
          <SortControl
            v-model:sort-by="noteSortBy"
            v-model:order-by="noteOrderBy"
            :options="noteSortOptions"
          />
          <div class="flex-1">
            <TagFilter v-model:tag-ids="tagIds" />
          </div>
        </div>

        <div
          v-if="notes.loading.value"
          class="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-4"
        >
          <NoteSkeleton v-for="n in 6" :key="n" />
        </div>
        <div
          v-else-if="notes.error.value"
          class="flex flex-col items-center gap-3 py-10 text-muted-foreground"
        >
          <p class="text-destructive">{{ notes.error.value.detail }}</p>
          <Button variant="outline" @click="loadNotes">
            <RefreshCw class="w-4 h-4 mr-2" />
            {{ t("common.retry") }}
          </Button>
        </div>
        <p
          v-else-if="notes.entities.value.length === 0"
          class="py-6 text-sm text-muted-foreground"
        >
          {{ t("search.noResults") }}
        </p>
        <div
          v-else
          class="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-4"
        >
          <NoteCard
            v-for="note in notes.entities.value"
            :key="note.id"
            :note="note"
            :include-menu="true"
            @deleted="handleNoteRemoved"
          />
        </div>
        <div ref="notesSentinel" class="h-1" />
        <div v-if="notes.loadingMore.value" class="flex justify-center py-2">
          <Spinner />
        </div>
      </section>

      <!-- Files -->
      <section class="flex flex-col gap-3">
        <div class="flex items-center gap-2">
          <Files class="w-4 h-4 text-muted-foreground" />
          <h2 class="text-sm font-medium">{{ t("search.files") }}</h2>
          <Spinner v-if="files.loading.value" class="ml-1" />
        </div>

        <div class="flex items-center gap-2">
          <SortControl
            v-model:sort-by="fileSortBy"
            v-model:order-by="fileOrderBy"
            :options="fileSortOptions"
          />
        </div>

        <div
          v-if="files.loading.value"
          class="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-4"
        >
          <Card v-for="n in 6" :key="n" class="p-4 flex flex-col gap-3">
            <Skeleton class="h-5 w-3/4" />
            <Skeleton class="h-3 w-1/2" />
          </Card>
        </div>
        <div
          v-else-if="files.error.value"
          class="flex flex-col items-center gap-3 py-10 text-muted-foreground"
        >
          <p class="text-destructive">{{ files.error.value.detail }}</p>
          <Button variant="outline" @click="loadFiles">
            <RefreshCw class="w-4 h-4 mr-2" />
            {{ t("common.retry") }}
          </Button>
        </div>
        <p
          v-else-if="files.entities.value.length === 0"
          class="py-6 text-sm text-muted-foreground"
        >
          {{ t("search.noResults") }}
        </p>
        <div
          v-else
          class="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-4"
        >
          <FileCard
            v-for="file in files.entities.value"
            :key="file.id"
            :file="file"
            @download="handleFileDownload"
            @delete="handleFileDelete"
          />
        </div>
        <div ref="filesSentinel" class="h-1" />
        <div v-if="files.loadingMore.value" class="flex justify-center py-2">
          <Spinner />
        </div>
      </section>
    </div>
  </AppLayout>
</template>
