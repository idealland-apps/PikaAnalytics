const config = {
  API_BASE_URL: process.env.REACT_APP_API_URL ||
    (process.env.NODE_ENV === 'production'
      ? `${window.location.protocol}//${window.location.host}/api`
      : 'http://localhost:8080/api')
};

export default config;
