<script setup lang="ts">
import { useI18n } from "vue-i18n"
import {
  Dialog,
  DialogContent,
  DialogHeader,
  DialogTitle,
} from "@/components/ui/dialog"
import { Tabs, TabsContent, TabsList, TabsTrigger } from "@/components/ui/tabs"
import AccountTab from "./tabs/AccountTab.vue"
import StopwordsTab from "./tabs/StopwordsTab.vue"
import TagsTab from "./tabs/TagsTab.vue"

const { t } = useI18n()
const open = defineModel<boolean>("open")
</script>

<template>
  <Dialog v-model:open="open">
    <DialogContent
      class="w-[calc(100vw-1rem)] max-w-2xl max-h-[80vh] flex flex-col"
    >
      <DialogHeader>
        <DialogTitle>{{ t("settings.title") }}</DialogTitle>
      </DialogHeader>
      <Tabs default-value="account">
        <TabsList class="w-full overflow-x-auto">
          <TabsTrigger value="account">
            {{ t("settings.tabs.account") }}
          </TabsTrigger>
          <TabsTrigger value="stopwords">
            {{ t("settings.tabs.stopwords") }}
          </TabsTrigger>
          <TabsTrigger value="tags">
            {{ t("settings.tabs.tags") }}
          </TabsTrigger>
        </TabsList>
        <div
          class="flex-1 overflow-y-auto overflow-x-hidden mt-4 max-h-[50vh] min-h-[50vh]"
        >
          <TabsContent value="account">
            <AccountTab />
          </TabsContent>
          <TabsContent value="stopwords">
            <StopwordsTab />
          </TabsContent>
          <TabsContent value="tags">
            <TagsTab />
          </TabsContent>
        </div>
      </Tabs>
    </DialogContent>
  </Dialog>
</template>
