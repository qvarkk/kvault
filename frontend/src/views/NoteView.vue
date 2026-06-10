<script setup lang="ts">
import { ref, computed, onMounted, watch } from "vue"
import { useRouter } from "vue-router"
import { useI18n } from "vue-i18n"
import { useDark, useDebounceFn } from "@vueuse/core"
import { useNote } from "@/composables/useNote"
import { toast } from "vue-sonner"
import { toastApiError } from "@/services/apiError"
import type { Tag } from "@/types"
import AppLayout from "@/components/layout/AppLayout.vue"
import { Skeleton } from "@/components/ui/skeleton"
import { Button } from "@/components/ui/button"
import {
  ArrowLeft,
  Check,
  Ellipsis,
  ExternalLink,
  Info,
  Link as LinkIcon,
  Loader2,
  RefreshCw,
  Sparkles,
  Sticker,
  Trash2,
  Undo2,
  Wand2,
  X,
} from "lucide-vue-next"
import {
  DropdownMenu,
  DropdownMenuContent,
  DropdownMenuItem,
  DropdownMenuSeparator,
  DropdownMenuTrigger,
} from "@/components/ui/dropdown-menu"
import DeleteNoteAlertModal from "@/components/notes/DeleteNoteAlertModal.vue"
import NoteInfoModal from "@/components/notes/NoteInfoModal.vue"
import NoteSourceModal from "@/components/notes/NoteSourceModal.vue"
import { Input } from "@/components/ui/input"
import { Badge } from "@/components/ui/badge"
import "md-editor-v3/lib/style.css"
import NoteEditor from "@/components/notes/NoteEditor.vue"
import NoteTagSelector from "@/components/notes/NoteTagSelector.vue"
import {
  Popover,
  PopoverContent,
  PopoverTrigger,
} from "@/components/ui/popover"
import {
  NumberField,
  NumberFieldContent,
  NumberFieldDecrement,
  NumberFieldIncrement,
  NumberFieldInput,
} from "@/components/ui/number-field"
import { notesService } from "@/services/notes"
import EditorShortcutsHelper from "@/components/notes/EditorShortcutsHelper.vue"
import { useHead } from "@vueuse/head"

const props = defineProps<{
  id: string
}>()

const router = useRouter()
const { t } = useI18n()
const isDark = useDark()

const {
  note,
  loading,
  saving,
  error,
  fetchNote,
  updateNote,
  bindTag,
  unbindTag,
} = useNote()

const title = ref("")
const content = ref("")
const sourceUrl = ref("")
const changed = ref(false)
const deleteDialogOpen = ref(false)
const infoOpen = ref(false)
const sourceOpen = ref(false)
const autotagCount = ref(5)
const autotagLoading = ref(false)
const refetchLoading = ref(false)

async function handleTitleChange() {
  if (!note.value) return
  const cleanTitle = title.value.trim()
  if (!cleanTitle || cleanTitle === note.value.title) return
  try {
    await updateNote(note.value.id, { title: cleanTitle })
    toast.success(t("notes.update.titleChanged"))
  } catch (e) {
    toastApiError(e)
    title.value = note.value.title
  }
}

async function handleBindTag(tag: Tag) {
  if (!note.value) return
  try {
    await bindTag(note.value.id, tag)
  } catch (e) {
    toastApiError(e)
  }
}

async function handleUnbindTag(tagId: string) {
  if (!note.value) return
  try {
    await unbindTag(note.value.id, tagId)
  } catch (e) {
    toastApiError(e)
  }
}

async function handleAutotag() {
  if (!note.value) return
  autotagLoading.value = true
  try {
    const response = await notesService.autotag(
      note.value.id,
      autotagCount.value,
    )
    for (const tag of response.data.data) {
      await new Promise((resolve) => setTimeout(resolve, 250))
      if (!note.value.tags.find((t) => t.id === tag.id)) {
        note.value.tags.push(tag)
      }
    }
  } catch (e) {
    toastApiError(e)
  } finally {
    autotagLoading.value = false
  }
}

const debouncedUrlSave = useDebounceFn(async () => {
  if (!note.value) return
  try {
    await notesService.updateSourceUrl(note.value.id, sourceUrl.value.trim())
  } catch (e) {
    toastApiError(e)
  }
}, 1500)

