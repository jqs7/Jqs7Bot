<template lang='jade'>
#pie
</template>

<style lang='stylus'>
#pie
  width 300px
  height 300px
  margin-left 100px
</style>

<script>
var echarts = require('echarts');
  require('echarts/chart/pie');
var option = {
  title: {
    text: '',
    subtext: '',
    x: 'center',
  },
  tooltip: {
    trigger: 'item',
    formatter: "{b} : {c} ({d}%)"
  },
  series: [
    {
      name:'访问来源',
      type:'pie',
      radius : '55%',
      center: ['50%', '60%'],
      data:[
      ]
    }
  ]
};
var e, v;
export default {
  ready (){
    v = this;
    e = echarts.init(document.getElementById('pie'));
    document.getElementById('pie').style.display = 'none';
  },
  methods: {
    show: function(date, total){
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
      });
    }
  }
}
</script>
