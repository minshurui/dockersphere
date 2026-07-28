<template>
  <div class="container-card" @click="$emit('click')">
    <div class="card-left">
      <span :class="['status-indicator', container.state]" :title="container.state"></span>
    </div>
    <div class="card-body">
      <div class="card-top">
        <span class="card-name">{{ container.name || container.id }}</span>
        <span :class="['card-badge', container.state]">{{ container.state }}</span>
      </div>
      <div class="card-meta">
        <span class="meta-image" :title="container.image">{{ container.image }}</span>
      </div>
      <div class="card-details">
        <span class="meta-status">{{ container.status }}</span>
        <span v-if="container.ports && container.ports.length" class="meta-ports">
          {{ container.ports.length }} 个端口
        </span>
      </div>
    </div>
    <div class="card-actions" @click.stop>
      <button class="act act-start" :disabled="container.state === 'running'" @click="$emit('action', container.id, 'start')" title="启动">▶</button>
      <button class="act act-stop" :disabled="container.state !== 'running'" @click="$emit('action', container.id, 'stop')" title="停止">■</button>
      <button class="act act-restart" @click="$emit('action', container.id, 'restart')" title="重启">↻</button>
    </div>
  </div>
</template>

<script>
export default {
  name: 'ContainerCard',
  props: { container: { type: Object, required: true } },
  emits: ['action', 'click']
}
</script>

<style scoped>
.container-card {
  display: flex; align-items: stretch; background: #fff; border-radius: 12px;
  overflow: hidden; cursor: pointer; transition: all 0.15s;
  box-shadow: 0 1px 3px rgba(0,0,0,0.06); border: 1px solid transparent;
}
.container-card:hover { box-shadow: 0 4px 12px rgba(0,0,0,0.1); border-color: #6c63ff; transform: translateY(-1px); }
.card-left {
  width: 6px; display: flex; align-items: center; justify-content: center; flex-shrink: 0;
}
.status-indicator { width: 8px; height: 8px; border-radius: 50%; display: block; margin: 0 8px; }
.status-indicator.running { background: #4caf50; box-shadow: 0 0 6px rgba(76,175,80,0.5); }
.status-indicator.exited { background: #f44336; }
.status-indicator.paused { background: #ff9800; }
.status-indicator.created { background: #2196f3; }
.card-body { flex: 1; padding: 0.9rem 0.75rem 0.9rem 0; min-width: 0; }
.card-top { display: flex; align-items: center; gap: 0.5rem; margin-bottom: 0.3rem; }
.card-name { font-weight: 600; font-size: 0.95rem; color: #1a1a2e; }
.card-badge { padding: 0.1rem 0.5rem; border-radius: 8px; font-size: 0.7rem; font-weight: 600; text-transform: uppercase; }
.card-badge.running { background: #e8f5e9; color: #2e7d32; }
.card-badge.exited { background: #fbe9e7; color: #c62828; }
.card-badge.paused { background: #fff3e0; color: #e65100; }
.card-badge.created { background: #e3f2fd; color: #1565c0; }
.card-meta { margin-bottom: 0.2rem; }
.meta-image { font-size: 0.8rem; color: #888; overflow: hidden; text-overflow: ellipsis; white-space: nowrap; display: block; }
.card-details { display: flex; gap: 1rem; font-size: 0.78rem; color: #aaa; }
.card-actions {
  display: flex; flex-direction: column; justify-content: center; gap: 0.25rem;
  padding: 0.5rem 0.75rem 0.5rem 0; flex-shrink: 0;
}
.act {
  width: 28px; height: 28px; border: none; border-radius: 6px; cursor: pointer;
  font-size: 0.8rem; transition: all 0.1s; display: flex; align-items: center; justify-content: center;
}
.act:hover:not(:disabled) { transform: scale(1.1); }
.act:disabled { opacity: 0.3; cursor: not-allowed; }
.act-start { background: #e8f5e9; color: #2e7d32; }
.act-stop { background: #fbe9e7; color: #c62828; }
.act-restart { background: #e3f2fd; color: #1565c0; }
</style>
