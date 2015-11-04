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

<script lang='babel'>
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
    smooth: true
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
let e;
let xAxisData = new Array();
let dailyData = new Array();
let dailyUsers = new Array();
export default{
  ready(){
    this.$http.get('/api',(data,status,req) => {
      for (let i of data['total']){
        if (i['total']==0){
          continue
        }
        xAxisData.push(i['date'].replace('T00:00:00+08:00',''));
        dailyData.push(i['total']);
      }
      for (let i of data['users']){
        if(i['userCount']==0){
          continue
        }
        dailyUsers.push(i['userCount']);
      }
      option.xAxis[0].data = xAxisData;
      option.series[0].data = dailyData;
      option.series[1].data = dailyUsers;
      document.getElementById('chart').style.width = '100%';
      e = echarts.init(document.getElementById('chart'));
      e.on(eConfig.EVENT.CLICK,param => {
          document.getElementById('chart').style.width = '70%'
          e.resize();
          pie.methods.show(xAxisData[param.dataIndex],dailyData[param.dataIndex]);
        }).setOption(option);
      window.onresize = e.resize;
    });
  },
  components: {
    pie
  },
  events: {
    resize(){
      document.getElementById('chart').style.width = '100%';
      e.resize();
    }
  }
}
</script>
