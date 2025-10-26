import axios from 'axios'

const apiClient = axios.create({
  baseURL: import.meta.env.VITE_API_URL || 'http://localhost:8080/api/v1',
  headers: {
    'Content-Type': 'application/json',
  },
})

// Request interceptor to add auth token
apiClient.interceptors.request.use(
  (config) => {
    const token = localStorage.getItem('token')
    if (token) {
      config.headers.Authorization = `Bearer ${token}`
    }
    return config
  },
  (error) => {
    return Promise.reject(error)
  }
)

// Response interceptor to handle errors
apiClient.interceptors.response.use(
  (response) => {
    return response
  },
  (error) => {
    if (error.response && error.response.status === 401) {
      // Handle unauthorized access
      localStorage.removeItem('token')
      window.location.href = '/login'
    }
    return Promise.reject(error)
  }
)

export default {
  // Auth
  register(data) {
    return apiClient.post('/auth/register', data)
  },
  login(data) {
    return apiClient.post('/auth/login', data)
  },
  logout() {
    return apiClient.post('/auth/logout')
  },
  getCurrentUser() {
    return apiClient.get('/auth/me')
  },

  // Documents
  uploadDocument(formData) {
    return apiClient.post('/documents/upload', formData, {
      headers: {
        'Content-Type': 'multipart/form-data',
      },
    })
  },
  getDocuments(params) {
    return apiClient.get('/documents', { params })
  },
  getDocument(id) {
    return apiClient.get(`/documents/${id}`)
  },
  deleteDocument(id) {
    return apiClient.delete(`/documents/${id}`)
  },

  // OCR Jobs
  submitOCRJob(data) {
    return apiClient.post('/ocr/submit', data)
  },
  getJobs(params) {
    return apiClient.get('/ocr/jobs', { params })
  },
  getJob(id) {
    return apiClient.get(`/ocr/jobs/${id}`)
  },
  cancelJob(id) {
    return apiClient.put(`/ocr/jobs/${id}/cancel`)
  },

  // Results
  getResult(id) {
    return apiClient.get(`/results/${id}`)
  },
  downloadResult(id, format) {
    return apiClient.get(`/results/${id}/download`, {
      params: { format },
      responseType: 'blob',
    })
  },

  // Settings
  getSettings() {
    return apiClient.get('/settings')
  },
  updateSettings(data) {
    return apiClient.put('/settings', data)
  },

  // Health
  healthCheck() {
    return apiClient.get('/health')
  },
}
