<template>
  <div class="app">
    <!-- Sidebar -->
    <aside class="sidebar">
      <div class="logo">
        <span class="logo-icon">◈</span>
        <span class="logo-text">Sphere</span>
      </div>
      <nav class="nav">
        <a v-for="tab in tabs" :key="tab.id"
           :class="['nav-item', { active: activeTab === tab.id }]"
           @click="activeTab = tab.id">
          <span class="nav-icon">{{ tab.icon }}</span>
          <span class="nav-label">{{ tab.label }}</span>
          <span v-if="tab.badge" class="nav-badge">{{ tab.badge }}</span>
        </a>
      </nav>
      <div class="sidebar-footer">
        <div class="connection-status">
          <span :class="['dot', connected ? 'online' : 'offline']"></span>
          {{ connected ? '已连接' : '已断开' }}
        </div>
        <div class="version">v0.3.0</div>
      </div>
    </aside>

    <!-- Main -->
    <div class="main-area">
      <header class="topbar">
        <h1 class="page-title">{{ currentTab.label }}</h1>
        <div class="topbar-actions">
          <span class="container-count">{{ containers.length }} 容器</span>
          <button class="btn-refresh" @click="fetchContainers" title="刷新">↻</button>
        </div>
      </header>

      <main class="content">
        <!-- Dashboard -->
        <div v-if="activeTab === 'dashboard'" class="tab-content">
          <div class="stats-grid">
            <div class="stat-card">
              <div class="stat-value">{{ stats.total }}</div>
              <div class="stat-label">全部容器</div>
            </div>
            <div class="stat-card running">
              <div class="stat-value">{{ stats.running }}</div>
              <div class="stat-label">运行中</div>
            </div>
            <div class="stat-card stopped">
              <div class="stat-value">{{ stats.stopped }}</div>
              <div class="stat-label">已停止</div>
            </div>
            <div class="stat-card">
              <div class="stat-value">{{ stats.images }}</div>
              <div class="stat-label">镜像数</div>
            </div>
          </div>

          <div class="section">
            <h3 class="section-title">最近事件</h3>
            <div class="mini-events">
              <div v-for="(e, i) in events.slice(0, 8)" :key="i" class="mini-event">
                <span :class="['event-dot', e.type.split('.')[0]]"></span>
                <span class="event-text">{{ e.summary || e.type }}</span>
                <span class="event-time">{{ formatTime(e.timestamp) }}</span>
              </div>
              <div v-if="events.length === 0" class="empty-state">暂无事件</div>
            </div>
          </div>
        </div>

        <!-- Containers -->
        <div v-if="activeTab === 'containers'" class="tab-content">
          <div v-if="loading" class="loading">加载中...</div>
          <div v-else-if="containers.length === 0" class="empty-state">未找到容器</div>
          <div v-else class="container-grid">
            <ContainerCard
              v-for="c in containers" :key="c.id"
              :container="c"
              @action="action"
              @click="selectedContainer = c"
            />
          </div>
        </div>

        <!-- Container Detail -->
        <div v-if="activeTab === 'detail' && selectedContainer" class="tab-content">
          <button class="btn-back" @click="activeTab = 'containers'">← 返回列表</button>
          <div class="detail-header">
            <h2>{{ selectedContainer.name }}</h2>
            <span :class="['badge-lg', selectedContainer.state]">{{ selectedContainer.state }}</span>
          </div>
          <div class="detail-grid">
            <div class="detail-card">
              <div class="detail-card-title">基本信息</div>
              <div class="detail-row"><span class="dl">ID</span><span class="dd mono">{{ selectedContainer.id }}</span></div>
              <div class="detail-row"><span class="dl">镜像</span><span class="dd">{{ selectedContainer.image }}</span></div>
              <div class="detail-row"><span class="dl">状态</span><span class="dd">{{ selectedContainer.status }}</span></div>
              <div class="detail-row"><span class="dl">创建</span><span class="dd">{{ selectedContainer.created }}</span></div>
            </div>
            <div class="detail-card">
              <div class="detail-card-title">端口映射</div>
              <div v-if="selectedContainer.ports && selectedContainer.ports.length > 0">
                <div v-for="p in selectedContainer.ports" :key="p.private_port" class="detail-row">
                  <span class="dl">{{ p.type }}</span>
                  <span class="dd">{{ p.public_port || '?' }} → {{ p.private_port }}</span>
                </div>
              </div>
              <div v-else class="empty-state-sm">无端口映射</div>
            </div>
            <div class="detail-card detail-card-wide">
              <div class="detail-card-title">操作</div>
              <div class="detail-actions">
                <button @click="action(selectedContainer.id, 'start')" :disabled="selectedContainer.state === 'running'" class="btn btn-start">▶ 启动</button>
                <button @click="action(selectedContainer.id, 'stop')" :disabled="selectedContainer.state !== 'running'" class="btn btn-stop">■ 停止</button>
                <button @click="action(selectedContainer.id, 'restart')" class="btn btn-restart">↻ 重启</button>
                <button @click="action(selectedContainer.id, 'remove')" class="btn btn-remove">✕ 删除</button>
              </div>
            </div>
          </div>
        </div>

        <!-- Events -->
        <div v-if="activeTab === 'events'" class="tab-content">
          <div class="events-panel">
            <div v-for="(e, i) in events" :key="i" class="event-row">
              <span :class="['event-badge', e.type.split('.')[0]]">{{ e.type.split('.')[0] }}</span>
              <span class="event-msg">{{ e.type }} {{ e.summary }}</span>
              <span class="event-ts">{{ formatTime(e.timestamp) }}</span>
            </div>
            <div v-if="events.length === 0" class="empty-state">暂无事件</div>
          </div>
        </div>
      </main>
    </div>
  </div>
