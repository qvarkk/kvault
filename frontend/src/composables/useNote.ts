import { ref } from "vue"
import { notesService } from "@/services/notes"
import type {
  Note,
  NoteUpdatePayload,
  BindTagPayload,
  ApiError,
  Tag,
} from "@/types"

export function useNote() {
  const note = ref<Note | null>(null)
  const loading = ref(false)
  const saving = ref(false)
  const error = ref<ApiError | null>(null)

  async function fetchNote(id: string) {
    loading.value = true
    error.value = null
    try {
      const response = await notesService.getById(id)
      note.value = response.data
    } catch (e) {
      error.value = e as ApiError
    } finally {
      loading.value = false
    }
  }

  async function updateNote(id: string, payload: NoteUpdatePayload) {
    saving.value = true
    try {
      const response = await notesService.update(id, payload)
      if (note.value !== null) {
        note.value.title = response.data.title
        note.value.content = response.data.content
      }
    } finally {
      saving.value = false
    }
  }

  async function bindTag(itemId: string, tag: Tag) {
    await notesService.bindTag(itemId, { tag_id: tag.id })
    if (note.value) {
      note.value = {
        ...note.value,
        tags: [...note.value.tags, tag],
      }
    }
  }

  async function unbindTag(itemId: string, tagId: string) {
    await notesService.unbindTag(itemId, tagId)
    if (note.value) {
      note.value.tags = note.value.tags.filter((t) => t.id !== tagId)
    }
  }

  return {
    note,
    loading,
    saving,
    error,
    fetchNote,
    updateNote,
    bindTag,
    unbindTag,
  }
}
