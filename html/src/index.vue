<template lang='jade'>
.center
  h1 群组娘第二舰队活动图
  searchbox
  chart
footer
  foot
</template>

<style lang='stylus'>
.center
  text-align center
footer
  position absolute
  left 0
  bottom 0
  width 100%
  background #DDE4E9
</style>

<script>
import chart from './chart.vue'
import searchbox from './searchbox.vue'
import foot from './footer.vue'

export default{
  components: {
    searchbox,
    chart,
    foot
  }
}
</script>
