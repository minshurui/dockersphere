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
            <div class="stat-card"><div class="stat-value">{{ stats.total }}</div><div class="stat-label">全部容器</div></div>
            <div class="stat-card running"><div class="stat-value">{{ stats.running }}</div><div class="stat-label">运行中</div></div>
            <div class="stat-card stopped"><div class="stat-value">{{ stats.stopped }}</div><div class="stat-label">已停止</div></div>
            <div class="stat-card"><div class="stat-value">{{ stats.projects }}</div><div class="stat-label">Compose 项目</div></div>
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
          <div v-else class="container-groups">
            <div v-for="group in composeGroups" :key="group.project" class="compose-group">
              <div class="compose-header">
                <span class="compose-icon">▦</span>
                <span class="compose-name">{{ group.project }}</span>
                <span class="compose-count">{{ group.containers.length }} 服务</span>
                <span class="compose-file" :title="group.configFiles">{{ group.configFiles ? group.configFiles.split('/').pop() : '' }}</span>
              </div>
              <div class="compose-containers">
                <ContainerCard
                  v-for="c in group.containers" :key="c.id"
                  :container="c"
                  @action="action"
                  @click="openDetail(c)"
                />
              </div>
            </div>
            <!-- 非 Compose 容器 -->
            <div v-if="standaloneContainers.length > 0" class="compose-group">
              <div class="compose-header">
                <span class="compose-icon">◯</span>
                <span class="compose-name">独立容器</span>
                <span class="compose-count">{{ standaloneContainers.length }}</span>
              </div>
              <div class="compose-containers">
                <ContainerCard
                  v-for="c in standaloneContainers" :key="c.id"
                  :container="c"
                  @action="action"
                  @click="openDetail(c)"
                />
              </div>
            </div>
          </div>
        </div>

        <!-- Container Detail -->
        <div v-if="activeTab === 'detail' && selectedContainer" class="tab-content">
          <button class="btn-back" @click="activeTab = 'containers'">← 返回列表</button>
          <div class="detail-header">
            <div class="detail-title-group">
              <span v-if="selectedContainer.composeProject" class="compose-tag">{{ selectedContainer.composeProject }}/{{ selectedContainer.composeService }}</span>
              <h2>{{ selectedContainer.name }}</h2>
            </div>
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
                <div v-for="p in mappedPorts(selectedContainer.ports)" :key="p.key" class="detail-row port-row">
                  <span class="port-type">{{ p.type }}</span>
                  <span class="port-mapping">
                    <span class="port-binding">{{ p.binding }}</span>
                    <span class="port-arrow">→</span>
                    <span class="port-private">{{ p.private }}</span>
                  </span>
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
            <div class="detail-card detail-card-wide">
              <div class="detail-card-title">日志 <button v-if="showLogs" @click="showLogs=false" class="btn-sm">收起</button></div>
              <div v-if="showLogs" class="log-viewer" ref="logViewer">
                <div v-for="(line, i) in logs" :key="i" class="log-line">{{ line }}</div>
                <div v-if="logs.length === 0" class="empty-state-sm">无日志</div>
              </div>
              <button v-if="!showLogs" @click="fetchLogs" class="btn btn-log">📋 查看日志</button>
            </div>
            <div class="detail-card detail-card-wide">
              <div class="detail-card-title">命令执行</div>
              <div class="exec-bar">
                <input v-model="execCmd" @keyup.enter="runExec" placeholder="输入命令，例如: ps aux" class="exec-input" />
                <button @click="runExec" class="btn btn-exec">运行</button>
              </div>
              <div v-if="execOutput" class="exec-output"><pre>{{ execOutput }}</pre></div>
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
      showLogs: false,
      logs: [],
      execCmd: '',
      execOutput: '',
      tabs: [
        { id: 'dashboard', label: '仪表盘', icon: '◉' },
        { id: 'containers', label: '容器', icon: '▣', badge: null },
        { id: 'events', label: '事件', icon: '⚡' },
      ]
    }
  },
  computed: {
    currentTab() { return this.tabs.find(t => t.id === this.activeTab) || this.tabs[0] },
    stats() {
      const total = this.containers.length
      const running = this.containers.filter(c => c.state === 'running').length
      const stopped = this.containers.filter(c => c.state !== 'running').length
      const projects = new Set(this.containers.map(c => c.labels?.['com.docker.compose.project'] || '_')).size
      return { total, running, stopped, projects }
    },
    composeGroups() {
      const groups = {}
      for (const c of this.containers) {
        const project = c.labels?.['com.docker.compose.project']
        if (project) {
          if (!groups[project]) {
            groups[project] = {
              project,
              configFiles: c.labels?.['com.docker.compose.project.config_files'] || '',
              containers: []
            }
          }
          groups[project].containers.push({
            ...c,
            composeProject: project,
            composeService: c.labels?.['com.docker.compose.service'] || ''
          })
        }
      }
      return Object.values(groups).sort((a, b) => a.project.localeCompare(b.project))
    },
    standaloneContainers() {
      return this.containers
        .filter(c => !c.labels?.['com.docker.compose.project'])
        .map(c => ({ ...c, composeProject: null, composeService: null }))
    }
  },
  mounted() {
    this.fetchContainers()
    this.connectWebSocket()
  },
  beforeUnmount() { if (this.ws) this.ws.close() },
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
      try { await axios.post(`/api/v1/containers/${id}/action`, { action: act })
      } catch (err) { console.error(err) }
    },
    openDetail(c) {
      this.selectedContainer = c
      this.activeTab = 'detail'
      this.showLogs = false
      this.logs = []
      this.execOutput = ''
    },
    async fetchLogs() {
      this.showLogs = true
      try {
        const res = await axios.get(`/api/v1/containers/${this.selectedContainer.id}/logs?tail=50`)
        if (res.data.code === 0) this.logs = res.data.data || []
      } catch (err) { this.logs = ['获取日志失败'] }
    },
    async runExec() {
      if (!this.execCmd.trim()) return
      this.execOutput = '执行中...'
      try {
        const parts = this.execCmd.trim().split(/\s+/)
        const res = await axios.post(`/api/v1/containers/${this.selectedContainer.id}/exec`, { cmd: parts[0], args: parts.slice(1) })
        if (res.data.code === 0) {
          this.execOutput = res.data.data.replace(/[\x00-\x08\x0e-\x1f]/g, '').trim()
        }
      } catch (err) { this.execOutput = '执行失败: ' + (err.response?.data?.message || err.message) }
    },
    mappedPorts(ports) {
      const seen = new Set()
      return ports.filter(p => {
        const key = `${p.private_port}-${p.type}`
        if (seen.has(key)) return false
        seen.add(key)
        return true
      }).map(p => ({
        key: `${p.private_port}-${p.type}`,
        type: p.type.toUpperCase(),
        binding: p.public_port ? `${p.ip || '0.0.0.0'}:${p.public_port}` : '未映射',
        private: `${p.private_port}`
      }))
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
  width: 200px; background: #1a1a2e; color: #e0e0e0;
  display: flex; flex-direction: column; position: fixed; top: 0; left: 0; bottom: 0; z-index: 100;
}
.logo { padding: 1.25rem 1.25rem; display: flex; align-items: center; gap: 0.5rem; border-bottom: 1px solid rgba(255,255,255,0.08); }
.logo-icon { font-size: 1.3rem; color: #6c63ff; }
.logo-text { font-size: 1.15rem; font-weight: 700; }
.nav { flex: 1; padding: 0.5rem 0; }
.nav-item { display: flex; align-items: center; gap: 0.75rem; padding: 0.65rem 1.25rem; cursor: pointer; color: #a0a0b8; transition: all 0.15s; font-size: 0.85rem; }
.nav-item:hover { background: rgba(108,99,255,0.08); color: #e0e0e0; }
.nav-item.active { background: rgba(108,99,255,0.15); color: #fff; border-right: 3px solid #6c63ff; }
.nav-icon { font-size: 1rem; width: 20px; text-align: center; }
.sidebar-footer { padding: 0.75rem 1.25rem; border-top: 1px solid rgba(255,255,255,0.08); font-size: 0.75rem; }
.connection-status { display: flex; align-items: center; gap: 0.5rem; margin-bottom: 0.25rem; }
.dot { width: 8px; height: 8px; border-radius: 50%; }
.dot.online { background: #4caf50; box-shadow: 0 0 6px rgba(76,175,80,0.5); }
.dot.offline { background: #f44336; }
.version { color: #555; }

/* Main */
.main-area { margin-left: 200px; flex: 1; display: flex; flex-direction: column; }
.topbar { background: #fff; padding: 0.75rem 1.5rem; display: flex; justify-content: space-between; align-items: center; border-bottom: 1px solid #e8e8e8; position: sticky; top: 0; z-index: 50; }
.page-title { font-size: 1.15rem; font-weight: 600; }
.topbar-actions { display: flex; align-items: center; gap: 1rem; }
.container-count { color: #888; font-size: 0.85rem; }
.btn-refresh { background: none; border: 1px solid #ddd; border-radius: 6px; padding: 0.35rem 0.5rem; cursor: pointer; font-size: 1rem; color: #666; }
.btn-refresh:hover { background: #f5f5f5; }
.content { padding: 1.25rem 1.5rem; flex: 1; }

/* Stats */
.stats-grid { display: grid; grid-template-columns: repeat(4, 1fr); gap: 0.75rem; margin-bottom: 1.5rem; }
.stat-card { background: #fff; border-radius: 10px; padding: 1.25rem; text-align: center; box-shadow: 0 1px 3px rgba(0,0,0,0.06); }
.stat-value { font-size: 1.75rem; font-weight: 700; color: #1a1a2e; }
.stat-card.running .stat-value { color: #4caf50; }
.stat-card.stopped .stat-value { color: #f44336; }
.stat-label { color: #888; font-size: 0.8rem; margin-top: 0.2rem; }

/* Sections */
.section { background: #fff; border-radius: 10px; padding: 1.25rem; box-shadow: 0 1px 3px rgba(0,0,0,0.06); }
.section-title { font-size: 0.9rem; font-weight: 600; margin-bottom: 0.75rem; color: #333; }

/* Mini events */
.mini-events { display: flex; flex-direction: column; gap: 0.35rem; }
.mini-event { display: flex; align-items: center; gap: 0.5rem; padding: 0.35rem 0; border-bottom: 1px solid #f5f5f5; font-size: 0.8rem; }
.mini-event:last-child { border-bottom: none; }
.event-dot { width: 7px; height: 7px; border-radius: 50%; flex-shrink: 0; }
.event-dot.container { background: #4caf50; }
.event-dot.task { background: #2196f3; }
.event-text { flex: 1; color: #555; overflow: hidden; text-overflow: ellipsis; white-space: nowrap; }
.event-time { color: #aaa; font-size: 0.75rem; }

/* Compose groups */
.compose-group { margin-bottom: 1.25rem; }
.compose-header {
  display: flex; align-items: center; gap: 0.5rem; padding: 0.5rem 0.75rem; margin-bottom: 0.5rem;
  background: #fff; border-radius: 8px; box-shadow: 0 1px 2px rgba(0,0,0,0.04);
}
.compose-icon { font-size: 0.9rem; color: #6c63ff; }
.compose-name { font-weight: 600; font-size: 0.9rem; color: #333; }
.compose-count { margin-left: auto; font-size: 0.78rem; color: #999; }
.compose-file { font-size: 0.75rem; color: #bbb; font-family: monospace; }
.compose-containers { display: grid; grid-template-columns: repeat(auto-fill, minmax(320px, 1fr)); gap: 0.6rem; }

/* Detail */
.btn-back { background: none; border: none; color: #6c63ff; cursor: pointer; font-size: 0.85rem; padding: 0; margin-bottom: 0.75rem; }
.detail-header { display: flex; align-items: center; gap: 0.75rem; margin-bottom: 1rem; }
.detail-title-group h2 { font-size: 1.35rem; font-weight: 600; }
.compose-tag { font-size: 0.75rem; color: #6c63ff; background: #e8e5ff; padding: 0.1rem 0.5rem; border-radius: 4px; display: inline-block; margin-bottom: 0.2rem; }
.badge-lg { padding: 0.2rem 0.75rem; border-radius: 10px; font-size: 0.75rem; font-weight: 600; text-transform: uppercase; }
.badge-lg.running { background: #e8f5e9; color: #2e7d32; }
.badge-lg.exited { background: #fbe9e7; color: #c62828; }
.detail-grid { display: grid; grid-template-columns: 1fr 1fr; gap: 0.75rem; }
.detail-card { background: #fff; border-radius: 10px; padding: 1.25rem; box-shadow: 0 1px 3px rgba(0,0,0,0.06); }
.detail-card-wide { grid-column: 1 / -1; }
.detail-card-title { font-weight: 600; font-size: 0.8rem; color: #999; text-transform: uppercase; letter-spacing: 0.5px; margin-bottom: 0.75rem; }
.detail-row { display: flex; padding: 0.35rem 0; font-size: 0.85rem; border-bottom: 1px solid #f5f5f5; }
.detail-row:last-child { border-bottom: none; }
.dl { width: 50px; color: #888; flex-shrink: 0; }
.dd { color: #333; }
.mono { font-family: 'SF Mono', 'Fira Code', monospace; font-size: 0.8rem; }
.port-row { display: flex; align-items: center; gap: 0.5rem; }
.port-type { font-size: 0.7rem; color: #999; background: #f5f5f5; padding: 0.1rem 0.4rem; border-radius: 3px; font-weight: 600; width: 36px; text-align: center; flex-shrink: 0; }
.port-mapping { display: flex; align-items: center; gap: 0.4rem; flex: 1; }
.port-binding { color: #6c63ff; font-family: monospace; font-size: 0.85rem; }
.port-arrow { color: #ccc; font-size: 0.8rem; }
.port-private { color: #333; font-family: monospace; }
.detail-actions { display: flex; gap: 0.5rem; flex-wrap: wrap; }
.btn { padding: 0.5rem 1rem; border: none; border-radius: 6px; cursor: pointer; font-size: 0.8rem; font-weight: 500; }
.btn:hover { opacity: 0.85; }
.btn:disabled { opacity: 0.35; cursor: not-allowed; }
.btn-log { background: #f3e5f5; color: #7b1fa2; }
.btn-exec { background: #e8eaf6; color: #283593; }
.btn-sm { background: none; border: 1px solid #ddd; border-radius: 4px; padding: 0.15rem 0.5rem; cursor: pointer; font-size: 0.75rem; color: #888; float: right; }
.log-viewer { background: #1a1a2e; color: #e0e0e0; border-radius: 6px; padding: 0.75rem; max-height: 300px; overflow-y: auto; font-family: 'SF Mono', 'Fira Code', monospace; font-size: 0.78rem; line-height: 1.4; }
.log-line { white-space: pre-wrap; word-break: break-all; }
.exec-bar { display: flex; gap: 0.5rem; margin-bottom: 0.5rem; }
.exec-input { flex: 1; padding: 0.5rem 0.75rem; border: 1px solid #ddd; border-radius: 6px; font-size: 0.85rem; font-family: 'SF Mono', 'Fira Code', monospace; outline: none; }
.exec-input:focus { border-color: #6c63ff; }
.exec-output { background: #1a1a2e; color: #e0e0e0; border-radius: 6px; padding: 0.75rem; max-height: 200px; overflow-y: auto; font-family: 'SF Mono', 'Fira Code', monospace; font-size: 0.78rem; }
.exec-output pre { margin: 0; white-space: pre-wrap; word-break: break-all; }
.btn-start { background: #e8f5e9; color: #2e7d32; }
.btn-stop { background: #fbe9e7; color: #c62828; }
.btn-restart { background: #e3f2fd; color: #1565c0; }
.btn-remove { background: #fce4ec; color: #ad1457; }

/* Events */
.events-panel { background: #fff; border-radius: 10px; padding: 0.75rem; box-shadow: 0 1px 3px rgba(0,0,0,0.06); }
.event-row { display: flex; align-items: center; gap: 0.5rem; padding: 0.5rem; border-bottom: 1px solid #f5f5f5; font-size: 0.82rem; }
.event-row:last-child { border-bottom: none; }
.event-badge { padding: 0.1rem 0.4rem; border-radius: 3px; font-size: 0.7rem; font-weight: 600; text-transform: uppercase; flex-shrink: 0; }
.event-badge.container { background: #e8f5e9; color: #2e7d32; }
.event-badge.task { background: #e3f2fd; color: #1565c0; }
.event-msg { flex: 1; color: #555; overflow: hidden; text-overflow: ellipsis; white-space: nowrap; }
.event-ts { color: #aaa; font-size: 0.75rem; flex-shrink: 0; }

.loading, .empty-state { text-align: center; padding: 2rem; color: #999; font-size: 0.85rem; }
.empty-state-sm { text-align: center; padding: 0.75rem; color: #999; font-size: 0.8rem; }
</style>
