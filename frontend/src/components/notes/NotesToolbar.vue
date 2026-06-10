<script setup lang="ts">
import { computed } from "vue"
import { SearchIcon } from "lucide-vue-next"
import { useI18n } from "vue-i18n"
import TagFilter from "./TagFilter.vue"
import SortControl from "../common/SortControl.vue"
import { InputGroup, InputGroupAddon, InputGroupInput } from "../ui/input-group"
import { Spinner } from "../ui/spinner"

const { t } = useI18n()

const query = defineModel<string>("query", { default: "" })
const sortBy = defineModel<string>("sortBy", { default: "" })
const orderBy = defineModel<string>("orderBy", { default: "" })
const tagIds = defineModel<string[]>("tagIds", { default: [] })
const loading = defineModel<boolean>("loading", { default: false })

const sortOptions = computed(() => [
  { value: "updated_at", label: t("notes.sort.updatedAt") },
  { value: "created_at", label: t("notes.sort.createdAt") },
  { value: "title", label: t("notes.sort.title") },
])
</script>

<template>
  <div class="flex flex-col gap-2">
    <InputGroup>
      <InputGroupInput v-model="query" :placeholder="t('notes.search')" />
      <InputGroupAddon>
        <SearchIcon />
      </InputGroupAddon>
      <InputGroupAddon v-if="loading" align="inline-end">
        <Spinner />
      </InputGroupAddon>
    </InputGroup>

    <div class="flex items-center gap-2 md:gap-4">
      <SortControl
        v-model:sort-by="sortBy"
        v-model:order-by="orderBy"
        :options="sortOptions"
      />

      <div class="flex-1">
        <TagFilter v-model:tag-ids="tagIds" />
      </div>
    </div>
  </div>
</template>
