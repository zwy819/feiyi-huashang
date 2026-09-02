<template>
  <div class="flex flex-col min-h-screen">
    <Navbar />
    <main class="flex-1">
      <router-view />
    </main>
    <Footer />
    <ToastContainer />
  </div>
</template>

<script setup lang="ts">
import Navbar from './Navbar.vue'
import Footer from './Footer.vue'
import ToastContainer from '@/components/Common/Toast.vue'
</script>
