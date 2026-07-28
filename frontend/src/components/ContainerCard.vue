<template>
  <div class="container-card" @click="$emit('click')">
    <div class="card-status-strip" :class="container.state"></div>
    <div class="card-body">
      <div class="card-row1">
        <span class="card-name">{{ container.name || container.id }}</span>
        <span :class="['card-state', container.state]">{{ container.state }}</span>
      </div>
      <div class="card-row2">{{ container.image }}</div>
      <div class="card-row3">
        <span class="card-up">{{ container.status }}</span>
        <span v-if="container.ports && container.ports.length" class="card-ports">{{ container.ports.filter(p => p.public_port).length }} 端口</span>
      </div>
    </div>
    <div class="card-actions" @click.stop>
      <button class="ca ca-start" :disabled="container.state === 'running'" @click="$emit('action', container.id, 'start')" title="启动">▶</button>
      <button class="ca ca-stop" :disabled="container.state !== 'running'" @click="$emit('action', container.id, 'stop')" title="停止">■</button>
      <button class="ca ca-reload" @click="$emit('action', container.id, 'restart')" title="重启">↻</button>
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
  display: flex; background: var(--card); border-radius: var(--radius); overflow: hidden;
  cursor: pointer; transition: all var(--transition);
  box-shadow: var(--shadow); border: 1px solid var(--border);
}
.container-card:hover { box-shadow: var(--shadow-lg); border-color: var(--primary); transform: translateY(-1px); }
.card-status-strip { width: 4px; flex-shrink: 0; }
.card-status-strip.running { background: linear-gradient(180deg, #22c55e, #16a34a); }
.card-status-strip.exited { background: linear-gradient(180deg, #ef4444, #dc2626); }
.card-status-strip.paused { background: linear-gradient(180deg, #f59e0b, #d97706); }
.card-status-strip.created { background: linear-gradient(180deg, #6366f1, #4f46e5); }
.card-body { flex: 1; padding: 0.75rem 0.75rem 0.75rem 0.75rem; min-width: 0; display: flex; flex-direction: column; gap: 0.2rem; }
.card-row1 { display: flex; align-items: center; gap: 0.5rem; }
.card-name { font-weight: 600; font-size: 0.9rem; color: var(--text); }
.card-state { padding: 0.1rem 0.45rem; border-radius: 20px; font-size: 0.65rem; font-weight: 600; text-transform: uppercase; }
.card-state.running { background: #f0fdf4; color: #16a34a; }
.card-state.exited { background: #fef2f2; color: #dc2626; }
.card-state.paused { background: #fffbeb; color: #d97706; }
.card-state.created { background: #eef2ff; color: #4f46e5; }
.card-row2 { font-size: 0.78rem; color: var(--text2); overflow: hidden; text-overflow: ellipsis; white-space: nowrap; }
.card-row3 { display: flex; gap: 0.75rem; font-size: 0.75rem; color: var(--text2); }
.card-actions { display: flex; flex-direction: column; justify-content: center; gap: 0.2rem; padding: 0.5rem 0.6rem 0.5rem 0; flex-shrink: 0; }
.ca { width: 26px; height: 26px; border: none; border-radius: 6px; cursor: pointer; font-size: 0.75rem; transition: all 0.15s; display: flex; align-items: center; justify-content: center; }
.ca:hover:not(:disabled) { transform: scale(1.15); }
.ca:disabled { opacity: 0.25; cursor: not-allowed; }
.ca-start { background: #f0fdf4; color: #16a34a; }
.ca-stop { background: #fef2f2; color: #dc2626; }
.ca-reload { background: #eff6ff; color: #2563eb; }
</style>
