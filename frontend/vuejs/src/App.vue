<template>
  <div id="app">
    <MainScreen v-bind:backend-url="backendUrl" v-bind:session-id="sessionId"/>
  </div>
</template>

<script>
import MainScreen from './components/MainScreen.vue'

export default {
  name: 'App',
  components: {
    MainScreen
  },
  data() {
    return {
      backendUrl: process.env.VUE_APP_BACKEND_URL,
      sessionId: parseInt(this.$route.params.sessionId) || null,
      userId: null,
      userName: null
    }
  },
  created() {
    document.title = "=^_^= scrum planning poker";
  }
}
</script>

<style>

@import './assets/styles/main.css';
@import url('https://fonts.googleapis.com/css2?family=Aldrich&display=swap');

#app {
  font-family: Aldrich, Helvetica, Avenir, Arial, sans-serif;
  -webkit-font-smoothing: antialiased;
  -moz-osx-font-smoothing: grayscale;
  color: #2c3e50;
}
</style>
