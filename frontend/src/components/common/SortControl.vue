<script setup lang="ts">
import { computed } from "vue"
import { ArrowDown, ArrowUp } from "lucide-vue-next"
import { Button } from "../ui/button"
import { ButtonGroup } from "../ui/button-group"
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from "../ui/select"

const props = defineProps<{
  options: { value: string; label: string }[]
}>()

const sortBy = defineModel<string>("sortBy", { default: "" })
const orderBy = defineModel<string>("orderBy", { default: "" })

// SelectValue caches the selected item's text at selection time; the content
// is teleported and unmounted when closed, so it can't re-read a re-translated
// label on locale change. Drive the trigger label explicitly from `options`
// (a computed in the parent) so it stays reactive.
const selectedLabel = computed(
  () => props.options.find((o) => o.value === sortBy.value)?.label ?? "",
)

function toggleOrder() {
  orderBy.value = orderBy.value === "DESC" ? "ASC" : "DESC"
}
</script>

<template>
  <ButtonGroup class="flex flex-1">
    <Button variant="outline" size="icon" @click="toggleOrder">
      <ArrowUp
        v-if="orderBy === 'ASC'"
        class="w-4 h-4 pointer-events-none"
      />
      <ArrowDown v-else class="w-4 h-4 pointer-events-none" />
    </Button>
    <Select v-model:model-value="sortBy">
      <SelectTrigger class="flex-1">
        <SelectValue>{{ selectedLabel }}</SelectValue>
      </SelectTrigger>
      <SelectContent>
        <SelectItem
          v-for="opt in options"
          :key="opt.value"
          :value="opt.value"
        >
          {{ opt.label }}
        </SelectItem>
      </SelectContent>
    </Select>
  </ButtonGroup>
</template>
