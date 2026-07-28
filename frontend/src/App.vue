<template>
  <div class="app">
    <header class="header">
      <h1>DockerSphere</h1>
      <div class="header-info">
        <span class="container-count">{{ containers.length }} 容器</span>
        <div class="status">
          <span :class="['indicator', connected ? 'online' : 'offline']"></span>
          {{ connected ? '已连接' : '已断开' }}
        </div>
      </div>
    </header>

    <main class="main">
      <div class="container-list">
        <h2>容器列表</h2>
        <div v-if="loading" class="loading">加载中...</div>
        <div v-else-if="containers.length === 0" class="empty">未找到容器</div>
        <div v-else class="cards">
          <ContainerCard
            v-for="container in containers"
            :key="container.id"
            :container="container"
            @action="action"
          />
        </div>
      </div>

      <div class="event-log">
        <h2>实时事件</h2>
        <EventLog :events="events" />
      </div>
    </main>
  </div>
</template>

<script>
import axios from 'axios'
import ContainerCard from './components/ContainerCard.vue'
import EventLog from './components/EventLog.vue'

export default {
  name: 'App',
  components: { ContainerCard, EventLog },
  data() {
    return {
      containers: [],
      events: [],
      loading: false,
      connected: false,
      ws: null
    }
  },
  mounted() {
    this.fetchContainers()
    this.connectWebSocket()
    setInterval(() => this.fetchContainers(), 5000)
  },
  beforeUnmount() {
    if (this.ws) this.ws.close()
  },
  methods: {
    async fetchContainers() {
      try {
        this.loading = true
        const res = await axios.get('/api/v1/containers')
        if (res.data.code === 0) {
          this.containers = res.data.data || []
        }
      } catch (err) {
        console.error('获取容器失败:', err)
      } finally {
        this.loading = false
      }
    },
    async action(id, action) {
      try {
        const res = await axios.post(`/api/v1/containers/${id}/action`, { action })
        if (res.data.code === 0) {
          console.log(`任务 ${res.data.data.task_id} 已创建: ${action}`)
        }
      } catch (err) {
        console.error(`操作失败:`, err.response?.data || err)
      }
    },
    connectWebSocket() {
      const protocol = window.location.protocol === 'https:' ? 'wss:' : 'ws:'
      this.ws = new WebSocket(`${protocol}//${window.location.host}/ws`)

      this.ws.onopen = () => { this.connected = true }
      this.ws.onmessage = (event) => {
        try {
          const msg = JSON.parse(event.data)
          const data = msg.data || {}
          this.events.unshift({
            type: msg.event,
            summary: data.name || data.container_id || data.target || '',
            data: data,
            timestamp: new Date()
          })
          if (this.events.length > 50) this.events.pop()
          if (msg.event.startsWith('container.')) this.fetchContainers()
        } catch (err) { /* ignore */ }
      }
      this.ws.onclose = () => {
        this.connected = false
        setTimeout(() => this.connectWebSocket(), 3000)
      }
    }
  }
}
</script>

<style scoped>
.app { min-height: 100vh; }
.header {
  background: #2c3e50; color: white; padding: 1rem 2rem;
  display: flex; justify-content: space-between; align-items: center;
}
.header h1 { font-size: 1.5rem; }
.header-info { display: flex; align-items: center; gap: 1rem; }
.container-count { font-size: 0.875rem; color: #bdc3c7; }
.status { display: flex; align-items: center; gap: 0.5rem; font-size: 0.875rem; }
.indicator { width: 10px; height: 10px; border-radius: 50%; }
.indicator.online { background: #27ae60; }
.indicator.offline { background: #e74c3c; }
.main {
  display: grid; grid-template-columns: 2fr 1fr; gap: 2rem;
  padding: 2rem; max-width: 1400px; margin: 0 auto;
}
.container-list, .event-log {
  background: white; border-radius: 8px; padding: 1.5rem;
  box-shadow: 0 2px 4px rgba(0,0,0,0.1);
}
h2 { margin-bottom: 1rem; color: #2c3e50; }
.loading, .empty { text-align: center; padding: 2rem; color: #7f8c8d; }
.cards { display: grid; gap: 1rem; }
@media (max-width: 1024px) { .main { grid-template-columns: 1fr; } }
</style>
