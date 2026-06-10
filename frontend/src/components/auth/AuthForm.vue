<script setup lang="ts">
import { useI18n } from "vue-i18n"
import {
  FormField,
  FormItem,
  FormLabel,
  FormControl,
  FormMessage,
} from "@/components/ui/form"
import { Button } from "@/components/ui/button"
import { Input } from "@/components/ui/input"
import { toastApiError } from "@/services/apiError"
import { useForm } from "vee-validate"
import { toTypedSchema } from "@vee-validate/zod"
import { z } from "zod"
import {
  Card,
  CardAction,
  CardContent,
  CardDescription,
  CardFooter,
  CardHeader,
  CardTitle,
} from "@/components/ui/card"
import { useRouter } from "vue-router"
import { useAuthStore } from "@/stores"

const { t } = useI18n()
const router = useRouter()
const auth = useAuthStore()
const props = defineProps<{
  mode: "login" | "signUp"
}>()
const baseI18n = `auth.${props.mode}`

const usernameSchema =
  props.mode === "login"
    ? z.string().max(255, { message: t("validation.username.max", { max: 255 }) })
    : z
        .string()
        .min(3, { message: t("validation.username.min", { min: 3 }) })
        .max(255, { message: t("validation.username.max", { max: 255 }) })

const formSchema = toTypedSchema(
  z.object({
    username: usernameSchema,
    password: z.string().min(8, {
      message: t("validation.password.min", { min: 8 }),
    }),
  }),
)

const { handleSubmit, resetForm } = useForm({
  validationSchema: formSchema,
  initialValues: {
    username: "",
    password: "",
  },
})

const onSubmit = handleSubmit(async (data) => {
  try {
    if (props.mode === "login") {
      await auth.login(data)
    } else {
      await auth.signup(data)
    }
    router.push({ name: "home" })
  } catch (error) {
    toastApiError(error)
  }
})

function switchMode() {
  router.push({ name: props.mode === "login" ? "signUp" : "login" })
}
</script>

<template>
  <Card>
    <CardHeader>
      <CardTitle>
        {{ t(`${baseI18n}.title`) }}
      </CardTitle>
      <CardDescription>
        {{ t(`${baseI18n}.message`) }}
      </CardDescription>
      <CardAction>
        <Button variant="link" @click="switchMode">
          {{
            mode === "login" ? t(`auth.signUp.action`) : t(`auth.login.action`)
          }}
        </Button>
      </CardAction>
    </CardHeader>
    <CardContent>
      <form
        id="login-form"
        @submit.prevent="onSubmit"
        class="flex flex-col gap-4"
      >
        <FormField v-slot="{ componentField }" name="username">
          <FormItem>
            <FormLabel>{{ t("auth.username") }}</FormLabel>
            <FormControl>
              <Input
                type="text"
                :placeholder="t('auth.usernamePlaceholder')"
                v-bind="componentField"
              />
            </FormControl>
            <FormMessage />
          </FormItem>
        </FormField>
        <FormField v-slot="{ componentField }" name="password">
          <FormItem>
            <FormLabel>{{ t("auth.password") }}</FormLabel>
            <FormControl>
              <Input
                type="password"
                placeholder="********"
                v-bind="componentField"
              />
            </FormControl>
            <FormMessage />
          </FormItem>
        </FormField>
      </form>
    </CardContent>
    <CardFooter>
      <Button type="submit" form="login-form" class="w-full">
        {{ t(`${baseI18n}.action`) }}
      </Button>
    </CardFooter>
  </Card>
</template>