async function handleRefetch() {
  if (!note.value) return
  refetchLoading.value = true
  try {
    await notesService.refetch(note.value.id)
    toast.success(t("notes.url.refetchStarted"))
  } catch (e) {
    toastApiError(e)
  } finally {
    refetchLoading.value = false
  }
}

function openSourceUrl() {
  if (sourceUrl.value) window.open(sourceUrl.value, "_blank")
}

function handleNoteDeleted() {
  router.push({ name: "notes" })
}

function openDelete() {
  deleteDialogOpen.value = true
}

onMounted(async () => {
  await fetchNote(props.id)
  if (note.value) {
    title.value = note.value.title ?? ""
    content.value = note.value.content ?? ""
    sourceUrl.value = note.value.sourceUrl ?? ""
  }
})

const debouncedSave = useDebounceFn(async () => {
  if (!note.value) return
  try {
    await updateNote(note.value.id, { content: content.value })
  } catch (e) {
    toastApiError(e)
  } finally {
    changed.value = false
  }
}, 1500)

watch(content, (newValue, oldValue) => {
  if (newValue === oldValue) return
  changed.value = true
  debouncedSave()
})

useHead({ title: computed(() => note.value?.title ?? t("head.note")) })
</script>

<template>
  <AppLayout>
    <div v-if="loading" class="flex flex-col container gap-4">
      <Skeleton class="h-12 w-full" />
      <Skeleton class="h-12 w-full" />
      <Skeleton class="h-full w-full" />
    </div>

    <div
      v-else-if="error"
      class="flex flex-col items-center justify-center gap-4 py-24 text-muted-foreground"
    >
      <template v-if="error.status === 404 || error.status === 422">
        <Sticker class="w-12 h-12" />
        <p class="text-muted-foreground">{{ t("errors.notFound.message") }}</p>
        <Button variant="outline" @click="router.push({ name: 'notes' })">
          <Undo2 class="h-4 w-4 pointer-events-none" />
          {{ t("notes.backToNotes") }}
        </Button>
      </template>
      <template v-else>
        <p class="text-destructive">{{ error.detail }}</p>
        <Button variant="outline" @click="fetchNote(props.id)">
          <RefreshCw class="h-4 w-4 pointer-events-none" />
          {{ t("common.retry") }}
        </Button>
      </template>
    </div>

    <div v-else-if="note" class="flex flex-col h-full max-h-screen container">
      <div class="flex items-center justify-between py-3 border-b gap-4">
        <Button
          variant="ghost"
          size="icon"
          @click="router.push({ name: 'notes' })"
        >
          <ArrowLeft class="w-4 h-4 pointer-events-none" />
        </Button>

        <div class="flex items-center gap-2">
          <template v-if="saving">
            <Loader2 class="w-4 h-4 animate-spin text-muted-foreground" />
            <span class="text-sm text-muted-foreground">
              {{ t("common.saving") }}
            </span>
          </template>
          <template v-else-if="changed">
            <X class="w-4 h-4 text-red-400" />
            <span class="text-sm text-red-400">
              {{ t("notes.update.unsaved") }}
            </span>
          </template>
          <template v-else>
            <Check class="w-4 h-4 text-emerald-400" />
            <span class="text-sm text-emerald-400">
              {{ t("notes.update.saved") }}
            </span>
          </template>
        </div>

        <div class="flex items-center gap-2">
          <Popover>
            <PopoverTrigger as-child>
              <Button variant="outline" size="sm">
                <Wand2 class="w-4 h-4 mr-2 pointer-events-none" />
                {{ t("notes.update.autotag") }}
              </Button>
            </PopoverTrigger>
            <PopoverContent class="w-48 flex flex-col gap-3">
              <NumberField
                v-model="autotagCount"
                :default-value="3"
                :min="1"
                :max="10"
              >
                <NumberFieldContent>
                  <NumberFieldDecrement />
                  <NumberFieldInput />
                  <NumberFieldIncrement />
                </NumberFieldContent>
              </NumberField>
              <Button
                size="sm"
                @click="handleAutotag"
                :disabled="autotagLoading"
              >
                <Loader2
                  v-if="autotagLoading"
                  class="w-4 h-4 mr-2 animate-spin pointer-events-none"
                />
                {{ t("notes.tags.generate") }}
              </Button>
            </PopoverContent>
          </Popover>
          <DropdownMenu>
            <DropdownMenuTrigger as-child>
              <Button variant="ghost" size="icon">
                <Ellipsis class="w-4 h-4 pointer-events-none" />
              </Button>
            </DropdownMenuTrigger>
            <DropdownMenuContent align="end">
              <DropdownMenuItem @click="infoOpen = true">
                <Info class="w-4 h-4 mr-2 pointer-events-none" />
                {{ t("common.additionalInfo") }}
              </DropdownMenuItem>
              <DropdownMenuItem
                v-if="note.type === 'url'"
                @click="sourceOpen = true"
              >
                <LinkIcon class="w-4 h-4 mr-2 pointer-events-none" />
                {{ t("notes.info.sourceInfo") }}
              </DropdownMenuItem>
              <DropdownMenuSeparator />
              <DropdownMenuItem class="text-destructive" @click="openDelete">
                <Trash2 class="w-4 h-4 mr-2 pointer-events-none" />
                {{ t("common.delete") }}
              </DropdownMenuItem>
            </DropdownMenuContent>
          </DropdownMenu>
        </div>
      </div>

      <div class="pt-6">
        <Input
          v-model="title"
          @blur="handleTitleChange"
          class="text-2xl font-bold focus-visible:ring-0"
        />
      </div>

      <div v-if="note.type === 'url'" class="flex items-center gap-2 pt-2">
        <Input
          v-model="sourceUrl"
          :placeholder="t('notes.url.source')"
          class="flex-1"
          @input="debouncedUrlSave"
        />
        <Button
          v-if="sourceUrl"
          variant="ghost"
          size="icon"
          @click="openSourceUrl"
        >
          <ExternalLink class="w-4 h-4 pointer-events-none" />
        </Button>
        <Button
          variant="ghost"
          size="icon"
          @click="handleRefetch"
          :disabled="refetchLoading"
        >
          <Loader2
            v-if="refetchLoading"
            class="w-4 h-4 animate-spin pointer-events-none"
          />
          <RefreshCw v-else class="w-4 h-4 pointer-events-none" />
        </Button>
      </div>

      <div
        v-if="note.type === 'url' && note.urlMetadata"
        class="flex flex-col gap-0.5 pt-2 text-sm text-muted-foreground"
      >
        <span v-if="note.urlMetadata.siteName" class="font-medium text-foreground">
          {{ note.urlMetadata.siteName }}
        </span>
        <span v-if="note.urlMetadata.title">{{ note.urlMetadata.title }}</span>
      </div>

      <div class="flex items-center gap-2 py-2 flex-wrap">
        <Badge
          v-for="tag in note.tags"
          :key="tag.id"
          variant="outline"
          class="flex items-center gap-1 py-0 h-6 cursor-pointer"
          :class="
            tag.source === 'auto'
              ? 'border-violet-400/60 text-violet-700 dark:text-violet-300 bg-linear-to-r from-violet-500/10 to-cyan-500/10'
              : ''
          "
          @click="handleUnbindTag(tag.id)"
        >
          <Sparkles
            v-if="tag.source === 'auto'"
            class="w-3 h-3 shrink-0 pointer-events-none text-violet-500"
          />
          <span class="pointer-events-none">
            {{ tag.name }}
          </span>
          <X class="w-4 h-4 shrink-0 pointer-events-none" />
        </Badge>
        <NoteTagSelector
          :note-id="note.id"
          :bound-tag-ids="note.tags.map((t) => t.id)"
          @bind="handleBindTag"
        />
      </div>

      <div class="h-full py-4 overflow-y-auto">
        <NoteEditor
          :theme="isDark ? 'dark' : 'light'"
          v-model:content="content"
        />
      </div>
    </div>

    <EditorShortcutsHelper />
    <DeleteNoteAlertModal
      v-model:open="deleteDialogOpen"
      :note-id="id"
      @deleted="handleNoteDeleted"
    />
    <NoteInfoModal v-if="note" v-model:open="infoOpen" :note="note" />
    <NoteSourceModal v-if="note" v-model:open="sourceOpen" :note="note" />
  </AppLayout>
</template>
