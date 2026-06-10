<script setup lang="ts">
import { ref, computed, onMounted, onUnmounted, watch, reactive } from "vue"
import { useI18n } from "vue-i18n"
import {
  watchDebounced,
  useIntersectionObserver,
  useDropZone,
} from "@vueuse/core"
import { useInfiniteEntities } from "@/composables/useInfiniteEntities"
import { safeUuid } from "@/lib/utils"
import { filesService } from "@/services/files"
import { toast } from "vue-sonner"
import { toastApiError } from "@/services/apiError"
import type { ApiError, UploadItem, File } from "@/types"
import { Check, FileIcon, Files, RefreshCw, Trash2, X } from "lucide-vue-next"
import { Button } from "@/components/ui/button"
import { Skeleton } from "@/components/ui/skeleton"
import { Card } from "@/components/ui/card"
import { Progress } from "@/components/ui/progress"
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
import AppLayout from "@/components/layout/AppLayout.vue"
import FileCard from "@/components/files/FileCard.vue"
import { Spinner } from "@/components/ui/spinner"
import FilesToolbar from "@/components/files/FilesToolbar.vue"
import { useHead } from "@vueuse/head"

const PAGE_SIZE = 20

const { t } = useI18n()
useHead({ title: t("head.files") })

const trashMode = ref(false)
const emptyTrashDialogOpen = ref(false)

const serviceProxy = {
  list: (params?: Record<string, unknown>) =>
    trashMode.value
      ? filesService.listDeleted(params)
      : filesService.list(params),
}

const { entities, loading, loadingMore, error, fetchEntities, loadMore } =
  useInfiniteEntities<File>(serviceProxy)

const query = ref("")
const sortBy = ref("created_at")
const orderBy = ref("DESC")
const sentinel = ref<HTMLElement | null>(null)
const dropZoneRef = ref<HTMLElement | null>(null)

const params = computed(() => ({
  q: query.value,
  sort_by: sortBy.value,
  order_by: orderBy.value,
  page_size: PAGE_SIZE,
}))

async function loadFiles() {
  await fetchEntities(params.value)
}

const uploadQueue = ref<UploadItem[]>([])
const uploadControllers = new Map<string, AbortController>()

const hasFinishedUploads = computed(() =>
  uploadQueue.value.some(i => i.status !== "uploading")
)

function clearFinished() {
  uploadQueue.value = uploadQueue.value.filter(i => i.status === "uploading")
}

function cancelUpload(itemId: string) {
  const item = uploadQueue.value.find(i => i.id === itemId)
  if (!item) return
  item.status = "cancelled"
  uploadControllers.get(itemId)?.abort()
  uploadControllers.delete(itemId)
}

async function uploadFile(file: globalThis.File) {
  const item = reactive<UploadItem>({
    id: safeUuid(),
    name: file.name,
    progress: 0,
    status: "uploading",
  })
  uploadQueue.value.push(item)

  const controller = new AbortController()
  uploadControllers.set(item.id, controller)

  try {
    const response = await filesService.upload(
      file,
      (percent) => { item.progress = percent },
      controller.signal,
    )
    uploadControllers.delete(item.id)
    item.status = "done"
    entities.value = [response.data, ...entities.value]
    toast.success(t("files.upload.uploaded"))
  } catch (e) {
    uploadControllers.delete(item.id)
    if (item.status === "cancelled") return
    item.status = "error"
    if (e && typeof e === "object" && "detail" in e)
      item.error = (e as ApiError).detail
    else item.error = t("errors.internal.title")
  }
}

onUnmounted(() => {
  uploadControllers.forEach(c => c.abort())
})

async function handleFiles(incoming: globalThis.File[] | FileList | null) {
  if (trashMode.value || !incoming) return
  const arr = Array.from(incoming).filter((f) => f.type === "application/pdf")
  if (arr.length === 0) {
    toast.error(t("files.onlyPdf"))
    return
  }
  await Promise.all(arr.map(uploadFile))
}

const { isOverDropZone } = useDropZone(dropZoneRef, {
  onDrop: handleFiles,
  dataTypes: ["application/pdf"],
})

function openFilePicker() {
  const input = document.createElement("input")
  input.type = "file"
  input.accept = "application/pdf"
  input.multiple = true
  input.onchange = () => handleFiles(input.files)
  input.click()
}

async function handleDownload(id: string) {
  try {
    const response = await filesService.getDownloadUrl(id)
    window.open(response.data.url, "_blank")
  } catch (e) {
    toastApiError(e)
  }
}

async function handleDelete(id: string) {
  try {
    if (trashMode.value) {
      await filesService.permanentDelete(id)
      toast.success(t("files.trash.deletedPermanently"))
    } else {
      await filesService.remove(id)
      toast.success(t("files.delete.deleted"))
    }
    entities.value = entities.value.filter((f) => f.id !== id)
  } catch (e) {
    toastApiError(e)
  }
}

async function handleRestore(id: string) {
  try {
    await filesService.restore(id)
    entities.value = entities.value.filter((f) => f.id !== id)
    toast.success(t("files.trash.restored"))
  } catch (e) {
    toastApiError(e)
  }
}

async function handleEmptyTrash() {
  try {
    await filesService.emptyTrash()
    toast.success(t("files.trash.emptied"))
    emptyTrashDialogOpen.value = false
    await loadFiles()
  } catch (e) {
    toastApiError(e)
  }
}

useIntersectionObserver(sentinel, ([entry]) => {
  if (entry?.isIntersecting) loadMore(params.value)
})

