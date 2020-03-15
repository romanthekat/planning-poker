import Vue from 'vue'
import VueRouter from 'vue-router'

import App from './App.vue'

Vue.use(VueRouter)

Vue.config.productionTip = false


const routes = [
  { path: '/', component: App },
  { path: '/:sessionId', props: true, component: App },
]

const router = new VueRouter({
  routes // short for `routes: routes`
})

new Vue({
  render: h => h(App),
  router: router,
}).$mount('#app')
