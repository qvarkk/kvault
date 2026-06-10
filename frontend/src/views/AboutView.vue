<script setup lang="ts">
import AppLayout from "@/components/layout/AppLayout.vue"
import { useI18n } from "vue-i18n"
import { Separator } from "@/components/ui/separator"
import { Badge } from "@/components/ui/badge"
import { AlertCircleIcon, ExternalLink } from "lucide-vue-next"
import { Alert, AlertDescription, AlertTitle } from "@/components/ui/alert"
import { useHead } from "@vueuse/head"

const { t } = useI18n()
useHead({ title: t("head.about") })

const stack = {
  frontend: [
    "Vue 3",
    "Vite",
    "TypeScript",
    "Tailwind CSS",
    "shadcn-vue",
    "Pinia",
  ],
  backend: ["Go", "PostgreSQL", "Redis", "S3"],
}

const repos = [
  {
    name: "qvarkk/kvault-frontend",
    description: "Vue 3 + Vite + TypeScript",
    url: "https://gitverse.ru/qvarkk/kvault-frontend",
  },
  {
    name: "qvarkk/kvault",
    description: "Go + PostgreSQL + Redis",
    url: "https://gitverse.ru/qvarkk/kvault",
  },
]
</script>

<template>
  <AppLayout>
    <div class="py-10 flex flex-col gap-8 container max-w-2xl">
      <div>
        <h1 class="text-3xl font-bold tracking-tight">[kvault]</h1>
        <p class="text-sm text-muted-foreground mt-1 uppercase tracking-widest">
          {{ t("about.tagline") }}
        </p>
      </div>

      <Separator />

      <div class="flex flex-col gap-2">
        <p class="text-xs uppercase tracking-widest text-muted-foreground">
          {{ t("about.what.title") }}
        </p>
        <p class="text-sm text-muted-foreground leading-relaxed">
          {{ t("about.what.description") }}
        </p>
      </div>

      <Separator />

      <div class="flex flex-col gap-2">
        <p class="text-xs uppercase tracking-widest text-muted-foreground">
          {{ t("about.details.title") }}
        </p>
        <div class="grid grid-cols-2 md:grid-cols-4 gap-3">
          <div class="bg-secondary rounded-md p-3 flex flex-col gap-1">
            <p class="text-xs text-muted-foreground">
              {{ t("about.details.version") }}
            </p>
            <p class="text-sm font-medium">v0.3.1</p>
          </div>
          <div class="bg-secondary rounded-md p-3 flex flex-col gap-1">
            <p class="text-xs text-muted-foreground">
              {{ t("about.details.license") }}
            </p>
            <p class="text-sm font-medium">MIT</p>
          </div>
          <div class="bg-secondary rounded-md p-3 flex flex-col gap-1">
            <p class="text-xs text-muted-foreground">
              {{ t("about.details.deployment") }}
            </p>
            <p class="text-sm font-medium">Docker</p>
          </div>
          <div class="bg-secondary rounded-md p-3 flex flex-col gap-1">
            <p class="text-xs text-muted-foreground">
              {{ t("about.details.storage") }}
            </p>
            <p class="text-sm font-medium">S3 compatible</p>
          </div>
        </div>
      </div>

      <Separator />

      <div class="flex flex-col gap-2">
        <p class="text-xs uppercase tracking-widest text-muted-foreground">
          {{ t("about.stack") }}
        </p>
        <div class="flex flex-wrap gap-2">
          <Badge
            v-for="tech in [...stack.frontend, ...stack.backend]"
            :key="tech"
            variant="secondary"
          >
            {{ tech }}
          </Badge>
        </div>
      </div>

      <Separator />

      <div class="flex flex-col gap-2">
        <p class="text-xs uppercase tracking-widest text-muted-foreground">
          {{ t("about.repositories") }}
        </p>
        <div class="flex flex-col gap-2">
          <a
            v-for="repo in repos"
            :key="repo.name"
            :href="repo.url"
            target="_blank"
            class="flex items-center justify-between p-3 rounded-md border hover:bg-secondary transition-colors"
          >
            <div class="flex flex-col gap-0.5 pointer-events-none">
              <p class="text-sm font-medium">{{ repo.name }}</p>
              <p class="text-xs text-muted-foreground">
                {{ repo.description }}
              </p>
            </div>
            <ExternalLink
              class="w-4 h-4 text-muted-foreground pointer-events-none"
            />
          </a>
        </div>
      </div>

      <Separator />

      <div class="flex flex-col gap-2">
        <p class="text-xs uppercase tracking-widest text-muted-foreground">
          {{ t("about.selfHosting.title") }}
        </p>
        <p class="text-sm text-muted-foreground leading-relaxed">
          {{ t("about.selfHosting.description") }}
          <code class="text-xs bg-secondary px-1 py-0.5 rounded">.env</code>
          {{ t("about.selfHosting.description2") }}
          <code class="text-xs bg-secondary px-1 py-0.5 rounded">
            docker compose up
          </code>
        </p>
      </div>

      <Separator />

      <Alert variant="destructive">
        <AlertCircleIcon />
        <AlertTitle>{{ t("about.alpha.title") }}</AlertTitle>
        <AlertDescription>
          <p>
            {{ t("about.alpha.warning") }}
          </p>
        </AlertDescription>
      </Alert>
    </div>
  </AppLayout>
</template>
