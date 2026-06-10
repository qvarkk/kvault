<script setup lang="ts">
import AuthForm from "@/components/auth/AuthForm.vue"
import AuthLayout from "@/components/layout/AuthLayout.vue"
import { useHead } from "@vueuse/head"
import { useI18n } from "vue-i18n"
const { t } = useI18n()
useHead({ title: t("head.login") })
</script>

<template>
  <AuthLayout>
    <AuthForm mode="login" />
  </AuthLayout>
</template>
