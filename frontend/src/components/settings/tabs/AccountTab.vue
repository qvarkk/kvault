<script setup lang="ts">
import { onMounted, ref } from "vue"
import { useI18n } from "vue-i18n"
import { useRouter } from "vue-router"
import { useAuthStore } from "@/stores/auth"
import { authService } from "@/services/auth"
import { toast } from "vue-sonner"
import { toastApiError } from "@/services/apiError"
import { timeAgo } from "@/services/format"
import type { ApiKey } from "@/types"
import { Button } from "@/components/ui/button"
import { Input } from "@/components/ui/input"
import { Badge } from "@/components/ui/badge"
import { Separator } from "@/components/ui/separator"
import { Check, Loader2, LogOut, Pencil, X } from "lucide-vue-next"
import {
  Dialog,
  DialogContent,
  DialogHeader,
  DialogTitle,
  DialogFooter,
} from "@/components/ui/dialog"
import DialogDescription from "@/components/ui/dialog/DialogDescription.vue"
import { Switch } from "@/components/ui/switch"
import { useEditorShortcuts } from "@/composables/useEditorShortcuts"

const { t } = useI18n()
const router = useRouter()
const auth = useAuthStore()
const { isHidden } = useEditorShortcuts()

const keys = ref<ApiKey[]>([])
const keysLoading = ref(false)
const logoutOthersLoading = ref(false)
const editingId = ref<string | null>(null)
const editingLabel = ref("")

async function loadKeys() {
  keysLoading.value = true
  try {
    const response = await authService.listKeys()
    // Tolerate both a bare array and a { data: [...] } envelope.
    const payload = response.data as unknown
    keys.value = Array.isArray(payload)
      ? (payload as ApiKey[])
      : ((payload as { data?: ApiKey[] })?.data ?? [])
  } catch (e) {
    toastApiError(e)
  } finally {
    keysLoading.value = false
  }
}

function startRename(key: ApiKey) {
  editingId.value = key.id
  editingLabel.value = key.label
}

function cancelRename() {
  editingId.value = null
  editingLabel.value = ""
}

async function saveRename(key: ApiKey) {
  const label = editingLabel.value.trim()
  if (!label || label === key.label) {
    cancelRename()
    return
  }
  try {
    await authService.renameKey(key.id, label)
    key.label = label
    toast.success(t("settings.account.sessions.renamed"))
  } catch (e) {
    toastApiError(e)
  } finally {
    cancelRename()
  }
}

async function logoutKey(key: ApiKey) {
  if (key.current) {
    await auth.logout()
    router.push({ name: "login" })
    return
  }
  try {
    await authService.deleteKey(key.id)
    keys.value = keys.value.filter((k) => k.id !== key.id)
    toast.success(t("settings.account.sessions.loggedOut"))
  } catch (e) {
    toastApiError(e)
  }
}

async function handleLogoutOthers() {
  logoutOthersLoading.value = true
  try {
    await authService.logoutOthers()
    keys.value = keys.value.filter((k) => k.current)
    toast.success(t("settings.account.sessions.loggedOutOthers"))
  } catch (e) {
    toastApiError(e)
  } finally {
    logoutOthersLoading.value = false
  }
}

onMounted(loadKeys)

const changePasswordOpen = ref(false)
const oldPassword = ref("")
const newPassword = ref("")
const changePasswordLoading = ref(false)

async function handleChangePassword() {
  changePasswordLoading.value = true
  try {
    await authService.updatePassword(oldPassword.value, newPassword.value)
    toast.success(t("settings.account.passwordChanged"))
    changePasswordOpen.value = false
    oldPassword.value = ""
    newPassword.value = ""
  } catch (e) {
    toastApiError(e)
  } finally {
    changePasswordLoading.value = false
  }
}

const deleteOpen = ref(false)
const deleteUnlocked = ref(false)
const deletePassword = ref("")
const deleteLoading = ref(false)

async function handleDeleteAccount() {
  deleteLoading.value = true
  try {
    await authService.deleteAccount(deletePassword.value)
    await auth.logout()
    router.push({ name: "signUp" })
  } catch (e) {
    toastApiError(e)
  } finally {
    deleteLoading.value = false
  }
}

function openDeleteModal() {
  deleteUnlocked.value = false
  deletePassword.value = ""
  deleteOpen.value = true
}
</script>

