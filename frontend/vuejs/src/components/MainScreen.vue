<template>
  <div class="main">
    <template v-if="!sessionFound()">
      <div class="grid-container">
        <div class="grid-item-header">
          <svg src="" width="64px" height="64px" xmlns="http://www.w3.org/2000/svg"
               xmlns:xlink="http://www.w3.org/1999/xlink" xmlns:serif="http://www.serif.com/"
               viewBox="0 0 9 11" version="1.1" xml:space="preserve"
               style="fill-rule:evenodd;clip-rule:evenodd;stroke-linejoin:round;stroke-miterlimit:2;">
            <g id="logo">
              <path
                  d="M6.997,0L6.996,-0L8.008,0L8.004,1.004L9,1.002C8.999,2.335 9,5 9,5L8,5L8,6L9,6C9,6 8.999,9.626 9,11L1,11L1,6L0,6L0,5L2,5L2,10L3,10L3,7L4,7L4,6L5,6L5,5L4,5L4,1L5,1L5,0L5.999,0L5.998,2L7.001,2.001C7.001,2.001 7.007,0.068 6.997,0ZM6,3L5,3L5,4L6,4L6,3ZM8,3L7,3L7,4L8,4L8,3Z"/>
            </g>
					</svg>
        </div>
        <div class="grid-item-main">
          <button class="button button-big" v-on:click="createSession">create</button>
          <div class="text text-regular">or</div>
          <div class="input-container">
            <input type="text" placeholder="session id..." v-model.lazy="sessionId">
            <button class="button button-big">join</button>
          </div>
          <div style="display: none" class="error text-regular">session with the id is not found</div>
        </div>
      </div>
    </template>

    <template v-if="sessionFound() && userId === ''">
      <label>
        Type name:
        <input v-model="name">
      </label>
      <button class='button' v-on:click="joinSession">Join Session</button>
    </template>
    <template v-if="sessionFound() && userId !== '' && session === null">
      <label>
        Type name:
        <input v-model="name">
      </label>
      <button v-on:click="joinSession">Join Session</button>
    </template>

    <template v-if="sessionFound() && userId !== '' && session !== null">
      <button class='button vote-button' v-on:click="voteInSession(1)">1</button>
      <button class='button vote-button' v-on:click="voteInSession(2)">2</button>
      <button class='button vote-button' v-on:click="voteInSession(3)">3</button>
      <button class='button vote-button' v-on:click="voteInSession(5)">5</button>
      <button class='button vote-button' v-on:click="voteInSession(8)">8</button>
      <button class='button vote-button' v-on:click="voteInSession(13)">13</button>
      <button class='button vote-button' v-on:click="voteInSession(20)">20</button>

      <br>
      <button class='button clear-button' v-on:click="clearVotes">Clear votes</button>
      <button class='button show-votes-button' v-on:click="showVotes">Show votes</button>

      <template v-if="session.votes_info !== null && session.votes_info !== undefined">
        <table>
          <thead>
          <tr>
            <th>Name</th>
            <th>Vote</th>
          </tr>
          </thead>
          <tbody>
          <tr v-bind:class="{ currentUser: vote.is_current_user}"
              v-for="vote in session.votes_info" :key="vote.name">
            <td>{{ vote.name }}</td>
            <template v-if="vote.is_voted">
              <template v-if="vote.vote !== null">
                <td>{{ vote.vote }}</td>
              </template>
              <template v-else>
                <td>Voted</td>
              </template>
            </template>
            <template v-else>
              <td>No vote</td>
            </template>
          </tr>
          </tbody>
        </table>
      </template>
    </template>
    <template v-if="session !== null && session.votes_hidden === false">
      <br>
      <strong>Average:</strong><label> {{ averageVote() }}</label>
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
      for (var i = 0; i < this.session.votes_info.length; i++) {
        let voteInfo = this.session.votes_info[i];

        let vote = parseInt(voteInfo.vote);
        if (!isNaN(vote)) {
          total += vote;
          count++;
        }
      }

      if (count === 0) {
        return 0
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

