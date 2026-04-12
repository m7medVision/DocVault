// API configuration for mobile app
// Uses environment variables or defaults

export const API_URL = process.env.EXPO_PUBLIC_API_URL || 'http://localhost:8080/api/v1';

export const CONFIG = {
  // API base URL for backend
  apiBaseUrl: API_URL,
  
  // Document upload max size (50MB as per PRD)
  maxUploadSize: 50 * 1024 * 1024,
  
  // Supported file types for upload
  supportedFileTypes: ['application/pdf', 'image/jpeg', 'image/png', 'image/heic'],
  
  // Presigned URL TTL (15 minutes as per PRD)
  presignedUrlTtl: 15 * 60,
};