watchDebounced(query, loadFiles, { debounce: 400 })
watch([sortBy, orderBy], loadFiles)
watch(trashMode, loadFiles)
onMounted(loadFiles)
</script>

<template>
  <AppLayout>
    <div
      ref="dropZoneRef"
      class="flex flex-col gap-4 min-h-full relative container"
    >
      <div
        v-if="isOverDropZone && !trashMode"
        class="absolute inset-0 z-50 bg-background/80 border-2 border-dashed border-primary rounded-lg flex items-center justify-center"
      >
        <p class="text-primary text-lg font-medium">
          {{ t("files.dropHere") }}
        </p>
      </div>

      <div class="flex items-center justify-between gap-2">
        <div class="flex gap-1">
          <Button
            :variant="trashMode ? 'outline' : 'secondary'"
            size="sm"
            @click="trashMode = false"
          >
            <Files class="w-3.5 h-3.5 mr-1.5 pointer-events-none" />
            {{ t("nav.files") }}
          </Button>
          <Button
            :variant="trashMode ? 'secondary' : 'outline'"
            size="sm"
            @click="trashMode = true"
          >
            <Trash2 class="w-3.5 h-3.5 mr-1.5 pointer-events-none" />
            {{ t("files.trash.toggle") }}
          </Button>
        </div>
        <Button
          v-if="trashMode && entities.length > 0 && !loading"
          variant="destructive"
          size="sm"
          @click="emptyTrashDialogOpen = true"
        >
          {{ t("files.trash.emptyAction") }}
        </Button>
      </div>

      <FilesToolbar
        v-model:query="query"
        v-model:order-by="orderBy"
        v-model:sort-by="sortBy"
        v-model:loading="loading"
        :trash="trashMode"
        @upload-click="openFilePicker"
      />

      <div v-if="uploadQueue.length > 0 && !trashMode" class="flex flex-col gap-1">
        <div class="flex justify-end">
          <Button
            v-if="hasFinishedUploads"
            variant="ghost"
            size="sm"
            class="text-muted-foreground"
            @click="clearFinished"
          >
            {{ t("files.upload.clearFinished") }}
          </Button>
        </div>
        <div
          v-for="item in uploadQueue"
          :key="item.id"
          class="flex items-center gap-3 p-3 rounded-lg border"
        >
          <FileIcon
            class="w-4 h-4 shrink-0 text-muted-foreground pointer-events-none"
          />
          <span class="text-sm flex-1 truncate">{{ item.name }}</span>
          <span v-if="item.status === 'error'" class="text-xs text-destructive">
            {{ item.error }}
          </span>
          <span v-else-if="item.status === 'cancelled'" class="text-xs text-muted-foreground">
            {{ t("common.cancel") }}
          </span>
          <Progress
            v-else-if="item.status === 'uploading'"
            :value="item.progress"
            class="w-24"
          />
          <Check v-else class="w-4 h-4 text-green-500 pointer-events-none" />
          <Button
            v-if="item.status === 'uploading'"
            variant="ghost"
            size="icon"
            class="w-6 h-6 shrink-0"
            @click="cancelUpload(item.id)"
          >
            <X class="w-3 h-3 pointer-events-none" />
          </Button>
        </div>
      </div>

      <div
        v-if="loading"
        class="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-4"
      >
        <Card v-for="n in 6" :key="n" class="p-4 flex flex-col gap-3">
          <Skeleton class="h-5 w-3/4" />
          <Skeleton class="h-3 w-1/2" />
          <Skeleton class="h-3 w-1/3" />
        </Card>
      </div>

      <div
        v-else-if="error"
        class="flex flex-col items-center justify-center gap-4 py-24 text-muted-foreground"
      >
        <p class="text-destructive">{{ error.detail }}</p>
        <Button variant="outline" @click="loadFiles">
          <RefreshCw class="w-4 h-4 mr-2 pointer-events-none" />
          {{ t("common.retry") }}
        </Button>
      </div>

      <div v-else>
        <div
          v-if="entities.length === 0"
          class="flex flex-col items-center justify-center gap-4 py-24 text-muted-foreground"
        >
          <Files class="w-12 h-12" />
          <p>{{ trashMode ? t("files.trash.empty") : t("files.empty") }}</p>
          <Button v-if="!trashMode" @click="openFilePicker">
            {{ t("files.upload.first") }}
          </Button>
        </div>

        <div
          v-else
          class="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-4"
        >
          <FileCard
            v-for="file in entities"
            :key="file.id"
            :file="file"
            :trash="trashMode"
            @download="handleDownload"
            @delete="handleDelete"
            @restore="handleRestore"
          />
        </div>

        <div ref="sentinel" class="h-1" />
        <div v-if="loadingMore" class="flex justify-center py-4">
          <Spinner />
        </div>
      </div>
    </div>
  </AppLayout>

  <AlertDialog v-model:open="emptyTrashDialogOpen">
    <AlertDialogContent>
      <AlertDialogHeader>
        <AlertDialogTitle>{{ t("files.trash.emptyConfirm.title") }}</AlertDialogTitle>
        <AlertDialogDescription>
          {{ t("files.trash.emptyConfirm.description") }}
        </AlertDialogDescription>
      </AlertDialogHeader>
      <AlertDialogFooter>
        <AlertDialogCancel>{{ t("common.cancel") }}</AlertDialogCancel>
        <AlertDialogAction
          @click="handleEmptyTrash"
          class="bg-destructive text-destructive-foreground hover:bg-destructive/90"
        >
          {{ t("files.trash.emptyConfirm.action") }}
        </AlertDialogAction>
      </AlertDialogFooter>
    </AlertDialogContent>
  </AlertDialog>
</template>
