<template>
  <div class="home">
    <el-container>
      <el-header>
        <h1>VisEkai - OCR Web Application</h1>
      </el-header>
      <el-main>
        <el-row :gutter="20" justify="center">
          <el-col :span="18">
            <el-card class="welcome-card">
              <h2>Welcome to VisEkai</h2>
              <p>
                Transform your scanned documents and handwritten notes into digital text 
                with our powerful OCR engine powered by DeepSeek-OCR.
              </p>
              <el-button-group>
                <el-button type="primary" @click="$router.push('/upload')">
                  <el-icon><Upload /></el-icon>
                  Upload Document
                </el-button>
                <el-button @click="$router.push('/documents')">
                  <el-icon><Document /></el-icon>
                  View Documents
                </el-button>
                <el-button @click="$router.push('/jobs')">
                  <el-icon><List /></el-icon>
                  View Jobs
                </el-button>
              </el-button-group>
            </el-card>
          </el-col>
        </el-row>
      </el-main>
    </el-container>
  </div>
</template>

<script setup>
import { Upload, Document, List } from '@element-plus/icons-vue'
</script>

<style scoped>
.home {
  min-height: 100vh;
  background: linear-gradient(135deg, #667eea 0%, #764ba2 100%);
}

.el-header {
  display: flex;
  align-items: center;
  justify-content: center;
  color: white;
}

.el-main {
  padding-top: 80px;
}

.welcome-card {
  text-align: center;
  padding: 40px;
}

.welcome-card h2 {
  margin-bottom: 20px;
  color: #333;
}

.welcome-card p {
  margin-bottom: 30px;
  font-size: 16px;
  color: #666;
  line-height: 1.6;
}

.el-button-group {
  display: flex;
  gap: 10px;
  justify-content: center;
}
</style>