</template>

<script>
import axios from 'axios'
import ContainerCard from './components/ContainerCard.vue'

export default {
  name: 'App',
  components: { ContainerCard },
  data() {
    return {
      activeTab: 'dashboard',
      containers: [],
      events: [],
      selectedContainer: null,
      loading: false,
      connected: false,
      ws: null,
      tabs: [
        { id: 'dashboard', label: '仪表盘', icon: '◉' },
        { id: 'containers', label: '容器', icon: '▣' },
        { id: 'events', label: '事件', icon: '⚡' },
      ]
    }
  },
  computed: {
    currentTab() {
      return this.tabs.find(t => t.id === this.activeTab) || this.tabs[0]
    },
    stats() {
      const total = this.containers.length
      const running = this.containers.filter(c => c.state === 'running').length
      const stopped = this.containers.filter(c => c.state !== 'running').length
      const images = new Set(this.containers.map(c => c.image)).size
      return { total, running, stopped, images }
    }
  },
  mounted() {
    this.fetchContainers()
    this.connectWebSocket()
  },
  beforeUnmount() {
    if (this.ws) this.ws.close()
  },
  methods: {
    async fetchContainers() {
      try {
        this.loading = true
        const res = await axios.get('/api/v1/containers')
        if (res.data.code === 0) this.containers = res.data.data || []
      } catch (err) { console.error(err)
      } finally { this.loading = false }
    },
    async action(id, act) {
      try {
        await axios.post(`/api/v1/containers/${id}/action`, { action: act })
      } catch (err) { console.error(err) }
    },
    connectWebSocket() {
      const proto = window.location.protocol === 'https:' ? 'wss:' : 'ws:'
      this.ws = new WebSocket(`${proto}//${window.location.host}/ws`)
      this.ws.onopen = () => { this.connected = true }
      this.ws.onmessage = (e) => {
        try {
          const msg = JSON.parse(e.data)
          this.events.unshift({ type: msg.event, summary: (msg.data || {}).name || '', timestamp: new Date() })
          if (this.events.length > 100) this.events.pop()
          if (msg.event.startsWith('container.')) this.fetchContainers()
        } catch (_) {}
      }
      this.ws.onclose = () => {
        this.connected = false
        setTimeout(() => this.connectWebSocket(), 5000)
      }
    },
    formatTime(d) { return d.toLocaleTimeString() }
  }
}
</script>

