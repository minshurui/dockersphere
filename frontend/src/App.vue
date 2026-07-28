<template>
  <div :class="['app', darkMode ? 'dark' : '']">
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
        <div class="theme-toggle" @click="toggleDark">{{ darkMode ? '☀️ 白天' : '🌙 夜间' }}</div>
        <div class="version">v0.3.0</div>
      </div>
    </aside>

    <div class="main-area">
      <header class="topbar">
        <h1 class="page-title">{{ currentTab.label }}</h1>
        <div class="topbar-actions">
          <div v-if="activeTab === 'containers'" class="search-box">
            <input v-model="searchQuery" placeholder="搜索容器..." class="search-input" />
          </div>
          <span class="container-count">{{ containers.length }} 容器</span>
          <button class="btn-refresh" @click="fetchContainers" title="刷新">↻</button>
        </div>
      </header>

      <main class="content">
        <!-- Dashboard -->
        <div v-if="activeTab === 'dashboard'" class="tab-content">
          <div class="stats-grid">
            <div class="stat-card"><div class="stat-value">{{ stats.total }}</div><div class="stat-label">容器</div></div>
            <div class="stat-card running"><div class="stat-value">{{ stats.running }}</div><div class="stat-label">运行中</div></div>
            <div class="stat-card"><div class="stat-value">{{ stats.images }}</div><div class="stat-label">镜像</div></div>
            <div class="stat-card"><div class="stat-value">{{ stats.projects }}</div><div class="stat-label">Compose 项目</div></div>
          </div>
          <div class="section">
            <h3 class="section-title">系统信息</h3>
            <div v-if="sysInfo" class="sys-info">
              <span><b>Docker</b> {{ sysInfo.ServerVersion }}</span>
              <span><b>OS</b> {{ sysInfo.OSType }} {{ sysInfo.Architecture }}</span>
              <span><b>CPU</b> {{ sysInfo.NCPU }} 核</span>
              <span><b>内存</b> {{ formatBytes(sysInfo.MemTotal) }}</span>
              <span><b>运行时间</b> {{ Math.floor(sysInfo.ContainersRunning) }} 容器</span>
            </div>
          </div>
          <div class="section">
            <h3 class="section-title">最近事件</h3>
            <div class="mini-events">
              <div v-for="(e, i) in events.slice(0, 10)" :key="i" class="mini-event">
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
          <!-- New Deploy -->
          <div class="section deploy-section">
            <div class="section-title-row">
              <h3 class="section-title" @click="showDeploy = !showDeploy" style="cursor:pointer">
                {{ showDeploy ? '▼' : '▶' }} 新建部署
              </h3>
              <span class="compose-count">粘贴 docker-compose.yml 直接启动</span>
            </div>
            <div v-if="showDeploy">
              <div class="deploy-form">
                <input v-model="deployProject" placeholder="项目名称 (例如: myapp)" class="search-input" style="width:300px;margin-bottom:0.5rem" />
                <textarea v-model="deployContent" placeholder="粘贴 docker-compose.yml 内容..." class="compose-editor" rows="8"></textarea>
                <div class="deploy-actions">
                  <button @click="deployStack" :disabled="deploying" class="btn btn-exec">
                    {{ deploying ? '部署中...' : '🚀 部署' }}
                  </button>
                </div>
              </div>
              <div v-if="deployOutput" :class="['deploy-output', deploySuccess ? 'deploy-ok' : 'deploy-fail']">
                <pre>{{ deployOutput }}</pre>
              </div>
            </div>
          </div>

          <div v-if="loading" class="loading">加载中...</div>
          <div v-else-if="filteredContainers.length === 0" class="empty-state">没有匹配的容器</div>
          <div v-else class="container-groups">
            <div v-for="group in filteredGroups" :key="group.project" class="compose-group">
              <div class="compose-header" @click="toggleCompose(group)">
                <span :class="['compose-arrow', { expanded: composeExpanded[group.project] }]">▶</span>
                <span class="compose-icon">▦</span>
                <span class="compose-name">{{ group.project }}</span>
                <span class="compose-count">{{ group.containers.length }} 服务</span>
                <span class="compose-file">{{ group.configFiles ? group.configFiles.split('/').pop() : '' }}</span>
                <div class="compose-actions" @click.stop>
                  <button @click="projectAction(group.project, 'start')" class="act-compose act-start" title="启动全部">▶</button>
                  <button @click="projectAction(group.project, 'stop')" class="act-compose act-stop" title="停止全部">■</button>
                  <button @click="projectAction(group.project, 'restart')" class="act-compose act-restart" title="重启全部">↻</button>
                  <button v-if="group.configFiles && composeExpanded[group.project] && group._content" @click.stop="editCompose(group)" class="btn-compose">编辑</button>
                </div>
              </div>
              <div v-if="composeExpanded[group.project] && group._content !== undefined" class="compose-inline">
                <div class="compose-inline-header">
                  <span class="compose-inline-title">docker-compose.yml</span>
                  <div class="compose-inline-actions">
                    <button v-if="composeEditingKey !== group.project" @click.stop="editCompose(group)" class="btn btn-log" style="padding:0.2rem 0.5rem;font-size:0.75rem">编辑</button>
                    <template v-if="composeEditingKey === group.project">
                      <button @click.stop="saveCompose()" class="btn btn-exec" style="padding:0.2rem 0.5rem;font-size:0.75rem">保存</button>
                      <button @click.stop="cancelCompose()" class="btn btn-stop" style="padding:0.2rem 0.5rem;font-size:0.75rem">取消</button>
                    </template>
                  </div>
                </div>
                <textarea v-if="composeEditingKey === group.project" v-model="composeEditContent" class="compose-editor" rows="15" @click.stop></textarea>
                <pre v-else class="compose-viewer">{{ group._content }}</pre>
              </div>
              <div v-if="composeExpanded[group.project] && group._content === undefined" class="compose-loading">加载中...</div>
              <div class="compose-containers">
                <ContainerCard v-for="c in group.containers" :key="c.id" :container="c" @action="action" @click="openDetail(c)" />
              </div>
            </div>
            <div v-if="standaloneContainers.length > 0" class="compose-group">
              <div class="compose-header"><span class="compose-icon">◯</span><span class="compose-name">独立容器</span><span class="compose-count">{{ standaloneContainers.length }}</span></div>
              <div class="compose-containers">
                <ContainerCard v-for="c in standaloneContainers" :key="c.id" :container="c" @action="action" @click="openDetail(c)" />
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
            <div class="detail-badges">
              <span :class="['badge-lg', healthClass(selectedContainer)]">{{ healthLabel(selectedContainer) }}</span>
              <span :class="['badge-lg', selectedContainer.state]">{{ selectedContainer.state }}</span>
            </div>
          </div>
          <div class="detail-grid">
            <div class="detail-card">
              <div class="detail-card-title">基本信息</div>
              <div class="detail-row"><span class="dl">ID</span><span class="dd mono">{{ selectedContainer.id }}</span></div>
              <div class="detail-row"><span class="dl">镜像</span><span class="dd">{{ selectedContainer.image }}</span></div>
              <div class="detail-row"><span class="dl">状态</span><span class="dd">{{ selectedContainer.status }}</span></div>
              <div class="detail-row"><span class="dl">创建</span><span class="dd">{{ selectedContainer.created }}</span></div>
              <div class="detail-row"><span class="dl">重启</span><span class="dd">{{ selectedContainer.labels?.['com.docker.compose.restart'] || (selectedContainer.status?.includes('unhealthy') ? '异常' : '默认') }}</span></div>
              <div class="detail-row"><span class="dl">项目</span><span class="dd">{{ selectedContainer.labels?.['com.docker.compose.project'] || '-' }}/{{ selectedContainer.labels?.['com.docker.compose.service'] || '-' }}</span></div>
            </div>
            <div class="detail-card">
              <div class="detail-card-title">标签 ({{ Object.keys(selectedContainer.labels || {}).length }})</div>
              <div class="label-list">
                <template v-for="(v, k) in selectedContainer.labels || {}" :key="k">
                  <div v-if="!k.startsWith('com.docker.compose')" class="label-item">
                    <span class="label-key">{{ k }}</span>
                    <span class="label-val">{{ v }}</span>
                  </div>
                </template>
                <div v-if="!selectedContainer.labels || Object.keys(selectedContainer.labels).length === 0" class="empty-state-sm">无标签</div>
              </div>
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
              <div class="detail-card-title">
                日志
                <button v-if="showLogs" @click="showLogs=false" class="btn-sm">收起</button>
                <span v-else class="log-hint">最近 50 行</span>
              </div>
              <button v-if="!showLogs" @click="fetchLogs" class="btn btn-log">📋 加载日志</button>
              <div v-if="showLogs" class="log-viewer" ref="logViewer">
                <div v-for="(line, i) in logs" :key="i" class="log-line">{{ line }}</div>
                <div v-if="logs.length === 0" class="empty-state-sm">无日志</div>
              </div>
            </div>
            <div class="detail-card detail-card-wide">
              <div class="detail-card-title">命令执行</div>
              <div class="exec-bar">
                <input v-model="execCmd" @keyup.enter="runExec" placeholder="例如: ps aux, df -h, top -bn1" class="exec-input" />
                <button @click="runExec" class="btn btn-exec">运行</button>
              </div>
              <div v-if="execOutput" class="exec-output"><pre>{{ execOutput }}</pre></div>
            </div>
          </div>
        </div>

        <!-- Images -->
        <div v-if="activeTab === 'images'" class="tab-content">
          <div v-if="imagesLoading" class="loading">加载中...</div>
          <div v-else class="section">
            <div class="section-title-row">
              <h3 class="section-title">镜像列表</h3>
              <span class="compose-count">{{ images.length }} 个镜像 · {{ formatBytes(totalImageSize) }}</span>
            </div>
            <div class="image-grid">
              <div v-for="img in images" :key="img.ID" class="image-card">
                <div class="image-info">
                  <span class="image-name">{{ img.RepoTags?.[0] || img.RepoDigests?.[0]?.split('@')[0] || img.ID?.slice(7,19) || '无标签' }}</span>
                  <span class="image-size">{{ formatBytes(img.Size) }}</span>
                  <span class="image-id mono">{{ img.ID?.slice(7,19) }}</span>
                </div>
                <button @click="removeImage(img.ID)" class="btn btn-remove-sm" :disabled="imageRemoving === img.ID">✕</button>
              </div>
              <div v-if="images.length === 0" class="empty-state">无镜像</div>
            </div>
          </div>
        </div>

        <!-- Events -->
        <div v-if="activeTab === 'events'" class="tab-content">
          <div class="section">
            <h3 class="section-title">事件日志</h3>
            <div class="events-panel">
              <div v-for="(e, i) in events" :key="i" class="event-row">
                <span :class="['event-badge', e.type.split('.')[0]]">{{ e.type.split('.')[0] }}</span>
                <span class="event-msg">{{ e.type }} {{ e.summary }}</span>
                <span class="event-ts">{{ formatTime(e.timestamp) }}</span>
              </div>
              <div v-if="events.length === 0" class="empty-state">暂无事件</div>
            </div>
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
    const dark = localStorage.getItem('dockersphere-dark') === 'true'
    return {
      activeTab: 'dashboard',
      containers: [], images: [], events: [],
      selectedContainer: null,
      loading: false, imagesLoading: false, connected: false,
      ws: null, darkMode: dark, searchQuery: '',
      showLogs: false, logs: [], execCmd: '', execOutput: '',
      sysInfo: null, imageRemoving: null,
      composeFile: '', composeFilePath: '', composeEditing: false, composeSaving: false,
      composeEditingKey: '', composeEditContent: '', composeEditPath: '',
      showDeploy: false, deployProject: '', deployContent: '', deployOutput: '', deploying: false, deploySuccess: false,
      composeExpanded: {},
      tabs: [
        { id: 'dashboard', label: '仪表盘', icon: '◉' },
        { id: 'containers', label: '容器', icon: '▣' },
        { id: 'images', label: '镜像', icon: '◻' },
        { id: 'events', label: '事件', icon: '⚡' },
      ]
    }
  },
  computed: {
    currentTab() { return this.tabs.find(t => t.id === this.activeTab) || this.tabs[0] },
    stats() {
      const total = this.containers.length
      const running = this.containers.filter(c => c.state === 'running').length
      const images = new Set(this.containers.map(c => c.image)).size
      const projects = new Set(this.containers.map(c => c.labels?.['com.docker.compose.project'] || '_')).size
      return { total, running, images, projects }
    },
    filteredContainers() {
      if (!this.searchQuery) return this.containers
      const q = this.searchQuery.toLowerCase()
      return this.containers.filter(c =>
        c.name.toLowerCase().includes(q) ||
        c.image.toLowerCase().includes(q) ||
        c.state.toLowerCase().includes(q) ||
        c.id.toLowerCase().includes(q)
      )
    },
    filteredGroups() {
      const groups = {}
      for (const c of this.filteredContainers) {
        const project = c.labels?.['com.docker.compose.project']
        if (project) {
          if (!groups[project]) groups[project] = { project, configFiles: c.labels?.['com.docker.compose.project.config_files'] || '', containers: [] }
          groups[project].containers.push({ ...c, composeProject: project, composeService: c.labels?.['com.docker.compose.service'] || '' })
        }
      }
      return Object.values(groups).sort((a, b) => a.project.localeCompare(b.project))
    },
    standaloneContainers() {
      return this.containers.filter(c => !c.labels?.['com.docker.compose.project']).map(c => ({ ...c, composeProject: null, composeService: null }))
    },
    totalImageSize() {
      return this.images.reduce((s, img) => s + img.Size, 0)
    }
  },
  mounted() {
    this.fetchContainers()
    this.fetchImages()
    this.fetchSysInfo()
    this.connectWebSocket()
  },
  beforeUnmount() { if (this.ws) this.ws.close() },
  methods: {
    async fetchContainers() {
      try { this.loading = true; const r = await axios.get('/api/v1/containers'); if (r.data.code === 0) this.containers = r.data.data || []
      } catch (e) { console.error(e) } finally { this.loading = false }
    },
    async fetchImages() {
      try { this.imagesLoading = true; const r = await axios.get('/api/v1/images'); if (r.data.code === 0) this.images = r.data.data || []
      } catch (e) { console.error(e) } finally { this.imagesLoading = false }
    },
    async fetchSysInfo() {
      try { const r = await axios.get('/api/v1/system/info'); if (r.data.code === 0) this.sysInfo = r.data.data
      } catch (e) { /* ignore */ }
    },
    toggleCompose(project) {
      const key = project.project
      this.composeExpanded[key] = !this.composeExpanded[key]
      if (this.composeExpanded[key] && !project._content && project.configFiles) {
        axios.get('/api/v1/compose/file', { params: { path: project.configFiles } }).then(r => {
          project._content = r.data.code === 0 ? r.data.data : '加载失败'
        }).catch(() => { project._content = '加载失败' })
      }
    },
    editCompose(group) {
      this.composeEditingKey = group.project
      this.composeEditContent = group._content
      this.composeEditPath = group.configFiles
    },
    cancelCompose() { this.composeEditingKey = '' },
    async saveCompose() {
      if (!this.composeEditPath) { alert('错误：没有文件路径'); return }
      try {
        const r = await axios.put('/api/v1/compose/file', { path: this.composeEditPath, content: this.composeEditContent })
        if (r.data.code !== 0) { alert('保存失败: ' + (r.data.message || '未知错误')); return }
        // 更新所有匹配项目的 _content
        for (const g of this.filteredGroups) {
          if (g.project === this.composeEditingKey) g._content = this.composeEditContent
        }
        this.composeEditingKey = ''
      } catch (e) { alert('保存失败: ' + (e.response?.data?.message || e.message)) }
    },
    async projectAction(project, action) {
      try {
        const r = await axios.post(`/api/v1/compose/project/${project}/${action}`, {})
        if (r.data.code === 0 && r.data.data.success) {
          setTimeout(() => this.fetchContainers(), 2000)
        } else {
          console.warn(r.data.data?.output || '操作完成')
        }
      } catch (e) { console.error(e) }
    },
    async deployStack() {
      if (!this.deployProject.trim() || !this.deployContent.trim()) { alert('请输入项目名称和内容'); return }
      this.deploying = true; this.deployOutput = ''; this.deploySuccess = false
      try {
        const r = await axios.post('/api/v1/compose/deploy', { project: this.deployProject, content: this.deployContent })
        if (r.data.code === 0) {
          this.deployOutput = r.data.data.output
          this.deploySuccess = r.data.data.success
          if (this.deploySuccess) { this.fetchContainers(); this.deployProject = ''; this.deployContent = '' }
        }
      } catch (e) {
        this.deployOutput = '请求失败: ' + (e.response?.data?.message || e.message)
        this.deploySuccess = false
      }
      this.deploying = false
    },
    async removeImage(id) {
      this.imageRemoving = id
      try { await axios.delete(`/api/v1/images/${id}`); this.images = this.images.filter(i => i.ID !== id)
      } catch (e) { alert('删除失败: ' + (e.response?.data?.message || e.message)) }
      this.imageRemoving = null
    },
    async action(id, act) { try { await axios.post(`/api/v1/containers/${id}/action`, { action: act }) } catch (e) { console.error(e) } },
    openDetail(c) { this.selectedContainer = c; this.activeTab = 'detail'; this.showLogs = false; this.logs = []; this.execOutput = '' },
    async fetchLogs() {
      this.showLogs = true
      try { const r = await axios.get(`/api/v1/containers/${this.selectedContainer.id}/logs?tail=50`); if (r.data.code === 0) this.logs = r.data.data || []
      } catch (e) { this.logs = ['获取日志失败'] }
    },
    async runExec() {
      if (!this.execCmd.trim()) return; this.execOutput = '执行中...'
      try {
        const parts = this.execCmd.trim().split(/\s+/)
        const r = await axios.post(`/api/v1/containers/${this.selectedContainer.id}/exec`, { cmd: parts[0], args: parts.slice(1) })
        if (r.data.code === 0) this.execOutput = r.data.data.replace(/[\x00-\x08\x0e-\x1f]/g, '').trim()
      } catch (e) { this.execOutput = '失败: ' + (e.response?.data?.message || e.message) }
    },
    mappedPorts(ports) {
      const seen = new Set()
      return ports.filter(p => { const k = `${p.private_port}-${p.type}`; if (seen.has(k)) return false; seen.add(k); return true })
        .map(p => ({ key: `${p.private_port}-${p.type}`, type: p.type.toUpperCase(), binding: p.public_port ? `${p.ip || '0.0.0.0'}:${p.public_port}` : '未映射', private: `${p.private_port}` }))
    },
    healthClass(c) {
      if (!c.status) return 'created'
      if (c.status.includes('healthy')) return 'running'
      if (c.status.includes('unhealthy')) return 'exited'
      return c.state
    },
    healthLabel(c) {
      if (!c.status) return '未知'
      if (c.status.includes('healthy')) return '健康'
      if (c.status.includes('unhealthy')) return '异常'
      return c.state
    },
    formatBytes(b) {
      if (!b) return '0B'
      const u = ['B','KB','MB','GB','TB'], i = Math.floor(Math.log(b)/Math.log(1024))
      return (b/Math.pow(1024,i)).toFixed(1) + u[i]
    },
    toggleDark() { this.darkMode = !this.darkMode; localStorage.setItem('dockersphere-dark', this.darkMode) },
    async openCompose(path) {
      this.composeFilePath = path
      this.composeEditing = false
      this.activeTab = 'compose'
      try {
        const r = await axios.get('/api/v1/compose/file', { params: { path } })
        if (r.data.code === 0) this.composeFile = r.data.data
      } catch (e) { this.composeFile = '加载失败: ' + (e.response?.data?.message || e.message) }
    },
    async saveCompose() {
      this.composeSaving = true
      try {
        await axios.put('/api/v1/compose/file', { path: this.composeFilePath, content: this.composeFile })
        this.composeEditing = false
      } catch (e) { alert('保存失败: ' + (e.response?.data?.message || e.message)) }
      this.composeSaving = false
    },
    connectWebSocket() {
      const proto = window.location.protocol === 'https:' ? 'wss:' : 'ws:'
      this.ws = new WebSocket(`${proto}//${window.location.host}/ws`)
      this.ws.onopen = () => { this.connected = true }
      this.ws.onmessage = (e) => {
        try { const msg = JSON.parse(e.data); this.events.unshift({ type: msg.event, summary: (msg.data || {}).name || '', timestamp: new Date() }); if (this.events.length > 200) this.events.pop(); if (msg.event.startsWith('container.')) this.fetchContainers()
        } catch (_) {}
      }
      this.ws.onclose = () => { this.connected = false; setTimeout(() => this.connectWebSocket(), 5000) }
    },
    formatTime(d) { return d.toLocaleTimeString() }
  }
}
</script>

