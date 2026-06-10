import { ref, watchEffect } from "vue"

const LOCAL_STORAGE_KEY = "hideEditorShortcuts"

const storedValue = localStorage.getItem(LOCAL_STORAGE_KEY)

const isHidden = ref<boolean>((storedValue ?? "false") === "true")

watchEffect(() => {
  localStorage.setItem(LOCAL_STORAGE_KEY, String(isHidden.value))
})

export function useEditorShortcuts() {
  const hide = () => {
    isHidden.value = true
  }
  const show = () => {
    isHidden.value = false
  }
  const toggle = () => {
    isHidden.value = !isHidden.value
  }

  return {
    isHidden,
    hide,
    show,
    toggle,
  }
}
