import axios from 'axios';

// Create a function to setup axios interceptors
export const setupAxiosInterceptors = (onUnauthorized) => {
  // Request interceptor to add auth header
  axios.interceptors.request.use(
    (config) => {
      const token = localStorage.getItem('token');
      if (token) {
        config.headers.Authorization = `Bearer ${token}`;
      }
      return config;
    },
    (error) => {
      return Promise.reject(error);
    }
  );

  // Response interceptor to handle 401 errors
  axios.interceptors.response.use(
    (response) => {
      return response;
    },
    (error) => {
      if (error.response?.status === 401) {
        // Clear stored auth data
        localStorage.removeItem('token');
        localStorage.removeItem('user');
        
        // Call the logout callback to update app state
        if (onUnauthorized) {
          onUnauthorized();
        }
      }
      return Promise.reject(error);
    }
  );
};

// Helper function to get auth headers (for backwards compatibility)
export const getAuthHeaders = () => ({
  headers: {
    Authorization: `Bearer ${localStorage.getItem('token')}`
  }
});