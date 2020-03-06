<template>
    <div class="main">
        <template v-if="sessionId === ''">
            <button v-on:click="createSession">Create Session</button>
            <label>
                or join by id:
                <input v-model.lazy="sessionId">
            </label>
        </template>
        <template v-if="sessionId !== '' && userId === ''">
            <label>
                Type name:
                <input v-model="name">
            </label>
            <button v-on:click="joinSession">Join Session</button>
        </template>
        <template v-if="sessionId !== '' && userId !== '' && session === null">
            <label>
                Type name:
                <input v-model="name">
            </label>
            <button v-on:click="joinSession">Join Session</button>
        </template>
        <template v-if="sessionId !== '' && userId !== '' && session !== null">
            <label>Session id: {{sessionId}}</label>
            <br>

            <input v-model="vote">
            <button v-on:click="voteInSession">Vote</button>

            <table>
                <thead>
                <tr>
                    <th>name</th>
                    <th>vote</th>
                </tr>
                </thead>
                <tbody>
                <tr v-for="item in Object.entries(this.session.votes_info)" :key="item.message">
                    <td>{{item[0]}}</td>
                    <td>{{item[1]}}</td>
                </tr>
                </tbody>
            </table>

            <ul id="example-1">

            </ul>
        </template>
    </div>
</template>

<script>
    import axios from "axios";

    export default {
        name: 'MainScreen',
        props: {
            backendUrl: String,

        },
        data() {
            return {
                sessionId: '',
                name: '',
                userId: '',
                timer: '',
                session: {},
                vote: 0.0
            }
        },
        created() {
            this.interval = setInterval(this.fetchSession, 1000)
        },
        beforeDestroy() {
            clearInterval(this.interval)
        },

        methods: {
            fetchSession() {
                if (this.userId !== '') {
                    axios({
                        method: 'get',
                        baseURL: this.backendUrl,
                        url: '/sessions/' + this.sessionId + '/get/' + this.userId,
                    })
                        .then(response => {
                            this.session = response.data
                        })
                        .catch(error => {
                            console.log(error);
                        });
                }
            },
            beforeDestroy() {
                clearInterval(this.timer)
            },

            createSession() {
                axios({
                    method: 'post',
                    baseURL: this.backendUrl,
                    url: '/sessions',
                })
                    .then(response => {
                        this.sessionId = response.data.id
                    })
                    .catch(error => {
                        console.log(error);
                    });
            },
            joinSession() {
                axios({
                        method: 'post',
                        baseURL: this.backendUrl,
                        url: '/sessions/' + this.sessionId + '/join',
                        data: JSON.stringify({name: this.name}),
                    }
                )
                    .then(response => {
                        this.userId = response.data.id
                    })
                    .catch(error => {
                        console.log(error);
                    });
            },
            voteInSession() {
                axios({
                        method: 'post',
                        baseURL: this.backendUrl,
                        url: '/sessions/' + this.sessionId + '/vote',
                        data: JSON.stringify({user_id: this.userId, vote: parseFloat(this.vote)}),
                    }
                )
                    .then(() => {
                        this.fetchSession()
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
