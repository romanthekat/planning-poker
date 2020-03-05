<template>
    <div class="hello">
        <template v-if="sessionId == null">
            <button v-on:click="createSession">Create Session</button>
        </template>
        <template v-else>
            <label>session id: {{ sessionId }}</label>
        </template>
    </div>
</template>

<script>
    import axios from "axios";

    export default {
        name: 'MainScreen',
        props: {
            backendUrl: String,
            sessionId: String
        },
        methods: {
            createSession() {
                axios({
                    method: 'post',
                    url: '/createSession',
                    baseURL: this.backendUrl,
                })
                    .then(response => {
                        this.sessionId = response.data.id
                    })
                    .catch(error => {
                        console.log(error);
                    });
            }
        }
    }
</script>

<!-- Add "scoped" attribute to limit CSS to this component only -->
<style scoped>
    h3 {
        margin: 40px 0 0;
    }

    ul {
        list-style-type: none;
        padding: 0;
    }

    li {
        display: inline-block;
        margin: 0 10px;
    }

    a {
        color: #42b983;
    }
</style>
