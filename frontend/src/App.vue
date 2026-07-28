<template>
  <div :class="['app', darkMode ? 'dark' : '']">
    <!-- Sidebar -->
    <aside class="sidebar">
      <div class="logo">
        <div class="logo-icon">
          <svg viewBox="0 0 24 24" width="28" height="28" fill="none" stroke="currentColor" stroke-width="2">
            <circle cx="12" cy="12" r="3"/><circle cx="12" cy="12" r="8" stroke-dasharray="4 3"/>
          </svg>
        </div>
        <span class="logo-text">Sphere</span>
      </div>

      <nav class="nav">
        <a v-for="tab in tabs" :key="tab.id" :class="['nav-item', { active: activeTab === tab.id }]" @click="activeTab = tab.id">
          <span class="nav-icon" v-html="tab.icon"></span>
          <span class="nav-label">{{ tab.label }}</span>
          <span v-if="tab.badge" class="nav-badge">{{ tab.badge }}</span>
        </a>
      </nav>

      <div class="sidebar-footer">
        <div class="status-row">
          <span :class="['status-dot', connected ? 'online' : 'offline']"></span>
          <span class="status-text">{{ connected ? '已连接' : '已断开' }}</span>
        </div>
        <div class="theme-btn" @click="toggleDark">{{ darkMode ? '☀️' : '🌙' }}</div>
      </div>
    </aside>

    <!-- Main -->
    <div class="main-area">
      <header class="topbar">
        <div class="topbar-left">
          <h1 class="page-title">{{ currentTab.label }}</h1>
          <div class="page-actions">
            <div v-if="activeTab === 'containers'" class="search-wrap">
              <span class="search-icon">🔍</span>
              <input v-model="searchQuery" placeholder="搜索容器..." class="search-input" />
            </div>
          </div>
        </div>
        <div class="topbar-right">
          <span class="count-badge">{{ containers.length }} 容器</span>
          <button class="icon-btn" @click="fetchContainers" title="刷新">↻</button>
        </div>
      </header>

      <main class="content">
        <!-- Dashboard -->
        <div v-if="activeTab === 'dashboard'" class="fade-in">
          <div class="stats-grid">
            <div class="stat-card stat-total">
              <div class="stat-icon">📦</div>
              <div class="stat-info"><div class="stat-value">{{ stats.total }}</div><div class="stat-label">总容器</div></div>
            </div>
            <div class="stat-card stat-running">
              <div class="stat-icon">▶</div>
              <div class="stat-info"><div class="stat-value">{{ stats.running }}</div><div class="stat-label">运行中</div></div>
            </div>
            <div class="stat-card stat-stopped">
              <div class="stat-icon">⏸</div>
              <div class="stat-info"><div class="stat-value">{{ stats.stopped }}</div><div class="stat-label">已停止</div></div>
            </div>
            <div class="stat-card stat-projects">
              <div class="stat-icon">▦</div>
              <div class="stat-info"><div class="stat-value">{{ stats.projects }}</div><div class="stat-label">项目</div></div>
            </div>
          </div>

          <div class="card-section">
            <div class="card-section-header">
              <h3>系统信息</h3>
            </div>
            <div v-if="sysInfo" class="sys-grid">
              <div class="sys-item" v-for="s in sysItems" :key="s.label">
                <span class="sys-label">{{ s.label }}</span>
                <span class="sys-value">{{ s.value }}</span>
              </div>
            </div>
          </div>

          <div class="card-section">
            <div class="card-section-header">
              <h3>实时事件</h3>
              <span class="header-badge">{{ events.length }}</span>
            </div>
            <div class="event-timeline">
              <div v-for="(e, i) in events.slice(0, 8)" :key="i" class="event-item">
                <span :class="['event-marker', e.type.split('.')[0]]"></span>
                <span class="event-text">{{ e.summary || e.type }}</span>
                <span class="event-time">{{ formatTime(e.timestamp) }}</span>
              </div>
              <div v-if="events.length === 0" class="empty-state">暂无事件</div>
            </div>
          </div>
        </div>

        <!-- Containers -->
        <div v-if="activeTab === 'containers'" class="fade-in">
          <div class="card-section deploy-section">
            <div class="card-section-header" style="cursor:pointer" @click="showDeploy = !showDeploy">
              <h3>{{ showDeploy ? '▼' : '▶' }} 快速部署</h3>
              <span class="header-hint">粘贴 compose 内容直接启动</span>
            </div>
            <div v-if="showDeploy" class="deploy-body">
              <input v-model="deployProject" placeholder="项目名称" class="input" />
              <textarea v-model="deployContent" placeholder="粘贴 docker-compose.yml 内容..." class="input textarea" rows="5"></textarea>
              <div class="deploy-actions">
                <button @click="deployStack" :disabled="deploying" class="btn btn-primary">
                  {{ deploying ? '部署中...' : '🚀 部署' }}
                </button>
              </div>
              <div v-if="deployOutput" :class="['deploy-output-box', deploySuccess ? 'success' : 'error']"><pre>{{ deployOutput }}</pre></div>
            </div>
          </div>

          <div v-if="loading" class="loading">加载中...</div>
          <template v-else>
            <div v-for="group in filteredGroups" :key="group.project" class="compose-group">
              <div class="group-header" @click="toggleCompose(group)">
                <span :class="['group-arrow', { open: composeExpanded[group.project] }]">▶</span>
                <span class="group-icon">▦</span>
                <span class="group-name">{{ group.project }}</span>
                <span class="group-count">{{ group.containers.length }} 服务</span>
                <span class="group-file">{{ group.configFiles ? group.configFiles.split('/').pop() : '' }}</span>
                <div class="group-actions" @click.stop>
                  <button @click="projectAction(group.project, 'start')" class="btn-icon-sm btn-start" title="启动">▶</button>
                  <button @click="projectAction(group.project, 'stop')" class="btn-icon-sm btn-stop" title="停止">■</button>
                  <button @click="projectAction(group.project, 'restart')" class="btn-icon-sm btn-reload" title="重启">↻</button>
                </div>
              </div>
              <div v-if="composeExpanded[group.project]" class="group-compose">
                <div class="compose-toolbar">
                  <span class="compose-filename">docker-compose.yml</span>
                  <div class="compose-tools">
                    <button v-if="composeEditingKey !== group.project" @click.stop="editCompose(group)" class="btn-tiny">编辑</button>
                    <template v-if="composeEditingKey === group.project">
                      <button @click.stop="saveCompose()" class="btn-tiny btn-tiny-primary">保存</button>
                      <button @click.stop="cancelCompose()" class="btn-tiny btn-tiny-danger">取消</button>
                    </template>
                  </div>
                </div>
                <textarea v-if="composeEditingKey === group.project" v-model="composeEditContent" class="compose-textarea" rows="12" @click.stop></textarea>
                <pre v-else class="compose-pre">{{ group._content || '加载中...' }}</pre>
              </div>
              <div class="group-cards">
                <ContainerCard v-for="c in group.containers" :key="c.id" :container="c" @action="action" @click="openDetail(c)" />
              </div>
            </div>
            <div v-if="standaloneContainers.length > 0" class="compose-group">
              <div class="group-header"><span class="group-icon">◯</span><span class="group-name">独立容器</span><span class="group-count">{{ standaloneContainers.length }}</span></div>
              <div class="group-cards">
                <ContainerCard v-for="c in standaloneContainers" :key="c.id" :container="c" @action="action" @click="openDetail(c)" />
              </div>
            </div>
          </template>
        </div>

        <!-- Detail -->
        <div v-if="activeTab === 'detail' && selectedContainer" class="fade-in">
          <button class="back-btn" @click="activeTab = 'containers'">← 返回</button>
          <div class="detail-hero">
            <div class="detail-hero-info">
              <span v-if="selectedContainer.composeProject" class="hero-tag">{{ selectedContainer.composeProject }}/{{ selectedContainer.composeService }}</span>
              <h2>{{ selectedContainer.name }}</h2>
            </div>
            <span :class="['hero-status', selectedContainer.state]">{{ selectedContainer.state }}</span>
          </div>

          <div class="detail-grid">
            <div class="info-card">
              <div class="info-card-title">基本信息</div>
              <div class="info-row"><span>ID</span><span class="mono">{{ selectedContainer.id }}</span></div>
              <div class="info-row"><span>镜像</span><span>{{ selectedContainer.image }}</span></div>
              <div class="info-row"><span>状态</span><span>{{ selectedContainer.status }}</span></div>
              <div class="info-row"><span>创建</span><span>{{ selectedContainer.created }}</span></div>
              <div class="info-row"><span>重启策略</span><span>{{ selectedContainer.labels?.['com.docker.compose.restart'] || 'unless-stopped' }}</span></div>
            </div>
            <div class="info-card">
              <div class="info-card-title">端口映射</div>
              <div v-if="selectedContainer.ports && selectedContainer.ports.length > 0">
                <div v-for="p in mappedPorts(selectedContainer.ports)" :key="p.key" class="port-row">
                  <span class="port-type-tag">{{ p.type }}</span>
                  <span class="port-val"><span class="port-host">{{ p.binding }}</span><span class="port-arrow">→</span><span>{{ p.private }}</span></span>
                </div>
              </div>
              <div v-else class="empty-sm">无端口映射</div>
            </div>
            <div class="info-card actions-card">
              <div class="info-card-title">操作</div>
              <div class="action-btns">
                <button @click="action(selectedContainer.id, 'start')" :disabled="selectedContainer.state === 'running'" class="btn action-start">▶ 启动</button>
                <button @click="action(selectedContainer.id, 'stop')" :disabled="selectedContainer.state !== 'running'" class="btn action-stop">■ 停止</button>
                <button @click="action(selectedContainer.id, 'restart')" class="btn action-reload">↻ 重启</button>
                <button @click="action(selectedContainer.id, 'remove')" class="btn action-remove">✕ 删除</button>
              </div>
            </div>
            <div class="info-card wide-card">
              <div class="info-card-title">日志 <span v-if="!showLogs" class="title-hint">最近 50 行</span></div>
              <button v-if="!showLogs" @click="fetchLogs" class="btn btn-log-btn">📋 加载日志</button>
              <div v-if="showLogs" class="log-box"><div v-for="(line, i) in logs" :key="i" class="log-line">{{ line }}</div></div>
            </div>
            <div class="info-card wide-card">
              <div class="info-card-title">执行命令</div>
              <div class="exec-row">
                <input v-model="execCmd" @keyup.enter="runExec" placeholder="例如: ps aux" class="input exec-input" />
                <button @click="runExec" class="btn btn-exec-btn">运行</button>
              </div>
              <div v-if="execOutput" class="exec-box"><pre>{{ execOutput }}</pre></div>
            </div>
          </div>
        </div>

        <!-- Images -->
        <div v-if="activeTab === 'images'" class="fade-in">
          <div class="card-section">
            <div class="card-section-header">
              <h3>镜像列表</h3>
              <span class="header-badge">{{ images.length }} 个 · {{ formatBytes(totalImageSize) }}</span>
            </div>
            <div v-if="imagesLoading" class="loading">加载中...</div>
            <div v-else class="image-list">
              <div v-for="img in images" :key="img.ID" class="image-row">
                <div class="image-info">
                  <span class="image-name">{{ img.RepoTags?.[0] || img.ID?.slice(7,19) || '无标签' }}</span>
                  <span class="image-size">{{ formatBytes(img.Size) }}</span>
                  <span class="image-id">{{ img.ID?.slice(7,19) }}</span>
                </div>
                <button @click="removeImage(img.ID)" class="btn-icon-sm btn-remove-img" :disabled="imageRemoving === img.ID">✕</button>
              </div>
              <div v-if="images.length === 0" class="empty-state">无镜像</div>
            </div>
          </div>
        </div>

        <!-- Events -->
        <div v-if="activeTab === 'events'" class="fade-in">
          <div class="card-section">
            <div class="card-section-header"><h3>事件日志</h3></div>
            <div class="event-list">
              <div v-for="(e, i) in events" :key="i" class="event-row">
                <span :class="['event-tag', e.type.split('.')[0]]">{{ e.type.split('.')[0] }}</span>
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
      activeTab: 'dashboard', containers: [], images: [], events: [],
      selectedContainer: null, loading: false, imagesLoading: false, connected: false,
      ws: null, darkMode: dark, searchQuery: '',
      showLogs: false, logs: [], execCmd: '', execOutput: '',
      sysInfo: null, imageRemoving: null,
      composeExpanded: {}, composeEditingKey: '', composeEditContent: '', composeEditPath: '',
      showDeploy: false, deployProject: '', deployContent: '', deployOutput: '', deploying: false, deploySuccess: false,
      tabs: [
        { id: 'dashboard', label: '仪表盘', icon: '&#9675;' },
        { id: 'containers', label: '容器', icon: '&#9635;' },
        { id: 'images', label: '镜像', icon: '&#9675;' },
        { id: 'events', label: '事件', icon: '&#9889;' },
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
    sysItems() {
      if (!this.sysInfo) return []
      return [
        { label: 'Docker 版本', value: this.sysInfo.ServerVersion },
        { label: '系统', value: `${this.sysInfo.OSType} ${this.sysInfo.Architecture}` },
        { label: 'CPU', value: `${this.sysInfo.NCPU} 核` },
        { label: '内存', value: this.formatBytes(this.sysInfo.MemTotal) },
        { label: '运行容器', value: this.sysInfo.ContainersRunning },
      ]
    },
    filteredContainers() {
      if (!this.searchQuery) return this.containers
      const q = this.searchQuery.toLowerCase()
      return this.containers.filter(c => c.name.toLowerCase().includes(q) || c.image.toLowerCase().includes(q) || c.state.includes(q))
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
    standaloneContainers() { return this.containers.filter(c => !c.labels?.['com.docker.compose.project']).map(c => ({ ...c, composeProject: null, composeService: null })) },
    totalImageSize() { return this.images.reduce((s, img) => s + img.Size, 0) }
  },
  mounted() { this.fetchContainers(); this.fetchImages(); this.fetchSysInfo(); this.connectWebSocket() },
  beforeUnmount() { if (this.ws) this.ws.close() },
  methods: {
    async fetchContainers() { try { this.loading = true; const r = await axios.get('/api/v1/containers'); if (r.data.code === 0) this.containers = r.data.data || [] } catch (e) { console.error(e) } finally { this.loading = false } },
    async fetchImages() { try { this.imagesLoading = true; const r = await axios.get('/api/v1/images'); if (r.data.code === 0) this.images = r.data.data || [] } catch (e) { console.error(e) } finally { this.imagesLoading = false } },
    async fetchSysInfo() { try { const r = await axios.get('/api/v1/system/info'); if (r.data.code === 0) this.sysInfo = r.data.data } catch (e) { /* ignore */ } },
    async removeImage(id) { this.imageRemoving = id; try { await axios.delete(`/api/v1/images/${id}`); this.images = this.images.filter(i => i.ID !== id) } catch (e) { alert('删除失败') }; this.imageRemoving = null },
    async action(id, act) { try { await axios.post(`/api/v1/containers/${id}/action`, { action: act }) } catch (e) { console.error(e) } },
    openDetail(c) { this.selectedContainer = c; this.activeTab = 'detail'; this.showLogs = false; this.logs = []; this.execOutput = '' },
    async fetchLogs() { this.showLogs = true; try { const r = await axios.get(`/api/v1/containers/${this.selectedContainer.id}/logs?tail=50`); if (r.data.code === 0) this.logs = r.data.data || [] } catch (e) { this.logs = ['获取失败'] } },
    async runExec() { if (!this.execCmd.trim()) return; this.execOutput = '执行中...'; try { const parts = this.execCmd.trim().split(/\s+/); const r = await axios.post(`/api/v1/containers/${this.selectedContainer.id}/exec`, { cmd: parts[0], args: parts.slice(1) }); if (r.data.code === 0) this.execOutput = r.data.data.replace(/[\x00-\x08\x0e-\x1f]/g, '').trim() } catch (e) { this.execOutput = '失败: ' + (e.response?.data?.message || e.message) } },
    mappedPorts(ports) {
      const seen = new Set(); return ports.filter(p => { const k = `${p.private_port}-${p.type}`; if (seen.has(k)) return false; seen.add(k); return true }).map(p => ({ key: `${p.private_port}-${p.type}`, type: p.type.toUpperCase(), binding: p.public_port ? `${p.ip || '0.0.0.0'}:${p.public_port}` : '未映射', private: `${p.private_port}` }))
    },
    toggleCompose(project) {
      const key = project.project; this.composeExpanded[key] = !this.composeExpanded[key]
      if (this.composeExpanded[key] && !project._content && project.configFiles) {
        axios.get('/api/v1/compose/file', { params: { path: project.configFiles } }).then(r => { project._content = r.data.code === 0 ? r.data.data : '加载失败' }).catch(() => { project._content = '加载失败' })
      }
    },
    editCompose(group) { this.composeEditingKey = group.project; this.composeEditContent = group._content; this.composeEditPath = group.configFiles },
    cancelCompose() { this.composeEditingKey = '' },
    async saveCompose() { if (!this.composeEditPath) { alert('错误'); return }; try { const r = await axios.put('/api/v1/compose/file', { path: this.composeEditPath, content: this.composeEditContent }); if (r.data.code !== 0) { alert('保存失败'); return }; for (const g of this.filteredGroups) { if (g.project === this.composeEditingKey) g._content = this.composeEditContent }; this.composeEditingKey = '' } catch (e) { alert('保存失败: ' + (e.response?.data?.message || e.message)) } },
    async projectAction(project, action) { try { const r = await axios.post(`/api/v1/compose/project/${project}/${action}`, {}); if (r.data.code === 0 && r.data.data.success) setTimeout(() => this.fetchContainers(), 2000) } catch (e) { console.error(e) } },
    async deployStack() { if (!this.deployProject.trim() || !this.deployContent.trim()) { alert('请输入项目名称和内容'); return }; this.deploying = true; this.deployOutput = ''; this.deploySuccess = false; try { const r = await axios.post('/api/v1/compose/deploy', { project: this.deployProject, content: this.deployContent }); if (r.data.code === 0) { this.deployOutput = r.data.data.output; this.deploySuccess = r.data.data.success; if (this.deploySuccess) { this.fetchContainers(); this.deployProject = ''; this.deployContent = '' } } } catch (e) { this.deployOutput = '请求失败'; this.deploySuccess = false }; this.deploying = false },
    formatBytes(b) { if (!b) return '0B'; const u = ['B','KB','MB','GB','TB']; const i = Math.floor(Math.log(b)/Math.log(1024)); return (b/Math.pow(1024,i)).toFixed(1) + u[i] },
    toggleDark() { this.darkMode = !this.darkMode; localStorage.setItem('dockersphere-dark', this.darkMode) },
    connectWebSocket() {
      const proto = window.location.protocol === 'https:' ? 'wss:' : 'ws:'; this.ws = new WebSocket(`${proto}//${window.location.host}/ws`)
      this.ws.onopen = () => { this.connected = true }
      this.ws.onmessage = (e) => { try { const msg = JSON.parse(e.data); this.events.unshift({ type: msg.event, summary: (msg.data || {}).name || '', timestamp: new Date() }); if (this.events.length > 200) this.events.pop(); if (msg.event.startsWith('container.')) this.fetchContainers() } catch (_) {} }
      this.ws.onclose = () => { this.connected = false; setTimeout(() => this.connectWebSocket(), 5000) }
    },
    formatTime(d) { return d.toLocaleTimeString() }
  }
}
</script>

