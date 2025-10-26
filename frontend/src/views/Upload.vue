<template>
  <div class="upload-container">
    <el-card class="upload-card">
      <template #header>
        <div class="card-header">
          <h2>Upload Documents</h2>
          <span class="subtitle">Upload images or PDFs for OCR processing</span>
        </div>
      </template>

      <el-upload
        class="upload-dragger"
        drag
        :action="uploadUrl"
        :headers="uploadHeaders"
        :on-success="handleSuccess"
        :on-error="handleError"
        :before-upload="beforeUpload"
        :show-file-list="true"
        multiple
        accept=".jpg,.jpeg,.png,.pdf,.tiff,.bmp"
      >
        <el-icon class="el-icon--upload"><upload-filled /></el-icon>
        <div class="el-upload__text">
          Drop files here or <em>click to upload</em>
        </div>
        <template #tip>
          <div class="el-upload__tip">
            Supported formats: JPG, PNG, PDF, TIFF, BMP (max 50MB per file)
          </div>
        </template>
      </el-upload>

      <div class="upload-info">
        <el-divider />
        <h3>Upload Settings</h3>
        <el-form :model="uploadForm" label-width="140px">
          <el-form-item label="OCR Mode">
            <el-select v-model="uploadForm.ocrMode" placeholder="Select OCR mode">
              <el-option label="Document" value="document" />
              <el-option label="Handwritten" value="handwritten" />
              <el-option label="General" value="general" />
              <el-option label="Figure" value="figure" />
            </el-select>
          </el-form-item>
          <el-form-item label="Resolution">
            <el-select v-model="uploadForm.resolution" placeholder="Select resolution">
              <el-option label="Tiny (Fast)" value="tiny" />
              <el-option label="Small (Balanced)" value="small" />
              <el-option label="Base (Recommended)" value="base" />
              <el-option label="Large (High Quality)" value="large" />
              <el-option label="Gundam (Dynamic)" value="gundam" />
            </el-select>
          </el-form-item>
        </el-form>
      </div>
    </el-card>
  </div>
</template>

<script setup>
import { ref } from 'vue'
import { ElMessage } from 'element-plus'
import { UploadFilled } from '@element-plus/icons-vue'
import api from '../services/api'

const uploadUrl = `${import.meta.env.VITE_API_URL}/documents/upload`
const uploadHeaders = {
  Authorization: `Bearer ${localStorage.getItem('token')}`
}

const uploadForm = ref({
  ocrMode: 'document',
  resolution: 'base'
})

const beforeUpload = (file) => {
  const maxSize = 50 * 1024 * 1024 // 50MB
  if (file.size > maxSize) {
    ElMessage.error('File size cannot exceed 50MB')
    return false
  }
  return true
}

const handleSuccess = (response, file) => {
  ElMessage.success(`${file.name} uploaded successfully`)
  // TODO: Navigate to jobs page or auto-submit OCR job
}

const handleError = (error, file) => {
  ElMessage.error(`Failed to upload ${file.name}`)
  console.error('Upload error:', error)
}
</script>

<style scoped>
.upload-container {
  padding: 20px;
  max-width: 900px;
  margin: 0 auto;
}

.upload-card {
  box-shadow: 0 2px 12px 0 rgba(0, 0, 0, 0.1);
}

.card-header {
  display: flex;
  flex-direction: column;
  gap: 5px;
}

.card-header h2 {
  margin: 0;
  color: #303133;
}

.subtitle {
  color: #909399;
  font-size: 14px;
}

.upload-dragger {
  width: 100%;
}

.el-icon--upload {
  font-size: 67px;
  color: #409eff;
  margin-bottom: 16px;
}

.upload-info {
  margin-top: 30px;
}

.upload-info h3 {
  color: #303133;
  margin-bottom: 20px;
}
</style>
