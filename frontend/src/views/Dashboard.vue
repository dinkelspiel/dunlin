<script setup lang="ts">
import Logo from '@/components/ui/Logo.vue'
import { useAuthUser } from '@/router/auth/AuthUserProvider'
import Button from '@/components/ui/Button.vue'
import {
  AlertCircle,
  ChartArea,
  ChevronDown,
  ChevronLeft,
  ChevronRight,
  File,
  Folder,
  Images,
  Loader2,
  RotateCw,
  Search,
  X,
} from 'lucide-vue-next'
import { Input } from '@/components/ui/input'
import {
  TableHeader,
  Table,
  TableRow,
  TableHead,
  TableBody,
  TableCell,
} from '@/components/ui/table'
import { Alert, AlertTitle, AlertDescription } from '@/components/ui/alert'
import DashboardLayout from '@/components/DashboardLayout.vue'
import { useMutation, useQuery, useQueryClient } from '@tanstack/vue-query'
import type { TeamProjectResponse, TeamResponse } from '@/lib/types'
import { useRoute, useRouter } from 'vue-router'
import { computed, ref, watch } from 'vue'
import { useEventListener } from '@vueuse/core'
import normalize from 'path-normalize'
import TeamsDropdown from '@/components/header/TeamsDropdown.vue'
import TeamProjectsDropdown from '@/components/header/TeamProjectsDropdown.vue'
import {
  Dialog,
  DialogContent,
  DialogFooter,
  DialogHeader,
  DialogTitle,
  DialogTrigger,
} from '@/components/ui/dialog'
import FileUploader from '@/components/FileUploader.vue'
import { humanFileSize } from '@/lib/bytes'
import Breadcrumbs from '@/components/Breadcrumbs.vue'

const { authUser } = useAuthUser()
const route = useRoute()
const router = useRouter()
const teamSlug = computed(() => route.params.team as string)
const projectSlug = computed(() => route.params.project as string)
const apiUrl = import.meta.env.VITE_API_URL

type File = {
  type: 'dir' | 'file'
  name: string
  lastModified: string
  size: number
}

type FilesResponse = {
  message: string
  files: File[]
}

type AlbumImage = {
  path: string
  name: string
  rotationDegrees: number
  lastModified: string
  size: number
}

type AlbumImagesResponse = {
  message: string
  images: AlbumImage[]
}

type RotateAlbumImageResponse = {
  message: string
  path: string
  rotationDegrees: number
}

const { data: team } = useQuery<TeamResponse>({
  queryKey: ['team', teamSlug],
  queryFn: async () => {
    const response = await fetch(`${import.meta.env.VITE_API_URL}/api/v1/teams/${teamSlug.value}`, {
      credentials: 'include',
    })
    if (!response.ok) {
      router.push('/auth/login')
      throw new Error((await response.json()).message)
    }
    return response.json() as Promise<TeamResponse>
  },
})

const { data: teamProject } = useQuery<TeamProjectResponse>({
  queryKey: ['teamProject', teamSlug, projectSlug],
  queryFn: async () => {
    const response = await fetch(
      `${import.meta.env.VITE_API_URL}/api/v1/teams/${teamSlug.value}/projects/${projectSlug.value}`,
      {
        credentials: 'include',
      },
    )
    if (!response.ok) {
      router.push('/auth/login')
      throw new Error((await response.json()).message)
    }
    return response.json() as Promise<TeamProjectResponse>
  },
})

const createFolderOpen = ref(false)
const folderPath = ref('')

const createFolder = useMutation({
  mutationKey: ['createProject'],
  mutationFn: async (path: string) => {
    const response = await fetch(
      `${import.meta.env.VITE_API_URL}/api/v1/teams/${teamSlug.value}/projects/${projectSlug.value}/folders`,
      {
        method: 'POST',
        credentials: 'include',
        body: JSON.stringify({
          path: `${filepathWithSlashes.value}/${path.startsWith('/') ? path.substring(1) : path}`,
        }),
      },
    )
    if (!response.ok) {
      throw new Error((await response.json()).message)
    }
    return response.json()
  },
  onSuccess() {
    queryClient.invalidateQueries({ queryKey: ['files'] })
    createFolderOpen.value = false
    folderPath.value = ''
  },
})