<template>
  <div class="flex flex-col gap-6">
    <DialogDescription class="text-center hidden">
      {{ t("settings.account.description") }}
    </DialogDescription>

    <div class="flex flex-col gap-1">
      <p class="text-sm font-medium">{{ t("settings.account.username") }}</p>
      <p class="text-xs text-muted-foreground">{{ auth.user?.username }}</p>
    </div>

    <Separator />

    <div class="flex flex-col items-start gap-2">
      <p class="text-sm font-medium mb-2">
        {{ t("settings.account.editor.title") }}
      </p>
      <div class="flex w-full items-center justify-between gap-2">
        <p class="text-xs text-muted-foreground">
          {{ t("settings.account.editor.hideShortcuts") }}
        </p>
        <Switch v-model:model-value="isHidden" />
      </div>
    </div>

    <Separator />

    <div class="flex flex-col gap-3">
      <div class="flex items-start justify-between gap-4">
        <div class="flex flex-col gap-1">
          <p class="text-sm font-medium">
            {{ t("settings.account.sessions.title") }}
          </p>
          <p class="text-xs text-muted-foreground">
            {{ t("settings.account.sessions.description") }}
          </p>
        </div>
        <Button
          variant="outline"
          size="sm"
          class="shrink-0"
          :disabled="logoutOthersLoading || keys.length < 2"
          @click="handleLogoutOthers"
        >
          <Loader2
            v-if="logoutOthersLoading"
            class="w-4 h-4 mr-2 animate-spin pointer-events-none"
          />
          {{ t("settings.account.sessions.logoutOthers") }}
        </Button>
      </div>

      <div v-if="keysLoading" class="flex justify-center py-4">
        <Loader2 class="w-4 h-4 animate-spin text-muted-foreground" />
      </div>
      <p
        v-else-if="keys.length === 0"
        class="text-xs text-muted-foreground py-2"
      >
        {{ t("settings.account.sessions.empty") }}
      </p>
      <div v-else class="flex flex-col divide-y rounded-md border">
        <div
          v-for="key in keys"
          :key="key.id"
          class="flex items-center gap-2 p-3"
        >
          <div class="flex flex-col gap-0.5 min-w-0 flex-1">
            <div v-if="editingId === key.id" class="flex items-center gap-1">
              <Input
                v-model="editingLabel"
                class="h-7 text-sm"
                @keydown.enter="saveRename(key)"
                @keydown.esc="cancelRename"
              />
              <Button
                variant="ghost"
                size="icon"
                class="w-7 h-7 shrink-0"
                @click="saveRename(key)"
              >
                <Check class="w-4 h-4 pointer-events-none" />
              </Button>
              <Button
                variant="ghost"
                size="icon"
                class="w-7 h-7 shrink-0"
                @click="cancelRename"
              >
                <X class="w-4 h-4 pointer-events-none" />
              </Button>
            </div>
            <div v-else class="flex items-center gap-2 min-w-0">
              <span class="text-sm truncate">{{ key.label }}</span>
              <Badge v-if="key.current" variant="secondary" class="shrink-0">
                {{ t("settings.account.sessions.current") }}
              </Badge>
              <button
                class="shrink-0 text-muted-foreground hover:text-foreground"
                @click="startRename(key)"
              >
                <Pencil class="w-3.5 h-3.5" />
              </button>
            </div>
            <span class="text-xs text-muted-foreground">
              {{ t("settings.account.sessions.lastUsed") }}:
              {{ timeAgo(key.lastUsedAt) }}
            </span>
          </div>
          <Button
            variant="ghost"
            size="icon"
            class="w-8 h-8 shrink-0 text-destructive"
            :title="
              key.current
                ? t('settings.account.sessions.logoutThis')
                : t('settings.account.sessions.logout')
            "
            @click="logoutKey(key)"
          >
            <LogOut class="w-4 h-4 pointer-events-none" />
          </Button>
        </div>
      </div>
    </div>

    <Separator />

    <div class="flex items-start justify-between gap-4">
      <div class="flex flex-col gap-1">
        <p class="text-sm font-medium">
          {{ t("settings.account.password.title") }}
        </p>
        <p class="text-xs text-muted-foreground">
          {{ t("settings.account.password.description") }}
        </p>
      </div>
      <Button
        variant="outline"
        class="shrink-0 min-w-32"
        @click="changePasswordOpen = true"
      >
        {{ t("settings.account.password.action") }}
      </Button>
    </div>

    <Separator />

    <div class="flex items-start justify-between gap-4">
      <div class="flex flex-col gap-1">
        <p class="text-sm font-medium text-destructive">
          {{ t("settings.account.delete.title") }}
        </p>
        <p class="text-xs text-muted-foreground">
          {{ t("settings.account.delete.description") }}
        </p>
      </div>
      <Button
        variant="destructive"
        class="shrink-0 min-w-32"
        @click="openDeleteModal"
      >
        {{ t("settings.account.delete.action") }}
      </Button>
    </div>
  </div>

  <Dialog v-model:open="changePasswordOpen">
    <DialogContent aria-describedby="">
      <DialogHeader>
        <DialogTitle>{{ t("settings.account.password.title") }}</DialogTitle>
      </DialogHeader>
      <div class="flex flex-col gap-3">
        <Input
          v-model="oldPassword"
          type="password"
          :placeholder="t('settings.account.password.old')"
        />
        <Input
          v-model="newPassword"
          type="password"
          :placeholder="t('settings.account.password.new')"
        />
      </div>
      <DialogFooter>
        <Button variant="outline" @click="changePasswordOpen = false">
          {{ t("common.cancel") }}
        </Button>
        <Button
          :disabled="
            changePasswordLoading || !oldPassword.trim() || !newPassword.trim()
          "
          @click="handleChangePassword"
        >
          <Loader2
            v-if="changePasswordLoading"
            class="w-4 h-4 mr-2 animate-spin pointer-events-none"
          />
          {{ t("common.save") }}
        </Button>
      </DialogFooter>
    </DialogContent>
  </Dialog>

  <Dialog v-model:open="deleteOpen">
    <DialogContent aria-describedby="">
      <DialogHeader>
        <DialogTitle class="text-destructive">{{
          t("settings.account.delete.title")
        }}</DialogTitle>
      </DialogHeader>
      <div class="flex flex-col gap-4">
        <p class="text-sm text-muted-foreground">
          {{ t("settings.account.delete.description") }}
        </p>

        <Button
          v-if="!deleteUnlocked"
          variant="destructive"
          @click="deleteUnlocked = true"
        >
          {{ t("settings.account.delete.iUnderstand") }}
        </Button>

        <div v-else class="flex flex-col gap-3">
          <Input
            v-model="deletePassword"
            type="password"
            :placeholder="t('settings.account.delete.password')"
          />
          <Button
            variant="destructive"
            :disabled="deleteLoading || !deletePassword.trim()"
            @click="handleDeleteAccount"
          >
            <Loader2
              v-if="deleteLoading"
              class="w-4 h-4 mr-2 animate-spin pointer-events-none"
            />
            {{ t("settings.account.delete.forever") }}
          </Button>
        </div>
      </div>
    </DialogContent>
  </Dialog>
</template>
