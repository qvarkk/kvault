<script setup lang="ts">
import { nextTick, ref, watch } from "vue"
import { useRouter } from "vue-router"
import { useI18n } from "vue-i18n"
import { watchDebounced } from "@vueuse/core"
import {
  ArrowRight,
  Clock,
  ExternalLink,
  FileText,
  NotebookPen,
  SearchIcon,
  X,
} from "lucide-vue-next"
import { Dialog, DialogContent } from "@/components/ui/dialog"
import {
  InputGroup,
  InputGroupAddon,
  InputGroupInput,
} from "@/components/ui/input-group"
import { Spinner } from "@/components/ui/spinner"
import { useSearchHistory } from "@/composables/useSearchHistory"
import { notesService } from "@/services/notes"
import { filesService } from "@/services/files"
import type { File, Note } from "@/types"

const PREVIEW_SIZE = 6

const open = defineModel<boolean>("open", { default: false })

const { t } = useI18n()
const router = useRouter()
const { history, add: addHistory, remove: removeHistory } = useSearchHistory()

const query = ref("")
const loading = ref(false)
const noteResults = ref<Note[]>([])
const fileResults = ref<File[]>([])
const inputWrap = ref<HTMLElement | null>(null)

const hasQuery = () => query.value.trim().length > 0
const hasResults = () =>
  noteResults.value.length > 0 || fileResults.value.length > 0

async function search() {
  if (!hasQuery()) {
    noteResults.value = []
    fileResults.value = []
    loading.value = false
    return
  }
  loading.value = true
  const params = { q: query.value, page_size: PREVIEW_SIZE }
  try {
    const [notes, files] = await Promise.all([
      notesService.list(params),
      filesService.list(params),
    ])
    noteResults.value = notes.data.data
    fileResults.value = files.data.data
  } catch {
    noteResults.value = []
    fileResults.value = []
  } finally {
    loading.value = false
  }
}

watchDebounced(query, search, { debounce: 300 })

watch(open, (isOpen) => {
  if (isOpen) {
    query.value = ""
    noteResults.value = []
    fileResults.value = []
    nextTick(() => inputWrap.value?.querySelector("input")?.focus())
  }
})

function goToNote(id: string) {
  addHistory(query.value)
  open.value = false
  router.push({ name: "note", params: { id } })
}

function goToFile(id: string) {
  addHistory(query.value)
  open.value = false
  router.push({ name: "file", params: { id } })
}

function seeAll() {
  const q = query.value.trim()
  if (!q) return
  addHistory(q)
  open.value = false
  router.push({ name: "search", query: { q } })
}

function applyHistory(entry: string) {
  query.value = entry
}
</script>

<template>
  <Dialog v-model:open="open">
    <DialogContent
      class="w-[calc(100vw-1rem)] max-w-xl gap-0 overflow-hidden p-0"
      aria-describedby=""
    >
      <div ref="inputWrap" class="border-b p-2">
        <InputGroup>
          <InputGroupInput
            v-model="query"
            :placeholder="t('search.placeholder')"
            @keydown.enter="seeAll"
          />
          <InputGroupAddon>
            <SearchIcon />
          </InputGroupAddon>
          <InputGroupAddon v-if="loading" align="inline-end">
            <Spinner />
          </InputGroupAddon>
        </InputGroup>
      </div>

      <div class="max-h-[60vh] overflow-y-auto p-2">
        <!-- History (empty query) -->
        <template v-if="!hasQuery()">
          <p
            v-if="history.length === 0"
            class="py-8 text-center text-sm text-muted-foreground"
          >
            {{ t("search.empty") }}
          </p>
          <template v-else>
            <p class="px-2 py-1.5 text-xs font-medium text-muted-foreground">
              {{ t("search.history") }}
            </p>
            <div
              v-for="entry in history"
              :key="entry"
              class="group flex items-center gap-2 rounded-md px-2 py-1.5 text-sm hover:bg-accent cursor-pointer"
              @click="applyHistory(entry)"
            >
              <Clock class="w-4 h-4 shrink-0 text-muted-foreground" />
              <span class="flex-1 truncate">{{ entry }}</span>
              <button
                class="shrink-0 opacity-0 group-hover:opacity-100 text-muted-foreground hover:text-foreground"
                @click.stop="removeHistory(entry)"
              >
                <X class="w-4 h-4" />
              </button>
            </div>
          </template>
        </template>

        <!-- Results -->
        <template v-else>
          <p
            v-if="!loading && !hasResults()"
            class="py-8 text-center text-sm text-muted-foreground"
          >
            {{ t("search.noResults") }}
          </p>

          <template v-if="noteResults.length > 0">
            <p class="px-2 py-1.5 text-xs font-medium text-muted-foreground">
              {{ t("search.notes") }}
            </p>
            <div
              v-for="note in noteResults"
              :key="note.id"
              class="flex items-center gap-2 rounded-md px-2 py-1.5 text-sm hover:bg-accent cursor-pointer"
              @click="goToNote(note.id)"
            >
              <component
                :is="note.type === 'url' ? ExternalLink : NotebookPen"
                class="w-4 h-4 shrink-0 text-muted-foreground"
              />
              <span class="flex-1 truncate">{{ note.title }}</span>
            </div>
          </template>

          <template v-if="fileResults.length > 0">
            <p class="px-2 py-1.5 text-xs font-medium text-muted-foreground">
              {{ t("search.files") }}
            </p>
            <div
              v-for="file in fileResults"
              :key="file.id"
              class="flex items-center gap-2 rounded-md px-2 py-1.5 text-sm hover:bg-accent cursor-pointer"
              @click="goToFile(file.id)"
            >
              <FileText class="w-4 h-4 shrink-0 text-muted-foreground" />
              <span class="flex-1 truncate">{{ file.originalName }}</span>
            </div>
          </template>

          <button
            v-if="hasResults()"
            class="mt-1 flex w-full items-center gap-2 rounded-md px-2 py-1.5 text-sm font-medium text-primary hover:bg-accent"
            @click="seeAll"
          >
            <ArrowRight class="w-4 h-4 shrink-0" />
            {{ t("search.seeAll") }}
          </button>
        </template>
      </div>
    </DialogContent>
  </Dialog>
</template>
