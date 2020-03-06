<template>
    <div id="app">
        <MainScreen :backend-url="backendUrl"/>
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
                sessionId: null,
                userId: null,
                userName: null
            }
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