<style>
:root {
  --bg: #f0f2f5; --card-bg: #fff; --text: #1a1a2e; --text2: #333; --muted: #888; --border: #e8e8e8;
  --sidebar-bg: #1a1a2e; --sidebar-text: #e0e0e0; --sidebar-muted: #a0a0b8;
  --primary: #6c63ff; --hover: #f5f5f5; --log-bg: #1a1a2e; --log-text: #e0e0e0;
  --shadow: rgba(0,0,0,0.06);
}
.dark {
  --bg: #121212; --card-bg: #1e1e2e; --text: #e0e0e0; --text2: #ccc; --muted: #888; --border: #2a2a3e;
  --sidebar-bg: #0d0d1a; --sidebar-text: #e0e0e0; --sidebar-muted: #666;
  --hover: #2a2a3e; --log-bg: #0d0d1a; --log-text: #e0e0e0; --shadow: rgba(0,0,0,0.2);
}
* { margin: 0; padding: 0; box-sizing: border-box; }
body { font-family: -apple-system, BlinkMacSystemFont, 'Segoe UI', Roboto, sans-serif; background: var(--bg); color: var(--text); transition: background 0.2s, color 0.2s; }
.app { display: flex; min-height: 100vh; }

/* Sidebar */
.sidebar { width: 200px; background: var(--sidebar-bg); color: var(--sidebar-text); display: flex; flex-direction: column; position: fixed; top: 0; left: 0; bottom: 0; z-index: 100; }
.logo { padding: 1.25rem; display: flex; align-items: center; gap: 0.5rem; border-bottom: 1px solid rgba(255,255,255,0.08); }
.logo-icon { font-size: 1.3rem; color: var(--primary); }
.logo-text { font-size: 1.15rem; font-weight: 700; }
.nav { flex: 1; padding: 0.5rem 0; }
.nav-item { display: flex; align-items: center; gap: 0.75rem; padding: 0.65rem 1.25rem; cursor: pointer; color: var(--sidebar-muted); transition: all 0.15s; font-size: 0.85rem; }
.nav-item:hover { background: rgba(108,99,255,0.08); color: var(--sidebar-text); }
.nav-item.active { background: rgba(108,99,255,0.15); color: #fff; border-right: 3px solid var(--primary); }
.nav-icon { font-size: 1rem; width: 20px; text-align: center; }
.sidebar-footer { padding: 0.75rem 1.25rem; border-top: 1px solid rgba(255,255,255,0.08); font-size: 0.75rem; }
.connection-status { display: flex; align-items: center; gap: 0.5rem; margin-bottom: 0.5rem; }
.dot { width: 8px; height: 8px; border-radius: 50%; }
.dot.online { background: #4caf50; box-shadow: 0 0 6px rgba(76,175,80,0.5); }
.dot.offline { background: #f44336; }
.theme-toggle { cursor: pointer; color: #888; margin-bottom: 0.25rem; }
.theme-toggle:hover { color: #ccc; }
.version { color: #555; }

/* Main */
.main-area { margin-left: 200px; flex: 1; display: flex; flex-direction: column; min-height: 100vh; }
.topbar { background: var(--card-bg); padding: 0.75rem 1.5rem; display: flex; justify-content: space-between; align-items: center; border-bottom: 1px solid var(--border); position: sticky; top: 0; z-index: 50; }
.page-title { font-size: 1.15rem; font-weight: 600; }
.topbar-actions { display: flex; align-items: center; gap: 0.75rem; }
.search-input { padding: 0.4rem 0.75rem; border: 1px solid var(--border); border-radius: 6px; font-size: 0.85rem; background: var(--card-bg); color: var(--text); outline: none; width: 200px; }
.search-input:focus { border-color: var(--primary); }
.container-count { color: var(--muted); font-size: 0.85rem; }
.btn-refresh { background: none; border: 1px solid var(--border); border-radius: 6px; padding: 0.35rem 0.5rem; cursor: pointer; font-size: 1rem; color: var(--muted); }
.btn-refresh:hover { background: var(--hover); }
.content { padding: 1.25rem 1.5rem; flex: 1; }

/* Stats */
.stats-grid { display: grid; grid-template-columns: repeat(4, 1fr); gap: 0.75rem; margin-bottom: 1.5rem; }
.stat-card { background: var(--card-bg); border-radius: 10px; padding: 1.25rem; text-align: center; box-shadow: 0 1px 3px var(--shadow); }
.stat-value { font-size: 1.75rem; font-weight: 700; color: var(--text); }
.stat-card.running .stat-value { color: #4caf50; }
.stat-value:last-child { margin-bottom: 0; }
.stat-label { color: var(--muted); font-size: 0.8rem; margin-top: 0.2rem; }

.section { background: var(--card-bg); border-radius: 10px; padding: 1.25rem; box-shadow: 0 1px 3px var(--shadow); margin-bottom: 1rem; }
.section-title { font-size: 0.9rem; font-weight: 600; margin-bottom: 0.75rem; }
.section-title-row { display: flex; justify-content: space-between; align-items: center; margin-bottom: 0.75rem; }
.section-title-row .section-title { margin-bottom: 0; }

/* Sys info */
.sys-info { display: flex; gap: 1.5rem; flex-wrap: wrap; font-size: 0.85rem; color: var(--muted); }
.sys-info b { color: var(--text); }

/* Mini events */
.mini-events { display: flex; flex-direction: column; gap: 0.35rem; }
.mini-event { display: flex; align-items: center; gap: 0.5rem; padding: 0.35rem 0; border-bottom: 1px solid var(--border); font-size: 0.8rem; }
.mini-event:last-child { border-bottom: none; }
.event-dot { width: 7px; height: 7px; border-radius: 50%; flex-shrink: 0; }
.event-dot.container { background: #4caf50; }
.event-dot.task { background: #2196f3; }
.event-text { flex: 1; color: var(--text2); overflow: hidden; text-overflow: ellipsis; white-space: nowrap; }
.event-time { color: var(--muted); font-size: 0.75rem; flex-shrink: 0; }

/* Compose groups */
.compose-group { margin-bottom: 1.25rem; }
.compose-header { display: flex; align-items: center; gap: 0.5rem; padding: 0.5rem 0.75rem; margin-bottom: 0.5rem; background: var(--card-bg); border-radius: 8px; box-shadow: 0 1px 2px var(--shadow); cursor: pointer; transition: background 0.15s; }
.compose-header:hover { background: var(--hover); }
.compose-icon { font-size: 0.9rem; color: var(--primary); }
.compose-name { font-weight: 600; font-size: 0.9rem; }
.compose-count { margin-left: auto; font-size: 0.78rem; color: var(--muted); }
.compose-file { font-size: 0.75rem; color: var(--muted); font-family: monospace; }
.compose-containers { display: grid; grid-template-columns: repeat(auto-fill, minmax(320px, 1fr)); gap: 0.6rem; }

/* Detail */
.btn-back { background: none; border: none; color: var(--primary); cursor: pointer; font-size: 0.85rem; padding: 0; margin-bottom: 0.75rem; }
.detail-header { display: flex; align-items: center; gap: 0.75rem; margin-bottom: 1rem; }
.detail-title-group h2 { font-size: 1.35rem; font-weight: 600; }
.detail-badges { margin-left: auto; display: flex; gap: 0.5rem; }
.compose-tag { font-size: 0.75rem; color: var(--primary); background: rgba(108,99,255,0.1); padding: 0.1rem 0.5rem; border-radius: 4px; }
.badge-lg { padding: 0.2rem 0.75rem; border-radius: 10px; font-size: 0.75rem; font-weight: 600; text-transform: uppercase; }
.badge-lg.running { background: #e8f5e9; color: #2e7d32; }
.badge-lg.exited { background: #fbe9e7; color: #c62828; }
.detail-grid { display: grid; grid-template-columns: 1fr 1fr; gap: 0.75rem; }
.detail-card { background: var(--card-bg); border-radius: 10px; padding: 1.25rem; box-shadow: 0 1px 3px var(--shadow); }
.detail-card-wide { grid-column: 1 / -1; }
.detail-card-title { font-weight: 600; font-size: 0.8rem; color: var(--muted); text-transform: uppercase; letter-spacing: 0.5px; margin-bottom: 0.75rem; }
.detail-row { display: flex; padding: 0.35rem 0; font-size: 0.85rem; border-bottom: 1px solid var(--border); }
.detail-row:last-child { border-bottom: none; }
.dl { width: 50px; color: var(--muted); flex-shrink: 0; }
.dd { color: var(--text2); }
.mono { font-family: 'SF Mono', 'Fira Code', monospace; font-size: 0.8rem; }
.port-row { display: flex; align-items: center; gap: 0.5rem; }
.port-type { font-size: 0.7rem; color: var(--muted); background: var(--bg); padding: 0.1rem 0.4rem; border-radius: 3px; font-weight: 600; width: 36px; text-align: center; flex-shrink: 0; }
.port-mapping { display: flex; align-items: center; gap: 0.4rem; flex: 1; }
.port-binding { color: var(--primary); font-family: monospace; font-size: 0.85rem; }
.port-arrow { color: var(--muted); font-size: 0.8rem; }
.port-private { color: var(--text2); font-family: monospace; }
.detail-actions { display: flex; gap: 0.5rem; flex-wrap: wrap; }
.log-hint { font-weight: 400; text-transform: none; font-size: 0.75rem; color: var(--muted); }
.log-viewer { background: var(--log-bg); color: var(--log-text); border-radius: 6px; padding: 0.75rem; max-height: 300px; overflow-y: auto; font-family: 'SF Mono', 'Fira Code', monospace; font-size: 0.78rem; line-height: 1.4; }
.log-line { white-space: pre-wrap; word-break: break-all; }
.exec-bar { display: flex; gap: 0.5rem; margin-bottom: 0.5rem; }
.exec-input { flex: 1; padding: 0.5rem 0.75rem; border: 1px solid var(--border); border-radius: 6px; font-size: 0.85rem; font-family: 'SF Mono', 'Fira Code', monospace; background: var(--card-bg); color: var(--text); outline: none; }
.exec-input:focus { border-color: var(--primary); }
.exec-output { background: var(--log-bg); color: var(--log-text); border-radius: 6px; padding: 0.75rem; max-height: 200px; overflow-y: auto; font-family: 'SF Mono', 'Fira Code', monospace; font-size: 0.78rem; }
.exec-output pre { margin: 0; white-space: pre-wrap; word-break: break-all; }

/* Buttons */
.btn { padding: 0.5rem 1rem; border: none; border-radius: 6px; cursor: pointer; font-size: 0.8rem; font-weight: 500; }
.btn:hover { opacity: 0.85; }
.btn:disabled { opacity: 0.35; cursor: not-allowed; }
.btn-start { background: #e8f5e9; color: #2e7d32; }
.btn-stop { background: #fbe9e7; color: #c62828; }
.btn-restart { background: #e3f2fd; color: #1565c0; }
.btn-remove { background: #fce4ec; color: #ad1457; }
.btn-log { background: #f3e5f5; color: #7b1fa2; }
.btn-exec { background: #e8eaf6; color: #283593; }
.btn-sm { background: none; border: 1px solid var(--border); border-radius: 4px; padding: 0.15rem 0.5rem; cursor: pointer; font-size: 0.75rem; color: var(--muted); float: right; }
.label-list { max-height: 200px; overflow-y: auto; font-size: 0.78rem; }
.label-item { display: flex; gap: 0.5rem; padding: 0.2rem 0; border-bottom: 1px solid var(--border); }
.label-key { color: var(--primary); font-family: monospace; font-size: 0.75rem; min-width: 80px; overflow: hidden; text-overflow: ellipsis; }
.label-val { color: var(--muted); font-family: monospace; overflow: hidden; text-overflow: ellipsis; }
.btn-remove-sm { background: none; border: none; color: #f44336; cursor: pointer; font-size: 0.9rem; padding: 0.25rem; flex-shrink: 0; }
.btn-remove-sm:disabled { opacity: 0.3; }
.btn-compose { background: none; border: 1px solid var(--border); border-radius: 4px; padding: 0.2rem 0.5rem; cursor: pointer; font-size: 0.75rem; color: var(--muted); flex-shrink: 0; }
.btn-compose:hover { background: var(--hover); color: var(--primary); }
.compose-actions { display: flex; align-items: center; gap: 0.2rem; flex-shrink: 0; }
.act-compose { width: 24px; height: 24px; border: none; border-radius: 4px; cursor: pointer; font-size: 0.7rem; display: flex; align-items: center; justify-content: center; opacity: 0.7; transition: opacity 0.15s; }
.act-compose:hover { opacity: 1; }
.act-compose.act-start { background: #e8f5e9; color: #2e7d32; }
.act-compose.act-stop { background: #fbe9e7; color: #c62828; }
.act-compose.act-restart { background: #e3f2fd; color: #1565c0; }
.compose-arrow { display: inline-block; transition: transform 0.15s; font-size: 0.7rem; color: var(--muted); }
.compose-arrow.expanded { transform: rotate(90deg); }
.compose-inline { margin: 0 0 0.5rem 0; border-radius: 8px; overflow: hidden; border: 1px solid var(--border); background: var(--card-bg); }
.compose-inline-header { display: flex; justify-content: space-between; align-items: center; padding: 0.5rem 0.75rem; background: var(--hover); border-bottom: 1px solid var(--border); }
.compose-inline-title { font-size: 0.8rem; font-weight: 600; color: var(--muted); }
.compose-inline-actions { display: flex; gap: 0.35rem; }
.compose-loading { text-align: center; padding: 1rem; color: var(--muted); font-size: 0.8rem; }
.compose-editor { width: 100%; font-family: 'SF Mono', 'Fira Code', monospace; font-size: 0.82rem; background: var(--log-bg); color: var(--log-text); border: none; padding: 0.75rem; resize: vertical; outline: none; line-height: 1.5; }
.compose-viewer { margin: 0; padding: 0.75rem; font-family: 'SF Mono', 'Fira Code', monospace; font-size: 0.8rem; line-height: 1.5; overflow-x: auto; white-space: pre-wrap; word-break: break-word; background: var(--log-bg); color: var(--log-text); }
.deploy-section { border-left: 3px solid var(--primary); }
.deploy-form { display: flex; flex-direction: column; gap: 0.5rem; }
.deploy-form .compose-editor { min-height: 150px; }
.deploy-actions { display: flex; gap: 0.5rem; }
.deploy-output { margin-top: 0.5rem; border-radius: 6px; padding: 0.75rem; font-family: 'SF Mono', 'Fira Code', monospace; font-size: 0.8rem; line-height: 1.4; max-height: 300px; overflow-y: auto; }
.deploy-output pre { margin: 0; white-space: pre-wrap; word-break: break-all; }
.deploy-ok { background: #1a1a2e; color: #4caf50; border: 1px solid #2e7d32; }
.deploy-fail { background: #1a1a2e; color: #f44336; border: 1px solid #c62828; }
.compose-editor-actions { display: flex; gap: 0.5rem; }
.compose-editor { width: 100%; font-family: 'SF Mono', 'Fira Code', monospace; font-size: 0.85rem; background: var(--log-bg); color: var(--log-text); border: 1px solid var(--border); border-radius: 6px; padding: 0.75rem; resize: vertical; outline: none; line-height: 1.5; }
.compose-editor:focus { border-color: var(--primary); }
.compose-viewer { background: var(--log-bg); color: var(--log-text); border-radius: 6px; padding: 0.75rem; font-family: 'SF Mono', 'Fira Code', monospace; font-size: 0.82rem; line-height: 1.5; overflow-x: auto; white-space: pre-wrap; word-break: break-word; }

/* Images */
.image-grid { display: flex; flex-direction: column; gap: 0.35rem; }
.image-card { display: flex; align-items: center; gap: 0.75rem; padding: 0.5rem 0.75rem; border-radius: 6px; border: 1px solid var(--border); }
.image-card:hover { background: var(--hover); }
.image-info { flex: 1; min-width: 0; display: flex; align-items: center; gap: 0.75rem; }
.image-name { font-size: 0.85rem; overflow: hidden; text-overflow: ellipsis; white-space: nowrap; flex: 1; }
.image-size { font-size: 0.78rem; color: var(--muted); white-space: nowrap; }
.image-id { font-size: 0.75rem; color: var(--muted); }

/* Events */
.events-panel { display: flex; flex-direction: column; }
.event-row { display: flex; align-items: center; gap: 0.5rem; padding: 0.5rem; border-bottom: 1px solid var(--border); font-size: 0.82rem; }
.event-row:last-child { border-bottom: none; }
.event-badge { padding: 0.1rem 0.4rem; border-radius: 3px; font-size: 0.7rem; font-weight: 600; text-transform: uppercase; flex-shrink: 0; }
.event-badge.container { background: #e8f5e9; color: #2e7d32; }
.event-badge.task { background: #e3f2fd; color: #1565c0; }
.event-msg { flex: 1; color: var(--text2); overflow: hidden; text-overflow: ellipsis; white-space: nowrap; }
.event-ts { color: var(--muted); font-size: 0.75rem; flex-shrink: 0; }

.loading, .empty-state { text-align: center; padding: 2rem; color: var(--muted); font-size: 0.85rem; }
.empty-state-sm { text-align: center; padding: 0.75rem; color: var(--muted); font-size: 0.8rem; }
</style>
