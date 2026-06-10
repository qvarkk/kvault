<script setup lang="ts">
import { ref } from "vue"
import {
  AlertDialog,
  AlertDialogAction,
  AlertDialogCancel,
  AlertDialogContent,
  AlertDialogDescription,
  AlertDialogFooter,
  AlertDialogHeader,
  AlertDialogTitle,
} from "../ui/alert-dialog"
import { useI18n } from "vue-i18n"
import { notesService } from "@/services/notes"
import { toast } from "vue-sonner"
import { toastApiError } from "@/services/apiError"
import type { ApiError } from "@/types"

const props = defineProps<{
  noteId: string
}>()

const emit = defineEmits<{
  deleted: [id: string]
}>()

const { t } = useI18n()

const open = defineModel<boolean>("open", { default: false })
const deleteLoading = ref(false)

async function confirmDelete() {
  deleteLoading.value = true
  try {
    await notesService.remove(props.noteId)
    emit("deleted", props.noteId)
    toast.success(t("notes.delete.deleted"))
    open.value = false
  } catch (e) {
    toastApiError(e)
  } finally {
    deleteLoading.value = false
  }
}
</script>

<template>
  <AlertDialog v-model:open="open">
    <AlertDialogContent @click.stop>
      <AlertDialogHeader>
        <AlertDialogTitle>{{ t("notes.delete.title") }}</AlertDialogTitle>
        <AlertDialogDescription>
          {{ t("notes.delete.description") }}
        </AlertDialogDescription>
      </AlertDialogHeader>
      <AlertDialogFooter>
        <AlertDialogCancel>{{ t("common.cancel") }}</AlertDialogCancel>
        <AlertDialogAction
          @click="confirmDelete"
          :disabled="deleteLoading"
          class="bg-destructive text-destructive-foreground hover:bg-destructive/90"
        >
          {{ t("common.delete") }}
        </AlertDialogAction>
      </AlertDialogFooter>
    </AlertDialogContent>
  </AlertDialog>
</template>
