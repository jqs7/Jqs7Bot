<template lang='jade'>
.center
  h1 群组娘第二舰队活动图
  searchbox
  chart
</template>

<style lang='stylus'>
.center
  text-align center
</style>

<script>
import chart from './chart.vue'
import searchbox from './searchbox.vue'

export default{
  components: {
    chart,
    searchbox
  }
}
</script>
