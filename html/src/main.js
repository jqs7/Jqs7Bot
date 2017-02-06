var Vue = require('vue');
var c = require('./index.vue');

Vue.use(require('vue-resource'));
Vue.http.options.root = 'http://localhost';
new Vue({
  el: 'body',
  components: {
    index: c
  }
})
