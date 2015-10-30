<template lang='jade'>
.container
  #chart
  pie
</template>

<style lang='stylus'>
.container
  text-align center
  margin-top 30px
  #chart
    float left
    height 400px
</style>

<script>
import pie from './pie.vue'
var echarts = require('echarts');
var eConfig = require('echarts/config');
require('echarts/chart/line');
var option = {
  legend:{
    data:["日发言量","日活跃用户"]
  },
  dataZoom: {
    show: true,
    realtime: true,
    start: 0,
    end:100
  },
  tooltip: {
    trigger: 'axis'
  },
  yAxis: [{},{
    scale: true,
    splitArea: {
      show: true
    }
  }],
  xAxis: [{}],
  series: [{
    name: "日发言量",
    type: 'line',
    yAxisIndex:0,
    smooth: true,
    markPoint: {
      data: [
        {type: 'max',name: 'max'},
        {type: 'min',name: 'min'},
      ]
    }
  },
  {
    name: "日活跃用户",
    type: "line",
    yAxisIndex:1,
    smooth: true,
    markLine: {
      data: [
        {type: 'average', name: 'avg'}
      ]
    },
    markPoint:{
      data: [
        {type: 'max',name:'max'},
        {type: 'min',name:'min'}
      ]
    }
  }]
};
var xAxisData = new Array();
var dailyData = new Array();
var dailyUsers = new Array();
export default{
  ready (){
    this.$http.get('/api',function(data,status,req){
      for (var i in data['total']){
        xAxisData.push(data['total'][i]['date'].replace('T00:00:00+08:00',''));
        dailyData.push(data['total'][i]['total']);
      }
      for (var i in data['users']){
        dailyUsers.push(data['users'][i]['userCount']);
      }
      option.xAxis[0].data = xAxisData;
      option.series[0].data = dailyData;
      option.series[1].data = dailyUsers;
      document.getElementById('chart').style.width = '100%';
      var e = echarts.init(document.getElementById('chart'));
      e.on(eConfig.EVENT.CLICK,function(param){
          document.getElementById('chart').style.width = '70%'
          e.resize();
          pie.methods.show(xAxisData[param.dataIndex],dailyData[param.dataIndex]);
        }).setOption(option);
    });
  },
  components: {
    pie
  }
}
</script>
