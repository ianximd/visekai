<template>
  <div class="jobs-container">
    <el-card>
      <template #header>
        <div class="card-header">
          <h2>OCR Jobs</h2>
          <el-space>
            <el-select v-model="statusFilter" placeholder="Filter by status" @change="fetchJobs">
              <el-option label="All" value="" />
              <el-option label="Pending" value="pending" />
              <el-option label="Processing" value="processing" />
              <el-option label="Completed" value="completed" />
              <el-option label="Failed" value="failed" />
              <el-option label="Cancelled" value="cancelled" />
            </el-select>
            <el-button @click="fetchJobs" :icon="Refresh">Refresh</el-button>
          </el-space>
        </div>
      </template>

      <el-table
        v-loading="loading"
        :data="jobs"
        style="width: 100%"
        stripe
      >
        <el-table-column prop="id" label="Job ID" width="100">
          <template #default="{ row }">
            <el-text truncated>{{ row.id?.substring(0, 8) }}</el-text>
          </template>
        </el-table-column>
        <el-table-column prop="document_name" label="Document" min-width="200" />
        <el-table-column prop="ocr_mode" label="Mode" width="120">
          <template #default="{ row }">
            <el-tag size="small">{{ row.ocr_mode }}</el-tag>
          </template>
        </el-table-column>
        <el-table-column prop="status" label="Status" width="120">
          <template #default="{ row }">
            <el-tag :type="getStatusType(row.status)">
              {{ row.status }}
            </el-tag>
          </template>
        </el-table-column>
        <el-table-column prop="progress_percentage" label="Progress" width="150">
          <template #default="{ row }">
            <el-progress
              :percentage="row.progress_percentage || 0"
              :status="getProgressStatus(row.status)"
              :stroke-width="8"
            />
          </template>
        </el-table-column>
        <el-table-column prop="created_at" label="Created" width="180">
          <template #default="{ row }">
            {{ formatDate(row.created_at) }}
          </template>
        </el-table-column>
        <el-table-column label="Actions" width="200" align="center">
          <template #default="{ row }">
            <el-button
              v-if="row.status === 'completed'"
              size="small"
              type="success"
              @click="viewResult(row)"
            >
              View Result
            </el-button>
            <el-button
              v-if="row.status === 'pending' || row.status === 'processing'"
              size="small"
              type="warning"
              @click="cancelJob(row)"
            >
              Cancel
            </el-button>
            <el-button
              v-if="row.status === 'failed'"
              size="small"
              type="primary"
              @click="retryJob(row)"
            >
              Retry
            </el-button>
            <el-button
              size="small"
              type="danger"
              @click="deleteJob(row)"
            >
              Delete
            </el-button>
          </template>
        </el-table-column>
      </el-table>

      <el-pagination
        v-if="total > pageSize"
        v-model:current-page="currentPage"
        v-model:page-size="pageSize"
        :total="total"
        :page-sizes="[10, 20, 50, 100]"
        layout="total, sizes, prev, pager, next, jumper"
        @size-change="handleSizeChange"
        @current-change="handleCurrentChange"
        style="margin-top: 20px; justify-content: center;"
      />
    </el-card>
  </div>
</template>

<script setup>
import { ref, onMounted, onUnmounted } from 'vue'
import { useRouter } from 'vue-router'
import { ElMessage, ElMessageBox } from 'element-plus'
import { Refresh } from '@element-plus/icons-vue'
import api from '../services/api'

const router = useRouter()
const loading = ref(false)
const jobs = ref([])
const currentPage = ref(1)
const pageSize = ref(20)
const total = ref(0)
const statusFilter = ref('')
let refreshInterval = null

const fetchJobs = async () => {
  loading.value = true
  try {
    const response = await api.get('/ocr/jobs', {
      params: {
        page: currentPage.value,
        per_page: pageSize.value,
        status: statusFilter.value
      }
    })
    jobs.value = response.data.data.items || []
    total.value = response.data.data.pagination?.total || 0
  } catch (error) {
    ElMessage.error('Failed to load jobs')
    console.error('Fetch jobs error:', error)
  } finally {
    loading.value = false
  }
}

const getStatusType = (status) => {
  const statusMap = {
    'pending': 'info',
    'processing': 'warning',
    'completed': 'success',
    'failed': 'danger',
    'cancelled': 'info'
  }
  return statusMap[status] || 'info'
}

const getProgressStatus = (status) => {
  if (status === 'completed') return 'success'
  if (status === 'failed') return 'exception'
  return undefined
}

const formatDate = (dateString) => {
  if (!dateString) return '-'
  return new Date(dateString).toLocaleString()
}

const viewResult = (job) => {
  router.push(`/results/${job.id}`)
}

const cancelJob = async (job) => {
  try {
    await api.put(`/ocr/jobs/${job.id}/cancel`)
    ElMessage.success('Job cancelled successfully')
    fetchJobs()
  } catch (error) {
    ElMessage.error('Failed to cancel job')
    console.error('Cancel job error:', error)
  }
}

const retryJob = async (job) => {
  ElMessage.info('Retry job feature coming soon')
  // TODO: Implement job retry
}

const deleteJob = async (job) => {
  try {
    await ElMessageBox.confirm(
      'This will permanently delete the job and its results. Continue?',
      'Warning',
      {
        confirmButtonText: 'Delete',
        cancelButtonText: 'Cancel',
        type: 'warning',
      }
    )
    
    await api.delete(`/ocr/jobs/${job.id}`)
    ElMessage.success('Job deleted successfully')
    fetchJobs()
  } catch (error) {
    if (error !== 'cancel') {
      ElMessage.error('Failed to delete job')
      console.error('Delete error:', error)
    }
  }
}

const handleSizeChange = (newSize) => {
  pageSize.value = newSize
  fetchJobs()
}

const handleCurrentChange = (newPage) => {
  currentPage.value = newPage
  fetchJobs()
}

onMounted(() => {
  fetchJobs()
  // Auto-refresh every 5 seconds for active jobs
  refreshInterval = setInterval(() => {
    if (jobs.value.some(job => job.status === 'pending' || job.status === 'processing')) {
      fetchJobs()
    }
  }, 5000)
})

onUnmounted(() => {
  if (refreshInterval) {
    clearInterval(refreshInterval)
  }
})
</script>

<style scoped>
.jobs-container {
  padding: 20px;
}

.card-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
}

.card-header h2 {
  margin: 0;
  color: #303133;
}
</style>
