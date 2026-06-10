<script setup lang="ts">
import AppLayout from "@/components/layout/AppLayout.vue"
import NoteCard from "@/components/notes/NoteCard.vue"
import NoteSkeleton from "@/components/notes/NoteSkeleton.vue"
import { Button } from "@/components/ui/button"
import { useEntities } from "@/composables/useEntities"
import { notesService } from "@/services/notes"
import type { Note } from "@/types"
import { useHead } from "@vueuse/head"
import { NotebookPen, RefreshCw } from "lucide-vue-next"
import { onMounted } from "vue"
import { useI18n } from "vue-i18n"
import { useRouter } from "vue-router"

const { t } = useI18n()
const router = useRouter()
useHead({ title: t("head.home") })

const {
  entities: notes,
  loading: notesLoading,
  error: notesError,
  fetchEntities: fetchNotes,
} = useEntities<Note>(notesService)

function loadNotes() {
  fetchNotes({ sort_by: "updated_at", order_by: "DESC", page_size: 9 })
}

onMounted(loadNotes)
</script>

<template>
  <AppLayout>
    <div
      v-if="notesLoading"
      class="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-4 container"
    >
      <NoteSkeleton v-for="n in 9" :key="n" />
    </div>

    <div
      v-else-if="notesError"
      class="flex flex-col items-center justify-center gap-4 pt-24 pb-8 text-muted-foreground container"
    >
      <p class="text-destructive">{{ notesError.detail }}</p>
      <Button variant="outline" @click="loadNotes">
        <RefreshCw class="w-4 h-4 mr-2" />
        {{ t("common.retry") }}
      </Button>
    </div>

    <div v-else class="flex flex-col gap-6 container">
      <div>
        <h1 class="text-2xl font-bold">{{ t("home.greeting") }}</h1>
        <p class="text-muted-foreground">{{ t("home.subtitle") }}</p>
      </div>

      <div
        v-if="notes.length === 0"
        class="flex flex-col items-center justify-center gap-4 pt-24 pb-8 text-muted-foreground"
      >
        <NotebookPen class="w-12 h-12" />
        <p>{{ t("notes.empty") }}</p>
        <Button @click="router.push({ name: 'notes' })">
          {{ t("notes.create.first") }}
        </Button>
      </div>

      <div v-else class="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-4">
        <NoteCard
          v-for="note in notes"
          :key="note.id"
          :note="note"
          :include-menu="false"
        />
      </div>

      <div v-if="notes.length > 0" class="flex justify-center">
        <Button variant="ghost" @click="router.push({ name: 'notes' })">
          {{ t("home.viewAll") }}
        </Button>
      </div>
    </div>
  </AppLayout>
</template>
