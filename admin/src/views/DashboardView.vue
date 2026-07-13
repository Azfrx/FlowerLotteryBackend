<template>
  <div>
    <div class="page-title">
      <div><h2>Dashboard</h2><p>Core activity and system metrics</p></div>
      <el-date-picker v-model="date" type="date" />
    </div>
    <el-row :gutter="16">
      <el-col v-for="card in cards" :key="card.label" :xs="12" :sm="12" :lg="6">
        <el-card class="metric"><span>{{ card.label }}</span><strong>{{ card.value }}</strong><small>{{ card.tip }}</small></el-card>
      </el-col>
    </el-row>
    <el-row :gutter="16" class="charts">
      <el-col :xs="24" :lg="16"><el-card><div ref="trendEl" class="chart" /></el-card></el-col>
      <el-col :xs="24" :lg="8"><el-card><div ref="progressEl" class="chart" /></el-card></el-col>
    </el-row>
  </div>
</template>
<script setup lang="ts">
import { onBeforeUnmount, onMounted, ref } from 'vue'
import * as echarts from 'echarts'
const date = ref(new Date())
const cards = [
  { label: 'Lottery users', value: '1,286', tip: '+12.6% day over day' },
  { label: 'Executed draws', value: '18,420', tip: 'Day pool 72%' },
  { label: 'Petals consumed', value: '92,100', tip: 'Added to leaderboard' },
  { label: 'Rewards granted', value: '18,537', tip: '3 failed grants' },
]
const trendEl = ref<HTMLElement>()
const progressEl = ref<HTMLElement>()
let charts: echarts.ECharts[] = []
function resize() { charts.forEach(chart => chart.resize()) }
onMounted(() => {
  const trend = echarts.init(trendEl.value!)
  trend.setOption({ tooltip: { trigger: 'axis' }, legend: { data: ['Draws', 'Petals'] }, xAxis: { type: 'category', data: ['07-07','07-08','07-09','07-10','07-11','07-12','07-13'] }, yAxis: { type: 'value' }, series: [{ name: 'Draws', type: 'line', smooth: true, data: [8200,9100,10500,12800,14300,16900,18420] }, { name: 'Petals', type: 'line', smooth: true, data: [41000,45500,52500,64000,71500,84500,92100] }] })
  const progress = echarts.init(progressEl.value!)
  progress.setOption({ tooltip: { trigger: 'item' }, series: [{ type: 'pie', radius: ['45%','70%'], data: [{ name: '1-5', value: 42 }, { name: '6-11', value: 31 }, { name: '12-17', value: 19 }, { name: '18', value: 8 }] }] })
  charts = [trend, progress]
  window.addEventListener('resize', resize)
})
onBeforeUnmount(() => { window.removeEventListener('resize', resize); charts.forEach(chart => chart.dispose()) })
</script>