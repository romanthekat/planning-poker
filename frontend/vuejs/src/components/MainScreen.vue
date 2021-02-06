<template>
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
    <template v-if="!sessionFound()">
      <div class="grid-item-main">
        <button class="button button-big" v-on:click="createSession">create</button>
        <div class="text text-regular">or</div>
        <div class="input-container">
          <input type="text" placeholder="session id..." v-model.lazy="sessionId">
          <button class="button button-big">join</button>
        </div>
        <div style="display: none" class="error text-regular">session with the id is not found</div>
      </div>
    </template>

    <template v-if="sessionFound() && userId === ''">
      <div class="grid-item-main">
        <div class="input-container">
          <input type="text" placeholder="your name..." v-model="name">
          <button class='button button-big' v-on:click="joinSession">join</button>
        </div>
        <div style="display: none" class="error text-regular">please enter a correct name</div>
      </div>
    </template>
    <template v-if="sessionFound() && userId !== '' && session === null">
      <div class="grid-item-main">
        <div class="input-container">
          <input type="text" placeholder="your name..." v-model="name">
          <button class='button button-big' v-on:click="joinSession">join</button>
        </div>
        <div style="display: none" class="error text-regular">please enter a correct name</div>
      </div>
    </template>

    <template v-if="sessionFound() && userId !== '' && session !== null">
      <svg display="none">
        <defs>
          <g id="vote-circle">
            <circle cx="7.5" cy="7.5" r="7.5"/>
          </g>
        </defs>
      </svg>

      <template v-if="session.votes_info !== null && session.votes_info !== undefined">
        <div class="grid-item-members">
          <div class="members-title text text-regular">X/X voted</div>

          <div class="members-list">
            <div v-bind:class="{ 'members-list-item': true, 'current-user': vote.is_current_user}"
                 v-for="vote in session.votes_info" :key="vote.name">
              <svg class="vote-indicator" width="15" height="15" viewBox="0 0 15 15" fill="none"
                   xmlns="http://www.w3.org/2000/svg">
                <use href="#vote-circle" v-bind:class="{'vote-yes': vote.is_voted, 'vote-no': !vote_is_voted } "/>
              </svg>
              <div class="text username">{{ vote.name }}</div>
              <div class="card">
                <template v-if="vote.vote !== null">
                  <label>{{ vote.vote }}</label>
                </template>
                <template v-else>
                  <label> </label>
                </template>
              </div>
            </div>
          </div>
        </div>
      </template>

      <div class="grid-item-main">
        <div class="result-container">
          <div class="result">
            <div class="text text-regular">average</div>
            <div class="card">{{ averageVote() }}</div>
          </div>
          <div class="text timer">00:00:00</div>
        </div>

        <textarea placeholder="story description..."></textarea>

        <div class="card-container">
          <div class='card card-button' v-on:click="voteInSession(0)"><label>0</label></div>
          <div class='card card-button' v-on:click="voteInSession(1)"><label>1</label></div>
          <div class='card card-button' v-on:click="voteInSession(2)"><label>2</label></div>
          <div class='card card-button' v-on:click="voteInSession(3)"><label>3</label></div>
          <div class='card card-button' v-on:click="voteInSession(5)"><label>5</label></div>
          <div class='card card-button' v-on:click="voteInSession(8)"><label>8</label></div>
          <div class='card card-button' v-on:click="voteInSession(13)"><label>13</label></div>
          <div class='card card-button' v-on:click="voteInSession(20)"><label>20</label></div>
        </div>

        <div class="controls-container">
          <button class='button button-small' v-on:click="showVotes">show</button>
          <button class='button button-small' v-on:click="clearVotes">clear</button>
        </div>

        <div class="copy-link-container">
          <button class="copy-link-button">
            <svg width="30" height="38" viewBox="0 0 30 38" fill="none" xmlns="http://www.w3.org/2000/svg">
              <rect x="1.5" y="5.5" width="22" height="31" rx="4.5"/>
              <mask id="mask0" mask-type="alpha" maskUnits="userSpaceOnUse" x="5" y="0" width="25" height="34">
                <path fill-rule="evenodd" clip-rule="evenodd" d="M30 0H5V6H23V34H30V6V4V0Z" fill="#C4C4C4"/>
              </mask>
              <g mask="url(#mask0)">
                <rect x="6.5" y="1.5" width="22" height="31" rx="4.5"/>
              </g>
            </svg>
          </button>
          <span class="text-small">session id {{ sessionId }}</span>
        </div>
      </div>
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
      if (this.session.votes_hidden) {
        return " "
      }

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

