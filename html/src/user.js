var Vue = require('vue');
var c = require('./user.vue');

Vue.use(require('vue-resource'));
Vue.http.options.root = 'http://localhost';
new Vue({
  el: 'body',
  components: {
    index: c
  }
})
