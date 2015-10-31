<template lang='jade'>
#pie
button#hide(v-on:click='hidePie') Hide
</template>

<style lang='stylus'>
#pie
  width 500px
  height 400px
  margin-left 30px
button
  float right
  margin-right 10%
</style>

<script>
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
var e, v;
export default {
  ready(){
    v = this;
    e = echarts.init(document.getElementById('pie'));
    e.on(eConfig.EVENT.CLICK,function(param){
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
      var others = total;
      var pieData = new Array();
      v.$http.get('/api/rank/' + date,function (data,status,req) {
        for (var i in data['rank']){
          pieData.push({
            value:data['rank'][i]['count'],
            name:data['rank'][i]['name']
          });
          others -= data['rank'][i]['count'];
        }
        if (others !== 0) {
          pieData.push({value:others,name:'其他'});
        }
        option.series[0].data = pieData;
        e.setOption(option);
        document.getElementById('pie').style.display = '';
        document.getElementById('hide').style.display = '';
      });
    },
    hidePie(){
      document.getElementById('pie').style.display = 'none';
      document.getElementById('hide').style.display = 'none';
      this.$dispatch('resize');
    }
  }
}
</script>
