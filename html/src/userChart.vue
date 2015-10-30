<template lang='jade'>
#chart
</template>

<style lang='stylus'>
#chart
  height 400px
</style>

<script>
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
var e,v;
var xAxisData = new Array();
var dailyData = new Array();
export default{
  ready (){
    v = this;
    e = echarts.init(document.getElementById('chart'));
    window.onresize = e.resize;
  },
  methods: {
    show: function(userName) {
      v.$http.get('/api/user/' + escape(userName),function(data,status,req){
        for (var i in data['result']){
          xAxisData.push(data['result'][i]['date'].replace('T00:00:00+08:00',''));
          dailyData.push(data['result'][i]['count']);
        }
        option.xAxis[0].data = xAxisData;
        option.series[0].data = dailyData;
        e.setOption(option);
      });
    }
  }
}
</script>
