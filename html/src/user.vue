<template lang='jade'>
corner
.center
  h1 {{userName}} 子舰活动图
  user-chart
foot
</template>

<style lang='stylus'>
.center
  overflow hidden
  margin-bottom 70px
  text-align center
</style>

<script>
import corner from './corner.vue'
import userChart from './userChart.vue'
import foot from './footer.vue'

export default{
  ready(){
    let tmp = document.URL.split('/');
    let userName = unescape(tmp[tmp.length - 1]);
    this.$data = {userName: userName};
    userChart.methods.show(userName);
  },
  components: {
    corner,
    userChart,
    foot
  }
}
</script>
