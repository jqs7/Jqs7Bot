<template lang='jade'>
#pie
button#hide(v-on:click='hidePie') Hide
</template>

<style lang='stylus'>
#pie
  float left
  width 27%
  height 377px
button
  float right
  margin-right 10%
</style>

<script lang='babel'>
var echarts = require('echarts');
var eConfig = require('echarts/config');
require('echarts/chart/pie');
var option = {
  title: {
    text: '',
    subtext: '',
    x: '100',
  },
  tooltip: {
    trigger: 'item',
    formatter: "{b} : {c} ({d}%)"
  },
  series: [
    {
      type:'pie',
      radius : '47%',
      center: ['37%', '50%'],
      data:[]
    }
  ]
};
let e, v;
export default {
  ready(){
    v = this;
    e = echarts.init(document.getElementById('pie'));
    e.on(eConfig.EVENT.CLICK,param => {
      if (param.name !== '其他'){
        window.location.href = '/user/' + param.name;
      }
    })
    document.getElementById('pie').style.display = 'none';
    document.getElementById('hide').style.display = 'none';
  },
  methods: {
    show(date, total){
      option.title.subtext = date;
      let others = total;
      let pieData = new Array();
      e.showLoading({text: 'Loading...', effect: 'whirling'});
      v.$http.get('/api/rank/' + date, (data,status,req) => {
        for (let i of data['rank']){
          pieData.push({
            value:i['count'],
            name:i['name']
          });
          others -= i['count'];
        }
        if (others !== 0) {
          pieData.push({value:others, name:'其他'});
        }
        option.series[0].data = pieData;
        e.setOption(option);
        e.hideLoading();
      });
      document.getElementById('pie').style.display = '';
      document.getElementById('hide').style.display = '';
    },
    hidePie(){
      document.getElementById('pie').style.display = 'none';
      document.getElementById('hide').style.display = 'none';
      this.$dispatch('resize');
    }
  }
}
</script>
