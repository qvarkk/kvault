<script setup lang="ts">
import { Card } from "../ui/card"
import { Skeleton } from "../ui/skeleton"
</script>

<template>
  <Card class="p-4 flex flex-col gap-3">
    <Skeleton class="h-5 w-3/4" />
    <Skeleton class="h-3 w-full" />
    <Skeleton class="h-3 w-2/3" />
    <div class="flex gap-2 mt-auto">
      <Skeleton class="h-5 w-12 rounded-full" />
      <Skeleton class="h-5 w-16 rounded-full" />
    </div>
    <Skeleton class="h-3 w-1/4" />
  </Card>
</template>
