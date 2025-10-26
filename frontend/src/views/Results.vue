<template>
  <div class="results-container">
    <el-card v-loading="loading">
      <template #header>
        <div class="card-header">
          <div>
            <h2>OCR Result</h2>
            <el-text type="info" size="small">Job ID: {{ jobId }}</el-text>
          </div>
          <el-space>
            <el-button @click="downloadResult('markdown')">
              <el-icon><download /></el-icon>
              Download Markdown
            </el-button>
            <el-button @click="downloadResult('text')">
              <el-icon><download /></el-icon>
              Download Text
            </el-button>
            <el-button @click="downloadResult('json')">
              <el-icon><download /></el-icon>
              Download JSON
            </el-button>
          </el-space>
        </div>
      </template>

      <div v-if="result" class="result-content">
        <el-descriptions :column="2" border>
          <el-descriptions-item label="Document">
            {{ result.document_name || 'N/A' }}
          </el-descriptions-item>
          <el-descriptions-item label="Processing Time">
            {{ formatProcessingTime(result.processing_time_ms) }}
          </el-descriptions-item>
          <el-descriptions-item label="Confidence Score">
            <el-progress
              :percentage="Math.round((result.confidence_score || 0) * 100)"
              :color="getConfidenceColor(result.confidence_score)"
            />
          </el-descriptions-item>
          <el-descriptions-item label="Pages">
            {{ result.num_pages || 1 }}
          </el-descriptions-item>
          <el-descriptions-item label="Processed At" :span="2">
            {{ formatDate(result.created_at) }}
          </el-descriptions-item>
        </el-descriptions>

        <el-divider />

        <el-tabs v-model="activeTab" class="result-tabs">
          <el-tab-pane label="Markdown Preview" name="markdown">
            <div class="markdown-preview" v-html="renderedMarkdown"></div>
          </el-tab-pane>
          <el-tab-pane label="Raw Markdown" name="raw-markdown">
            <el-input
              v-model="result.markdown_text"
              type="textarea"
              :rows="20"
              readonly
              class="raw-text"
            />
          </el-tab-pane>
          <el-tab-pane label="Plain Text" name="text">
            <el-input
              v-model="result.raw_text"
              type="textarea"
              :rows="20"
              readonly
              class="raw-text"
            />
          </el-tab-pane>
          <el-tab-pane label="JSON Data" name="json">
            <el-input
              :model-value="formatJSON(result.json_data)"
              type="textarea"
              :rows="20"
              readonly
              class="raw-text"
            />
          </el-tab-pane>
        </el-tabs>
      </div>

      <el-empty v-else description="No result data available" />
    </el-card>
  </div>
</template>

<script setup>
import { ref, computed, onMounted } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { ElMessage } from 'element-plus'
import { Download } from '@element-plus/icons-vue'
import { marked } from 'marked'
import api from '../services/api'

const route = useRoute()
const router = useRouter()
const loading = ref(false)
const result = ref(null)
const activeTab = ref('markdown')
const jobId = computed(() => route.params.id)

const fetchResult = async () => {
  loading.value = true
  try {
    const response = await api.get(`/results/${jobId.value}`)
    result.value = response.data.data
  } catch (error) {
    ElMessage.error('Failed to load result')
    console.error('Fetch result error:', error)
    setTimeout(() => router.push('/jobs'), 2000)
  } finally {
    loading.value = false
  }
}

const renderedMarkdown = computed(() => {
  if (!result.value?.markdown_text) return '<p>No markdown content</p>'
  return marked(result.value.markdown_text)
})

const formatProcessingTime = (ms) => {
  if (!ms) return 'N/A'
  if (ms < 1000) return `${ms}ms`
  return `${(ms / 1000).toFixed(2)}s`
}

const formatDate = (dateString) => {
  if (!dateString) return 'N/A'
  return new Date(dateString).toLocaleString()
}

const getConfidenceColor = (score) => {
  if (!score) return '#909399'
  if (score >= 0.9) return '#67c23a'
  if (score >= 0.7) return '#e6a23c'
  return '#f56c6c'
}

const formatJSON = (data) => {
  if (!data) return '{}'
  return JSON.stringify(data, null, 2)
}

const downloadResult = async (format) => {
  try {
    const response = await api.get(`/results/${jobId.value}/download`, {
      params: { format },
      responseType: 'blob'
    })
    
    const url = window.URL.createObjectURL(new Blob([response.data]))
    const link = document.createElement('a')
    link.href = url
    link.setAttribute('download', `result-${jobId.value}.${format === 'json' ? 'json' : 'txt'}`)
    document.body.appendChild(link)
    link.click()
    link.remove()
    window.URL.revokeObjectURL(url)
    
    ElMessage.success(`Downloaded as ${format}`)
  } catch (error) {
    ElMessage.error('Failed to download result')
    console.error('Download error:', error)
  }
}

onMounted(() => {
  fetchResult()
})
</script>

<style scoped>
.results-container {
  padding: 20px;
  max-width: 1200px;
  margin: 0 auto;
}

.card-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
}

.card-header h2 {
  margin: 0 0 5px 0;
  color: #303133;
}

.result-content {
  margin-top: 20px;
}

.result-tabs {
  margin-top: 20px;
}

.markdown-preview {
  padding: 20px;
  background: #f5f7fa;
  border-radius: 4px;
  min-height: 400px;
  font-family: -apple-system, BlinkMacSystemFont, 'Segoe UI', Roboto, Oxygen, Ubuntu, Cantarell, sans-serif;
  line-height: 1.6;
}

.markdown-preview :deep(h1) {
  border-bottom: 2px solid #dcdfe6;
  padding-bottom: 10px;
}

.markdown-preview :deep(h2) {
  border-bottom: 1px solid #dcdfe6;
  padding-bottom: 8px;
}

.markdown-preview :deep(code) {
  background: #fff;
  padding: 2px 6px;
  border-radius: 3px;
  font-family: 'Courier New', monospace;
}

.markdown-preview :deep(pre) {
  background: #282c34;
  color: #abb2bf;
  padding: 16px;
  border-radius: 4px;
  overflow-x: auto;
}

.markdown-preview :deep(table) {
  border-collapse: collapse;
  width: 100%;
  margin: 16px 0;
}

.markdown-preview :deep(th),
.markdown-preview :deep(td) {
  border: 1px solid #dcdfe6;
  padding: 8px 12px;
}

.markdown-preview :deep(th) {
  background: #f5f7fa;
  font-weight: bold;
}

.raw-text {
  font-family: 'Courier New', monospace;
  font-size: 13px;
}
</style>
