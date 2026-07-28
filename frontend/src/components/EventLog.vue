<template>
  <div class="events">
    <div v-for="(event, index) in events" :key="index" class="event">
      <span class="time">{{ formatTime(event.timestamp) }}</span>
      <span :class="['type', event.type.split('.')[0]]">{{ event.type }}</span>
      <span class="data">{{ event.summary }}</span>
    </div>
    <div v-if="events.length === 0" class="empty">暂无事件</div>
  </div>
</template>

<script>
export default {
  name: 'EventLog',
  props: {
    events: { type: Array, default: () => [] }
  },
  methods: {
    formatTime(date) {
      return date.toLocaleTimeString()
    }
  }
}
</script>

<style scoped>
.events { max-height: 600px; overflow-y: auto; }
.event { padding: 0.5rem; border-bottom: 1px solid #ecf0f1; font-size: 0.875rem; }
.event:last-child { border-bottom: none; }
.time { color: #95a5a6; margin-right: 0.5rem; }
.type { font-weight: 600; margin-right: 0.5rem; }
.type.container { color: #27ae60; }
.type.task { color: #3498db; }
.type.audit { color: #f39c12; }
.data { color: #7f8c8d; font-family: monospace; font-size: 0.75rem; }
.empty { text-align: center; padding: 2rem; color: #7f8c8d; }
</style>