const error = ref('')
const queryClient = useQueryClient()

const rawFilepath = computed(() =>
  normalize(Array.isArray(route.params.filepath) ? route.params.filepath.join('/') : ''),
)

const filepathWithSlashes = computed(() => (rawFilepath.value ? `/${rawFilepath.value}/` : '/'))
const albumMode = computed(() => route.query.mode === 'album')

const setAlbumMode = (enabled: boolean) => {
  const query = { ...route.query }
  if (enabled) {
    query.mode = 'album'
  } else {
    delete query.mode
  }

  router.replace({ query })
}

const encodeFilePath = (path: string) => path.split('/').map(encodeURIComponent).join('/')
const getImageUrl = (image: AlbumImage, width?: number) => {
  const params = new URLSearchParams({ rotation: String(image.rotationDegrees) })
  if (width) {
    params.set('w', String(width))
  }

  return `${apiUrl}/files/${teamSlug.value}/${projectSlug.value}/${encodeFilePath(image.path)}?${params}`
}

const { data: files } = useQuery<FilesResponse>({
  queryKey: ['files', teamSlug, projectSlug, filepathWithSlashes],
  queryFn: async () => {
    const url = `${import.meta.env.VITE_API_URL}/api/v1/teams/${teamSlug.value}/projects/${projectSlug.value}/files/${filepathWithSlashes.value}`
    const response = await fetch(url, {
      credentials: 'include',
    })
    if (!response.ok) {
      error.value = (await response.json()).error
      throw new Error((await response.json()).error)
    }
    return response.json() as Promise<FilesResponse>
  },
})

const {
  data: albumImages,
  isLoading: albumImagesIsLoading,
  error: albumImagesError,
} = useQuery<AlbumImagesResponse>({
  queryKey: ['albumImages', teamSlug, projectSlug],
  enabled: albumMode,
  queryFn: async () => {
    const response = await fetch(
      `${import.meta.env.VITE_API_URL}/api/v1/teams/${teamSlug.value}/projects/${projectSlug.value}/album`,
      {
        credentials: 'include',
      },
    )
    if (!response.ok) {
      throw new Error((await response.json()).error)
    }
    return response.json() as Promise<AlbumImagesResponse>
  },
})

const albumImagesList = computed(() => albumImages.value?.images ?? [])
const albumDialogOpen = ref(false)
const selectedImageIndex = ref(0)
const selectedImage = computed(() => albumImagesList.value[selectedImageIndex.value])

const rotateAlbumImage = useMutation({
  mutationKey: ['rotateAlbumImage', teamSlug, projectSlug],
  mutationFn: async (image: AlbumImage) => {
    const response = await fetch(
      `${import.meta.env.VITE_API_URL}/api/v1/teams/${teamSlug.value}/projects/${projectSlug.value}/album/rotate`,
      {
        method: 'POST',
        credentials: 'include',
        body: JSON.stringify({
          path: image.path,
        }),
      },
    )
    if (!response.ok) {
      throw new Error((await response.json()).error)
    }
    return response.json() as Promise<RotateAlbumImageResponse>
  },
  onSuccess(data) {
    queryClient.setQueryData<AlbumImagesResponse>(
      ['albumImages', teamSlug, projectSlug],
      (currentAlbumImages) => {
        if (!currentAlbumImages) return currentAlbumImages

        return {
          ...currentAlbumImages,
          images: currentAlbumImages.images.map((image) =>
            image.path === data.path ? { ...image, rotationDegrees: data.rotationDegrees } : image,
          ),
        }
      },
    )
  },
})

const rotateSelectedImage = () => {
  if (!authUser.value) return
  if (!selectedImage.value) return
  rotateAlbumImage.mutate(selectedImage.value)
}

