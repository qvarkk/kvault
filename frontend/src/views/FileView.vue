<script setup lang="ts">
import { ref, computed, onMounted, onUnmounted } from "vue"
import { useRouter } from "vue-router"
import { useI18n } from "vue-i18n"
import { filesService } from "@/services/files"
import { toast } from "vue-sonner"
import { toastApiError } from "@/services/apiError"
import type { ApiError, File, PresignedUrl } from "@/types"
import AppLayout from "@/components/layout/AppLayout.vue"
import { Button } from "@/components/ui/button"
import { Skeleton } from "@/components/ui/skeleton"
import { ArrowLeft, Download, Ellipsis, Info, RefreshCw, Sticker, Trash2 } from "lucide-vue-next"
import {
  DropdownMenu,
  DropdownMenuContent,
  DropdownMenuItem,
  DropdownMenuSeparator,
  DropdownMenuTrigger,
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
import FileInfoModal from "@/components/files/FileInfoModal.vue"
import { useHead } from "@vueuse/head"

const props = defineProps<{
  id: string
}>()

const router = useRouter()
const { t } = useI18n()

const loading = ref(true)
const error = ref<ApiError | null>(null)
const viewData = ref<PresignedUrl | null>(null)
const blobUrl = ref<string | null>(null)
const fileInfo = ref<File | null>(null)
const infoOpen = ref(false)
const deleteOpen = ref(false)

useHead({ title: computed(() => viewData.value?.filename ?? t("head.file")) })

function revokeBlobUrl() {
  if (blobUrl.value) {
    URL.revokeObjectURL(blobUrl.value)
    blobUrl.value = null
  }
}

async function fetchViewUrl() {
  loading.value = true
  error.value = null
  viewData.value = null
  revokeBlobUrl()
  try {
    const { data } = await filesService.getViewUrl(props.id)

    const response = await fetch(data.url)
    if (!response.ok) {
      error.value = {
        title: t("errors.internal.title"),
        detail: `${response.status} ${response.statusText}`,
        type: "",
        instance: "",
        status: response.status,
      }
      return
    }

    const blob = await response.blob()
    blobUrl.value = URL.createObjectURL(new Blob([blob], { type: "application/pdf" }))
    viewData.value = data
  } catch (e) {
    if (e && typeof e === "object" && "detail" in e) {
      error.value = e as ApiError
    } else {
      error.value = {
        title: t("errors.network.title"),
        detail: t("errors.network.detail"),
        type: "",
        instance: "",
        status: 0,
      }
    }
  } finally {
    loading.value = false
  }
}

async function fetchInfo() {
  try {
    const { data } = await filesService.getInfo(props.id)
    fileInfo.value = data
  } catch {
    // Non-fatal: the viewer still works without the metadata panel.
    fileInfo.value = null
  }
}

async function handleDownload() {
  try {
    const response = await filesService.getDownloadUrl(props.id)
    window.open(response.data.url, "_blank")
  } catch (e) {
    toastApiError(e)
  }
}

async function handleDelete() {
  try {
    await filesService.remove(props.id)
    toast.success(t("files.delete.deleted"))
    router.push({ name: "files" })
  } catch (e) {
    toastApiError(e)
  }
}

onMounted(() => {
  fetchViewUrl()
  fetchInfo()
})
onUnmounted(revokeBlobUrl)
</script>

<template>
  <AppLayout>
    <div v-if="loading" class="flex flex-col container gap-4 h-full">
      <Skeleton class="h-12 w-full shrink-0" />
      <Skeleton class="flex-1 w-full" />
    </div>

    <div
      v-else-if="error"
      class="flex flex-col items-center justify-center gap-4 py-24 text-muted-foreground"
    >
      <Sticker class="w-12 h-12" />
      <p class="text-destructive">{{ error.detail }}</p>
      <div class="flex gap-2">
        <Button variant="outline" @click="router.push({ name: 'files' })">
          <ArrowLeft class="h-4 w-4 mr-2 pointer-events-none" />
          {{ t("nav.files") }}
        </Button>
        <Button variant="outline" @click="fetchViewUrl">
          <RefreshCw class="h-4 w-4 mr-2 pointer-events-none" />
          {{ t("common.retry") }}
        </Button>
      </div>
    </div>

    <div
      v-else-if="viewData && blobUrl"
      class="flex flex-col h-full max-h-screen container"
    >
      <div
        class="flex items-center justify-between py-3 border-b gap-4 shrink-0"
      >
        <Button
          variant="ghost"
          size="icon"
          @click="router.push({ name: 'files' })"
        >
          <ArrowLeft class="w-4 h-4 pointer-events-none" />
        </Button>

        <span class="text-sm font-medium truncate flex-1 text-center">
          {{ viewData.filename }}
        </span>

        <div class="flex items-center gap-2">
          <Button variant="outline" size="icon" @click="handleDownload">
            <Download class="w-4 h-4 pointer-events-none" />
          </Button>
          <DropdownMenu>
            <DropdownMenuTrigger as-child>
              <Button variant="ghost" size="icon">
                <Ellipsis class="w-4 h-4 pointer-events-none" />
              </Button>
            </DropdownMenuTrigger>
            <DropdownMenuContent align="end">
              <DropdownMenuItem :disabled="!fileInfo" @click="infoOpen = true">
                <Info class="w-4 h-4 mr-2 pointer-events-none" />
                {{ t("common.additionalInfo") }}
              </DropdownMenuItem>
              <DropdownMenuItem @click="handleDownload">
                <Download class="w-4 h-4 mr-2 pointer-events-none" />
                {{ t("files.download") }}
              </DropdownMenuItem>
              <DropdownMenuSeparator />
              <DropdownMenuItem class="text-destructive" @click="deleteOpen = true">
                <Trash2 class="w-4 h-4 mr-2 pointer-events-none" />
                {{ t("common.delete") }}
              </DropdownMenuItem>
            </DropdownMenuContent>
          </DropdownMenu>
        </div>
      </div>

      <iframe :src="blobUrl" class="flex-1 w-full" frameborder="0" />
    </div>

    <FileInfoModal v-if="fileInfo" v-model:open="infoOpen" :file="fileInfo" />

    <AlertDialog v-model:open="deleteOpen">
      <AlertDialogContent>
        <AlertDialogHeader>
          <AlertDialogTitle>{{ t("files.delete.title") }}</AlertDialogTitle>
          <AlertDialogDescription>
            {{ t("files.delete.description") }}
          </AlertDialogDescription>
        </AlertDialogHeader>
        <AlertDialogFooter>
          <AlertDialogCancel>{{ t("common.cancel") }}</AlertDialogCancel>
          <AlertDialogAction
            class="bg-destructive text-destructive-foreground hover:bg-destructive/90"
            @click="handleDelete"
          >
            {{ t("common.delete") }}
          </AlertDialogAction>
        </AlertDialogFooter>
      </AlertDialogContent>
    </AlertDialog>
  </AppLayout>
</template>
