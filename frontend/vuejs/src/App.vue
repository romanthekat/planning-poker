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
                backendUrl: "http://127.0.0.1:10080/api",
                sessionId: parseInt(this.$route.params.sessionId) || null,
                userId: null,
                userName: null
            }
        },
        created () {
            document.title = "=^_^= scrum planning";
        }
    }
</script>

<style>
    #app {
        font-family: Avenir, Helvetica, Arial, sans-serif;
        -webkit-font-smoothing: antialiased;
        -moz-osx-font-smoothing: grayscale;
        color: #2c3e50;
    }
</style>
