<script setup lang="ts">
import { computed } from "vue"
import { SearchIcon, Upload } from "lucide-vue-next"
import { useI18n } from "vue-i18n"
import SortControl from "../common/SortControl.vue"
import { InputGroup, InputGroupAddon, InputGroupInput } from "../ui/input-group"
import { Spinner } from "../ui/spinner"
import { Button } from "../ui/button"

const { t } = useI18n()

const query = defineModel<string>("query", { default: "" })
const sortBy = defineModel<string>("sortBy", { default: "" })
const orderBy = defineModel<string>("orderBy", { default: "" })
const loading = defineModel<boolean>("loading", { default: false })

defineProps<{
  trash?: boolean
}>()

const emits = defineEmits<{
  uploadClick: []
}>()

const sortOptions = computed(() => [
  { value: "created_at", label: t("files.sort.createdAt") },
  { value: "original_name", label: t("files.sort.name") },
  { value: "size", label: t("files.sort.size") },
])
</script>

<template>
  <div class="flex flex-col gap-2">
    <InputGroup>
      <InputGroupInput v-model="query" :placeholder="t('files.search')" />
      <InputGroupAddon>
        <SearchIcon />
      </InputGroupAddon>
      <InputGroupAddon v-if="loading" align="inline-end">
        <Spinner />
      </InputGroupAddon>
    </InputGroup>

    <div class="flex items-center gap-2">
      <SortControl
        v-model:sort-by="sortBy"
        v-model:order-by="orderBy"
        :options="sortOptions"
      />
      <Button v-if="!trash" @click="emits('uploadClick')">
        <Upload class="w-4 h-4 mr-2 pointer-events-none" />
        {{ t("common.upload") }}
      </Button>
    </div>
  </div>
</template>