<style>
* { margin: 0; padding: 0; box-sizing: border-box; }
body { font-family: -apple-system, BlinkMacSystemFont, 'SF Pro Display', 'Segoe UI', Roboto, sans-serif; background: #f0f2f6; color: #1e293b; -webkit-font-smoothing: antialiased; }

/* Theme - Mecha Cyberpunk */
:root {
  --sidebar-w: 220px;
  --primary: #00f5ff; --primary-light: #66f9ff;
  --accent: #ff00aa; --accent2: #ffd700;
  --success: #00ff88; --danger: #ff0044; --warning: #ffaa00;
  --bg: #0a0e17; --card: rgba(15,23,42,0.9);
  --text: #e2e8f0; --text2: #64748b;
  --border: rgba(0,245,255,0.15);
  --shadow: 0 0 15px rgba(0,245,255,0.08);
  --shadow-lg: 0 0 30px rgba(0,245,255,0.15);
  --radius: 8px; --radius-sm: 4px;
  --transition: 0.2s ease;
  --glow-cyan: 0 0 10px rgba(0,245,255,0.3);
  --glow-pink: 0 0 10px rgba(255,0,170,0.3);
}
.dark { /* same as default - already dark mecha theme */ }

/* Grid background overlay */
body::before {
  content: ''; position: fixed; inset: 0; z-index: -1;
  background-image:
    linear-gradient(rgba(0,245,255,0.03) 1px, transparent 1px),
    linear-gradient(90deg, rgba(0,245,255,0.03) 1px, transparent 1px);
  background-size: 40px 40px;
  pointer-events: none;
}
body::after {
  content: ''; position: fixed; inset: 0; z-index: -1;
  background: radial-gradient(ellipse at 20% 50%, rgba(0,245,255,0.06) 0%, transparent 50%),
              radial-gradient(ellipse at 80% 20%, rgba(255,0,170,0.04) 0%, transparent 50%),
              radial-gradient(ellipse at 50% 80%, rgba(255,215,0,0.03) 0%, transparent 50%);
  pointer-events: none;
}

.app { display: flex; min-height: 100vh; background: var(--bg); color: var(--text); transition: background var(--transition), color var(--transition); }

/* Sidebar - Mecha style */
.sidebar {
  width: var(--sidebar-w);
  background: linear-gradient(180deg, #050a14 0%, #0a0e17 100%);
  color: #e2e8f0;
  display: flex; flex-direction: column;
  position: fixed; top: 0; left: 0; bottom: 0; z-index: 100;
  border-right: 1px solid rgba(0,245,255,0.1);
}
.sidebar::after {
  content: ''; position: absolute; top: 0; right: -1px; width: 1px; height: 100%;
  background: linear-gradient(180deg, transparent 0%, var(--primary) 50%, transparent 100%);
  opacity: 0.4;
}
.logo { padding: 1.25rem; display: flex; align-items: center; gap: 0.75rem; border-bottom: 1px solid rgba(0,245,255,0.08); }
.logo-icon { color: var(--primary); filter: drop-shadow(0 0 6px rgba(0,245,255,0.5)); display: flex; }
.logo-text { font-size: 1.2rem; font-weight: 700; background: linear-gradient(135deg, var(--primary), var(--accent)); -webkit-background-clip: text; -webkit-text-fill-color: transparent; }
.nav { flex: 1; padding: 0.75rem 0; }
.nav-item { display: flex; align-items: center; gap: 0.75rem; padding: 0.7rem 1.25rem; cursor: pointer; color: #4a5578; transition: all var(--transition); font-size: 0.88rem; border-left: 2px solid transparent; position: relative; }
.nav-item:hover { color: var(--primary); background: rgba(0,245,255,0.04); }
.nav-item.active { color: var(--primary); background: rgba(0,245,255,0.08); border-left-color: var(--primary); box-shadow: inset 0 0 20px rgba(0,245,255,0.05); }
.nav-item.active::after { content: ''; position: absolute; right: 0; top: 50%; transform: translateY(-50%); width: 3px; height: 16px; background: var(--primary); border-radius: 2px 0 0 2px; box-shadow: 0 0 8px rgba(0,245,255,0.5); }
.nav-icon { font-size: 1rem; width: 20px; text-align: center; }
.sidebar-footer { padding: 1rem 1.25rem; border-top: 1px solid rgba(0,245,255,0.08); display: flex; align-items: center; justify-content: space-between; }
.status-row { display: flex; align-items: center; gap: 0.5rem; font-size: 0.8rem; }
.status-dot { width: 8px; height: 8px; border-radius: 50%; }
.status-dot.online { background: var(--success); box-shadow: 0 0 10px rgba(0,255,136,0.6); }
.status-dot.offline { background: var(--danger); box-shadow: 0 0 10px rgba(255,0,68,0.6); }
.theme-btn { cursor: pointer; font-size: 1rem; opacity: 0.6; transition: opacity var(--transition); }
.theme-btn:hover { opacity: 1; }

/* Main */
.main-area { margin-left: var(--sidebar-w); flex: 1; display: flex; flex-direction: column; min-height: 100vh; }
.topbar {
  background: rgba(10,14,23,0.95);
  padding: 0.75rem 1.5rem; display: flex; justify-content: space-between; align-items: center;
  border-bottom: 1px solid rgba(0,245,255,0.08);
  position: sticky; top: 0; z-index: 50;
  backdrop-filter: blur(10px);
}
.topbar-left { display: flex; align-items: center; gap: 1.5rem; }
.page-title { font-size: 1.15rem; font-weight: 600; }
.search-wrap { display: flex; align-items: center; gap: 0.5rem; }
.search-icon { font-size: 0.85rem; }
.search-input { padding: 0.4rem 0.75rem; border: 1px solid var(--border); border-radius: var(--radius-sm); font-size: 0.85rem; background: rgba(0,245,255,0.03); color: var(--text); outline: none; width: 200px; transition: all var(--transition); }
.search-input:focus { border-color: var(--primary); box-shadow: 0 0 8px rgba(0,245,255,0.15); }
.topbar-right { display: flex; align-items: center; gap: 0.75rem; }
.count-badge { font-size: 0.8rem; color: var(--text2); background: var(--bg); padding: 0.2rem 0.6rem; border-radius: 20px; }
.icon-btn { background: none; border: 1px solid var(--border); border-radius: var(--radius-sm); padding: 0.35rem 0.5rem; cursor: pointer; font-size: 1rem; color: var(--text2); transition: all var(--transition); }
.icon-btn:hover { background: var(--bg); color: var(--text); }

.content { padding: 1.5rem; flex: 1; }
.fade-in { animation: fadeIn 0.25s ease; }
@keyframes fadeIn { from { opacity: 0; transform: translateY(8px); } to { opacity: 1; transform: translateY(0); } }

/* Stats - Mecha style */
.stats-grid { display: grid; grid-template-columns: repeat(4, 1fr); gap: 1rem; margin-bottom: 1.5rem; }
.stat-card {
  background: var(--card); border-radius: var(--radius); padding: 1.25rem;
  display: flex; align-items: center; gap: 1rem;
  box-shadow: var(--shadow); border: 1px solid var(--border);
  transition: all var(--transition); position: relative; overflow: hidden;
}
.stat-card::before {
  content: ''; position: absolute; top: 0; left: 0; right: 0; height: 2px;
  opacity: 0.6; transition: opacity var(--transition);
}
.stat-total::before { background: linear-gradient(90deg, var(--primary), var(--accent)); }
.stat-running::before { background: linear-gradient(90deg, var(--success), var(--primary)); }
.stat-stopped::before { background: linear-gradient(90deg, var(--danger), var(--accent)); }
.stat-projects::before { background: linear-gradient(90deg, var(--warning), var(--accent2)); }
.stat-card:hover { transform: translateY(-2px); box-shadow: var(--glow-cyan); border-color: rgba(0,245,255,0.3); }
.stat-card:hover::before { opacity: 1; }
.stat-icon { font-size: 1.3rem; width: 44px; height: 44px; display: flex; align-items: center; justify-content: center; border-radius: var(--radius-sm); border: 1px solid rgba(0,245,255,0.15); background: rgba(0,245,255,0.05); }
.stat-total .stat-icon { color: var(--primary); }
.stat-running .stat-icon { color: var(--success); }
.stat-stopped .stat-icon { color: var(--danger); }
.stat-projects .stat-icon { color: var(--warning); }
.stat-value { font-size: 1.6rem; font-weight: 700; line-height: 1.2; font-family: 'SF Mono', 'Fira Code', monospace; }
.stat-label { font-size: 0.78rem; color: var(--text2); text-transform: uppercase; letter-spacing: 1px; }

/* Card sections */
.card-section { background: var(--card); border-radius: var(--radius); padding: 1.25rem; box-shadow: var(--shadow); margin-bottom: 1rem; }
.card-section-header { display: flex; align-items: center; justify-content: space-between; margin-bottom: 1rem; }
.card-section-header h3 { font-size: 0.95rem; font-weight: 600; }
.header-badge { font-size: 0.78rem; color: var(--text2); background: var(--bg); padding: 0.15rem 0.5rem; border-radius: 20px; }
.header-hint { font-size: 0.78rem; color: var(--text2); }

/* System info */
.sys-grid { display: flex; flex-wrap: wrap; gap: 1.5rem; }
.sys-item { display: flex; flex-direction: column; gap: 0.2rem; }
.sys-label { font-size: 0.75rem; color: var(--text2); text-transform: uppercase; letter-spacing: 0.3px; }
.sys-value { font-size: 0.95rem; font-weight: 500; }

/* Events timeline */
.event-timeline { display: flex; flex-direction: column; }
.event-item { display: flex; align-items: center; gap: 0.75rem; padding: 0.5rem 0; border-bottom: 1px solid var(--border); font-size: 0.83rem; }
.event-item:last-child { border-bottom: none; }
.event-marker { width: 8px; height: 8px; border-radius: 50%; flex-shrink: 0; }
.event-marker.container { background: var(--success); }
.event-marker.task { background: var(--primary); }
.event-text { flex: 1; color: var(--text2); overflow: hidden; text-overflow: ellipsis; white-space: nowrap; }
.event-time { font-size: 0.75rem; color: var(--text2); flex-shrink: 0; }

/* Compose groups - Mecha */
.compose-group { margin-bottom: 1rem; }
.group-header {
  display: flex; align-items: center; gap: 0.5rem; padding: 0.6rem 0.75rem;
  background: var(--card); border-radius: var(--radius);
  box-shadow: var(--shadow); border: 1px solid var(--border);
  cursor: pointer; transition: all var(--transition);
}
.group-header:hover { box-shadow: var(--glow-cyan); border-color: rgba(0,245,255,0.25); }
.group-arrow { font-size: 0.7rem; color: var(--text2); transition: transform var(--transition); }
.group-arrow.open { transform: rotate(90deg); color: var(--primary); }
.group-icon { color: var(--primary); font-size: 0.9rem; }
.group-name { font-weight: 600; font-size: 0.9rem; }
.group-count { margin-left: auto; font-size: 0.78rem; color: var(--text2); }
.group-file { font-size: 0.72rem; color: var(--text2); font-family: monospace; letter-spacing: 0.5px; }
.group-actions { display: flex; gap: 0.2rem; margin-left: 0.5rem; flex-shrink: 0; }
.group-cards { display: grid; grid-template-columns: repeat(auto-fill, minmax(320px, 1fr)); gap: 0.6rem; margin-top: 0.6rem; }

/* Compose inline */
.group-compose { margin: 0.5rem 0; border-radius: var(--radius-sm); overflow: hidden; border: 1px solid var(--border); }
.compose-toolbar { display: flex; justify-content: space-between; align-items: center; padding: 0.5rem 0.75rem; background: var(--bg); border-bottom: 1px solid var(--border); }
.compose-filename { font-size: 0.8rem; font-weight: 500; color: var(--text2); }
.compose-tools { display: flex; gap: 0.3rem; }
.btn-tiny { padding: 0.15rem 0.5rem; border: 1px solid var(--border); border-radius: 4px; background: var(--card); cursor: pointer; font-size: 0.72rem; color: var(--text2); transition: all var(--transition); }
.btn-tiny:hover { background: var(--bg); }
.btn-tiny-primary { border-color: var(--primary); color: var(--primary); }
.btn-tiny-danger { border-color: var(--danger); color: var(--danger); }
.compose-textarea { width: 100%; border: none; padding: 0.75rem; font-family: 'SF Mono', 'Fira Code', monospace; font-size: 0.8rem; background: #1e1b4b; color: #e2e8f0; resize: vertical; outline: none; line-height: 1.5; min-height: 120px; }
.compose-pre { margin: 0; padding: 0.75rem; font-family: 'SF Mono', 'Fira Code', monospace; font-size: 0.8rem; line-height: 1.5; background: #1e1b4b; color: #e2e8f0; overflow-x: auto; white-space: pre-wrap; word-break: break-word; }

/* Detail - Mecha */
.back-btn { background: none; border: 1px solid var(--border); color: var(--primary); cursor: pointer; font-size: 0.82rem; padding: 0.3rem 0.75rem; border-radius: var(--radius-sm); margin-bottom: 0.75rem; transition: all var(--transition); }
.back-btn:hover { border-color: var(--primary); box-shadow: var(--glow-cyan); }
.detail-hero {
  display: flex; align-items: center; gap: 1rem; margin-bottom: 1.25rem; padding: 1.25rem;
  background: var(--card); border-radius: var(--radius); box-shadow: var(--shadow);
  border: 1px solid var(--border); position: relative; overflow: hidden;
}
.detail-hero::before {
  content: ''; position: absolute; top: 0; left: 0; right: 0; height: 2px;
  background: linear-gradient(90deg, var(--primary), var(--accent), var(--primary));
  opacity: 0.6;
}
.detail-hero-info { flex: 1; }
.hero-tag { font-size: 0.7rem; color: var(--primary); border: 1px solid rgba(0,245,255,0.2); padding: 0.1rem 0.5rem; display: inline-block; margin-bottom: 0.25rem; letter-spacing: 0.5px; }
.detail-hero h2 { font-size: 1.4rem; font-weight: 700; }
.hero-status { padding: 0.25rem 0.75rem; border-radius: 4px; font-size: 0.78rem; font-weight: 600; text-transform: uppercase; letter-spacing: 1px; border: 1px solid; }
.hero-status.running { background: rgba(0,255,136,0.1); color: var(--success); border-color: rgba(0,255,136,0.3); box-shadow: 0 0 10px rgba(0,255,136,0.15); }
.hero-status.exited { background: rgba(255,0,68,0.1); color: var(--danger); border-color: rgba(255,0,68,0.3); box-shadow: 0 0 10px rgba(255,0,68,0.15); }
.detail-grid { display: grid; grid-template-columns: 1fr 1fr; gap: 0.75rem; }
.info-card { background: var(--card); border-radius: var(--radius); padding: 1.25rem; box-shadow: var(--shadow); }
.wide-card { grid-column: 1 / -1; }
.info-card-title { font-size: 0.8rem; font-weight: 600; color: var(--text2); text-transform: uppercase; letter-spacing: 0.3px; margin-bottom: 0.75rem; display: flex; align-items: center; gap: 0.5rem; }
.title-hint { font-weight: 400; text-transform: none; font-size: 0.75rem; color: var(--text2); }
.info-row { display: flex; padding: 0.35rem 0; font-size: 0.85rem; border-bottom: 1px solid var(--border); }
.info-row:last-child { border-bottom: none; }
.info-row span:first-child { width: 70px; color: var(--text2); flex-shrink: 0; }
.info-row span:last-child { color: var(--text); }
.mono { font-family: 'SF Mono', 'Fira Code', monospace; font-size: 0.82rem; }

/* Ports */
.port-row { display: flex; align-items: center; gap: 0.5rem; padding: 0.3rem 0; border-bottom: 1px solid var(--border); font-size: 0.85rem; }
.port-row:last-child { border-bottom: none; }
.port-type-tag { font-size: 0.65rem; padding: 0.1rem 0.35rem; border-radius: 3px; background: var(--bg); color: var(--text2); font-weight: 600; width: 32px; text-align: center; flex-shrink: 0; }
.port-val { display: flex; align-items: center; gap: 0.4rem; }
.port-host { color: var(--primary); font-family: monospace; }
.port-arrow { color: var(--text2); }

/* Action buttons - Mecha */
.action-btns { display: flex; gap: 0.5rem; flex-wrap: wrap; }
.btn {
  padding: 0.5rem 1rem; border: 1px solid transparent;
  border-radius: var(--radius-sm); cursor: pointer;
  font-size: 0.82rem; font-weight: 500;
  transition: all var(--transition); font-family: inherit;
}
.btn:hover { transform: translateY(-1px); }
.btn:disabled { opacity: 0.25; cursor: not-allowed; transform: none; }
.btn-primary { background: rgba(0,245,255,0.1); color: var(--primary); border-color: rgba(0,245,255,0.3); }
.btn-primary:hover { background: rgba(0,245,255,0.15); box-shadow: var(--glow-cyan); }
.action-start { background: rgba(0,255,136,0.08); color: var(--success); border-color: rgba(0,255,136,0.2); }
.action-start:hover { box-shadow: 0 0 10px rgba(0,255,136,0.2); }
.action-stop { background: rgba(255,0,68,0.08); color: var(--danger); border-color: rgba(255,0,68,0.2); }
.action-stop:hover { box-shadow: 0 0 10px rgba(255,0,68,0.2); }
.action-reload { background: rgba(0,245,255,0.08); color: var(--primary); border-color: rgba(0,245,255,0.2); }
.action-reload:hover { box-shadow: var(--glow-cyan); }
.action-remove { background: rgba(255,0,68,0.08); color: var(--danger); border-color: rgba(255,0,68,0.2); }
.btn-log-btn { background: rgba(255,0,170,0.08); color: var(--accent); border-color: rgba(255,0,170,0.2); }
.btn-log-btn:hover { box-shadow: var(--glow-pink); }
.btn-exec-btn { background: rgba(0,245,255,0.1); color: var(--primary); border-color: rgba(0,245,255,0.3); }
.btn-exec-btn:hover { box-shadow: var(--glow-cyan); }
.btn-log-btn, .btn-exec-btn { padding: 0.4rem 0.8rem; border-radius: var(--radius-sm); cursor: pointer; font-size: 0.8rem; }

/* Log/Exec boxes */
.log-box { background: #1e1b4b; color: #e2e8f0; border-radius: var(--radius-sm); padding: 0.75rem; max-height: 250px; overflow-y: auto; font-family: 'SF Mono', 'Fira Code', monospace; font-size: 0.78rem; line-height: 1.5; }
.log-line { white-space: pre-wrap; word-break: break-all; }
.exec-row { display: flex; gap: 0.5rem; margin-bottom: 0.5rem; }
.exec-input { flex: 1; }
.exec-box { background: #1e1b4b; color: #e2e8f0; border-radius: var(--radius-sm); padding: 0.75rem; max-height: 200px; overflow-y: auto; font-family: 'SF Mono', 'Fira Code', monospace; font-size: 0.78rem; }
.exec-box pre { margin: 0; white-space: pre-wrap; word-break: break-all; }

/* Inputs */
.input { padding: 0.5rem 0.75rem; border: 1px solid var(--border); border-radius: var(--radius-sm); font-size: 0.85rem; background: var(--card); color: var(--text); outline: none; transition: border-color var(--transition); }
.input:focus { border-color: var(--primary); }
.input.textarea { resize: vertical; font-family: 'SF Mono', 'Fira Code', monospace; min-height: 100px; }

/* Deploy */
.deploy-section { border-left: 3px solid var(--primary); }
.deploy-body { display: flex; flex-direction: column; gap: 0.5rem; }
.deploy-actions { display: flex; gap: 0.5rem; }
.deploy-output-box { padding: 0.75rem; border-radius: var(--radius-sm); font-family: 'SF Mono', 'Fira Code', monospace; font-size: 0.78rem; max-height: 200px; overflow-y: auto; }
.deploy-output-box pre { margin: 0; white-space: pre-wrap; }
.deploy-output-box.success { background: #1e1b4b; color: var(--success); border: 1px solid var(--success); }
.deploy-output-box.error { background: #1e1b4b; color: var(--danger); border: 1px solid var(--danger); }

/* Images */
.image-list { display: flex; flex-direction: column; gap: 0.3rem; }
.image-row { display: flex; align-items: center; gap: 0.75rem; padding: 0.5rem 0.75rem; border-radius: var(--radius-sm); border: 1px solid var(--border); transition: background var(--transition); }
.image-row:hover { background: var(--bg); }
.image-info { flex: 1; min-width: 0; display: flex; align-items: center; gap: 0.75rem; }
.image-name { font-size: 0.85rem; overflow: hidden; text-overflow: ellipsis; white-space: nowrap; flex: 1; }
.image-size { font-size: 0.78rem; color: var(--text2); white-space: nowrap; }
.image-id { font-size: 0.75rem; color: var(--text2); font-family: monospace; }
.btn-icon-sm { width: 24px; height: 24px; border: none; border-radius: 4px; cursor: pointer; font-size: 0.7rem; display: flex; align-items: center; justify-content: center; opacity: 0.7; transition: opacity var(--transition); }
.btn-icon-sm:hover { opacity: 1; }
.btn-icon-sm.btn-start { background: #f0fdf4; color: #16a34a; }
.btn-icon-sm.btn-stop { background: #fef2f2; color: #dc2626; }
.btn-icon-sm.btn-reload { background: #eff6ff; color: #2563eb; }
.btn-remove-img { color: var(--danger); }

/* Events page */
.event-list { display: flex; flex-direction: column; }
.event-row { display: flex; align-items: center; gap: 0.5rem; padding: 0.5rem; border-bottom: 1px solid var(--border); font-size: 0.83rem; }
.event-row:last-child { border-bottom: none; }
.event-tag { padding: 0.1rem 0.4rem; border-radius: 3px; font-size: 0.68rem; font-weight: 600; text-transform: uppercase; flex-shrink: 0; }
.event-tag.container { background: #f0fdf4; color: #16a34a; }
.event-tag.task { background: #eff6ff; color: #2563eb; }
.event-msg { flex: 1; color: var(--text2); overflow: hidden; text-overflow: ellipsis; white-space: nowrap; }
.event-ts { font-size: 0.75rem; color: var(--text2); flex-shrink: 0; }

.loading, .empty-state { text-align: center; padding: 2rem; color: var(--text2); font-size: 0.85rem; }
.empty-sm { text-align: center; padding: 0.75rem; color: var(--text2); font-size: 0.8rem; }
</style>
