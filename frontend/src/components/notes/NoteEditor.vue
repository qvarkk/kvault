<script setup lang="ts">
import { ExportPDF, PreviewThemeSwitch, Emoji } from "@vavt/v3-extension"
import "@vavt/v3-extension/lib/asset/ExportPDF.css"
import "@vavt/v3-extension/lib/asset/PreviewThemeSwitch.css"
import "@vavt/v3-extension/lib/asset/Emoji.css"

import { useMediaQuery } from "@vueuse/core"
import {
  MdEditor,
  type PreviewThemes,
  type Themes,
  type ToolbarNames,
} from "md-editor-v3"
import { computed, ref } from "vue"
import { useI18n } from "vue-i18n"
import { toast } from "vue-sonner"
import { Smile } from "lucide-vue-next"
import i18n from "@/i18n"

const props = defineProps<{
  theme: Themes
}>()

const content = defineModel<string>("content")

const { t } = useI18n()
const isMobile = useMediaQuery("(max-width: 768px)")
const isDesktop = useMediaQuery("(min-width: 1280px)")

const mobileToolbars: ToolbarNames[] = [
  "bold",
  "italic",
  "-",
  "unorderedList",
  "task",
  "-",
  "revoke",
  "next",
  "=",
  "pageFullscreen",
  "previewOnly",
  1,
]

const mediumToolbars: ToolbarNames[] = [
  "bold",
  "italic",
  "-",
  "title",
  "unorderedList",
  "orderedList",
  "task",
  "-",
  "code",
  "-",
  "revoke",
  "next",
  "=",
  "pageFullscreen",
  "preview",
  "-",
  0,
  1,
]

const desktopToolbars: ToolbarNames[] = [
  "bold",
  "underline",
  "italic",
  "strikeThrough",
  "-",
  "title",
  "sub",
  "sup",
  "quote",
  "unorderedList",
  "orderedList",
  "task",
  "-",
  "codeRow",
  "code",
  "link",
  "image",
  "table",
  "mermaid",
  "katex",
  "-",
  "revoke",
  "next",
  "=",
  "pageFullscreen",
  "preview",
  "catalog",
  "-",
  0,
  1,
]

const toolbars = computed(() => {
  if (isMobile.value) return mobileToolbars
  if (isDesktop.value) return desktopToolbars
  return mediumToolbars
})

const getEditorLanguage = computed(() => {
  const currentLocale = i18n.global.locale.value

  if (currentLocale === "en") return "en_US"
  return currentLocale
})

const previewTheme = ref<PreviewThemes>("default")

function handlePdfExportSuccess() {
  toast.success(t("notes.editor.exportPdfSuccess"))
}

function handlePdfExportError(err: unknown) {
  toast.error(
    t("notes.editor.exportPdfErrorTitle", {
      description: t("notes.editor.exportPdfErrorDescription"),
    }),
  )
  console.log(err)
}
</script>

<template>
  <MdEditor
    v-model="content"
    :theme="theme"
    :toolbars="toolbars"
    :preview="!isMobile"
    :preview-theme="previewTheme"
    :language="getEditorLanguage"
    code-theme="github"
    no-upload-img
    style="height: 100% !important"
  >
    <template #defToolbars>
      <ExportPDF
        :modelValue="content"
        :title="t('notes.editor.exportPdfTitle')"
        :modal-title="t('notes.editor.exportPdfTitle')"
        :export-btn-text="t('notes.editor.exportPdfButton')"
        @success="handlePdfExportSuccess"
        @error="handlePdfExportError"
      />
      <PreviewThemeSwitch v-model="previewTheme" />
      <Emoji>
        <template #trigger>
          <Smile class="size-4" />
        </template>
      </Emoji>
    </template>
  </MdEditor>
</template>
