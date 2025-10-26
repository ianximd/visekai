<template>
  <div class="documents-container">
    <el-card>
      <template #header>
        <div class="card-header">
          <h2>My Documents</h2>
          <el-button type="primary" @click="navigateToUpload">
            <el-icon><upload-filled /></el-icon>
            Upload New
          </el-button>
        </div>
      </template>

      <el-table
        v-loading="loading"
        :data="documents"
        style="width: 100%"
        stripe
      >
        <el-table-column prop="filename" label="Filename" min-width="200">
          <template #default="{ row }">
            <div class="filename-cell">
              <el-icon><document /></el-icon>
              <span>{{ row.original_filename }}</span>
            </div>
          </template>
        </el-table-column>
        <el-table-column prop="file_size" label="Size" width="120">
          <template #default="{ row }">
            {{ formatFileSize(row.file_size) }}
          </template>
        </el-table-column>
        <el-table-column prop="uploaded_at" label="Uploaded" width="180">
          <template #default="{ row }">
            {{ formatDate(row.uploaded_at) }}
          </template>
        </el-table-column>
        <el-table-column prop="status" label="Status" width="120">
          <template #default="{ row }">
            <el-tag :type="getStatusType(row.status)">
              {{ row.status || 'Uploaded' }}
            </el-tag>
          </template>
        </el-table-column>
        <el-table-column label="Actions" width="200" align="center">
          <template #default="{ row }">
            <el-button size="small" @click="viewDocument(row)">View</el-button>
            <el-button size="small" type="primary" @click="processDocument(row)">
              Process
            </el-button>
            <el-button size="small" type="danger" @click="deleteDocument(row)">
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
import { ref, onMounted } from 'vue'
import { useRouter } from 'vue-router'
import { ElMessage, ElMessageBox } from 'element-plus'
import { Document, UploadFilled } from '@element-plus/icons-vue'
import api from '../services/api'

const router = useRouter()
const loading = ref(false)
const documents = ref([])
const currentPage = ref(1)
const pageSize = ref(20)
const total = ref(0)

const fetchDocuments = async () => {
  loading.value = true
  try {
    const response = await api.get('/documents', {
      params: {
        page: currentPage.value,
        per_page: pageSize.value
      }
    })
    documents.value = response.data.data.items || []
    total.value = response.data.data.pagination?.total || 0
  } catch (error) {
    ElMessage.error('Failed to load documents')
    console.error('Fetch documents error:', error)
  } finally {
    loading.value = false
  }
}

const formatFileSize = (bytes) => {
  if (!bytes) return '0 B'
  const k = 1024
  const sizes = ['B', 'KB', 'MB', 'GB']
  const i = Math.floor(Math.log(bytes) / Math.log(k))
  return Math.round(bytes / Math.pow(k, i) * 100) / 100 + ' ' + sizes[i]
}

const formatDate = (dateString) => {
  if (!dateString) return '-'
  return new Date(dateString).toLocaleString()
}

const getStatusType = (status) => {
  const statusMap = {
    'uploaded': 'info',
    'processing': 'warning',
    'completed': 'success',
    'failed': 'danger'
  }
  return statusMap[status] || 'info'
}

const navigateToUpload = () => {
  router.push('/upload')
}

const viewDocument = (document) => {
  ElMessage.info('View document feature coming soon')
  // TODO: Implement document viewer
}

const processDocument = async (document) => {
  ElMessage.info('OCR processing feature coming soon')
  // TODO: Implement OCR job submission
}

const deleteDocument = async (document) => {
  try {
    await ElMessageBox.confirm(
      'This will permanently delete the document. Continue?',
      'Warning',
      {
        confirmButtonText: 'Delete',
        cancelButtonText: 'Cancel',
        type: 'warning',
      }
    )
    
    await api.delete(`/documents/${document.id}`)
    ElMessage.success('Document deleted successfully')
    fetchDocuments()
  } catch (error) {
    if (error !== 'cancel') {
      ElMessage.error('Failed to delete document')
      console.error('Delete error:', error)
    }
  }
}

const handleSizeChange = (newSize) => {
  pageSize.value = newSize
  fetchDocuments()
}

const handleCurrentChange = (newPage) => {
  currentPage.value = newPage
  fetchDocuments()
}

onMounted(() => {
  fetchDocuments()
})
</script>

<style scoped>
.documents-container {
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

.filename-cell {
  display: flex;
  align-items: center;
  gap: 8px;
}

.filename-cell .el-icon {
  color: #409eff;
}
</style>
