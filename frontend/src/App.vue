<template>
  <div class="app">
    <header class="header">
      <h1>DockerSphere</h1>
      <div class="status">
        <span :class="['indicator', connected ? 'online' : 'offline']"></span>
        {{ connected ? 'Connected' : 'Disconnected' }}
      </div>
    </header>

    <main class="main">
      <div class="container-list">
        <h2>Containers</h2>
        <div v-if="loading" class="loading">Loading...</div>
        <div v-else-if="containers.length === 0" class="empty">No containers found</div>
        <div v-else class="cards">
          <div v-for="container in containers" :key="container.id" class="card">
            <div class="card-header">
              <span class="name">{{ container.name || container.id }}</span>
              <span :class="['badge', container.state]">{{ container.state }}</span>
            </div>
            <div class="card-body">
              <div class="info">
                <span class="label">Image:</span>
                <span class="value">{{ container.image }}</span>
              </div>
              <div class="info">
                <span class="label">Status:</span>
                <span class="value">{{ container.status }}</span>
              </div>
              <div class="info">
                <span class="label">ID:</span>
                <span class="value mono">{{ container.id }}</span>
              </div>
            </div>
            <div class="card-actions">
              <button @click="action(container.id, 'start')" :disabled="container.state === 'running'">Start</button>
              <button @click="action(container.id, 'stop')" :disabled="container.state !== 'running'">Stop</button>
              <button @click="action(container.id, 'restart')">Restart</button>
              <button @click="action(container.id, 'remove')" class="danger">Remove</button>
            </div>
          </div>
        </div>
      </div>

      <div class="event-log">
        <h2>Recent Events</h2>
        <div class="events">
          <div v-for="(event, index) in events" :key="index" class="event">
            <span class="time">{{ formatTime(event.timestamp) }}</span>
            <span class="type">{{ event.type }}</span>
            <span class="data">{{ JSON.stringify(event.data) }}</span>
          </div>
          <div v-if="events.length === 0" class="empty">No events yet</div>
        </div>
      </div>
    </main>
  </div>
</template>

<script>
import axios from 'axios'

export default {
  name: 'App',
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
    if (this.ws) {
      this.ws.close()
    }
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
        console.error('Failed to fetch containers:', err)
      } finally {
        this.loading = false
      }
    },
    async action(id, action) {
      try {
        const res = await axios.post(`/api/v1/containers/${id}/action`, { action })
        if (res.data.code === 0) {
          console.log(`Task ${res.data.data.task_id} created for ${action} on ${id}`)
        }
      } catch (err) {
        console.error(`Failed to ${action} container:`, err)
      }
    },
    connectWebSocket() {
      const protocol = window.location.protocol === 'https:' ? 'wss:' : 'ws:'
      const wsUrl = `${protocol}//${window.location.host}/ws`
      
      this.ws = new WebSocket(wsUrl)
      
      this.ws.onopen = () => {
        this.connected = true
        console.log('WebSocket connected')
      }
      
      this.ws.onmessage = (event) => {
        try {
          const msg = JSON.parse(event.data)
          this.events.unshift({
            type: msg.event,
            data: msg.data,
            timestamp: new Date()
          })
          if (this.events.length > 50) {
            this.events.pop()
          }
          // Refresh containers on state changes
          if (msg.event.startsWith('container.')) {
            this.fetchContainers()
          }
        } catch (err) {
          console.error('Failed to parse WebSocket message:', err)
        }
      }
      
      this.ws.onclose = () => {
        this.connected = false
        console.log('WebSocket disconnected, reconnecting...')
        setTimeout(() => this.connectWebSocket(), 3000)
      }
      
      this.ws.onerror = (err) => {
        console.error('WebSocket error:', err)
      }
    },
    formatTime(date) {
      return date.toLocaleTimeString()
    }
  }
}
</script>

<style scoped>
.app {
  min-height: 100vh;
}

.header {
  background: #2c3e50;
  color: white;
  padding: 1rem 2rem;
  display: flex;
  justify-content: space-between;
  align-items: center;
}

.header h1 {
  font-size: 1.5rem;
}

.status {
  display: flex;
  align-items: center;
  gap: 0.5rem;
}

.indicator {
  width: 10px;
  height: 10px;
  border-radius: 50%;
}

.indicator.online {
  background: #27ae60;
}

.indicator.offline {
  background: #e74c3c;
}

.main {
  display: grid;
  grid-template-columns: 2fr 1fr;
  gap: 2rem;
  padding: 2rem;
  max-width: 1400px;
  margin: 0 auto;
}

.container-list, .event-log {
  background: white;
  border-radius: 8px;
  padding: 1.5rem;
  box-shadow: 0 2px 4px rgba(0,0,0,0.1);
}

h2 {
  margin-bottom: 1rem;
  color: #2c3e50;
}

.loading, .empty {
  text-align: center;
  padding: 2rem;
  color: #7f8c8d;
}

.cards {
  display: grid;
  gap: 1rem;
}

.card {
  border: 1px solid #e0e0e0;
  border-radius: 6px;
  overflow: hidden;
}

.card-header {
  background: #ecf0f1;
  padding: 0.75rem 1rem;
  display: flex;
  justify-content: space-between;
  align-items: center;
}

.name {
  font-weight: 600;
  color: #2c3e50;
}

.badge {
  padding: 0.25rem 0.75rem;
  border-radius: 12px;
  font-size: 0.75rem;
  font-weight: 600;
  text-transform: uppercase;
}

.badge.running {
  background: #27ae60;
  color: white;
}

.badge.exited, .badge.stopped {
  background: #95a5a6;
  color: white;
}

.badge.created, .badge.paused {
  background: #f39c12;
  color: white;
}

.card-body {
  padding: 1rem;
}

.info {
  display: flex;
  gap: 0.5rem;
  margin-bottom: 0.5rem;
}

.label {
  font-weight: 600;
  color: #7f8c8d;
  min-width: 60px;
}

.value {
  color: #2c3e50;
}

.mono {
  font-family: 'Courier New', monospace;
  font-size: 0.875rem;
}

.card-actions {
  padding: 0.75rem 1rem;
  background: #f8f9fa;
  display: flex;
  gap: 0.5rem;
}

.card-actions button {
  padding: 0.5rem 1rem;
  border: none;
  border-radius: 4px;
  cursor: pointer;
  font-size: 0.875rem;
  transition: all 0.2s;
}

.card-actions button:hover:not(:disabled) {
  opacity: 0.9;
}

.card-actions button:disabled {
  opacity: 0.5;
  cursor: not-allowed;
}

.card-actions button:not(.danger) {
  background: #3498db;
  color: white;
}

.card-actions .danger {
  background: #e74c3c;
  color: white;
}

.events {
  max-height: 600px;
  overflow-y: auto;
}

.event {
  padding: 0.5rem;
  border-bottom: 1px solid #ecf0f1;
  font-size: 0.875rem;
}

.event:last-child {
  border-bottom: none;
}

.time {
  color: #95a5a6;
  margin-right: 0.5rem;
}

.type {
  font-weight: 600;
  color: #2c3e50;
  margin-right: 0.5rem;
}

.data {
  color: #7f8c8d;
  font-family: monospace;
  font-size: 0.75rem;
}

@media (max-width: 1024px) {
  .main {
    grid-template-columns: 1fr;
  }
}
</style>