const openAlbumImage = (index: number) => {
  selectedImageIndex.value = index
  albumDialogOpen.value = true
}

const showPreviousImage = () => {
  const imageCount = albumImagesList.value.length
  if (imageCount === 0) return
  selectedImageIndex.value = (selectedImageIndex.value - 1 + imageCount) % imageCount
}

const showNextImage = () => {
  const imageCount = albumImagesList.value.length
  if (imageCount === 0) return
  selectedImageIndex.value = (selectedImageIndex.value + 1) % imageCount
}

useEventListener(window, 'keydown', (event: KeyboardEvent) => {
  if (!albumDialogOpen.value) return

  if (event.key === 'ArrowLeft') {
    event.preventDefault()
    showPreviousImage()
  }

  if (event.key === 'ArrowRight') {
    event.preventDefault()
    showNextImage()
  }

  if (event.key.toLowerCase() === 'r') {
    event.preventDefault()
    rotateSelectedImage()
  }
})

watch(albumImagesList, (images) => {
  if (selectedImageIndex.value >= images.length) {
    selectedImageIndex.value = 0
  }
  if (images.length === 0) {
    albumDialogOpen.value = false
  }
})

watch([team, teamProject], () => {
  if (!team.value && !teamProject.value) {
    document.title = 'Index of'
    return
  }
  document.title = `Index of /${team.value?.team.slug}/${teamProject.value?.teamProject.slug}${filepathWithSlashes.value}`
})
</script>

