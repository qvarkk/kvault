<script setup lang="ts">
import { ref, computed, onMounted, watch } from "vue"
import { useI18n } from "vue-i18n"
import { watchDebounced } from "@vueuse/core"
import { stopwordsService } from "@/services/stopwords"
import { toast } from "vue-sonner"
import { toastApiError } from "@/services/apiError"
import type { Stopword, ApiError } from "@/types"
import {
  Loader2,
  RefreshCw,
  Trash2,
  Plus,
  ArrowUp,
  ArrowDown,
} from "lucide-vue-next"
import { Input } from "@/components/ui/input"
import { Button } from "@/components/ui/button"
import { Switch } from "@/components/ui/switch"
import { Badge } from "@/components/ui/badge"
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from "@/components/ui/select"
import { ButtonGroup } from "@/components/ui/button-group"
import DialogDescription from "@/components/ui/dialog/DialogDescription.vue"
import Separator from "@/components/ui/separator/Separator.vue"

const { t } = useI18n()

const stopwords = ref<Stopword[]>([])
const loading = ref(false)
const error = ref<ApiError | null>(null)

const query = ref("")
const source = ref<string>("all")
const sortBy = ref("source")
const orderBy = ref("ASC")

const newWord = ref("")
const addLoading = ref(false)

const params = computed(() => ({
  q: query.value || undefined,
  source: source.value === "all" ? undefined : source.value,
  sort_by: sortBy.value,
  order_by: orderBy.value,
}))

async function fetchStopwords() {
  loading.value = true
  error.value = null
  try {
    const response = await stopwordsService.list(params.value)
    stopwords.value = response.data.data
  } catch (e) {
    error.value = e as ApiError
  } finally {
    loading.value = false
  }
}

async function handleToggle(word: Stopword) {
  try {
    if (word.isEnabled) {
      await stopwordsService.disable(word.word)
    } else {
      await stopwordsService.enable(word.word)
    }
    await fetchStopwords()
  } catch (e) {
    toastApiError(e)
  }
}

async function handleDelete(word: string) {
  try {
    await stopwordsService.remove(word)
    await fetchStopwords()
  } catch (e) {
    toastApiError(e)
  }
}

async function handleAdd() {
  if (!newWord.value.trim()) return
  addLoading.value = true
  try {
    await stopwordsService.create({ word: newWord.value.trim() })
    newWord.value = ""
    await fetchStopwords()
  } catch (e) {
    toastApiError(e)
  } finally {
    addLoading.value = false
  }
}

watchDebounced(query, fetchStopwords, { debounce: 400 })
watch([source, sortBy, orderBy], fetchStopwords)
onMounted(fetchStopwords)
</script>

<template>
  <div class="flex flex-col gap-4 h-full">
    <DialogDescription class="text-center">
      {{ t("settings.stopwords.description") }}
    </DialogDescription>

    <div class="flex flex-col gap-2">
      <Input
        v-model="query"
        :placeholder="t('settings.stopwords.search')"
        class="focus-visible:ring-1"
      />
      <div class="flex gap-2">
        <ButtonGroup>
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
          <Select v-model="sortBy" class="flex-1">
            <SelectTrigger class="w-36">
              <SelectValue />
            </SelectTrigger>
            <SelectContent>
              <SelectItem value="word">
                {{ t("settings.stopwords.sort.word") }}
              </SelectItem>
              <SelectItem value="source">
                {{ t("settings.stopwords.sort.source") }}
              </SelectItem>
              <SelectItem value="updated_at">
                {{ t("settings.stopwords.sort.updatedAt") }}
              </SelectItem>
            </SelectContent>
          </Select>
        </ButtonGroup>
        <Select v-model="source">
          <SelectTrigger class="w-32 flex-1">
            <SelectValue :placeholder="t('settings.stopwords.sources.all')" />
          </SelectTrigger>
          <SelectContent>
            <SelectItem value="all">
              {{ t("settings.stopwords.sources.all") }}
            </SelectItem>
            <SelectItem value="user">
              {{ t("settings.stopwords.sources.user") }}
            </SelectItem>
            <SelectItem value="default">
              {{ t("settings.stopwords.sources.default") }}
            </SelectItem>
          </SelectContent>
        </Select>
      </div>

      <div class="flex gap-2">
        <Input
          v-model="newWord"
          :placeholder="t('settings.stopwords.add.placeholder')"
          class="flex-1 focus-visible:ring-1"
          @keydown.enter="handleAdd"
        />
        <Button :disabled="addLoading || !newWord.trim()" @click="handleAdd">
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
      <Button variant="outline" size="sm" @click="fetchStopwords">
        <RefreshCw class="w-4 h-4 mr-2 pointer-events-none" />
        {{ t("common.retry") }}
      </Button>
    </div>

    <div
      v-else-if="stopwords.length === 0"
      class="text-center py-8 text-sm text-muted-foreground"
    >
      {{ t("settings.stopwords.empty") }}
    </div>

    <div v-else class="flex flex-col divide-y">
      <div
        v-for="word in stopwords"
        :key="word.word"
        class="flex items-center justify-between py-2 gap-3"
      >
        <div class="flex items-center gap-2 min-w-0">
          <Switch
            :model-value="word.isEnabled"
            @update:model-value="handleToggle(word)"
          />
          <span
            class="text-sm truncate"
            :class="{ 'text-muted-foreground line-through': !word.isEnabled }"
          >
            {{ word.word }}
          </span>
        </div>
        <div class="flex items-center gap-2 shrink-0">
          <Badge variant="secondary" class="text-xs">
            {{ t(`settings.stopwords.sources.${word.source}`) }}
          </Badge>
          <Button
            v-if="word.source === 'user'"
            variant="ghost"
            size="icon"
            class="w-6 h-6"
            @click="handleDelete(word.word)"
          >
            <Trash2 class="w-3 h-3 pointer-events-none" />
          </Button>
        </div>
      </div>
    </div>
  </div>
</template>
