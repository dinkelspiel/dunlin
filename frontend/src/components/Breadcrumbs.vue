<template>
  <DropdownMenu v-if="pathSegments.length > 0">
    <DropdownMenuTrigger
      class="flex gap-2 items-center cursor-pointer text-neutral-400 hover:text-neutral-200 transition-colors"
    >
      <span class="text-neutral-400">/</span>
      <span>{{ triggerLabel }}</span>
      <ChevronDown class="size-4" />
    </DropdownMenuTrigger>
    <DropdownMenuContent align="start">
      <DropdownMenuItem
        v-for="(path, index) in pathSegments"
        :key="index"
        class="cursor-pointer"
        :style="{ paddingLeft: `${index * 12 + 8}px` }"
        @click="navigateTo(index)"
      >
        <span class="text-neutral-400 mr-1">/</span>{{ path }}
      </DropdownMenuItem>
    </DropdownMenuContent>
  </DropdownMenu>
</template>

<script setup lang="ts">
import { computed } from 'vue'
import { useRouter } from 'vue-router'
import { ChevronDown } from 'lucide-vue-next'
import {
  DropdownMenu,
  DropdownMenuTrigger,
  DropdownMenuContent,
  DropdownMenuItem,
} from '@/components/ui/dropdown-menu'

const props = defineProps<{
  filepath: string
  teamSlug: string
  projectSlug: string
}>()

const router = useRouter()

const pathSegments = computed(() => props.filepath.split('/').filter((e) => !!e))

const triggerLabel = computed(() => {
  const last = pathSegments.value[pathSegments.value.length - 1]
  return last.length > 22 ? last.slice(0, 22) + '...' : last
})

const getPathLink = (index: number) => {
  const partialPath = pathSegments.value.slice(0, index + 1).join('/')
  return `/-/${props.teamSlug}/${props.projectSlug}/${partialPath}`
}

const navigateTo = (index: number) => {
  router.push(getPathLink(index))
}
</script>
