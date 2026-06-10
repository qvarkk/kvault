<script setup lang="ts">
import { ref, watch } from "vue"
import { useRouter } from "vue-router"
import { useI18n } from "vue-i18n"
import { notesService } from "@/services/notes"
import type { ApiError } from "@/types"
import { toast } from "vue-sonner"
import { toastApiError } from "@/services/apiError"
import {
  Dialog,
  DialogContent,
  DialogHeader,
  DialogTitle,
  DialogFooter,
} from "@/components/ui/dialog"
import { Tabs, TabsContent, TabsList, TabsTrigger } from "@/components/ui/tabs"
import { Input } from "@/components/ui/input"
import { Button } from "@/components/ui/button"

const { t } = useI18n()
const router = useRouter()

const open = defineModel<boolean>("open")
const activeTab = ref<"text" | "url">("text")
const title = ref("")
const url = ref("")
const loading = ref(false)
const urlInputRef = ref<InstanceType<typeof Input> | null>(null)

watch(open, (val) => {
  if (!val) {
    title.value = ""
    url.value = ""
    activeTab.value = "text"
  }
})

const isValid = () => {
  if (!title.value.trim()) return false
  if (activeTab.value === "url" && !url.value.trim()) return false
  return true
}

function focusUrlInput() {
  urlInputRef.value?.$el?.focus()
}

async function handleCreate() {
  if (!isValid()) return
  loading.value = true

  try {
    const response = activeTab.value === "url"
      ? await notesService.createUrl(title.value.trim(), url.value.trim())
      : await notesService.createText(title.value.trim())
    toast.success(t("notes.create.created"))
    open.value = false
    router.push({ name: "note", params: { id: response.data.id } })
  } catch (e) {
    toastApiError(e)
  } finally {
    loading.value = false
  }
}
</script>

<template>
  <Dialog v-model:open="open">
    <DialogContent aria-describedby="">
      <DialogHeader>
        <DialogTitle>{{ t("notes.create.title") }}</DialogTitle>
      </DialogHeader>
      <Tabs v-model="activeTab">
        <TabsList class="w-full">
          <TabsTrigger value="text">{{ t("notes.create.tabs.text") }}</TabsTrigger>
          <TabsTrigger value="url">{{ t("notes.create.tabs.url") }}</TabsTrigger>
        </TabsList>
        <TabsContent value="text" class="mt-3">
          <Input
            v-model="title"
            :placeholder="t('notes.create.placeholder')"
            @keydown.enter="handleCreate"
          />
        </TabsContent>
        <TabsContent value="url" class="flex flex-col gap-2 mt-3">
          <Input
            v-model="title"
            :placeholder="t('notes.create.placeholder')"
            @keydown.enter="focusUrlInput"
          />
          <Input
            ref="urlInputRef"
            v-model="url"
            :placeholder="t('notes.create.urlPlaceholder')"
            @keydown.enter="handleCreate"
          />
        </TabsContent>
      </Tabs>
      <DialogFooter>
        <Button variant="outline" @click="open = false">
          {{ t("common.cancel") }}
        </Button>
        <Button @click="handleCreate" :disabled="loading || !isValid()">
          {{ loading ? t("common.creating") : t("common.create") }}
        </Button>
      </DialogFooter>
    </DialogContent>
  </Dialog>
</template>
