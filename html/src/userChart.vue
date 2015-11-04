<template lang='jade'>
#chart
</template>

<style lang='stylus'>
#chart
  height 400px
</style>

<script language='babel'>
var echarts = require('echarts');
require('echarts/chart/line');
var option = {
  dataZoom: {
    show: true,
    realtime: true,
    start: 0,
    end: 100
  },
  tooltip: {
    trigger: 'axis'
  },
  yAxis: [],
  xAxis: [{}],
  series: [{
    name: "total",
    type: "line",
    smooth: true,
    data: [],
    markLine: {
      data: [
        {type: 'average', name: 'avg'}
      ]
    },
    markPoint: {
      data: [
        {type: 'min',name: 'min'},
        {type: 'max',name: 'max'}
      ]
    }
  }]
}
let e,v;
let xAxisData = new Array();
let dailyData = new Array();
export default{
  ready(){
    v = this;
    e = echarts.init(document.getElementById('chart'));
    window.onresize = e.resize;
  },
  methods: {
    show(userName) {
      v.$http.get('/api/user/' + escape(userName),(data,status,req) => {
        for (let i of data['result']){
          xAxisData.push(i['date'].replace('T00:00:00+08:00',''));
          dailyData.push(i['count']);
        }
        option.xAxis[0].data = xAxisData;
        option.series[0].data = dailyData;
        e.setOption(option);
      });
    }
  }
}
</script>
