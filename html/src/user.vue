<template lang='jade'>
.center
  h1 {{{userName}}} 子舰活动图
  user-chart
</template>

<style lang='stylus'>
.center
  text-align center
</style>

<script>
import userChart from './userChart.vue'

export default{
  ready (){
    var tmp = document.URL.split('/');
    var userName = tmp[tmp.length - 1];
    this.$data = {userName: userName};
    userChart.methods.show(userName);
  },
  components: {
    userChart
  }
}
</script>
