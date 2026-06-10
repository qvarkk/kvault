<script setup lang="ts">
import { ref, onMounted, computed } from "vue"
import { useI18n } from "vue-i18n"
import { Plus, RefreshCw } from "lucide-vue-next"
import { watchDebounced, useIntersectionObserver } from "@vueuse/core"
import { tagsService } from "@/services/tags"
import { Input } from "@/components/ui/input"
import { Button } from "@/components/ui/button"
import {
  Popover,
  PopoverContent,
  PopoverTrigger,
} from "@/components/ui/popover"
import type { ApiError, Tag } from "@/types"
import { toast } from "vue-sonner"
import { toastApiError } from "@/services/apiError"
import { useInfiniteEntities } from "@/composables/useInfiniteEntities"
import { Spinner } from "../ui/spinner"

const PAGE_SIZE = 8

const props = defineProps<{
  noteId: string
  boundTagIds: string[]
}>()

const emit = defineEmits<{
  bind: [tag: Tag]
}>()

const { t } = useI18n()
const { entities, loading, loadingMore, error, fetchEntities, loadMore } =
  useInfiniteEntities<Tag>(tagsService)

const search = ref("")
const params = computed(() => ({
  q: search.value,
  page_size: PAGE_SIZE,
}))

const sentinel = ref<HTMLElement | null>(null)

async function fetch() {
  await fetchEntities(params.value)
}

async function handleCreate() {
  if (!search.value.trim()) return
  try {
    const response = await tagsService.create(search.value.trim())
    emit("bind", response.data)
    search.value = ""
    await fetch()
  } catch (e) {
    toastApiError(e)
  }
}

async function handleBind(tag: Tag) {
  emit("bind", tag)
}

const unbound = computed(() =>
  entities.value.filter((t) => !props.boundTagIds.includes(t.id)),
)

const exactMatch = computed(() =>
  entities.value.some(
    (t) => t.name.toLowerCase() === search.value.toLowerCase(),
  ),
)

useIntersectionObserver(sentinel, ([entry]) => {
  if (entry?.isIntersecting) loadMore(params.value)
})

watchDebounced(search, fetch, { debounce: 400 })
onMounted(fetch)
</script>

<template>
  <Popover>
    <PopoverTrigger as-child>
      <Button
        variant="outline"
        size="sm"
        class="text-xs h-6 gap-1 rounded-full"
      >
        <Plus class="size-3 pointer-events-none" />
        {{ t("notes.tags.add") }}
      </Button>
    </PopoverTrigger>
    <PopoverContent class="w-56 p-2 flex flex-col gap-2">
      <Input
        v-model="search"
        :placeholder="t('notes.tags.search')"
        class="h-8"
      />

      <div v-if="loading" class="flex justify-center py-4">
        <Spinner />
      </div>

      <div v-else-if="error" class="flex flex-col items-center gap-2 py-2">
        <p class="text-xs text-destructive">{{ error.detail }}</p>
        <Button variant="ghost" size="sm" @click="fetch">
          <RefreshCw class="w-3 h-3 mr-1 pointer-events-none" />
          {{ t("common.retry") }}
        </Button>
      </div>

      <div v-else class="flex flex-col max-h-48 overflow-y-auto">
        <button
          v-for="tag in unbound"
          :key="tag.id"
          @click="handleBind(tag)"
          class="flex items-center px-2 py-1.5 rounded-md hover:bg-accent text-sm cursor-pointer text-left"
        >
          {{ tag.name }}
        </button>
        <div v-if="unbound.length === 0">
          <Button
            v-if="search.trim() && !exactMatch"
            @click="handleCreate"
            class="w-full m-0"
          >
            <Plus class="w-3 h-3 pointer-events-none" />
            {{ t("notes.tags.create", { name: search }) }}
          </Button>
          <p v-else-if="!search.trim()" class="text-xs text-muted-foreground text-center py-2">
            {{ t("notes.tags.typeToAdd") }}
          </p>
        </div>

        <div ref="sentinel" class="h-1" />

        <div v-if="loadingMore" class="flex justify-center py-2">
          <Spinner />
        </div>
      </div>
    </PopoverContent>
  </Popover>
</template>
