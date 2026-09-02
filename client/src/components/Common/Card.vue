<template>
  <component
    :is="to ? 'router-link' : 'div'"
    :to="to"
    :class="[
      'card-mogao overflow-hidden cursor-pointer group transition-shadow duration-300 hover:-translate-y-1',
      className,
    ]"
    :style="{ transition: 'transform 0.25s ease, box-shadow 0.25s ease' }"
    @click="$emit('click')"
  >
    <!-- 图片区域 -->
    <div v-if="image" class="aspect-[4/3] overflow-hidden bg-mogao-ochre-100">
      <img
        :src="image"
        :alt="title"
        class="w-full h-full object-cover transition-transform duration-500 group-hover:scale-105"
        loading="lazy"
      />
    </div>

    <!-- 内容区域 -->
    <div class="p-4">
      <h3 v-if="title" class="font-serif text-base font-semibold text-mogao-lapis-800 line-clamp-1">
        {{ title }}
      </h3>
      <p v-if="subtitle" class="mt-1 text-xs text-mogao-ochre-600 line-clamp-2">{{ subtitle }}</p>
      <div v-if="$slots.default" class="mt-3">
        <slot />
      </div>
    </div>
  </component>
</template>

<script setup lang="ts">
defineProps<{
  title?: string
  subtitle?: string
  image?: string
  to?: string
  className?: string
}>()

defineEmits<{
  click: []
}>()
</script>
