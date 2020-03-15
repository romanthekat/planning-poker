<template>
    <div class="main">
        <template v-if="!sessionFound()">
            <button v-on:click="createSession">Create Session</button>
            <label>
                or join by id:
                <input v-model.lazy="sessionId">
            </label>
        </template>
        <template v-if="sessionFound() && userId === ''">
            <label>
                Type name:
                <input v-model="name">
            </label>
            <button v-on:click="joinSession">Join Session</button>
        </template>
        <template v-if="sessionFound() && userId !== '' && session === null">
            <label>
                Type name:
                <input v-model="name">
            </label>
            <button v-on:click="joinSession">Join Session</button>
        </template>
        <template v-if="sessionFound() && userId !== '' && session !== null">
            <button class='vote-btn' v-on:click="voteInSession(1)">1</button>
            <button class='vote-btn' v-on:click="voteInSession(2)">2</button>
            <button class='vote-btn' v-on:click="voteInSession(3)">3</button>
            <button class='vote-btn' v-on:click="voteInSession(5)">5</button>
            <button class='vote-btn' v-on:click="voteInSession(8)">8</button>
            <button class='vote-btn' v-on:click="voteInSession(13)">13</button>
            <button class='vote-btn' v-on:click="voteInSession(20)">20</button>

            <br>
            <button class='clear-btn' v-on:click="clearVotes">Clear votes</button>
            <button class='show-votes-btn' v-on:click="showVotes">Show votes</button>

            <template v-if="session.votes_info !== null && session.votes_info !== undefined">
                <table>
                    <thead>
                    <tr>
                        <th>Name</th>
                        <th>Vote</th>
                    </tr>
                    </thead>
                    <tbody>
                    <tr v-for="item in Object.entries(this.session.votes_info)" :key="item.message">
                        <td>{{item[0]}}</td>
                        <td>{{item[1]}}</td>
                    </tr>
                    </tbody>
                </table>
            </template>
        </template>
        <template v-if="session !== null && session.votes_hidden === false">
            <br>
            <strong>Average:</strong><label> {{averageVote()}}</label>
        </template>
    </div>
</template>

<script>
    import axios from "axios";

    export default {
        name: 'MainScreen',
        props: {
            backendUrl: String,
            sessionId: Number,
        },
        data() {
            return {
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
            sessionFound() {
                return this.sessionId != null && !isNaN(this.sessionId)
            },
            averageVote() {
                let total = 0;
                let count = 0;
                for (let key in this.session.votes_info) {
                    total += parseInt(this.session.votes_info[key])
                    count++;
                }
                return total / count;
            },
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
                        this.$router.push('/' + response.data.id)
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
            voteInSession(vote) {
                this.vote = vote

                axios({
                        method: 'post',
                        baseURL: this.backendUrl,
                        url: '/sessions/' + this.sessionId + '/vote',
                        data: JSON.stringify({user_id: this.userId, vote: parseFloat(vote)}),
                    }
                )
                    .then(() => {
                        this.fetchSession()
                    })
                    .catch(error => {
                        console.log(error);
                    });
            },
            clearVotes() {
                axios({
                        method: 'post',
                        baseURL: this.backendUrl,
                        url: '/sessions/' + this.sessionId + '/clear',
                        data: JSON.stringify({user_id: this.userId}),
                    }
                )
                    .then(() => {
                        this.fetchSession()
                    })
                    .catch(error => {
                        console.log(error);
                    });
            },
            showVotes() {
                axios({
                        method: 'post',
                        baseURL: this.backendUrl,
                        url: '/sessions/' + this.sessionId + '/show',
                        data: JSON.stringify({user_id: this.userId}),
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
