<template lang='jade'>
.center
  h1 群组娘第二舰队活动图
  searchbox
  chart
foot
</template>

<style lang='stylus'>
.center
  text-align center
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
