<template lang='jade'>
corner
.center
  h1 群组娘第二舰队活动图
  searchbox
  chart
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
import chart from './chart.vue'
import searchbox from './searchbox.vue'
import foot from './footer.vue'

export default{
  components: {
    corner,
    searchbox,
    chart,
    foot
  }
}
</script>
