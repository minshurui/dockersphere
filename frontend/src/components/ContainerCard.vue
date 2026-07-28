<template>
  <div :class="['card', container.state]">
    <div class="card-header">
      <span class="name">{{ container.name || container.id }}</span>
      <span :class="['badge', container.state]">{{ container.state }}</span>
    </div>
    <div class="card-body">
      <div class="info">
        <span class="label">镜像:</span>
        <span class="value">{{ container.image }}</span>
      </div>
      <div class="info">
        <span class="label">状态:</span>
        <span class="value">{{ container.status }}</span>
      </div>
      <div class="info">
        <span class="label">ID:</span>
        <span class="value mono">{{ container.id }}</span>
      </div>
      <div v-if="container.ports && container.ports.length > 0" class="info">
        <span class="label">端口:</span>
        <span class="value">
          <span v-for="p in container.ports" :key="p.private_port" class="port">
            {{ p.public_port || '?' }}:{{ p.private_port }}/{{ p.type }}
          </span>
        </span>
      </div>
    </div>
    <div class="card-actions">
      <button @click="$emit('action', container.id, 'start')" :disabled="container.state === 'running'">启动</button>
      <button @click="$emit('action', container.id, 'stop')" :disabled="container.state !== 'running'">停止</button>
      <button @click="$emit('action', container.id, 'restart')">重启</button>
      <button @click="$emit('action', container.id, 'remove')" class="danger">删除</button>
    </div>
  </div>
</template>

<script>
export default {
  name: 'ContainerCard',
  props: {
    container: { type: Object, required: true }
  },
  emits: ['action']
}
</script>

<style scoped>
.card { border: 1px solid #e0e0e0; border-radius: 6px; overflow: hidden; }
.card-header {
  background: #ecf0f1; padding: 0.75rem 1rem;
  display: flex; justify-content: space-between; align-items: center;
}
.name { font-weight: 600; color: #2c3e50; }
.badge { padding: 0.25rem 0.75rem; border-radius: 12px; font-size: 0.75rem; font-weight: 600; text-transform: uppercase; }
.badge.running { background: #27ae60; color: white; }
.badge.exited, .badge.paused { background: #f39c12; color: white; }
.badge.created { background: #3498db; color: white; }
.card-body { padding: 1rem; }
.info { display: flex; gap: 0.5rem; margin-bottom: 0.5rem; }
.label { font-weight: 600; color: #7f8c8d; min-width: 50px; }
.value { color: #2c3e50; }
.mono { font-family: 'Courier New', monospace; font-size: 0.875rem; }
.port { margin-right: 0.5rem; }
.card-actions { padding: 0.75rem 1rem; background: #f8f9fa; display: flex; gap: 0.5rem; }
.card-actions button { padding: 0.5rem 1rem; border: none; border-radius: 4px; cursor: pointer; font-size: 0.875rem; }
.card-actions button:hover:not(:disabled) { opacity: 0.9; }
.card-actions button:disabled { opacity: 0.5; cursor: not-allowed; }
.card-actions button:not(.danger) { background: #3498db; color: white; }
.card-actions .danger { background: #e74c3c; color: white; }
</style>
