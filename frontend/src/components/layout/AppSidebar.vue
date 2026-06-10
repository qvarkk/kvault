<script setup lang="ts">
import {
  ChevronsUpDown,
  Files,
  Home,
  Info,
  LogOut,
  NotebookPen,
  PanelLeft,
  Search,
  Settings,
  X,
} from "lucide-vue-next"
import {
  Sidebar,
  SidebarContent,
  SidebarFooter,
  SidebarGroup,
  SidebarGroupContent,
  SidebarHeader,
  SidebarMenu,
  SidebarMenuButton,
  SidebarMenuItem,
  useSidebar,
} from "../ui/sidebar"
import { useI18n } from "vue-i18n"
import { useRouter } from "vue-router"
import { useAuthStore } from "@/stores"
import {
  DropdownMenu,
  DropdownMenuContent,
  DropdownMenuItem,
  DropdownMenuLabel,
  DropdownMenuSeparator,
  DropdownMenuTrigger,
} from "../ui/dropdown-menu"
import { Avatar, AvatarFallback } from "../ui/avatar"
import { computed, ref, type FunctionalComponent } from "vue"
import { onKeyStroke } from "@vueuse/core"
import SettingsModal from "../settings/SettingsModal.vue"
import SearchPalette from "../search/SearchPalette.vue"

interface sidebarItem {
  name: string
  routeName: string
  iconComponent: FunctionalComponent
}

const { t } = useI18n()
const router = useRouter()
const auth = useAuthStore()
const { state, isMobile, openMobile, toggleSidebar } = useSidebar()

const settingsOpen = ref(false)
const searchOpen = ref(false)

function openSearch() {
  if (isMobile.value) toggleSidebar()
  searchOpen.value = true
}

onKeyStroke(["k", "K"], (e) => {
  if (e.metaKey || e.ctrlKey) {
    e.preventDefault()
    searchOpen.value = true
  }
})

const isExpanded = computed(() =>
  isMobile.value ? openMobile.value : state.value === "expanded",
)

// Computed so labels re-translate when the locale changes; a plain array
// would freeze the strings at setup time.
const sidebarItems = computed<sidebarItem[]>(() => [
  {
    name: t("nav.home"),
    routeName: "home",
    iconComponent: Home,
  },
  {
    name: t("nav.notes"),
    routeName: "notes",
    iconComponent: NotebookPen,
  },
  {
    name: t("nav.files"),
    routeName: "files",
    iconComponent: Files,
  },
  {
    name: t("nav.about"),
    routeName: "about",
    iconComponent: Info,
  },
])

async function logout() {
  await auth.logout()
  router.push({ name: "login" })
}

function openSettings() {
  if (isMobile.value) toggleSidebar()
  settingsOpen.value = true
}
</script>

<template>
  <Sidebar :collapsible="isMobile ? 'offcanvas' : 'icon'">
    <SidebarHeader>
      <SidebarMenu>
        <SidebarMenuItem class="flex flex-row items-center justify-between">
          <span
            v-if="isExpanded"
            class="text-lg font-bold tracking-tight pl-2 cursor-default"
          >
            [kv]
          </span>
          <SidebarMenuButton class="self-end w-auto" @click="toggleSidebar">
            <X v-if="isMobile" />
            <PanelLeft v-else class="pointer-events-none" />
          </SidebarMenuButton>
        </SidebarMenuItem>
      </SidebarMenu>
    </SidebarHeader>
    <SidebarContent>
      <SidebarGroup>
        <SidebarGroupContent>
          <SidebarMenu>
            <SidebarMenuItem>
              <SidebarMenuButton
                @click="openSearch"
                :tooltip="t('nav.search')"
              >
                <Search class="pointer-events-none" />
                <span class="pointer-events-none">{{ t("nav.search") }}</span>
              </SidebarMenuButton>
            </SidebarMenuItem>
            <SidebarMenuItem v-for="item in sidebarItems">
              <SidebarMenuButton
                @click="router.push({ name: item.routeName })"
                :tooltip="item.name"
              >
                <component
                  :is="item.iconComponent"
                  class="pointer-events-none"
                />
                <span class="pointer-events-none">{{ item.name }}</span>
              </SidebarMenuButton>
            </SidebarMenuItem>
          </SidebarMenu>
        </SidebarGroupContent>
      </SidebarGroup>
    </SidebarContent>
    <SidebarFooter>
      <SidebarMenu>
        <SidebarMenuItem>
          <DropdownMenu>
            <!-- 
              The only solution I came up with to keep the avatar
              from shifting to the right in collapsed state.
              Not clean, but gets the job done.
            -->
            <DropdownMenuTrigger
              as-child
              :style="!isExpanded ? 'padding: 4px !important' : ''"
            >
              <SidebarMenuButton :tooltip="t('nav.account')" class="py-5">
                <Avatar class="w-6 h-6 pointer-events-none">
                  <AvatarFallback
                    class="bg-primary text-primary-foreground text-xs"
                  >
                    {{ auth.user?.username?.slice(0, 2).toUpperCase() }}
                  </AvatarFallback>
                </Avatar>
                <span v-if="isExpanded" class="truncate pointer-events-none">
                  {{ auth.user?.username }}
                </span>
                <ChevronsUpDown
                  v-if="isExpanded"
                  class="ml-auto pointer-events-none"
                />
              </SidebarMenuButton>
            </DropdownMenuTrigger>
            <DropdownMenuContent :class="!isMobile ? 'ml-2' : ''" side="top">
              <DropdownMenuLabel class="flex flex-col">
                <span class="truncate">{{ auth.user?.username }}</span>
                <span class="text-xs text-muted-foreground truncate">
                  ID: {{ auth.user?.id }}
                </span>
              </DropdownMenuLabel>
              <DropdownMenuSeparator />
              <DropdownMenuItem @click="openSettings">
                <Settings class="pointer-events-none" />
                <span class="pointer-events-none">{{ t("nav.settings") }}</span>
              </DropdownMenuItem>
              <DropdownMenuSeparator />
              <DropdownMenuItem @click="logout" class="text-destructive">
                <LogOut class="pointer-events-none" />
                <span class="pointer-events-none">{{ t("nav.logout") }}</span>
              </DropdownMenuItem>
            </DropdownMenuContent>
          </DropdownMenu>
        </SidebarMenuItem>
      </SidebarMenu>
    </SidebarFooter>
  </Sidebar>
  <SettingsModal v-model:open="settingsOpen" />
  <SearchPalette v-model:open="searchOpen" />
</template>