<style>
* { margin: 0; padding: 0; box-sizing: border-box; }
body { font-family: -apple-system, BlinkMacSystemFont, 'Segoe UI', Roboto, 'Helvetica Neue', sans-serif; background: #f0f2f5; color: #1a1a2e; }
.app { display: flex; min-height: 100vh; }

/* Sidebar */
.sidebar {
  width: 220px; background: #1a1a2e; color: #e0e0e0;
  display: flex; flex-direction: column; position: fixed; top: 0; left: 0; bottom: 0; z-index: 100;
}
.logo { padding: 1.5rem 1.25rem; display: flex; align-items: center; gap: 0.5rem; border-bottom: 1px solid rgba(255,255,255,0.08); }
.logo-icon { font-size: 1.5rem; color: #6c63ff; }
.logo-text { font-size: 1.25rem; font-weight: 700; letter-spacing: 0.5px; }
.nav { flex: 1; padding: 0.75rem 0; }
.nav-item { display: flex; align-items: center; gap: 0.75rem; padding: 0.75rem 1.25rem; cursor: pointer; color: #a0a0b8; transition: all 0.15s; font-size: 0.9rem; }
.nav-item:hover { background: rgba(108,99,255,0.08); color: #e0e0e0; }
.nav-item.active { background: rgba(108,99,255,0.15); color: #fff; border-right: 3px solid #6c63ff; }
.nav-icon { font-size: 1.1rem; width: 24px; text-align: center; }
.nav-badge { margin-left: auto; background: #6c63ff; color: #fff; font-size: 0.7rem; padding: 0.15rem 0.5rem; border-radius: 10px; }
.sidebar-footer { padding: 1rem 1.25rem; border-top: 1px solid rgba(255,255,255,0.08); font-size: 0.8rem; }
.connection-status { display: flex; align-items: center; gap: 0.5rem; margin-bottom: 0.25rem; }
.dot { width: 8px; height: 8px; border-radius: 50%; }
.dot.online { background: #4caf50; box-shadow: 0 0 6px rgba(76,175,80,0.5); }
.dot.offline { background: #f44336; }
.version { color: #666; }

/* Main area */
.main-area { margin-left: 220px; flex: 1; display: flex; flex-direction: column; }
.topbar {
  background: #fff; padding: 1rem 2rem; display: flex; justify-content: space-between; align-items: center;
  border-bottom: 1px solid #e8e8e8; position: sticky; top: 0; z-index: 50;
}
.page-title { font-size: 1.25rem; font-weight: 600; }
.topbar-actions { display: flex; align-items: center; gap: 1rem; }
.container-count { color: #888; font-size: 0.85rem; }
.btn-refresh { background: none; border: 1px solid #ddd; border-radius: 6px; padding: 0.4rem 0.6rem; cursor: pointer; font-size: 1.1rem; color: #666; transition: all 0.15s; }
.btn-refresh:hover { background: #f5f5f5; color: #333; }

/* Content */
.content { padding: 1.5rem 2rem; flex: 1; }

/* Stats */
.stats-grid { display: grid; grid-template-columns: repeat(4, 1fr); gap: 1rem; margin-bottom: 2rem; }
.stat-card {
  background: #fff; border-radius: 12px; padding: 1.5rem; text-align: center;
  box-shadow: 0 1px 3px rgba(0,0,0,0.06); transition: transform 0.15s;
}
.stat-card:hover { transform: translateY(-2px); }
.stat-value { font-size: 2rem; font-weight: 700; color: #1a1a2e; }
.stat-card.running .stat-value { color: #4caf50; }
.stat-card.stopped .stat-value { color: #f44336; }
.stat-label { color: #888; font-size: 0.85rem; margin-top: 0.25rem; }

/* Sections */
.section { background: #fff; border-radius: 12px; padding: 1.5rem; box-shadow: 0 1px 3px rgba(0,0,0,0.06); margin-bottom: 1.5rem; }
.section-title { font-size: 1rem; font-weight: 600; margin-bottom: 1rem; color: #333; }

/* Mini events */
.mini-events { display: flex; flex-direction: column; gap: 0.5rem; }
.mini-event { display: flex; align-items: center; gap: 0.75rem; padding: 0.5rem 0; border-bottom: 1px solid #f5f5f5; font-size: 0.85rem; }
.mini-event:last-child { border-bottom: none; }
.event-dot { width: 8px; height: 8px; border-radius: 50%; flex-shrink: 0; }
.event-dot.container { background: #4caf50; }
.event-dot.task { background: #2196f3; }
.event-dot.audit { background: #ff9800; }
.event-text { flex: 1; color: #555; }
.event-time { color: #aaa; font-size: 0.8rem; }

/* Container grid */
.container-grid { display: grid; grid-template-columns: repeat(auto-fill, minmax(340px, 1fr)); gap: 1rem; }

/* Detail */
.btn-back { background: none; border: none; color: #6c63ff; cursor: pointer; font-size: 0.9rem; padding: 0; margin-bottom: 1rem; }
.detail-header { display: flex; align-items: center; gap: 1rem; margin-bottom: 1.5rem; }
.detail-header h2 { font-size: 1.5rem; font-weight: 600; }
.badge-lg { padding: 0.25rem 0.75rem; border-radius: 12px; font-size: 0.8rem; font-weight: 600; text-transform: uppercase; }
.badge-lg.running { background: #e8f5e9; color: #2e7d32; }
.badge-lg.exited, .badge-lg.stopped { background: #fbe9e7; color: #c62828; }
.badge-lg.created { background: #e3f2fd; color: #1565c0; }
.detail-grid { display: grid; grid-template-columns: 1fr 1fr; gap: 1rem; }
.detail-card {
  background: #fff; border-radius: 12px; padding: 1.5rem; box-shadow: 0 1px 3px rgba(0,0,0,0.06);
}
.detail-card-wide { grid-column: 1 / -1; }
.detail-card-title { font-weight: 600; font-size: 0.9rem; color: #888; text-transform: uppercase; letter-spacing: 0.5px; margin-bottom: 1rem; }
.detail-row { display: flex; padding: 0.4rem 0; font-size: 0.9rem; border-bottom: 1px solid #f5f5f5; }
.detail-row:last-child { border-bottom: none; }
.dl { width: 60px; color: #888; flex-shrink: 0; }
.dd { color: #333; }
.mono { font-family: 'SF Mono', 'Fira Code', monospace; font-size: 0.85rem; }
.detail-actions { display: flex; gap: 0.75rem; flex-wrap: wrap; }

/* Buttons */
.btn { padding: 0.6rem 1.2rem; border: none; border-radius: 8px; cursor: pointer; font-size: 0.85rem; font-weight: 500; transition: all 0.15s; }
.btn:hover { transform: translateY(-1px); }
.btn:disabled { opacity: 0.4; cursor: not-allowed; transform: none; }
.btn-start { background: #e8f5e9; color: #2e7d32; }
.btn-stop { background: #fbe9e7; color: #c62828; }
.btn-restart { background: #e3f2fd; color: #1565c0; }
.btn-remove { background: #fce4ec; color: #ad1457; }

/* Events panel */
.events-panel { background: #fff; border-radius: 12px; padding: 1rem; box-shadow: 0 1px 3px rgba(0,0,0,0.06); }
.event-row { display: flex; align-items: center; gap: 0.75rem; padding: 0.6rem 0.5rem; border-bottom: 1px solid #f5f5f5; font-size: 0.85rem; }
.event-row:last-child { border-bottom: none; }
.event-badge { padding: 0.15rem 0.5rem; border-radius: 4px; font-size: 0.75rem; font-weight: 600; text-transform: uppercase; flex-shrink: 0; }
.event-badge.container { background: #e8f5e9; color: #2e7d32; }
.event-badge.task { background: #e3f2fd; color: #1565c0; }
.event-badge.audit { background: #fff3e0; color: #e65100; }
.event-msg { flex: 1; color: #555; }
.event-ts { color: #aaa; font-size: 0.8rem; flex-shrink: 0; }

/* Shared */
.loading, .empty-state { text-align: center; padding: 3rem; color: #999; font-size: 0.9rem; }
.empty-state-sm { text-align: center; padding: 1rem; color: #999; font-size: 0.85rem; }
</style>
