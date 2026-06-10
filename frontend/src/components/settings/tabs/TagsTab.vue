<script setup lang="ts">
import { ref, onMounted, watch, computed } from "vue"
import { useI18n } from "vue-i18n"
import { useIntersectionObserver, watchDebounced } from "@vueuse/core"
import { tagsService } from "@/services/tags"
import { toast } from "vue-sonner"
import { toastApiError } from "@/services/apiError"
import type { Tag, ApiError } from "@/types"
import {
  Loader2,
  RefreshCw,
  Trash2,
  Plus,
  Check,
  X,
  ArrowUp,
  ArrowDown,
} from "lucide-vue-next"
import { Input } from "@/components/ui/input"
import { Button } from "@/components/ui/button"
import { ButtonGroup } from "@/components/ui/button-group"
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from "@/components/ui/select"
import { useInfiniteEntities } from "@/composables/useInfiniteEntities"
import DialogDescription from "@/components/ui/dialog/DialogDescription.vue"
import Separator from "@/components/ui/separator/Separator.vue"

const PAGE_SIZE = 20

const { t } = useI18n()

const { entities, loading, loadingMore, error, fetchEntities, loadMore } =
  useInfiniteEntities<Tag>(tagsService)

const query = ref("")
const sortBy = ref("name")
const orderBy = ref("ASC")

const sentinel = ref<HTMLElement | null>(null)

const params = computed(() => ({
  q: query.value,
  sort_by: sortBy.value,
  order_by: orderBy.value,
  page_size: PAGE_SIZE,
}))

const newTag = ref("")
const addLoading = ref(false)

const editingId = ref<string | null>(null)
const editingName = ref("")

async function loadTags() {
  await fetchEntities(params.value)
}

async function handleAdd() {
  if (!newTag.value.trim()) return
  addLoading.value = true
  try {
    const result = await tagsService.create(newTag.value.trim())
    entities.value = [result.data, ...entities.value]
    newTag.value = ""
    toast.success(t("settings.tags.messages.added"))
  } catch (e) {
    toastApiError(e)
  } finally {
    addLoading.value = false
  }
}

async function handleDelete(id: string) {
  try {
    await tagsService.remove(id)
    entities.value = entities.value.filter((t) => t.id !== id)
    toast.info(t("settings.tags.messages.deleted"))
  } catch (e) {
    toastApiError(e)
  }
}

function startEdit(tag: Tag) {
  editingId.value = tag.id
  editingName.value = tag.name
}

function cancelEdit() {
  editingId.value = null
  editingName.value = ""
}

async function confirmEdit(tag: Tag) {
  if (!editingName.value.trim() || editingName.value === tag.name) {
    cancelEdit()
    return
  }
  try {
    await tagsService.update(tag.id, editingName.value.trim())
    tag.name = editingName.value.trim()
    cancelEdit()
  } catch (e) {
    toastApiError(e)
  }
}

useIntersectionObserver(sentinel, ([entry]) => {
  if (entry?.isIntersecting) loadMore(params.value)
})

watchDebounced(query, loadTags, { debounce: 400 })
watch([sortBy, orderBy], loadTags)
onMounted(loadTags)
</script>

<template>
  <div class="flex flex-col gap-4">
    <DialogDescription class="text-center">
      {{ t("settings.tags.description") }}
    </DialogDescription>

    <div class="flex flex-col gap-2">
      <Input
        v-model="query"
        class="focus-visible:ring-1"
        :placeholder="t('settings.tags.search')"
      />
      <ButtonGroup class="w-full flex">
        <Button
          variant="outline"
          size="icon"
          @click="orderBy = orderBy === 'ASC' ? 'DESC' : 'ASC'"
        >
          <ArrowUp
            v-if="orderBy === 'ASC'"
            class="w-4 h-4 pointer-events-none"
          />
          <ArrowDown v-else class="w-4 h-4 pointer-events-none" />
        </Button>
        <Select v-model="sortBy">
          <SelectTrigger class="w-full">
            <SelectValue />
          </SelectTrigger>
          <SelectContent>
            <SelectItem value="name">
              {{ t("settings.tags.sort.name") }}
            </SelectItem>
            <SelectItem value="updated_at">
              {{ t("settings.tags.sort.updatedAt") }}
            </SelectItem>
            <SelectItem value="created_at">
              {{ t("settings.tags.sort.createdAt") }}
            </SelectItem>
          </SelectContent>
        </Select>
      </ButtonGroup>
      <div class="flex gap-2">
        <Input
          v-model="newTag"
          :placeholder="t('settings.tags.add.placeholder')"
          class="flex-1 focus-visible:ring-1"
          @keydown.enter="handleAdd"
        />
        <Button :disabled="addLoading || !newTag.trim()" @click="handleAdd">
          <Loader2 v-if="addLoading" class="animate-spin pointer-events-none" />
          <Plus v-else class="pointer-events-none" />
          {{ t("common.add") }}
        </Button>
      </div>
    </div>

    <div v-if="loading" class="flex items-center justify-center min-h-80">
      <Loader2 class="w-5 h-5 animate-spin text-muted-foreground" />
    </div>

    <div
      v-else-if="error"
      class="flex flex-col items-center justify-center gap-2 min-h-80"
    >
      <p class="text-sm text-destructive">{{ error.detail }}</p>
      <Button variant="outline" size="sm" @click="loadTags">
        <RefreshCw class="w-4 h-4 mr-2 pointer-events-none" />
        {{ t("common.retry") }}
      </Button>
    </div>

    <div
      v-else-if="entities.length === 0"
      class="text-center py-8 text-sm text-muted-foreground"
    >
      {{ t("settings.tags.empty") }}
    </div>

    <div v-else class="flex flex-col divide-y">
      <div
        v-for="tag in entities"
        :key="tag.id"
        class="flex items-center justify-between py-2 gap-3"
      >
        <div
          v-if="editingId === tag.id"
          class="flex items-center gap-2 flex-1 min-w-0"
        >
          <Input
            v-model="editingName"
            class="h-7 text-sm"
            @keydown.enter="confirmEdit(tag)"
            @keydown.escape="cancelEdit"
            autofocus
          />
          <Button
            variant="ghost"
            size="icon"
            class="w-6 h-6 shrink-0"
            @click="confirmEdit(tag)"
          >
            <Check class="w-3 h-3 pointer-events-none" />
          </Button>
          <Button
            variant="ghost"
            size="icon"
            class="w-6 h-6 shrink-0"
            @click="cancelEdit"
          >
            <X class="w-3 h-3 pointer-events-none" />
          </Button>
        </div>

        <div v-else class="flex items-center gap-2 flex-1 min-w-0">
          <span
            class="text-sm truncate cursor-pointer hover:text-foreground"
            @click="startEdit(tag)"
          >
            {{ tag.name }}
          </span>
          <span class="text-xs text-muted-foreground shrink-0">
            {{ t("settings.tags.fields.noteCount", { count: tag.itemCount }) }}
          </span>
        </div>

        <Button
          v-if="editingId !== tag.id"
          variant="ghost"
          size="icon"
          class="w-6 h-6 shrink-0"
          @click="handleDelete(tag.id)"
        >
          <Trash2 class="w-3 h-3 pointer-events-none" />
        </Button>
      </div>

      <div ref="sentinel" class="h-1" />

      <div v-if="loadingMore" class="flex justify-center py-4">
        <Loader2 class="w-4 h-4 animate-spin text-muted-foreground" />
      </div>
    </div>
  </div>
</template>