<template>
  <DashboardLayout>
    <header class="h-[72px] py-4 px-6 flex justify-between items-center">
      <div class="flex gap-4 font-medium items-center">
        <Logo />

        <router-link :to="`/-`" v-if="authUser.value">
          <div class="text-neutral-400">/</div>
        </router-link>
        <div class="text-neutral-400" v-if="!authUser.value">/</div>

        <TeamsDropdown v-if="authUser.value">
          <div class="flex gap-2 items-center cursor-pointer text-neutral-400">
            {{ team && team.team.name }}
            <ChevronDown class="size-4 stroke-neutral-400" />
          </div>
        </TeamsDropdown>
        <div v-if="!authUser.value" class="flex gap-2 items-center text-neutral-400">
          {{ team && team.team.name }}
        </div>

        <router-link :to="`/-/${team && team.team.slug}`" v-if="authUser.value">
          <div class="text-neutral-400">/</div>
        </router-link>
        <div class="text-neutral-400" v-if="!authUser.value">/</div>

        <TeamProjectsDropdown v-if="authUser.value">
          <div class="flex gap-2 items-center cursor-pointer">
            {{ teamProject && teamProject.teamProject.name }}
            <ChevronDown class="size-4 stroke-neutral-600" />
          </div>
        </TeamProjectsDropdown>
        <div v-if="!authUser.value" class="flex gap-2 items-center">
          {{ teamProject && teamProject.teamProject.name }}
        </div>
        <Breadcrumbs
          :team-slug="route.params.team as string"
          :project-slug="route.params.project as string"
          :filepath="filepathWithSlashes"
        />
      </div>
      <div class="flex items-center gap-4">
        <router-link to="/auth/login" v-if="!authUser.value">
          <Button> Log in </Button>
        </router-link>
      </div>
      <div class="flex items-center gap-4" v-if="authUser.value">
        <Dialog>
          <DialogTrigger>
            <div class="relative w-[350px] h-8">
              <Search class="size-4 stroke-neutral-400 absolute top-1/2 -translate-y-1/2 left-2" />
              <Input class="px-8" placeholder="Search" />
            </div>
          </DialogTrigger>
          <DialogContent
            :show-close="false"
            class="p-0 gap-0 divide-y divide-y-neutral-200 border-0"
          >
            <div class="relative h-12">
              <Search class="size-4 stroke-neutral-400 absolute top-1/2 -translate-y-1/2 left-2" />
              <Input class="px-8 h-12 rounded-b-none" placeholder="Search" />
            </div>
            <div class="h-[350px] flex flex-col gap-1 overflow-y-scroll p-2 no-scrollbar">
              <Button class="w-full" size="sm" variant="secondary"> Test </Button>
            </div>
          </DialogContent>
        </Dialog>
        <FileUploader
          :team-slug="route.params.team as string"
          :project-slug="route.params.project as string"
          :target-path="filepathWithSlashes"
        />
        <Dialog v-model:open="createFolderOpen">
          <DialogTrigger :as-child="true">
            <Button size="sm"><Folder class="size-4" /> New Folder </Button>
          </DialogTrigger>
          <DialogContent :show-close="true">
            <DialogHeader>
              <DialogTitle> Create a folder </DialogTitle>
            </DialogHeader>
            <form
              @submit.prevent="() => createFolder.mutate(folderPath)"
              class="flex flex-col gap-4"
            >
              <Input v-model="folderPath" placeholder="Name" />
              <DialogFooter>
                <Button> Create </Button>
              </DialogFooter>
            </form>
          </DialogContent>
        </Dialog>
        <Button
          size="sm"
          :variant="albumMode ? 'secondary' : 'outline'"
          @click="setAlbumMode(!albumMode)"
        >
          <Images class="size-4" />
          {{ albumMode ? 'Files' : 'Album' }}
        </Button>
        <router-link :to="`/statistics`">
          <Button size="sm"
            ><ChartArea class="size-4" />
            <div class="sr-only">To Statistics</div>
          </Button>
        </router-link>
      </div>
    </header>
    <div class="p-4">
      <div v-if="albumMode" class="max-w-full overflow-hidden space-y-4">
        <div class="flex items-end justify-between gap-4">
          <div>
            <h1 class="text-xl font-semibold">Album</h1>
            <p class="text-sm text-neutral-500">
              {{ albumImagesList.length }} image{{ albumImagesList.length === 1 ? '' : 's' }} in
              this project
            </p>
          </div>
        </div>

        <div v-if="albumImagesIsLoading" class="flex items-center gap-2 text-sm text-neutral-500">
          <Loader2 class="size-4 animate-spin" /> Loading album
        </div>

        <Alert v-else-if="albumImagesError" variant="destructive">
          <AlertCircle class="w-4 h-4" />
          <AlertTitle>Error</AlertTitle>
          <AlertDescription> {{ albumImagesError.message }} </AlertDescription>
        </Alert>

        <div
          v-else-if="albumImagesList.length === 0"
          class="rounded-lg border border-dashed p-10 text-center text-sm text-neutral-500"
        >
          No PNG or JPG images found in this project.
        </div>

        <div v-else class="max-w-full columns-2 gap-3 sm:columns-3 lg:columns-4 xl:columns-5">
          <button
            v-for="(image, index) in albumImagesList"
            :key="image.path"
            class="mb-3 block w-full break-inside-avoid overflow-hidden rounded-lg border bg-neutral-50 text-left shadow-sm transition hover:-translate-y-0.5 hover:shadow-md"
            @click="openAlbumImage(index)"
          >
            <img
              :src="getImageUrl(image, 480)"
              :alt="image.path"
              class="w-full object-cover"
              loading="lazy"
            />
          </button>
        </div>

        <Dialog v-model:open="albumDialogOpen">
          <DialogContent
            :show-close="false"
            class="h-screen max-h-screen w-screen max-w-none border-0 bg-neutral-950 p-0 text-white sm:rounded-none"
          >
            <div
              v-if="selectedImage"
              class="relative flex h-screen w-screen items-center justify-center overflow-hidden"
            >
              <button
                class="absolute right-4 top-4 z-10 rounded-full bg-white/10 p-3 text-white backdrop-blur transition hover:bg-white/20"
                @click="albumDialogOpen = false"
              >
                <X class="size-5" />
                <span class="sr-only">Close album image</span>
              </button>

              <Button
                v-if="authUser.value"
                variant="ghost"
                size="icon"
                class="absolute right-20 top-4 z-10 size-11 rounded-full bg-white/10 text-white backdrop-blur hover:bg-white/20 hover:text-white"
                :disabled="rotateAlbumImage.isPending.value"
                @click="rotateSelectedImage"
              >
                <RotateCw class="size-5" />
                <span class="sr-only">Rotate image clockwise</span>
              </Button>

              <Button
                variant="ghost"
                size="icon"
                class="absolute left-4 z-10 size-12 rounded-full bg-white/10 text-white backdrop-blur hover:bg-white/20 hover:text-white"
                @click="showPreviousImage"
              >
                <ChevronLeft class="size-7" />
                <span class="sr-only">Previous image</span>
              </Button>

              <img
                :src="getImageUrl(selectedImage)"
                :alt="selectedImage.path"
                class="max-h-screen max-w-full object-contain"
              />

              <Button
                variant="ghost"
                size="icon"
                class="absolute right-4 z-10 size-12 rounded-full bg-white/10 text-white backdrop-blur hover:bg-white/20 hover:text-white"
                @click="showNextImage"
              >
                <ChevronRight class="size-7" />
                <span class="sr-only">Next image</span>
              </Button>

              <div
                class="absolute bottom-4 left-1/2 max-w-[calc(100vw-2rem)] -translate-x-1/2 rounded-full bg-black/60 px-4 py-2 text-center text-sm backdrop-blur"
              >
                <div class="truncate">{{ selectedImage.path }}</div>
                <div class="text-xs text-neutral-300">
                  {{ selectedImageIndex + 1 }} / {{ albumImagesList.length }}
                </div>
              </div>
            </div>
          </DialogContent>
        </Dialog>
      </div>

      <Alert v-if="!albumMode && error" variant="destructive">
        <AlertCircle class="w-4 h-4" />
        <AlertTitle>Error</AlertTitle>
        <AlertDescription> {{ error }} </AlertDescription>
      </Alert>
      <Table class="rounded-t-lg overflow-clip" v-if="!albumMode && !error">
        <TableHeader>
          <TableRow>
            <TableHead> Name </TableHead>
            <TableHead> Last Changed </TableHead>
            <TableHead> Size </TableHead>
          </TableRow>
        </TableHeader>
        <TableBody v-if="files">
          <TableRow
            class="hover:underline cursor-pointer"
            v-for="file in [
              rawFilepath !== '.'
                ? {
                    type: 'dir',
                    name: '..',
                    lastModified: '',
                    size: '',
                  }
                : null,
              ...files.files.slice().sort((a, b) => {
                if (a.type === 'dir' && b.type !== 'dir') return -1
                if (a.type !== 'dir' && b.type === 'dir') return 1
                return a.name.localeCompare(b.name, undefined, { sensitivity: 'base' })
              }),
            ].filter((a) => !!a)"
            v-bind:key="`${filepathWithSlashes}${file.name}`"
            @click="
              () => {
                if (file.type === 'dir') {
                  queryClient.invalidateQueries({ queryKey: ['files'] })
                  router.replace(
                    `/-/${route.params.team}/${route.params.project}${filepathWithSlashes}${file.name}`,
                  )
                } else {
                  // Window is added but the types don't seem to work see /frontend/env.d.ts
                  // @ts-ignore
                  window.location.href = `${apiUrl}/files/${route.params.team}/${route.params.project}${filepathWithSlashes}${file.name}`
                }
              }
            "
          >
            <TableCell>
              <div class="flex gap-2 items-center">
                <File class="size-4 stroke-neutral-600" v-if="file.type === 'file'" />
                <Folder class="size-4 stroke-neutral-600" v-if="file.type === 'dir'" />
                {{ file.name }}
              </div>
            </TableCell>
            <TableCell>{{
              file.lastModified !== '' ? new Date(file.lastModified).toDateString() : ''
            }}</TableCell>
            <TableCell
              ><div v-if="file.type === 'file'">
                {{ humanFileSize(file.size as number) }}
              </div></TableCell
            >
          </TableRow>
        </TableBody>
      </Table>
    </div>
  </DashboardLayout>
</template>
