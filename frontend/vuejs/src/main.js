import Vue from 'vue'
import VueRouter from 'vue-router'
import VueClipboard from 'vue-clipboard2'
import Vuelidate from 'vuelidate'

import App from './App.vue'

Vue.use(VueRouter)

VueClipboard.config.autoSetContainer = true
Vue.use(VueClipboard)

Vue.use(Vuelidate)

Vue.config.productionTip = false


const routes = [
    {path: '/', component: App},
    {path: '/:sessionId', props: true, component: App},
]

const router = new VueRouter({
    routes // short for `routes: routes`
})

new Vue({
    render: h => h(App),
    router: router,
}).$mount('#app')
