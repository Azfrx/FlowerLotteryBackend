<template>
  <div>
    <div class="page-title">
      <div><h2>{{ title }}</h2><p>数据来自真实管理 API</p></div>
      <el-button @click="load">刷新</el-button>
    </div>
    <el-card>
      <el-table :data="rows" v-loading="loading" border stripe>
        <el-table-column v-for="column in columns" :key="column" :prop="column" :label="column" min-width="140">
          <template #default="scope">{{ format(scope.row[column]) }}</template>
        </el-table-column>
      </el-table>
      <el-empty v-if="!loading && !rows.length" description="暂无数据" />
    </el-card>
  </div>
</template>
<script setup lang="ts">
import { computed, onMounted, ref, watch } from 'vue'
import { useRoute } from 'vue-router'
import { ElMessage } from 'element-plus'
import { http } from '@/api/http'
const route = useRoute()
const config: Record<string, { title: string; url: string }> = {
  users: { title: '用户管理', url: '/users' },
  assets: { title: '资产流水', url: '/asset-transactions' },
  lottery: { title: '抽奖记录', url: '/lottery-orders' },
  rewards: { title: '奖励记录', url: '/rewards' },
  flowers: { title: '花朵进度', url: '/rounds' },
  leaderboard: { title: '榜单管理', url: '/leaderboard' },
  activities: { title: '活动管理', url: '/lottery-orders' },
  pools: { title: '奖池与概率', url: '/lottery-orders' },
  system: { title: '系统治理', url: '/users' },
}
const key = computed(() => String(route.path.split('/').pop()))
const current = computed(() => config[key.value] ?? config.users)
const title = computed(() => current.value.title)
const rows = ref<Record<string, unknown>[]>([])
const loading = ref(false)
const columns = computed(() => rows.value.length ? Object.keys(rows.value[0]).filter(value => !['PasswordHash', 'DeletedAt'].includes(value)) : [])
function format(value: unknown) {
  if (value === null || value === undefined) return '-'
  if (typeof value === 'object') return JSON.stringify(value)
  return String(value)
}
async function load() {
  loading.value = true
  try {
    const result: any = await http.get(current.value.url)
    rows.value = Array.isArray(result.data) ? result.data : Array.isArray(result) ? result : []
  } catch {
    ElMessage.error('加载失败')
    rows.value = []
  } finally {
    loading.value = false
  }
}
watch(key, load)
onMounted(load)
</script>