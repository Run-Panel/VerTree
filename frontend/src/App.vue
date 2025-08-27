<template>
  <div id="app">
    <router-view />
  </div>
</template>

<script setup>
// This is the root component
</script>

<style>
#app {
  font-family: Avenir, Helvetica, Arial, sans-serif;
  -webkit-font-smoothing: antialiased;
  -moz-osx-font-smoothing: grayscale;
}

/* Global styles */
.app-container {
  min-height: 100vh;
  background-color: #f5f7fa;
}

.main-content {
  margin-left: 250px;
  padding: 20px;
  min-height: calc(100vh - 60px);
}

@media (max-width: 768px) {
  .main-content {
    margin-left: 0;
  }
}
</style>
